package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Postgres struct {
	conn *sql.DB
	addr string
}

func (p *Postgres) Init() error {
	p.addr = os.Getenv("DATABASE_URL")
	if len(p.addr) == 0 {
		return errors.New("DATABASE_URL env var is empty")
	}

	var err error
	p.conn, err = sql.Open("postgres", p.addr)
	if err != nil {
		return err
	}

	if err = p.Ping(); err != nil {
		return err
	}

	return p.Migrate()
}

func (p *Postgres) Ping() error {
	return p.conn.Ping()
}

func (p *Postgres) Close() error {
	return p.conn.Close()
}

func (p *Postgres) Migrate() error {
	migrationsDir := "file://migrations/postgres"

	driver, err := migratepg.WithInstance(p.conn, &migratepg.Config{
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		return fmt.Errorf("error creating migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(migrationsDir, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	err = m.Up()
	if err != nil {
		if err != migrate.ErrNoChange {
			return fmt.Errorf("failed to run migrations: %w", err)
		}
	}
	return nil
}
