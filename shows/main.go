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
	addSeason := addCommand.Int("season", 0, "Show season to add")
	addEpisode := addCommand.Int("episode", 0, "Show episode to add")
	addOne := addCommand.Bool("one", false, "Add only one episode, don't pad missing")

	getCommand := flag.NewFlagSet("get", flag.ExitOnError)
	getConfigFile := getCommand.String("config", defaultconfig, "Config file to use (Default: "+defaultconfig+")")
	getName := getCommand.String("name", "", "Show name to get")

	removeCommand := flag.NewFlagSet("remove", flag.ExitOnError)
	removeConfigFile := removeCommand.String("config", defaultconfig, "Config file to use (Default: "+defaultconfig+")")
	removeName := removeCommand.String("name", "", "Show to remvoe")

	// TODO: error handling for no command line arguments
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
	case "remove":
		removeCommand.Parse(os.Args[2:])
		conf, err = config.Get(*removeConfigFile)
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
			os.Exit(0)
		}

		for _, v := range shows {
			fmt.Printf("%v S%02dE%02d\n", v.Name, v.Season, v.Episode)
		}
	}

	if addCommand.Parsed() {
		added, err := mydb.AddShow(*addName, *addSeason, *addEpisode, *addOne)
		if err != nil {
			fmt.Printf("Something went wrong Adding Show: %s\n", err)
			os.Exit(0)
		}

		if added == 0 {
			fmt.Printf("Episode for %s not added\n", *addName)
		} else {
			fmt.Printf("Added %s with %d episodes\n", *addName, added)
		}
	}

	if getCommand.Parsed() {
		exists, err := mydb.ShowExists(*getName)
		if err != nil {
			fmt.Printf("Something went wrong Getting Show: %s\n", err)
			os.Exit(0)
		}
		if exists {
			fmt.Printf("%s exists in the database\n", *getName)
		} else {
			fmt.Printf("%s does not exist in the database\n", *getName)
		}

	}

	if removeCommand.Parsed() {
		count, err := mydb.RemoveShow(*removeName)
		if err != nil {
			fmt.Printf("Something went wrong Removing Show: %s\n", err)
			os.Exit(0)
		}
		fmt.Printf("Deleted %d episodes of %s\n", count, *removeName)
	}
}
