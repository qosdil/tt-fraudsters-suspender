package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

const (
	driverName = "postgres"
)

func (db *Database) DisableUser(userID string) error {
	query := `UPDATE public.users SET is_enabled=FALSE WHERE id = $1`
	_, err := db.SqlDB.Exec(query, userID)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) CreateUser(userID string, email string) error {
	query := `INSERT INTO public.users (id, email, is_enabled) VALUES($1, $2, TRUE)`
	_, err := db.SqlDB.Exec(query, userID, email)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) Open() (*sql.DB, error) {
	dataSourceName := fmt.Sprintf("%s://%s:%s@%s/%s?sslmode=%s", driverName, db.Config.User,
		db.Config.Password, db.Config.Host, db.Config.Name, db.Config.SSLMode)
	sqlDB, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	return sqlDB, nil
}

func NewDatabase(config Config) *Database {
	db := new(Database)
	db.Config = config
	return db
}

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type Database struct {
	Config Config
	SqlDB  *sql.DB
}
