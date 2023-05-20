package migrate

import (
	migrate2 "github.com/golang-migrate/migrate/v4"
	"github.com/stretchr/testify/require"
	"github.com/zedisdog/ty/database/migrate"
	"github.com/zedisdog/tydm/migrate/migration"
	"testing"
)

func TestMigrate(t *testing.T) {
	migratorSourceDriver := migrate.NewFsDriver()
	migratorSourceDriver.Add(&migration.Migration)
	migrator, err := migrate2.NewWithSourceInstance("", migratorSourceDriver, "dm://fbook:fbook@172.21.32.1:5236")
	require.Nil(t, err)

	err = migrator.Up()
	require.Nil(t, err)
}
