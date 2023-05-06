package tydm

import (
	"github.com/zedisdog/ty/application"
	"github.com/zedisdog/ty/errx"
	"github.com/zedisdog/tydm/dm"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

type Module struct{}

func (d Module) Name() string {
	return "tydm"
}

func (d *Module) Register() (err error) {
	dsn := application.Config().GetString("database.dm.dsn")
	if dsn == "" {
		return errx.New("no dameng database config")
	}
	db, err := gorm.Open(dm.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             200 * time.Millisecond, // Slow SQL threshold
				LogLevel:                  logger.Warn,            // Log level
				IgnoreRecordNotFoundError: true,                   // Ignore ErrRecordNotFound error for logger
				Colorful:                  false,                  // Disable color
			},
		),
	})
	if err != nil {
		return
	}
	application.RegisterDatabase("dm", db)
	return nil
}

func (d Module) Boot() error {
	return nil
}
