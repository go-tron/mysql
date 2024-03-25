package mysql

import (
	"github.com/go-tron/config"
	goLogger "github.com/go-tron/logger"
	"github.com/go-tron/types/stringUtil"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"time"
)

type Config struct {
	Dialect        string `json:"dialect"`
	Url            string `json:"url"`
	MaxIdleConns   int    `json:"maxIdleConns"`
	MaxOpenConns   int    `json:"maxOpenConns"`
	Debug          bool   `json:"debug"`
	Logger         goLogger.Logger
	NamingStrategy *schema.NamingStrategy
}

type ConfigOption func(*Config)

func WithNamingStrategy(val *schema.NamingStrategy) ConfigOption {
	return func(config *Config) {
		config.NamingStrategy = val
	}
}
func NewWithConfig(c *config.Config, opts ...ConfigOption) *DB {
	return New(&Config{
		Dialect:      c.GetString("database.dialect"),
		Url:          c.GetString("database.url"),
		MaxIdleConns: c.GetInt("database.maxIdleConns"),
		MaxOpenConns: c.GetInt("database.maxOpenConns"),
		Debug:        c.GetBool("database.debug"),
		Logger:       goLogger.NewZapWithConfig(c, "mysql", "error"),
	}, opts...)
}

func New(c *Config, opts ...ConfigOption) *DB {
	if c == nil {
		panic("c 必须设置")
	}
	for _, apply := range opts {
		if apply != nil {
			apply(c)
		}
	}
	if c.Url == "" {
		panic("Url 必须设置")
	}
	if c.Logger == nil {
		panic("Logger 必须设置")
	}

	if c.NamingStrategy == nil {
		c.NamingStrategy = &schema.NamingStrategy{
			SingularTable: true,
		}
	}
	conf := &gorm.Config{
		NamingStrategy: c.NamingStrategy,
	}
	if c.Debug {
		conf.Logger = DefaultLogger.LogMode(Info)
	} else {
		conf.Logger = NewLogger(&DBLogger{c.Logger}, gormLogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  Error,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		})
	}
	db, err := gorm.Open(mysql.Open(c.Url), conf)
	if err != nil {
		panic(err)
	}
	dbConfig, _ := db.DB()
	dbConfig.SetConnMaxLifetime(time.Minute * 10)
	dbConfig.SetMaxIdleConns(c.MaxIdleConns)
	dbConfig.SetMaxOpenConns(c.MaxOpenConns)

	if err := dbConfig.Ping(); err != nil {
		panic(err)
	}
	return &DB{Config: c, DB: db}
}

type CamelCaseReplacer struct {
}

func (s *CamelCaseReplacer) Replace(name string) string {
	return stringUtil.FirstCharToLower(name)
}
