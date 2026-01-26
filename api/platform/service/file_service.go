package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/platform"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// FileService 文件服务接口
type FileService interface {
	UploadFile(filename string, reader io.Reader, size int64, contentType string) (*platform.FileInfo, error)
	DeleteFile(filePath string) error
}

// NewFileService 根据配置创建对应的文件服务
func NewFileService(config lib.Config, logger lib.Logger) FileService {
	ossConfig := config.OSS
	if ossConfig == nil {
		logger.Zap.Warn("OSS config not found, using local storage")
		return NewLocalFileService("./uploads", logger)
	}

	switch ossConfig.Type {
	case "minio":
		if ossConfig.Minio == nil {
			logger.Zap.Warn("Minio config not found, using local storage")
			return NewLocalFileService("./uploads", logger)
		}
		svc, err := NewMinioFileService(ossConfig.Minio, logger)
		if err != nil {
			logger.Zap.Errorf("Failed to create minio service: %v, using local storage", err)
			return NewLocalFileService("./uploads", logger)
		}
		return svc
	case "aliyun":
		if ossConfig.Aliyun == nil {
			logger.Zap.Warn("Aliyun config not found, using local storage")
			return NewLocalFileService("./uploads", logger)
		}
		return NewAliyunFileService(ossConfig.Aliyun, logger)
	default:
		storagePath := "./uploads"
		if ossConfig.Local != nil && ossConfig.Local.StoragePath != "" {
			storagePath = ossConfig.Local.StoragePath
		}
		return NewLocalFileService(storagePath, logger)
	}
}

// ==================== Local File Service ====================

// LocalFileService 本地文件存储服务
type LocalFileService struct {
	storagePath string
	logger      lib.Logger
}

// NewLocalFileService 创建本地文件服务
func NewLocalFileService(storagePath string, logger lib.Logger) *LocalFileService {
	// 确保存储目录存在
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		logger.Zap.Errorf("Failed to create storage directory: %v", err)
	}
	return &LocalFileService{
		storagePath: storagePath,
		logger:      logger,
	}
}

// UploadFile 上传文件到本地
func (s *LocalFileService) UploadFile(filename string, reader io.Reader, size int64, contentType string) (*platform.FileInfo, error) {
	// 获取文件后缀
	ext := filepath.Ext(filename)
	// 生成新文件名
	newFilename := uuid.New().String() + ext
	// 按日期分目录
	dateFolder := time.Now().Format("20060102")
	folderPath := filepath.Join(s.storagePath, dateFolder)

	// 创建日期目录
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// 完整文件路径
	fullPath := filepath.Join(folderPath, newFilename)

	// 创建文件
	file, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// 写入文件
	if _, err := io.Copy(file, reader); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// 返回相对路径，前端需要自行处理访问前缀
	fileURL := "/" + dateFolder + "/" + newFilename

	return &platform.FileInfo{
		Name: filename,
		URL:  fileURL,
	}, nil
}

// DeleteFile 删除本地文件
func (s *LocalFileService) DeleteFile(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("file path is empty")
	}

	// 安全检查：防止删除存储目录外的文件
	fullPath := filepath.Join(s.storagePath, filePath)
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	absStoragePath, _ := filepath.Abs(s.storagePath)
	if !strings.HasPrefix(absPath, absStoragePath) {
		return fmt.Errorf("invalid file path: access denied")
	}

	// 检查是否为目录
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在，视为删除成功
		}
		return fmt.Errorf("failed to stat file: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("cannot delete directory")
	}

	return os.Remove(absPath)
}

// ==================== MinIO File Service ====================

// MinioFileService MinIO文件存储服务
type MinioFileService struct {
	client       *minio.Client
	bucketName   string
	customDomain string
	endpoint     string
	logger       lib.Logger
}

