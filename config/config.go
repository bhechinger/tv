package config

import (
	"github.com/BurntSushi/toml"
	"gopkg.in/gomail.v2"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"runtime"
	"strings"
)

type Config struct {
	Database     DB
	RSSFeed      RSS
	Options      Options
	Transmission Transmission
	EMail        EMail
	Donescript   Donescript
}

type RSS struct {
	LogFile string
	BaseURL string
	Key     string
}

type Options struct {
	Prefer string
}

type Transmission struct {
	URI      string
	Username string
	Password string
}

type DB struct {
	User     string
	Host     string
	Port     string
	Password string
	DBName   string
	SSLMode  string
}

type EMail struct {
	Username      string
	Password      string
	Server        string
	Port          int
	From          string
	RecipientList string
}

type Donescript struct {
	LogFile       string
	TVLocation    string
	MovieLocation string
}

func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func Get(filename string) (Config, error) {
	var config Config

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, err
	}

	// TODO: Figure out how best to set defaults
	if _, err := toml.Decode(string(data), &config); err != nil {
		return config, err
	}
	return config, nil
}

func (conf Config) Sendmail(subject, body string) {
	var recipient_list []string

	for _, recipient := range strings.Split(conf.EMail.RecipientList, ",") {
		recipient_list = append(recipient_list, recipient)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", conf.EMail.From)
	m.SetHeader("To", recipient_list...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(conf.EMail.Server, conf.EMail.Port, conf.EMail.Username, conf.EMail.Password)

	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		log.Printf("d.DialAndSend(): %v", err)
		os.Exit(1)
	}
}

func (conf Config) GetDestination(name string) string {
	matched, err := regexp.MatchString(`[Ss]\d{2}[Ee]\d{2}`, name)
	if err != nil {
		log.Printf("regexp.MatchString(): %v", err)
		os.Exit(1)
	}

	if matched {
		return conf.Donescript.TVLocation
	} else {
		return conf.Donescript.MovieLocation
	}

}
