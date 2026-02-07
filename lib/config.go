package lib

import (
	"fmt"

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
	Cache:      &CacheConfig{Type: "memory"},
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
		panic(fmt.Sprintf("Failed to read configuration file: %s\nError: %v\nPlease ensure the config file exists and is valid YAML format.", configPath, err))
	}

	if err := viper.Unmarshal(&config); err != nil {
		panic(fmt.Sprintf("Failed to parse configuration file: %s\nError: %v\nPlease check your config file syntax.", configPath, err))
	}

	config.Casbin.Model = casbinModelPath
	return config
}

func SetConfigPath(path string) {
	if !file.IsFile(path) {
		panic(fmt.Sprintf("Configuration file not found: %s\nPlease create a config file or specify a valid path using the -c flag.\nExample: ./app runserver -c /path/to/config.yaml", path))
	}

	configPath = path
}

func SetConfigCasbinModelPath(path string) {
	if !file.IsFile(path) {
		panic(fmt.Sprintf("Casbin model file not found: %s\nPlease create the casbin model file or specify a valid path using the -m flag.\nExample: ./app runserver -m /path/to/casbin_model.conf", path))
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
	Cache      *CacheConfig      `mapstructure:"Cache"`
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
	Engine      string `mapstructure:"Engine"` // mysql, sqlite, postgres
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

// IsSQLite returns true if the database engine is SQLite
func (a *DatabaseConfig) IsSQLite() bool {
	return a.Engine == "sqlite"
}

// IsMySQL returns true if the database engine is MySQL
func (a *DatabaseConfig) IsMySQL() bool {
	return a.Engine == "" || a.Engine == "mysql"
}

// IsPostgreSQL returns true if the database engine is PostgreSQL
func (a *DatabaseConfig) IsPostgreSQL() bool {
	return a.Engine == "postgres"
}

// CacheConfig cache configuration
// Type: memory, redis
type CacheConfig struct {
	Type      string `mapstructure:"Type"` // memory or redis
	KeyPrefix string `mapstructure:"KeyPrefix"`

	// Redis specific settings (only used when Type is "redis")
	Host     string `mapstructure:"Host"`
	Port     int    `mapstructure:"Port"`
	Password string `mapstructure:"Password"`
}

// IsRedis returns true if cache type is Redis
func (c *CacheConfig) IsRedis() bool {
	return c.Type == "redis"
}

// IsMemory returns true if cache type is Memory (default)
func (c *CacheConfig) IsMemory() bool {
	return c.Type == "" || c.Type == "memory"
}

// Addr returns Redis address
func (c *CacheConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (a *DatabaseConfig) DSN() string {
	if a.IsPostgreSQL() {
		return a.PostgresDSN()
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", a.Username, a.Password, a.Host, a.Port, a.Name, a.Parameters)
}

// PostgresDSN returns the PostgreSQL connection string
func (a *DatabaseConfig) PostgresDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
		a.Host, a.Port, a.Username, a.Password, a.Name)
}

func (a *HttpConfig) ListenAddr() string {
	if err := validator.New().Struct(a); err != nil {
		return "0.0.0.0:5100"
	}

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
