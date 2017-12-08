package showsdb

import (
    . "github.com/smartystreets/goconvey/convey"
    "gopkg.in/DATA-DOG/go-sqlmock.v1"
    "testing"
    "github.com/jmoiron/sqlx"
    "strings"
)

var test_queries = map[string]string{
	"ListShows": "SELECT shows.name, MAX(episodes.season) as season, MAX(episodes.episode) as episode FROM episodes LEFT JOIN shows ON (episodes.show = shows.id) WHERE shows.active = $1 AND episodes.season = (select MAX(season) from episodes ep1 where ep1.show = shows.id) GROUP BY shows.name",
	"AddShow":   "CM",
}

func TestDatabase(t *testing.T) {
    Convey("Setup DB", t, func() {
        db, mock, err := sqlmock.New()
        So(err, ShouldBeNil)
        myDb := &DBInfo{Conn: sqlx.NewDb(db, "postgres")}

        mock.ExpectPrepare(
            makeQueryStringRegex(test_queries["ListShows"]),
            ).ExpectQuery().WithArgs(true).WillReturnRows(
            sqlmock.NewRows(
                []string{
                    "name",
                    "season",
                    "episode",
                },
            ).AddRow("Test Show 1", 3, 2).AddRow("Test Show 2", 2, 3))

        Convey("Test ListShows()", func(){
            shows, err := myDb.ListShows()
            So(err, ShouldBeNil)

            So(shows[0].Name, ShouldEqual, "Test Show 1")
			So(shows[0].Season, ShouldEqual, 3)
			So(shows[0].Episode, ShouldEqual, 2)

            So(shows[1].Name, ShouldEqual, "Test Show 2")
			So(shows[1].Season, ShouldEqual, 2)
			So(shows[1].Episode, ShouldEqual,3)
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