// NewMinioFileService 创建MinIO文件服务
func NewMinioFileService(config *lib.MinioOSSConfig, logger lib.Logger) (*MinioFileService, error) {
	// 解析endpoint
	endpoint := config.Endpoint
	useSSL := strings.HasPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(strings.TrimPrefix(endpoint, "https://"), "http://")

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	svc := &MinioFileService{
		client:       client,
		bucketName:   config.BucketName,
		customDomain: config.CustomDomain,
		endpoint:     config.Endpoint,
		logger:       logger,
	}

	// 确保bucket存在
	if err := svc.ensureBucket(); err != nil {
		logger.Zap.Warnf("Failed to ensure bucket: %v", err)
	}

	return svc, nil
}

// ensureBucket 确保bucket存在
func (s *MinioFileService) ensureBucket() error {
	ctx := context.Background()
	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return err
	}

	if !exists {
		if err := s.client.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{}); err != nil {
			return err
		}

		// 设置公共读策略
		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Principal": {"AWS": ["*"]},
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/*"]
			}]
		}`, s.bucketName)

		if err := s.client.SetBucketPolicy(ctx, s.bucketName, policy); err != nil {
			s.logger.Zap.Warnf("Failed to set bucket policy: %v", err)
		}
	}

	return nil
}

// UploadFile 上传文件到MinIO
func (s *MinioFileService) UploadFile(filename string, reader io.Reader, size int64, contentType string) (*platform.FileInfo, error) {
	// 获取文件后缀
	ext := filepath.Ext(filename)
	// 生成新文件名
	newFilename := uuid.New().String() + ext
	// 按日期分目录
	dateFolder := time.Now().Format("20060102")
	objectName := dateFolder + "/" + newFilename

	ctx := context.Background()
	// 上传文件
	_, err := s.client.PutObject(ctx, s.bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// 构建文件URL
	var fileURL string
	if s.customDomain != "" {
		fileURL = s.customDomain + "/" + s.bucketName + "/" + objectName
	} else {
		fileURL = s.endpoint + "/" + s.bucketName + "/" + objectName
	}

	return &platform.FileInfo{
		Name: filename,
		URL:  fileURL,
	}, nil
}

// DeleteFile 删除MinIO文件
func (s *MinioFileService) DeleteFile(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("file path is empty")
	}

	// 从URL中提取对象名
	var objectName string
	if s.customDomain != "" {
		prefix := s.customDomain + "/" + s.bucketName + "/"
		objectName = strings.TrimPrefix(filePath, prefix)
	} else {
		prefix := s.endpoint + "/" + s.bucketName + "/"
		objectName = strings.TrimPrefix(filePath, prefix)
	}

	ctx := context.Background()
	return s.client.RemoveObject(ctx, s.bucketName, objectName, minio.RemoveObjectOptions{})
}

// ==================== Aliyun OSS File Service ====================

// AliyunFileService 阿里云OSS文件存储服务
type AliyunFileService struct {
	config *lib.AliyunOSSConfig
	logger lib.Logger
}

// NewAliyunFileService 创建阿里云OSS文件服务
func NewAliyunFileService(config *lib.AliyunOSSConfig, logger lib.Logger) *AliyunFileService {
	return &AliyunFileService{
		config: config,
		logger: logger,
	}
}

// UploadFile 上传文件到阿里云OSS
func (s *AliyunFileService) UploadFile(filename string, reader io.Reader, size int64, contentType string) (*platform.FileInfo, error) {
	// 阿里云OSS需要引入阿里云SDK，这里提供一个简化实现
	// 实际使用时需要: go get github.com/aliyun/aliyun-oss-go-sdk/oss
	return nil, fmt.Errorf("aliyun OSS not implemented, please install aliyun-oss-go-sdk")
}

// DeleteFile 删除阿里云OSS文件
func (s *AliyunFileService) DeleteFile(filePath string) error {
	return fmt.Errorf("aliyun OSS not implemented, please install aliyun-oss-go-sdk")
}
