package db

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
	"path"

	"github.com/Erexo/Ventana/infrastructure/config"
	"github.com/georgysavva/scany/sqlscan"
	_ "modernc.org/sqlite"
)

func Initialize() error {
	_, err := os.Stat(getDatabasePath())
	if err == nil {
		log.Println("Database is already initialized")
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	conn, err := GetConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	schema := `
		CREATE TABLE user (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,	
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			salt TEXT,
			role INTEGER NOT NULL
			);
		CREATE TABLE thermometer (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			sensor TEXT UNIQUE NOT NULL
			);
		CREATE TABLE thermaldata (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			thermometerid INTEGER NOT NULL,
			celsius REAL NOT NULL,
			timestamp DATETIME NOT NULL,
			CONSTRAINT fk_thermometer FOREIGN KEY (thermometerid)
				REFERENCES thermometer(id) ON DELETE CASCADE
		);`
	statement, err := conn.Prepare(schema)
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	if err != nil {
		return err
	}

	initialData := `INSERT INTO user (username, password, role) VALUES ('admin', 'admin1', 3);
		INSERT INTO thermometer (name, sensor) VALUES ('bedroom', '28-011876e3d3ff');`
	statement, err = conn.Prepare(initialData)
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	if err != nil {
		return err
	}

	log.Println("Initialized database")
	return nil
}

func Scan(query string, args interface{}, dest ...interface{}) error {
	conn, err := GetConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	conn.QueryRow(query, args).Scan(dest)
	return nil
}

func Get(dest interface{}, query string, args ...interface{}) error {
	conn, err := GetConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	return sqlscan.Get(context.Background(), conn, dest, query, args...)
}

func Select(dest interface{}, query string, args ...interface{}) error {
	conn, err := GetConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	err = sqlscan.Select(context.Background(), conn, dest, query, args...)
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	return nil
}

func Exec(query string, args ...interface{}) (sql.Result, error) {
	conn, err := GetConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	st, err := conn.Prepare(query)
	if err != nil {
		return nil, err
	}
	return st.Exec(args...)
}

func GetConnection() (*sql.DB, error) {
	return sql.Open("sqlite", getDatabasePath())
}

func getDatabasePath() string {
	cfgDb := config.GetConfig().DatabaseFile
	if cfgDb.Valid {
		return cfgDb.String
	}
	wd, _ := os.Getwd()
	return path.Join(wd, "ventana.db")
}
