package sqldb

import (
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/gorm"
)

var dbDriverName = "postgres"

// MigratePostgresDB applies database migrations to a PostgreSQL database using the provided GORM DB instance and migration files path.
func MigratePostgresDB(db *gorm.DB, migrationPath string) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationPath,
		dbDriverName,
		driver,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

// MigratePostgresDBWithSteps applies a specified number of migration steps in a given direction (up or down) to a PostgreSQL database using the provided GORM DB instance and migration files path.
func MigratePostgresDBWithSteps(db *gorm.DB, migrationPath string, direction string, steps int) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationPath,
		dbDriverName,
		driver,
	)
	if err != nil {
		return err
	}

	switch direction {
	case "up":
		if err := m.Steps(steps); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
	case "down":
		if err := m.Steps(-steps); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
	default:
		return errors.New("invalid migration direction: must be 'up' or 'down'")
	}

	return nil
}

// RunMigration applies all pending migrations and returns the database client.
func RunMigration(db *gorm.DB, migrationPath string) (*gorm.DB, error) {
	if err := MigratePostgresDB(db, migrationPath); err != nil {
		return nil, err
	}
	return db, nil
}
