package migrate

import (
	migrate2 "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/zedisdog/ty/database/migrate"
)

type DmMigrator struct {
	sourceUrl      string
	sourceInstance source.Driver
	databaseUrl    string
}

func (d *DmMigrator) SetSourceURL(url string) migrate.IMigrator {
	d.sourceUrl = url
	return d
}

func (d *DmMigrator) SetSourceInstance(instance source.Driver) migrate.IMigrator {
	d.sourceInstance = instance
	return d
}

func (d *DmMigrator) GetSourceInstance() source.Driver {
	return d.sourceInstance
}

func (d *DmMigrator) SetDatabaseURL(dsn string) migrate.IMigrator {
	d.databaseUrl = dsn
	return d
}

func (d *DmMigrator) Migrate() (err error) {
	migrator, err := migrate2.NewWithSourceInstance("", d.sourceInstance, d.databaseUrl)
	if err != nil {
		return
	}
	return migrator.Up()
}
