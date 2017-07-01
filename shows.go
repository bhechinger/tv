package main

import (
	"fmt"
	"github.com/bhechinger/tv/config"
	"github.com/bhechinger/tv/showsdb"
)

func main() {
	mydb := showsdb.DBInfo{}
	homedir := config.UserHomeDir()

	conf, err := config.Get(homedir + "/.tv/shows.conf")
	if err != nil {
		fmt.Printf("Something went wrong: %s\n", err)
	}

	if err := mydb.Init("postgres", conf); err != nil {
		fmt.Printf("Something went wrong Initializing DB Connection: %s\n", err)
	}

	if err := mydb.Ping(5); err != nil {
		fmt.Printf("Something went wrong Pinging DB: %s\n", err)
	}

	if err := mydb.ListShows(); err != nil {
		fmt.Printf("Something went wrong Listing Shows: %s\n", err)
	}
}
