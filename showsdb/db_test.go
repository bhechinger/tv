package showsdb

import (
	"errors"
	"github.com/bhechinger/tv/config"
	"github.com/jmoiron/sqlx"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"strings"
	"testing"
)

var test_queries = map[string]string{
	"ListShows":      "SELECT shows.name, MAX(episodes.season) as season, MAX(episodes.episode) as episode FROM episodes LEFT JOIN shows ON (episodes.show = shows.id) WHERE shows.active = $1 AND episodes.season = (select MAX(season) from episodes ep1 where ep1.show = shows.id) GROUP BY shows.name",
	"AddNewShow":     "INSERT INTO shows (name, active) VALUES ($1, $2)",
	"AddShow":        "INSERT INTO episodes (show, season, episode) VALUES ((SELECT id FROM shows WHERE name = $1), $2, $3)",
	"ShowExists":     "SELECT name, active FROM shows WHERE name = $1",
	"RemoveEpisodes": "DELETE FROM episodes WHERE show = (SELECT id FROM shows WHERE name = $1)",
	"RemoveShow":     "DELETE FROM shows WHERE name = $1",
}

func TestDatabase(t *testing.T) {
	Convey("Setup DB", t, func() {
		db, mock, err := sqlmock.New()
		So(err, ShouldBeNil)
		myDb := &DBInfo{Conn: sqlx.NewDb(db, "postgres")}

		Convey("Test Good DB Connection", func() {
			myRealDb := &DBInfo{}
			conf := config.Config{Database: config.DB{
				User:     "shows_test",
				Host:     "localhost",
				Port:     "5432",
				Password: "shows_test",
				DBName:   "shows_test",
				SSLMode:  "disable",
			}}

			err := myRealDb.Init("postgres", conf)
			Convey("Test Init()", func() {
				So(err, ShouldBeNil)
			})

			Convey("Test Ping()", func() {
				err := myRealDb.Ping(5)
				So(err, ShouldBeNil)
			})

			myRealDb.Conn.Close()
		})

		Convey("Test Bad DB Connection", func() {
			myRealDb := &DBInfo{}
			conf := config.Config{Database: config.DB{
				User:     "test",
				Host:     "localhost",
				Port:     "5432",
				Password: "shows_test",
				DBName:   "shows_test",
				SSLMode:  "disable",
			}}

			err := myRealDb.Init("postgres", conf)
			Convey("Test Init()", func() {
				// TODO: Figure out how to make this fail
				SkipSo(err, ShouldNotBeNil)
			})

			Convey("Test Ping()", func() {
				err := myRealDb.Ping(1)
				So(err, ShouldNotBeNil)
			})

			myRealDb.Conn.Close()
		})

		Convey("Test ListShows()", func() {
			mock.ExpectPrepare(
				makeQueryStringRegex(test_queries["ListShows"]),
			).ExpectQuery().WithArgs(true).WillReturnRows(
				sqlmock.NewRows(
					[]string{"name", "season", "episode"},
				).AddRow("Test Show 1", 3, 2).AddRow("Test Show 2", 2, 3))

			shows, err := myDb.ListShows()
			So(err, ShouldBeNil)

			So(shows[0].Name, ShouldEqual, "Test Show 1")
			So(shows[0].Season, ShouldEqual, 3)
			So(shows[0].Episode, ShouldEqual, 2)

			So(shows[1].Name, ShouldEqual, "Test Show 2")
			So(shows[1].Season, ShouldEqual, 2)
			So(shows[1].Episode, ShouldEqual, 3)

			// Test connection failure
			mock.ExpectPrepare(
				makeQueryStringRegex(test_queries["ListShows"]),
			).ExpectQuery().WithArgs(true).WillReturnError(errors.New("Database connection lost"))

			shows, err = myDb.ListShows()
			So(err, ShouldNotBeNil)
		})

		Convey("Test ShowExists() - Show exists", func() {
			mock.ExpectPrepare(
				makeQueryStringRegex(test_queries["ShowExists"]),
			).ExpectQuery().WithArgs("Test Show 1").WillReturnRows(sqlmock.NewRows(
				[]string{"name", "active"},
			).AddRow("Test Show 1", true).AddRow("Test Show 2", true))

			exists, err := myDb.ShowExists("Test Show 1")
			So(err, ShouldBeNil)
			So(exists, ShouldBeTrue)
		})

		Convey("Test ShowExists() - Show doesn't exist", func() {
			mock.ExpectPrepare(
				makeQueryStringRegex(test_queries["ShowExists"]),
			).ExpectQuery().WithArgs("Test Show 1").WillReturnRows(sqlmock.NewRows(
				[]string{"name", "active"},
			))

			exists, err := myDb.ShowExists("Test Show 1")
			So(err, ShouldBeNil)
			So(exists, ShouldBeFalse)
		})

		Convey("Test AddShow() - Show exists - one", func() {
			mock.ExpectPrepare(
				makeQueryStringRegex(test_queries["ShowExists"]),
			).ExpectQuery().WithArgs("Test Show 2").WillReturnRows(sqlmock.NewRows(
				[]string{"name", "active"},
			).AddRow("Test Show 1", true).AddRow("Test Show 2", true))

			mock.ExpectExec(
				makeQueryStringRegex(test_queries["AddShow"]),
			).WithArgs("Test Show 2", 1, 2).WillReturnResult(
				sqlmock.NewResult(0, 2))

			num_added, err := myDb.AddShow("Test Show 2", 1, 2, true)
			So(err, ShouldBeNil)
			So(num_added, ShouldEqual, 1)
		})

		Convey("Test AddShow() - Show exists - multi", func() {
			mock.ExpectPrepare(
				makeQueryStringRegex(test_queries["ShowExists"]),
			).ExpectQuery().WithArgs("Test Show 2").WillReturnRows(sqlmock.NewRows(
				[]string{"name", "active"},
			).AddRow("Test Show 1", true).AddRow("Test Show 2", true))

			mock.ExpectExec(
				makeQueryStringRegex(test_queries["AddShow"]),
			).WithArgs("Test Show 2", 1, 1).WillReturnResult(
				sqlmock.NewResult(0, 2))

			mock.ExpectExec(
				makeQueryStringRegex(test_queries["AddShow"]),
			).WithArgs("Test Show 2", 1, 2).WillReturnResult(
				sqlmock.NewResult(0, 2))

			num_added, err := myDb.AddShow("Test Show 2", 1, 2, false)
			So(err, ShouldBeNil)
			So(num_added, ShouldEqual, 2)
		})

		Convey("Test AddShow() - Show doesn't exists - one", func() {
			mock.ExpectPrepare(
				makeQueryStringRegex(test_queries["ShowExists"]),
			).ExpectQuery().WithArgs("Test Show 2").WillReturnRows(sqlmock.NewRows(
				[]string{"name", "active"},
			))

			mock.ExpectExec(
				makeQueryStringRegex(test_queries["AddNewShow"]),
			).WithArgs("Test Show 2", true).WillReturnResult(
				sqlmock.NewResult(0, 1))

			mock.ExpectExec(
				makeQueryStringRegex(test_queries["AddShow"]),
			).WithArgs("Test Show 2", 1, 2).WillReturnResult(
				sqlmock.NewResult(0, 2))

			num_added, err := myDb.AddShow("Test Show 2", 1, 2, true)
			So(err, ShouldBeNil)
			So(num_added, ShouldEqual, 1)
		})

		Convey("Test AddShow() - Show doesn't exists - multi", func() {
			mock.ExpectPrepare(
				makeQueryStringRegex(test_queries["ShowExists"]),
			).ExpectQuery().WithArgs("Test Show 2").WillReturnRows(sqlmock.NewRows(
				[]string{"name", "active"},
			))

			mock.ExpectExec(
				makeQueryStringRegex(test_queries["AddNewShow"]),
			).WithArgs(
				"Test Show 2",
				true,
			).WillReturnResult(sqlmock.NewResult(0, 1))

			mock.ExpectExec(
				makeQueryStringRegex(test_queries["AddShow"]),
			).WithArgs("Test Show 2", 1, 1).WillReturnResult(
				sqlmock.NewResult(0, 2))

			mock.ExpectExec(
				makeQueryStringRegex(test_queries["AddShow"]),
			).WithArgs("Test Show 2", 1, 2).WillReturnResult(
				sqlmock.NewResult(0, 2))

			numAdded, err := myDb.AddShow("Test Show 2", 1, 2, false)
			So(err, ShouldBeNil)
			So(numAdded, ShouldEqual, 2)
		})

		Convey("TestRemoveShow()", func() {
			mock.ExpectExec(
				makeQueryStringRegex(test_queries["RemoveEpisodes"]),
			).WithArgs("Test Show 2").WillReturnResult(
				sqlmock.NewResult(0, 1))

			mock.ExpectExec(
				makeQueryStringRegex(test_queries["RemoveShow"]),
			).WithArgs("Test Show 2").WillReturnResult(
				sqlmock.NewResult(0, 1))

			numRemoved, err := myDb.RemoveShow("Test Show 2")
			So(err, ShouldBeNil)
			So(numRemoved, ShouldEqual, 1)
		})
	})
}

func makeQueryStringRegex(queryString string) string {
	sqlRegex := strings.Replace(queryString, "(", ".", -1)
	sqlRegex = strings.Replace(sqlRegex, ")", ".", -1)
	sqlRegex = strings.Replace(sqlRegex, "?", ".", -1)
	sqlRegex = strings.Replace(sqlRegex, ":", ".", -1)
	sqlRegex = strings.Replace(sqlRegex, "$", ".", -1)

	return sqlRegex
}
