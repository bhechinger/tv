package main

import (
	"github.com/bhechinger/tv/config"
	"flag"
	"fmt"
	"os"
	"net/http"
	"io/ioutil"
	"encoding/xml"
)

type Query struct {
	channel Channel
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
	var conf config.Config
	defaultconfig := config.UserHomeDir() + "/.tv/shows.conf"
	configFile := flag.String("config", defaultconfig, "Config file to use (Default: "+defaultconfig+")")
	flag.Parse()
	conf, err := config.Get(*configFile)
	if err != nil {
		fmt.Printf("Something went wrong with the config file: %s\n", err)
		os.Exit(3)
	}

	resp, err := http.Get(conf.RSSFeed.BaseURL+"/"+conf.RSSFeed.Key)
	if err != nil {
		fmt.Printf("Something has gone terribly wrong connecting to the server: %s\n", err)
		os.Exit(2)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Something has gone quite wrong fetching the body content: %s\n", err)
		os.Exit(1)
	}

	var q Query
	xml.Unmarshal(body, &q)

	for _, item := range q.channel.ItemList {
		fmt.Printf("\t%s\n", item)
	}
}