package showsdb

import (
	. "github.com/smartystreets/goconvey/convey"
	"github.com/bhechinger/tv/config"
	"testing"
)

func TestDatabase(t *testing.T) {
	Convey("Setup DB", t, func() {
		conf := config.Config{Database: config.DB{DBName: "shows_test", Host: "localhost", Port: "5432", User: "shows_test", Password: "shows_test", SSLMode: "disable"}}
		mydb := DBInfo{}

		So(mydb.Init("postgres", conf), ShouldBeNil)
		So(mydb.Ping(5), ShouldBeNil)
		Convey("Test ListShows()", func(){
			shows, err := mydb.ListShows()
			So(err, ShouldBeNil)
			So(shows[0].Name, ShouldEqual, "Test Show 1")
			So(shows[1].Name, ShouldEqual, "Test Show 2")
			So(shows[2].Name, ShouldEqual, "Test Show 3")
		})

	})
}
