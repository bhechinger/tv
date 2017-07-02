package showsdb

import (
	"errors"
	"fmt"
	"github.com/bhechinger/tv/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"time"
)

type DB interface {
	Init(driver string) error
	Ping(timeout int) error
}

type DBInfo struct {
	Driver string
	DSN    string
	Conn   *sqlx.DB
}

type Shows struct {
	Name    string
	Season  int
	Episode int
	Active  bool
}

func (db *DBInfo) Init(driver string, config config.Config) error {
	var err error
	db.Driver = driver
	db.DSN = fmt.Sprintf("user=%s host=%s password=%s dbname=%s sslmode=%s",
		config.Database.User,
		config.Database.Host,
		config.Database.Password,
		config.Database.DBName,
		config.Database.SSLMode)

	if db.Conn, err = sqlx.Open(db.Driver, db.DSN); err != nil {
		return errors.New("Error opening connection")
	}

	return nil
}

func (db *DBInfo) Ping(timeout int) error {
	for try := 0; ; try++ {
		if try > timeout {
			return errors.New("Timing out attempting to ping database")
		}
		if err := db.Conn.Ping(); err == nil {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
}

func (db *DBInfo) ListShows() ([]Shows, error) {
	stmt, err := db.Conn.PrepareNamed("SELECT shows.name, MAX(episodes.season) as season, MAX(episodes.episode) as episode FROM episodes LEFT JOIN shows ON (episodes.show = shows.id) WHERE shows.active = :active AND episodes.season = (select MAX(season) from episodes ep1 where ep1.show = shows.id) GROUP BY shows.name")
	s := Shows{Active: true}
	shows := []Shows{}

	if err = stmt.Select(&shows, s); err != nil {
		return shows, fmt.Errorf("Error selecting shows/seasons: %s", err)
	}

	return shows, nil
}
