package main

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2" // register database/sql driver
	"github.com/golang-migrate/migrate"
	_ "github.com/golang-migrate/migrate/database/clickhouse" // register golang-migrate driver
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
)

const (
	databaseNotExistErrorCode = 81
)

// NewDB return updated db.
func NewDB(dsn string, maxIdleConns, maxOpenConns int, log *logrus.Entry, migrationFolder string) *sqlx.DB {
	db, err := connect(dsn, log)
	if err != nil {
		log.Fatalln(err)
	}

	// TODO: find solution with better performance
	db.Mapper = reflectx.NewMapperTagFunc("json", strings.ToUpper, func(value string) string {
		if strings.Contains(value, ",") {
			return strings.Split(value, ",")[0]
		}
		return value
	})

	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetMaxOpenConns(maxOpenConns)

	if err := runMigrations(dsn, log, migrationFolder); err != nil {
		log.Fatal("Migrations: ", err)
	}
	log.Infof("Migrations applied.")
	return db
}

func createDB(dsn string, log *logrus.Entry) error {
	log.Infof("Creating database")
	parsedDSN, err := clickhouse.ParseDSN(dsn)
	if err != nil {
		return err
	}
	databaseName := parsedDSN.Auth.Database
	defaultDsn := strings.Replace(dsn, databaseName, "default", 1)

	defaultDB, err := sqlx.Connect("clickhouse", defaultDsn)
	if err != nil {
		return err
	}
	defer defaultDB.Close()

	result, err := defaultDB.Exec(fmt.Sprintf(`CREATE DATABASE %s`, databaseName))
	if err != nil {
		log.Printf("Result: %v", result)
		return err
	}
	log.Infof("Database was created")
	return nil
}

func connect(dsn string, log *logrus.Entry) (db *sqlx.DB, err error) {
	ctx := context.Background()
	connErrCh := make(chan error, 1)
	defer close(connErrCh)

	log.Infof("connecting to database %v", dsn)
	retries := 5
	go func() {
		try := 0
		for {
			try++
			if retries > 0 && retries <= try {
				err = errors.Errorf("could not connect, dsn=%s, tries=%d", dsn, try)
				break
			}

			db, err = sqlx.Connect("clickhouse", dsn)
			if err != nil {
				log.Infof("can't connect, dsn=%s, err=%s, try=%d", dsn, err, try)
				if exception, ok := err.(*clickhouse.Exception); ok && exception.Code == databaseNotExistErrorCode {
					err = createDB(dsn, log)
					if err != nil {
						log.Infof("Database wasn't created: %v", err)
					}
				} else {
					log.Infof("Connection: %v", err)
				}

				select {
				case <-ctx.Done():
					break
				case <-time.After(3 * time.Second):
					continue
				}
			}
			break
		}
		connErrCh <- err
	}()

	select {
	case err = <-connErrCh:
		break
	case <-time.After(30 * time.Second):
		return nil, errors.Errorf("db connect timed out, dsn=%s", dsn)
	case <-ctx.Done():
		return nil, errors.Errorf("db connection cancelled, dsn=%s", dsn)
	}

	return db, err
}

func runMigrations(dsn string, log *logrus.Entry, migrationFolder string) error {
	log.Infof("dsn: %v", dsn)
	m, err := migrate.New(fmt.Sprintf("file://%v", migrationFolder), dsn)
	if err != nil {
		return err
	}

	// run up to the latest migration
	err = m.Up()
	if err == migrate.ErrNoChange {
		return nil
	}
	return err
}

// DropOldPartition drops number of days old partitions.
func DropOldPartition(db *sqlx.DB, table string, days uint, log *logrus.Entry) {
	partitions := []string{}
	const query = `
		SELECT DISTINCT partition
		FROM system.parts
		WHERE toUInt32(partition) < toYYYYMMDD(now() - toIntervalDay(?)) ORDER BY partition
	`
	err := db.Select(
		&partitions,
		query,
		days,
	)
	if err != nil {
		log.Errorf("Select %d days old partitions of system.parts. Result: %v, Error: %v", days, partitions, err)
		return
	}
	for _, part := range partitions {
		result, err := db.Exec(fmt.Sprintf(`ALTER TABLE %v DROP PARTITION %s`, table, part))
		log.Infof("Drop %s partitions of %v. Result: %v, Error: %v", part, table, result, err)
	}
}
