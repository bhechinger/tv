package main

import (
	"flag"
	"fmt"
	"github.com/bhechinger/tv/config"
	"github.com/bhechinger/tv/showsdb"
	"os"
)

func main() {
	mydb := showsdb.DBInfo{}
	var conf config.Config
	var err error
	var shows []showsdb.Shows
	defaultconfig := config.UserHomeDir() + "/.tv/shows.conf"

	listCommand := flag.NewFlagSet("list", flag.ExitOnError)
	listConfigFile := listCommand.String("config", defaultconfig, "Config file to use (Default: "+defaultconfig+")")

	addCommand := flag.NewFlagSet("add", flag.ExitOnError)
	addConfigFile := addCommand.String("config", defaultconfig, "Config file to use (Default: "+defaultconfig+")")
	addName := addCommand.String("name", "", "Show name to add")
	addSeason := addCommand.String("season", "", "Show season to add")
	addEpisode := addCommand.String("episode", "", "Show episode to add")

	getCommand := flag.NewFlagSet("get", flag.ExitOnError)
	getConfigFile := getCommand.String("config", defaultconfig, "Config file to use (Default: "+defaultconfig+")")
	getName := getCommand.String("name", "", "Show name to get")

	switch os.Args[1] {
	case "list":
		listCommand.Parse(os.Args[2:])
		conf, err = config.Get(*listConfigFile)
	case "add":
		addCommand.Parse(os.Args[2:])
		conf, err = config.Get(*addConfigFile)
	case "get":
		getCommand.Parse(os.Args[2:])
		conf, err = config.Get(*getConfigFile)
	default:
		fmt.Printf("%q is not a valid command.\n", os.Args[1])
		os.Exit(2)
	}

	if err != nil {
		fmt.Printf("Something went wrong with the config file: %s\n", err)
		os.Exit(3)
	}

	if err = mydb.Init("postgres", conf); err != nil {
		fmt.Printf("Something went wrong Initializing DB Connection: %s\n", err)
		os.Exit(3)
	}

	if err = mydb.Ping(5); err != nil {
		fmt.Printf("Something went wrong Pinging DB: %s\n", err)
		os.Exit(3)
	}

	if listCommand.Parsed() {
		shows, err = mydb.ListShows()
		if err != nil {
			fmt.Printf("Something went wrong Listing Shows: %s\n", err)
		}

		for _, v := range shows {
			fmt.Printf("%v S%02dE%02d\n", v.Name, v.Season, v.Episode)
		}
		os.Exit(0)
	}

	if addCommand.Parsed() {
		if err := mydb.AddShow(*addName, *addSeason, *addEpisode); err != nil {
			fmt.Printf("Something went wrong Adding Show: %s\n", err)
		}
		os.Exit(0)
	}

	if getCommand.Parsed() {
		exists, err := mydb.GetShow(*getName)
		if err != nil {
			fmt.Printf("Something went wrong Getting Show: %s\n", err)
		}
		if exists {
			fmt.Printf("%s exists in the database", *getName)
		} else {
			fmt.Printf("%s does not exist in the database", *getName)
		}

		os.Exit(0)
	}
}
