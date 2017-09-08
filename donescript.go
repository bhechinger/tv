package main

import (
	"flag"
	"fmt"
	"github.com/bhechinger/tv/config"
	unarr "github.com/gen2brain/go-unarr"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"gopkg.in/gomail.v2"
	"strings"
)

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

	f, err := os.OpenFile(conf.Donescript.LogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Printf("Error opening file: %s", err)
		os.Exit(4)
	}

	defer f.Close()

	app_version := os.Getenv("TR_APP_VERSION")
	time_localtime := os.Getenv("TR_TIME_LOCALTIME")
	torrent_dir := os.Getenv("TR_TORRENT_DIR")
	torrent_hash := os.Getenv("TR_TORRENT_HASH")
	torrent_id := os.Getenv("TR_TORRENT_ID")
	torrent_name := os.Getenv("TR_TORRENT_NAME")

	output := fmt.Sprintf("Environment:\n"+
		"\tTR_APP_VERSION: %s\n"+
		"\tTR_TIME_LOCALTIME: %s\n"+
		"\tTR_TORRENT_DIR: %s\n"+
		"\tTR_TORRENT_HASH: %s\n"+
		"\tTR_TORRENT_ID: %s\n"+
		"\tTR_TORRENT_NAME: %s\n\n",
		app_version,
		time_localtime,
		torrent_dir,
		torrent_hash,
		torrent_id,
		torrent_name)

	if _, err = f.WriteString(output); err != nil {
		panic(err)
	}

	var glob []string

	glob = get_glob("rar")
	if len(glob) > 0 {
		dest_dir := get_dest(filepath.Base(glob[0]))
		msg := fmt.Sprintf("uncompressing '%+v' to '%s'\n", glob[0], dest_dir)
		send_mail(fmt.Sprintf("%s downloaded!", torrent_name), msg, conf)
		_, err = f.WriteString(msg)
		if err != nil {
			panic(err)
		}

		a, err := unarr.NewArchive(glob[0])
		if err != nil {
			panic(err)
		}

		defer a.Close()

		err = a.Extract(dest_dir)
		if err != nil {
			panic(err)
		}
	}

	ext_list := [...]string{"mkv", "avi", "mpg", "mp4"}

	for _, ext := range ext_list {
		glob = get_glob(ext)
		if len(glob) > 0 {
			for _, srcname := range glob {
				dest_dir := get_dest(filepath.Base(srcname))
				msg := fmt.Sprintf("copying '%+v' to '%s'\n", srcname, dest_dir)
				send_mail(fmt.Sprintf("%s downloaded!", torrent_name), msg, conf)
				_, err = f.WriteString(msg)
				if err != nil {
					panic(err)
				}

				in, err := os.Open(srcname)
				if err != nil {
					panic(err)
				}
				defer in.Close()

				out, err := os.Create(fmt.Sprintf("%s/%s", dest_dir, filepath.Base(srcname)))
				if err != nil {
					panic(err)
				}
				defer out.Close()

				_, err = io.Copy(out, in)
				cerr := out.Close()
				if err != nil {
					panic(err)
				}
				fmt.Sprintf("cerr: %+v\n", cerr)
			}
		}

	}

}

func get_dest(name string) string {
	matched, err := regexp.MatchString(`[Ss]\d{2}[Ee]\d{2}`, name)
	if err != nil {
		panic(err)
	}

	if matched {
		return "/tank/Plex/TV/1New/"
	} else {
		return "/tank/Plex/Movies/"
	}

}

func get_glob(ext string) []string {
	glob_output, err := filepath.Glob(
		fmt.Sprintf("%s/%s/*.%s",
			os.Getenv("TR_TORRENT_DIR"),
			os.Getenv("TR_TORRENT_NAME"),
			ext))

	if err != nil {
		panic(err)
	}

	return glob_output
}

func send_mail(subject, body string, conf config.Config) {
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
		panic(err)
	}
}
