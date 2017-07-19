package showsdb

import (
	"errors"
	"fmt"
	"github.com/bhechinger/tv/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"time"
)

const MAXEPISODES = 99

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
	db.DSN = fmt.Sprintf("user=%s host=%s port=%s password=%s dbname=%s sslmode=%s",
		config.Database.User,
		config.Database.Host,
		config.Database.Port,
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

func (db *DBInfo) AddShow(name string, season int, episode int, one bool) (int, error) {
	added := 0
	exists, err := db.GetShow(name)
	if err != nil {
		return added, fmt.Errorf("Error checking if show exists: %s", err)
	}

	if !exists {
		sh := Shows{Name: name, Active: true}
		// We need to create the show first in the show table
		if _, err := db.Conn.NamedExec("INSERT INTO shows (name, active) VALUES (:name, :active)", sh); err != nil {
			return added, fmt.Errorf("Error Inserting shows: %s", err)
		}
	}

	if one {
		added = 1
		sh := Shows{Name: name, Season: season, Episode: episode}
		if _, err := db.Conn.NamedExec("INSERT INTO episodes (show, season, episode) VALUES ((SELECT id FROM shows WHERE name = :name), :season, :episode)", sh); err != nil {
			if err.Error() == "pq: duplicate key value violates unique constraint \"episodes_show_season_episode_key\"" {
				return 0, nil
			}
			return 0, fmt.Errorf("Error Inserting episode: %s", err)
		}
		return 1, nil
	}

	for s := 1; s <= season; s++ {
		e_limit := MAXEPISODES
		if s == season {
			e_limit = episode
		}

		for e := 1; e <= e_limit; e++ {
			sh := Shows{Name: name, Season: s, Episode: e}
			if _, err := db.Conn.NamedExec("INSERT INTO episodes (show, season, episode) VALUES ((SELECT id FROM shows WHERE name = :name), :season, :episode)", sh); err != nil {
				return added, fmt.Errorf("Error Inserting seasons: %s", err)
			}
		}
	}

	return added, nil
}

func (db *DBInfo) RemoveShow(name string) (int64, error) {
	s := Shows{Name: name}
	result, err := db.Conn.NamedExec("DELETE FROM episodes WHERE show = (SELECT id FROM shows WHERE name = :name)", s)
	if err != nil {
		return 0, fmt.Errorf("Error deleting shows from episide table: %s", err)
	}

	episodes_deleted, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("Error getting RowsAffected(): %s", err)
	}

	if _, err := db.Conn.NamedExec("DELETE FROM shows WHERE name = :name", s); err != nil {
		return 0, fmt.Errorf("Error deleting shows from shows table: %s", err)
	}

	return episodes_deleted, nil
}

func (db *DBInfo) GetShow(name string) (bool, error) {
	shows := []Shows{}
	if err := db.Conn.Select(&shows, "SELECT name, active FROM shows"); err != nil {
		return false, err
	}

	for _, v := range shows {
		if name == v.Name {
			return true, nil
		}
	}

	return false, nil
}
