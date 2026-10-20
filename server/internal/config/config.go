package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 全局配置，结构对应 configs/config.yaml
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	HTTP     HTTPConfig     `mapstructure:"http"`
	Log      LogConfig      `mapstructure:"log"`
	MySQL    MySQLConfig    `mapstructure:"mysql"`
	Redis    RedisConfig    `mapstructure:"redis"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	SMS      SMSConfig      `mapstructure:"sms"`
	Wx       WxConfig       `mapstructure:"wx"`
	Qiniu    QiniuConfig    `mapstructure:"qiniu"`
	Finance  FinanceConfig  `mapstructure:"finance"`
}

type AppConfig struct {
	Name    string `mapstructure:"name"`
	Mode    string `mapstructure:"mode"`
	Version string `mapstructure:"version"`
}

type HTTPConfig struct {
	Addr string `mapstructure:"addr"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	File       string `mapstructure:"file"`
	MaxSizeMB  int    `mapstructure:"max_size_mb"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAgeDays int    `mapstructure:"max_age_days"`
}

type MySQLConfig struct {
	DSN          string `mapstructure:"dsn"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	// AutoMigrate 启动时自动建表/改表：开发期 true；生产置 false，
	// 表结构由 golang-migrate 迁移文件管理（task server:migrate，先于服务启动执行）
	AutoMigrate bool `mapstructure:"auto_migrate"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type RabbitMQConfig struct {
	URL string `mapstructure:"url"`
}

type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
	Issuer      string `mapstructure:"issuer"`
}

type SMSConfig struct {
	DevMode bool   `mapstructure:"dev_mode"`
	DevCode string `mapstructure:"dev_code"`
}

type WxConfig struct {
	MP  MPConfig  `mapstructure:"mp"`
	Pay PayConfig `mapstructure:"pay"`
}

type MPConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	AppID   string `mapstructure:"app_id"`
	Secret  string `mapstructure:"secret"`
}

type PayConfig struct {
	Enabled        bool   `mapstructure:"enabled"`
	Mock           bool   `mapstructure:"mock"`
	AppID          string `mapstructure:"app_id"`
	MchID          string `mapstructure:"mch_id"`
	SerialNo       string `mapstructure:"serial_no"`
	PrivateKeyPath string `mapstructure:"private_key_path"`
	APIv3Key       string `mapstructure:"apiv3_key"`
	NotifyURL      string `mapstructure:"notify_url"`
}

type QiniuConfig struct {
	AK     string `mapstructure:"ak"`
	SK     string `mapstructure:"sk"`
	Bucket string `mapstructure:"bucket"`
	Domain string `mapstructure:"domain"`
	Region string `mapstructure:"region"`
}

// FinanceConfig 费率统一用基点 BP：3% = 300
type FinanceConfig struct {
	ChefCommissionBP     int   `mapstructure:"chef_commission_bp"`
	SupplierCommissionBP int   `mapstructure:"supplier_commission_bp"`
	DepositBP            int   `mapstructure:"deposit_bp"`
	SettleDelayDays      int   `mapstructure:"settle_delay_days"`       // 结算冻结期（天）：完成后满 N 天自动可提，盖住售后窗口
	AutoReleaseMaxAmount int64 `mapstructure:"auto_release_max_amount"` // 单笔提现 ≤ 此金额（分）系统自动放行，0=不限；超过转人工审核
}

// Load 加载配置；支持 XC_ 前缀环境变量覆盖，如 XC_MYSQL_DSN
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix("XC")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置失败: %w", err)
	}
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}
	if err := validate(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func validate(c *Config) error {
	if c.JWT.Secret == "" {
		return fmt.Errorf("jwt.secret 不能为空")
	}
	if c.JWT.ExpireHours <= 0 {
		c.JWT.ExpireHours = 72
	}
	if c.Finance.DepositBP <= 0 {
		c.Finance.DepositBP = 3000
	}
	return nil
}
