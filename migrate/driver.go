package migrate

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4/database"
	_ "github.com/zedisdog/tydm/dm/sqldriver"
	"io"
)

var _ database.Driver = (*Dm)(nil)

const DefaultMigrationsTableName = "schema_migrations"

type Dm struct {
	db              *sql.DB
	MigrationsTable string
	tableSpaceName  string
}

func WithInstance(db *sql.DB) (driver database.Driver, err error) {
	if err = db.Ping(); err != nil {
		return
	}
	d := NewDmDriver(db)
	if err = d.ensureVersionTable(); err != nil {
		return
	}

	driver = d
	return
}

func NewDmDriver(db *sql.DB) *Dm {
	return &Dm{
		db:              db,
		MigrationsTable: DefaultMigrationsTableName,
	}
}

func (d Dm) Open(url string) (driver database.Driver, err error) {
	db, err := sql.Open("dm", url)
	if err != nil {
		return
	}
	return WithInstance(db)
}

func (d Dm) Close() error {
	return d.db.Close()
}

func (d Dm) Lock() error {
	//TODO implement me
	return nil
}

func (d Dm) Unlock() error {
	//TODO implement me
	return nil
}

func (d Dm) Run(migration io.Reader) (err error) {
	mig, err := io.ReadAll(migration)
	if err != nil {
		return
	}
	_, err = d.db.Exec(string(mig))
	if err != nil {
		return
	}
	return
}

func (d Dm) SetVersion(version int, dirty bool) (err error) {
	tx, err := d.db.Begin()
	if err != nil {
		return
	}

	query := fmt.Sprintf("DELETE FROM %s;", d.MigrationsTable)
	if _, err = tx.Exec(query); err != nil {
		tx.Rollback()
		return
	}

	// Also re-write the schema version for nil dirty versions to prevent
	// empty schema version for failed down migration on the first migration
	// See: https://github.com/golang-migrate/migrate/issues/330
	if version >= 0 || (version == database.NilVersion && dirty) {
		query := fmt.Sprintf("INSERT INTO %s (version, dirty) VALUES (?, ?)", d.MigrationsTable)
		if _, err = tx.Exec(query, version, dirty); err != nil {
			tx.Rollback()
			return
		}
	}

	err = tx.Commit()
	return
}

func (d Dm) Version() (version int, dirty bool, err error) {
	query := fmt.Sprintf("SELECT version, dirty FROM %s LIMIT 1", d.MigrationsTable)
	err = d.db.QueryRow(query).Scan(&version, &dirty)

	if errors.Is(err, sql.ErrNoRows) {
		return database.NilVersion, false, nil
	}

	if err != nil {
		if e, ok := err.(*mysql.MySQLError); ok && e.Number == 0 {
			return database.NilVersion, false, nil
		}
	} else {
		return 0, false, err
	}

	return
}

func (d Dm) Drop() (err error) {
	var count int
	query := `SELECT count(*) FROM USER_TABLES WHERE TABLESPACE_NAME = ?`
	err = d.db.QueryRow(query, d.tableSpaceName).Scan(&count)
	if err != nil {
		return
	}
	if count < 1 {
		return
	}

	tables := make([]string, 0, count)
	query = `SELECT TABLE_NAME FROM USER_TABLES WHERE TABLESPACE_NAME = ?`
	rows, err := d.db.Query(query, d.tableSpaceName)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return
		}

		tables = append(tables, name)
	}

	err = d.foreignKeyCheck(false)
	if err != nil {
		return
	}
	defer d.foreignKeyCheck(true)

	for _, table := range tables {
		query = fmt.Sprintf(`DROP TABLE IF EXISTS %s`, table)
		_, err = d.db.Exec(query)
		if err != nil {
			return
		}
	}

	return
}

type Col struct {
	TableName      string
	ConstraintName string
}

func (d Dm) foreignKeyCheck(enable bool) (err error) {
	query := `
SELECT count(*)
FROM SYSCONS a, SYSOBJECTS b, ALL_CONS_COLUMNS c
WHERE a.id = b.id AND a.TYPE$ = 'F' AND b.name = c.CONSTRAINT_NAME AND c.owner NOT IN ('SYS')
`
	var count int
	err = d.db.QueryRow(query).Scan(&count)
	if err != nil {
		return
	}

	if count < 1 {
		return
	}

	list := make([]Col, 0, count)
	query = `
SELECT TABLE_NAME, CONSTRAINT_NAME
FROM SYSCONS a, SYSOBJECTS b, ALL_CONS_COLUMNS c
WHERE a.id = b.id AND a.TYPE$ = 'F' AND b.name = c.CONSTRAINT_NAME AND c.owner NOT IN ('SYS')
`
	rows, err := d.db.Query(query)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var col Col
		if err = rows.Scan(&col.TableName, &col.ConstraintName); err != nil {
			return
		}
		list = append(list, col)
	}

	tx, err := d.db.Begin()
	if err != nil {
		return
	}

	str := "ENABLE"
	if !enable {
		str = "DISABLE"
	}
	for _, col := range list {
		query := fmt.Sprintf(`alter table %s %s CONSTRAINT %s`, col.TableName, str, col.ConstraintName)
		_, err = tx.Exec(query)
		if err != nil {
			tx.Rollback()
			return
		}
	}

	return tx.Commit()
}

func (d Dm) ensureVersionTable() (err error) {
	var count int

	query := `select count(*) from user_tables where table_name=?`
	err = d.db.QueryRow(query, d.MigrationsTable).Scan(&count)
	if err != nil {
		return
	}

	if count == 0 {
		query := fmt.Sprintf(`CREATE TABLE %s (version BIGINT NOT NULL, dirty TINYINT NOT NULL, PRIMARY KEY(version));`, d.MigrationsTable)
		_, err = d.db.Exec(query)
		if err != nil {
			return
		}
	}

	return
}
