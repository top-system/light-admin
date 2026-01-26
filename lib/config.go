package lib

import (
	"fmt"

	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/pkg/file"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

var configPath = "./config.yml"
var casbinModelPath = "./casbin_model.conf"

var defaultConfig = Config{
	Name: "app",
	Http: &HttpConfig{
		Host: "0.0.0.0",
		Port: 9999,
	},
	Log: &LogConfig{
		Level:       "debug",
		Directory:   "/tmp/app",
		Development: true,
	},
	SuperAdmin: &SuperAdminConfig{},
	Auth:       &AuthConfig{},
	Captcha:    &CaptchaConfig{Enable: true},
	Casbin:     &CasbinConfig{Enable: false},
	Redis:      &RedisConfig{Host: "127.0.0.1", Port: 6379},
	Database: &DatabaseConfig{
		Parameters:   "charset=utf8mb4&parseTime=True&loc=Local&allowNativePasswords=true&timeout=5s",
		MaxLifetime:  7200,
		MaxOpenConns: 150,
		MaxIdleConns: 50,
	},
	OSS: &OSSConfig{Type: "local", Local: &LocalOSSConfig{StoragePath: "./uploads"}},
}

func NewConfig() Config {
	config := defaultConfig

	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		panic(errors.Wrap(err, "failed to read config"))
	}

	if err := viper.Unmarshal(&config); err != nil {
		panic(errors.Wrap(err, "failed to marshal config"))
	}

	config.Casbin.Model = casbinModelPath
	return config
}

func SetConfigPath(path string) {
	if !file.IsFile(path) {
		panic("config filepath does not exist")
	}

	configPath = path
}

func SetConfigCasbinModelPath(path string) {
	if !file.IsFile(path) {
		panic("casbin model filepath does not exist")
	}

	casbinModelPath = path
}

// Configuration are the available config values
type Config struct {
	Name       string            `mapstructure:"Name"`
	Http       *HttpConfig       `mapstructure:"Http"`
	Log        *LogConfig        `mapstructure:"Log"`
	SuperAdmin *SuperAdminConfig `mapstructure:"SuperAdmin"`
	Auth       *AuthConfig       `mapstructure:"Auth"`
	Captcha    *CaptchaConfig    `mapstructure:"Captcha"`
	Casbin     *CasbinConfig     `mapstructure:"Casbin"`
	Redis      *RedisConfig      `mapstructure:"Redis"`
	Database   *DatabaseConfig   `mapstructure:"Database"`
	OSS        *OSSConfig        `mapstructure:"OSS"`

	// ====== 扩展功能配置 (可选) ======
	Queue      *QueueConfig      `mapstructure:"Queue"`
	Crontab    *CrontabConfig    `mapstructure:"Crontab"`
	Downloader *DownloaderConfig `mapstructure:"Downloader"`
}

type CaptchaConfig struct {
	Enable bool `mapstructure:"Enable"`
}

type HttpConfig struct {
	Host string `mapstructure:"Host" validate:"ipv4"`
	Port int    `mapstructure:"Port" validate:"gte=1,lte=65535"`
}

// LogLevel     : debug,info,warn,error,dpanic,panic,fatal
//                default info
// Format       : json, console
//                default json
// Directory    : Log storage path
//                default "./"
type LogConfig struct {
	Level       string `mapstructure:"Level"`
	Format      string `mapstructure:"Format"`
	Directory   string `mapstructure:"Directory"`
	Development bool   `mapstructure:"Development"`
}

type SuperAdminConfig struct {
	Username string `mapstructure:"Username"`
	Realname string `mapstructure:"Realname"`
	Password string `mapstructure:"Password"`
}

type AuthConfig struct {
	Enable             bool     `mapstructure:"Enable"`
	TokenExpired       int      `mapstructure:"TokenExpired"`
	IgnorePathPrefixes []string `mapstructure:"IgnorePathPrefixes"`
}

type CasbinConfig struct {
	Enable             bool     `mapstructure:"Enable"`
	Debug              bool     `mapstructure:"Debug"`
	Model              string   `mapstructure:"Model"`
	AutoLoad           bool     `mapstructure:"AutoLoad"`
	AutoLoadInternal   int      `mapstructure:"AutoLoadInternal"`
	IgnorePathPrefixes []string `mapstructure:"IgnorePathPrefixes"`
}

