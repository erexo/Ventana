package db

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"

	"github.com/georgysavva/scany/sqlscan"
	_ "modernc.org/sqlite"
)

const databasePath = "./ventana.db"

func Initialize() error {
	_, err := os.Stat(databasePath)
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

	schema := `CREATE TABLE user (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,		
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		salt TEXT,
		role INTEGER NOT NULL
		);`
	statement, err := conn.Prepare(schema)
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

func Get(dest interface{}, query string, args ...interface{}) error {
	conn, err := GetConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	return sqlscan.Get(context.Background(), conn, dest, query, args...)
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
	return sql.Open("sqlite", databasePath)
}
