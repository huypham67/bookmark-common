package sqldb

import (
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/gorm"
)

var dbDriverName = "postgres"

// newMigrator builds a golang-migrate instance backed by the given GORM
// PostgreSQL connection and the migration files at migrationPath.
func newMigrator(db *gorm.DB, migrationPath string) (*migrate.Migrate, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return nil, err
	}

	return migrate.NewWithDatabaseInstance("file://"+migrationPath, dbDriverName, driver)
}

// MigratePostgresDB applies all pending migrations. It is a no-op when the
// database is already up to date.
func MigratePostgresDB(db *gorm.DB, migrationPath string) error {
	m, err := newMigrator(db, migrationPath)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

// MigratePostgresDBWithSteps applies a fixed number of migration steps in the
// given direction ("up" or "down"). steps must be a positive integer.
func MigratePostgresDBWithSteps(db *gorm.DB, migrationPath string, direction string, steps int) error {
	if steps <= 0 {
		return errors.New("steps must be a positive integer")
	}

	m, err := newMigrator(db, migrationPath)
	if err != nil {
		return err
	}

	switch direction {
	case "up":
		err = m.Steps(steps)
	case "down":
		err = m.Steps(-steps)
	default:
		return errors.New("invalid migration direction: must be 'up' or 'down'")
	}

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

// ForcePostgresDB forcibly sets the migration version and clears the dirty flag
// WITHOUT running any migration SQL. Use it to recover from a failed (dirty)
// migration: inspect/repair the schema by hand, then force to the version that
// matches the actual schema so normal migrations can resume.
func ForcePostgresDB(db *gorm.DB, migrationPath string, version int) error {
	m, err := newMigrator(db, migrationPath)
	if err != nil {
		return err
	}

	return m.Force(version)
}

// PostgresMigrationVersion reports the current migration version and whether the
// database is dirty (a migration failed partway and must be resolved with
// ForcePostgresDB before further migrations can run). A fresh database with no
// migrations applied returns version 0, dirty false.
func PostgresMigrationVersion(db *gorm.DB, migrationPath string) (version uint, dirty bool, err error) {
	m, err := newMigrator(db, migrationPath)
	if err != nil {
		return 0, false, err
	}

	version, dirty, err = m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, nil
	}

	return version, dirty, err
}

// RunMigration applies all pending migrations and returns the database client.
func RunMigration(db *gorm.DB, migrationPath string) (*gorm.DB, error) {
	if err := MigratePostgresDB(db, migrationPath); err != nil {
		return nil, err
	}
	return db, nil
}