type DatabaseConfig struct {
	Engine      string `mapstructure:"Engine"`
	Name        string `mapstructure:"Name"`
	Host        string `mapstructure:"Host"`
	Port        int    `mapstructure:"Port"`
	Username    string `mapstructure:"Username"`
	Password    string `mapstructure:"Password"`
	TablePrefix string `mapstructure:"TablePrefix"`
	Parameters  string `mapstructure:"Parameters"`

	MaxLifetime  int `mapstructure:"MaxLifetime"`
	MaxOpenConns int `mapstructure:"MaxOpenConns"`
	MaxIdleConns int `mapstructure:"MaxIdleConns"`
}

type RedisConfig struct {
	Host      string `mapstructure:"Host"`
	Port      int    `mapstructure:"Port"`
	Password  string `mapstructure:"Password"`
	KeyPrefix string `mapstructure:"KeyPrefix"`
}

func (a *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", a.Username, a.Password, a.Host, a.Port, a.Name, a.Parameters)
}

func (a *HttpConfig) ListenAddr() string {
	if err := validator.New().Struct(a); err != nil {
		return "0.0.0.0:5100"
	}

	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}

func (a *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}

// OSSConfig 对象存储配置
type OSSConfig struct {
	Type   string          `mapstructure:"Type"` // local, minio, aliyun
	Local  *LocalOSSConfig `mapstructure:"Local"`
	Minio  *MinioOSSConfig `mapstructure:"Minio"`
	Aliyun *AliyunOSSConfig `mapstructure:"Aliyun"`
}

// LocalOSSConfig 本地存储配置
type LocalOSSConfig struct {
	StoragePath string `mapstructure:"StoragePath"` // 存储路径
}

// MinioOSSConfig MinIO配置
type MinioOSSConfig struct {
	Endpoint     string `mapstructure:"Endpoint"`
	AccessKey    string `mapstructure:"AccessKey"`
	SecretKey    string `mapstructure:"SecretKey"`
	BucketName   string `mapstructure:"BucketName"`
	CustomDomain string `mapstructure:"CustomDomain"` // 自定义域名
}

// AliyunOSSConfig 阿里云OSS配置
type AliyunOSSConfig struct {
	Endpoint        string `mapstructure:"Endpoint"`
	AccessKeyID     string `mapstructure:"AccessKeyID"`
	AccessKeySecret string `mapstructure:"AccessKeySecret"`
	BucketName      string `mapstructure:"BucketName"`
}

// ============================================================================
// 扩展功能配置 (可选)
// ============================================================================

// QueueConfig 任务队列配置
type QueueConfig struct {
	Enable    bool   `mapstructure:"Enable"`    // 是否启用
	Name      string `mapstructure:"Name"`      // 队列名称
	WorkerNum int    `mapstructure:"WorkerNum"` // 工作线程数
	MaxRetry  int    `mapstructure:"MaxRetry"`  // 最大重试次数
}

// CrontabConfig 定时任务配置
type CrontabConfig struct {
	Enable bool `mapstructure:"Enable"` // 是否启用
}

// DownloaderConfig 下载器配置
type DownloaderConfig struct {
	Enable      bool               `mapstructure:"Enable"` // 是否启用
	Type        string             `mapstructure:"Type"`   // 类型: aria2, qbittorrent
	Aria2       *Aria2Config       `mapstructure:"Aria2"`
	QBittorrent *QBittorrentConfig `mapstructure:"QBittorrent"`
}

// Aria2Config aria2 配置
type Aria2Config struct {
	Server   string                 `mapstructure:"Server"`   // RPC 服务器地址
	Token    string                 `mapstructure:"Token"`    // RPC 密钥
	TempPath string                 `mapstructure:"TempPath"` // 临时下载路径
	Options  map[string]interface{} `mapstructure:"Options"`  // 额外选项
}

// QBittorrentConfig qBittorrent 配置
type QBittorrentConfig struct {
	Server   string                 `mapstructure:"Server"`   // Web UI 地址
	User     string                 `mapstructure:"User"`     // 用户名
	Password string                 `mapstructure:"Password"` // 密码
	TempPath string                 `mapstructure:"TempPath"` // 临时下载路径
	Options  map[string]interface{} `mapstructure:"Options"`  // 额外选项
}
