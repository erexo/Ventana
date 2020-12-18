package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
		CREATE TABLE thermometerorder (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			userid INTEGER NOT NULL REFERENCES user(id) ON DELETE CASCADE,
			thermometerid INTEGER NOT NULL REFERENCES thermometer(id) ON DELETE CASCADE
		);
		CREATE TABLE thermaldata (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			thermometerid INTEGER NOT NULL REFERENCES thermometer(id) ON DELETE CASCADE,
			celsius REAL NOT NULL,
			timestamp DATETIME NOT NULL
		);
		CREATE TABLE sunblind (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			inputdownpin INTEGER NOT NULL,
			inputuppin INTEGER NOT NULL,
			outputdownpin INTEGER NOT NULL,
			outputuppin INTEGER NOT NULL
		);
		CREATE TABLE sunblindorder (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			userid INTEGER NOT NULL REFERENCES user(id) ON DELETE CASCADE,
			sunblindid INTEGER NOT NULL REFERENCES sunblind(id) ON DELETE CASCADE
		);
		CREATE TABLE light (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			inputpin INTEGER NOT NULL,
			outputpin INTEGER NOT NULL
		);
		CREATE TABLE lightorder (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			userid INTEGER NOT NULL REFERENCES user(id) ON DELETE CASCADE,
			lightid INTEGER NOT NULL REFERENCES light(id) ON DELETE CASCADE
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
		INSERT INTO thermometer (name, sensor) VALUES ('bedroom', '28-011876e3d3ff');
		INSERT INTO sunblind (name, inputdownpin, inputuppin, outputdownpin, outputuppin) VALUES ('livingroom', 0, 1, 2, 3);
		`
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

func GetConnection() (*sql.DB, error) {
	conn, err := sql.Open("sqlite", getDatabasePath())
	if err != nil {
		return nil, err
	}
	_, err = conn.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

func GetTransaction() (*sql.Tx, func() error, error) {
	conn, err := GetConnection()
	if err != nil {
		return nil, nil, err
	}
	tx, err := conn.Begin()
	if err != nil {
		return nil, nil, err
	}
	return tx, func() error {
		terr := tx.Rollback()
		if terr != nil && errors.Is(terr, sql.ErrTxDone) {
			terr = nil
		}
		cerr := conn.Close()
		var err error
		if terr == nil {
			err = cerr
		} else if cerr == nil {
			err = terr
		} else {
			err = fmt.Errorf("tx: %s, conn: %w", terr.Error(), cerr)
		}
		if err != nil {
			log.Println(err)
			return err
		}
		return nil
	}, nil
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
	return sqlscan.Get(Ctx(), conn, dest, query, args...)
}

func Select(dest interface{}, query string, args ...interface{}) error {
	conn, err := GetConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	err = sqlscan.Select(Ctx(), conn, dest, query, args...)
	if IsError(err) {
		return err
	}
	return nil
}

func Exec(query string, args ...interface{}) (sql.Result, error) {
	conn, err := GetConnection()
	if err != nil {
		return nil, err
	}
	x, _ := conn.Begin()
	x.Commit()
	defer conn.Close()
	st, err := conn.Prepare(query)
	if err != nil {
		return nil, err
	}
	return st.Exec(args...)
}

func IsError(err error) bool {
	return err != nil && !errors.Is(err, sql.ErrNoRows)
}

func Ctx() context.Context {
	return context.Background()
}

func getDatabasePath() string {
	cfgDb := config.GetConfig().DatabaseFile
	if cfgDb.Valid {
		return cfgDb.String
	}
	wd, _ := os.Getwd()
	return path.Join(wd, "ventana.db")
}
