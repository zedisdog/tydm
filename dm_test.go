package tydm

import (
	"github.com/stretchr/testify/require"
	"github.com/zedisdog/tydm/dm"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"testing"
	"time"
)

type Product struct {
	ID        uint64 `gorm:"primary key"`
	Name      string
	CreatedAt int64 `gorm:"timestamp"`
}

func TestConnectDatabase(t *testing.T) {
	db, err := gorm.Open(dm.Open("dm://fbook:fbook@172.21.32.1:5236"), &gorm.Config{
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
	require.Nil(t, err)
	d, err := db.DB()
	require.Nil(t, err)
	err = d.Ping()
	require.Nil(t, err)
}
