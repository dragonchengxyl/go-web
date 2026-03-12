package configs

import (
	"time"
)

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	OSS        OSSConfig        `mapstructure:"oss"`
	RateLimit  RateLimitConfig  `mapstructure:"ratelimit"`
	Email      EmailConfig      `mapstructure:"email"`
	Payment    PaymentConfig    `mapstructure:"payment"`
	Moderation ModerationConfig `mapstructure:"moderation"`
	Sponsor    SponsorConfig    `mapstructure:"sponsor"`
	GRPC       GRPCConfig       `mapstructure:"grpc"`
}

type ServerConfig struct {
	Port         int      `mapstructure:"port"`
	Mode         string   `mapstructure:"mode"`
	AllowOrigins []string `mapstructure:"allow_origins"`
	FrontendURL  string   `mapstructure:"frontend_url"`
}

type DatabaseConfig struct {
	DSN      string `mapstructure:"dsn"`
	MaxConns int    `mapstructure:"max_conns"`
	MinConns int    `mapstructure:"min_conns"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	Secret        string        `mapstructure:"secret"`
	AccessExpiry  time.Duration `mapstructure:"access_expiry"`
	RefreshExpiry time.Duration `mapstructure:"refresh_expiry"`
}

type OSSConfig struct {
	Provider        string   `mapstructure:"provider"` // "aliyun" or "r2"
	AccessKeyID     string   `mapstructure:"access_key_id"`
	AccessKeySecret string   `mapstructure:"access_key_secret"`
	Bucket          string   `mapstructure:"bucket"`
	Endpoint        string   `mapstructure:"endpoint"`
	Region          string   `mapstructure:"region"`        // "auto" for R2
	AllowedHosts    []string `mapstructure:"allowed_hosts"` // validated media URL hosts
}

type RateLimitConfig struct {
	Unauthenticated int `mapstructure:"unauthenticated"`
	Authenticated   int `mapstructure:"authenticated"`
	Admin           int `mapstructure:"admin"`
}

type EmailConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}

type PaymentConfig struct {
	Alipay AlipayConfig `mapstructure:"alipay"`
	Wechat WechatConfig `mapstructure:"wechat"`
}

type AlipayConfig struct {
	AppID      string `mapstructure:"app_id"`
	PrivateKey string `mapstructure:"private_key"`
	PublicKey  string `mapstructure:"public_key"`
	NotifyURL  string `mapstructure:"notify_url"`
	Sandbox    bool   `mapstructure:"sandbox"`
}

type WechatConfig struct {
	AppID     string `mapstructure:"app_id"`
	MchID     string `mapstructure:"mch_id"`
	APIKey    string `mapstructure:"api_key"`
	NotifyURL string `mapstructure:"notify_url"`
	Sandbox   bool   `mapstructure:"sandbox"`
}

// ModerationConfig holds settings for Aliyun content safety service.
type ModerationConfig struct {
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	Endpoint        string `mapstructure:"endpoint"` // default: green-cip.cn-shanghai.aliyuncs.com
}

// SponsorConfig holds sponsor dashboard display data.
type SponsorConfig struct {
	MonthlyGoal   float64 `mapstructure:"monthly_goal"`
	CurrentRaised float64 `mapstructure:"current_raised"`
	AlipayQRURL   string  `mapstructure:"alipay_qr_url"`
	WechatQRURL   string  `mapstructure:"wechat_qr_url"`
	Message       string  `mapstructure:"message"`
}

// GRPCConfig holds gRPC service addresses and ports.
type GRPCConfig struct {
	// Client addresses — empty string means use local in-process implementation.
	StatsAddr        string `mapstructure:"stats_addr"`
	NotificationAddr string `mapstructure:"notification_addr"`
	ModerationAddr   string `mapstructure:"moderation_addr"`
	// Server ports — used by the individual service binaries.
	StatsPort        int `mapstructure:"stats_port"`
	NotificationPort int `mapstructure:"notification_port"`
	ModerationPort   int `mapstructure:"moderation_port"`
}
