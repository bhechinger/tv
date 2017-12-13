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

var queries = map[string]string{
	"ListShows":     "SELECT shows.name, MAX(episodes.season) as season, MAX(episodes.episode) as episode FROM episodes LEFT JOIN shows ON (episodes.show = shows.id) WHERE shows.active = :active AND episodes.season = (select MAX(season) from episodes ep1 where ep1.show = shows.id) GROUP BY shows.name",
	"AddNewShow":    "INSERT INTO shows (name, active) VALUES (:name, :active)",
	"AddShow":       "INSERT INTO episodes (show, season, episode) VALUES ((SELECT id FROM shows WHERE name = :name), :season, :episode)",
	"ShowExists":    "SELECT name, active FROM shows WHERE name = :name",
	"RemoveEpisode": "DELETE FROM episodes WHERE show = (SELECT id FROM shows WHERE name = :name)",
	"RemoveShow":    "DELETE FROM shows WHERE name = :name",
}

func (db *DBInfo) Init(driver string, config config.Config) (err error) {
	var dbErr error
	db.Driver = driver
	db.DSN = fmt.Sprintf("user=%s host=%s port=%s password=%s dbname=%s sslmode=%s",
		config.Database.User,
		config.Database.Host,
		config.Database.Port,
		config.Database.Password,
		config.Database.DBName,
		config.Database.SSLMode)

	if db.Conn, err = sqlx.Open(db.Driver, db.DSN); dbErr != nil {
		return dbErr
	}

	return nil
}

func (db *DBInfo) Ping(timeout int) (err error) {
	for try := 0; ; try++ {
		if try > timeout {
			return errors.New("Timing out attempting to ping database")
		}
		if pingErr := db.Conn.Ping(); pingErr == nil {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
}

func (db *DBInfo) ListShows() (showList []Shows, err error) {
	stmt, dbErr := db.Conn.PrepareNamed(queries["ListShows"])
	s := Shows{Active: true}
	shows := []Shows{}

	if dbErr = stmt.Select(&shows, s); dbErr != nil {
		return shows, fmt.Errorf("Error selecting shows/seasons: %s", dbErr)
	}

	return shows, nil
}

func (db *DBInfo) AddShow(name string, season int, episode int, one bool) (numAdded int, err error) {
	added := 0
	exists, dbErr := db.ShowExists(name)
	if dbErr != nil {
		return added, fmt.Errorf("Error checking if show exists: %s", dbErr)
	}

	if !exists {
		sh := Shows{Name: name, Active: true}
		// We need to create the show first in the show table
		if _, dbErr := db.Conn.NamedExec(queries["AddNewShow"], sh); dbErr != nil {
			return added, fmt.Errorf("Error Inserting shows: %s", dbErr)
		}
	}

	if one {
		added = 1
		sh := Shows{Name: name, Season: season, Episode: episode}
		if _, dbErr := db.Conn.NamedExec(queries["AddShow"], sh); dbErr != nil {
			if dbErr.Error() == "pq: duplicate key value violates unique constraint \"episodes_show_season_episode_key\"" {
				return 0, nil
			}
			return 0, fmt.Errorf("Error Inserting episode: %s", dbErr)
		}
		return 1, nil
	}

	for s := 1; s <= season; s++ {
		eLimit := MAXEPISODES
		if s == season {
			eLimit = episode
		}

		for e := 1; e <= eLimit; e++ {
			sh := Shows{Name: name, Season: s, Episode: e}
			if _, dbErr := db.Conn.NamedExec(queries["AddShow"], sh); dbErr != nil {
				return added, fmt.Errorf("Error Inserting seasons: %s", dbErr)
			}
			added++
		}
	}

	return added, nil
}

func (db *DBInfo) RemoveShow(name string) (numRemoved int64, err error) {
	s := Shows{Name: name}
	result, dbErr := db.Conn.NamedExec(queries["RemoveEpisode"], s)
	if dbErr != nil {
		return 0, fmt.Errorf("Error deleting shows from episide table: %s", dbErr)
	}

	episodesDeleted, dbErr := result.RowsAffected()
	if dbErr != nil {
		return 0, fmt.Errorf("Error getting RowsAffected(): %s", dbErr)
	}

	if _, dbErr := db.Conn.NamedExec(queries["RemoveShow"], s); dbErr != nil {
		return 0, fmt.Errorf("Error deleting shows from shows table: %s", dbErr)
	}

	return episodesDeleted, nil
}

func (db *DBInfo) ShowExists(name string) (exists bool, err error) {
	stmt, dbErr := db.Conn.PrepareNamed(queries["ShowExists"])
	s := Shows{Name: name}
	shows := []Shows{}

	if dbErr = stmt.Select(&shows, s); dbErr != nil {
		return false, fmt.Errorf("Error checking if the show exists: %s", dbErr)
	}

	if len(shows) == 0 {
		return false, nil
	}

	return true, nil
}
