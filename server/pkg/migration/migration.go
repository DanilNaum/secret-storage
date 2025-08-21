package migration

import (
	"database/sql"
	// "errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type migrator struct {
	migrationFolderPath string
	dbName              string
	dsn                 string
}

type migrationFolderPathOpt func() string

// WithAbsolutePath returns a migrationFolderPathOpt function that converts an absolute file path
// to a file URL scheme for use with database migration sources. It prepends "file:///"
// to the provided path, ensuring compatibility with the golang-migrate library's file source.
func WithAbsolutePath(path string) migrationFolderPathOpt {
	return func() string {
		return "file:///" + path
	}
}

// WithRelativePath returns a migrationFolderPathOpt function that converts a relative file path
// to a file URL scheme for use with database migration sources. It prepends "file://"
// to the provided path, ensuring compatibility with the golang-migrate library's file source.
func WithRelativePath(path string) migrationFolderPathOpt {
	return func() string {
		return "file://" + path
	}
}

// NewMigrator creates and returns a new migrator instance configured with the provided database connection string
// and migration folder path. It sets up the migration configuration with a default database name of "postgres".
// The migrationFolderPathOpt allows specifying the migration source path using either absolute or relative path options.
func NewMigrator(dsn string, migrationFolderPathOpt migrationFolderPathOpt) *migrator {
	mig := &migrator{
		migrationFolderPath: migrationFolderPathOpt(),
		dbName:              "postgres",
		dsn:                 dsn,
	}
	return mig
}

// Migrate executes database migrations using the configured migration path.
// It opens a connection to the PostgreSQL database, sets up the migration driver,
// and applies all pending migrations in the forward direction.
// Returns an error if any step in the migration process fails.
func (mig *migrator) Migrate() error {
	db, err := sql.Open("postgres", mig.dsn)
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		mig.migrationFolderPath,
		mig.dbName, driver)
	if err != nil {
		return err
	}
	m.Up()

	return nil
}
