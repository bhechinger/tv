package main

import (
	"github.com/bhechinger/tv/config"
	"flag"
	"fmt"
	"os"
	"net/http"
	"io/ioutil"
	"encoding/xml"
	"github.com/bhechinger/tv/showsdb"
	"github.com/tubbebubbe/transmission"
	"regexp"
	"strconv"
	"log"
)

type Query struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	ItemList []Item `xml:"item"`
}

type Item struct {
	Title  string `xml:"title"`
	Link   string `xml:"link"`
}

func (i Item) String() string {
	return fmt.Sprintf("%s - %s", i.Title, i.Link)
}

func main() {
	mydb := showsdb.DBInfo{}
	var shows []showsdb.Shows
	var conf config.Config

	defaultconfig := config.UserHomeDir() + "/.tv/shows.conf"
	configFile := flag.String("config", defaultconfig, "Config file to use (Default: "+defaultconfig+")")
	flag.Parse()

	conf, err := config.Get(*configFile)
	if err != nil {
		fmt.Printf("Something went wrong with the config file: %s\n", err)
		os.Exit(3)
	}

	f, err := os.OpenFile(conf.RSSFeed.LogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Printf("os.OpenFile(): Error opening file: %s", err)
		os.Exit(4)
	}

	defer f.Close()

	log.SetOutput(f)
	log.Println("Running rss_feed")

	if err = mydb.Init("postgres", conf); err != nil {
		log.Printf("Something went wrong Initializing DB Connection: %s\n", err)
		os.Exit(3)
	}

	if err = mydb.Ping(5); err != nil {
		log.Printf("Something went wrong Pinging DB: %s\n", err)
		os.Exit(3)
	}

	shows, err = mydb.ListShows()
	if err != nil {
		log.Printf("Something went wrong Listing Shows: %s\n", err)
		os.Exit(0)
	}

	resp, err := http.Get(conf.RSSFeed.BaseURL+"/"+conf.RSSFeed.Key)
	if err != nil {
		log.Printf("Something has gone terribly wrong connecting to the server: %s\n", err)
		os.Exit(2)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Something has gone quite wrong fetching the body content: %s\n", err)
		os.Exit(1)
	}

	var q Query
	xml.Unmarshal(body, &q)

	client := transmission.New(conf.Transmission.URI, conf.Transmission.Username, conf.Transmission.Password)

	for _, item := range q.Channel.ItemList {
		for _, show := range shows {
			re := regexp.MustCompile(fmt.Sprintf("^%s S(?P<season>[0-9]+)E(?P<episode>[0-9]+)", show.Name))
			n1 := re.SubexpNames()
			r2 := re.FindAllStringSubmatch(item.Title, -1)

			if len(r2) > 0 {
				md := map[string]string{}
				for i, n := range r2[0] {
					md[n1[i]] = n
				}
				season, err := strconv.Atoi(md["season"])
				if err != nil {
					log.Printf("Something went wrong: %v\n", err)
				}
				episode, err := strconv.Atoi(md["episode"])
				if err != nil {
					log.Printf("Something went wrong: %v\n", err)
				}
				log.Printf("Checking %s S%dE%d\n", show.Name, season, episode)

				added, err := mydb.AddShow(show.Name, season, episode, true)
				if err != nil {
					log.Printf("Something went wrong: %v\n", err)
				}
				if added == 1 {
					addCommand, err := transmission.NewAddCmdByURL(item.Link)
					if err != nil {
						log.Printf("Something went wrong creating the addCommand: %s\n", err)
						continue
					}

					result, err := client.ExecuteAddCommand(addCommand)
					if err != nil {
						log.Printf("Something went wrong adding the torrent: %s\n", err)
						continue
					}

					log.Printf("Added Torrent: %s\n", result.Name)
				}
			}
		}
	}
}