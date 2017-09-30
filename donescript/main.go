package main

import (
	"flag"
	"fmt"
	"github.com/bhechinger/tv/config"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	var conf config.Config

	defaultconfig := config.UserHomeDir() + "/.tv/shows.conf"
	configFile := flag.String("config", defaultconfig, "Config file to use (Default: "+defaultconfig+")")
	flag.Parse()

	conf, err := config.Get(*configFile)
	if err != nil {
		log.Printf("config.Get(): Something went wrong with the config file: %s\n", err)
		os.Exit(3)
	}

	f, err := os.OpenFile(conf.Donescript.LogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Printf("os.OpenFile(): Error opening file: %s", err)
		os.Exit(4)
	}

	defer f.Close()

	log.SetOutput(f)

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

	log.Println(output)

	var glob []string

	glob = get_glob("rar")
	if len(glob) > 0 {
		dest_dir := conf.GetDestination(filepath.Base(glob[0]))

		//binary, lookErr := exec.LookPath("unrar")
		//if lookErr != nil {
		//	panic(lookErr)
		//}

		unrar := exec.Command("unrar", "e", "-y", glob[0], dest_dir)
		_, err := unrar.Output()
		if err != nil {
			panic(err)
		}

		msg := fmt.Sprintf("uncompressing '%+v' to '%s'\n", glob[0], dest_dir)
		conf.Sendmail(fmt.Sprintf("%s downloaded!", torrent_name), msg)
		log.Println(msg)

		// We're done
		os.Exit(0)
	}

	ext_list := [...]string{"mkv", "avi", "mpg", "mp4"}

	for _, ext := range ext_list {
		glob = get_glob(ext)
		if len(glob) > 0 {
			for _, srcname := range glob {
				dest_dir := conf.GetDestination(filepath.Base(srcname))

				in, err := os.Open(srcname)
				if err != nil {
					log.Printf("os.Open(): %v", err)
					os.Exit(1)
				}
				defer in.Close()

				out, err := os.Create(fmt.Sprintf("%s/%s", dest_dir, filepath.Base(srcname)))
				if err != nil {
					log.Printf("os.Create(): %v", err)
					os.Exit(1)
				}
				defer out.Close()

				_, err = io.Copy(out, in)
				if err != nil {
					log.Printf("is.Copy(): %v", err)
					os.Exit(1)
				}

				msg := fmt.Sprintf("copying '%+v' to '%s'\n", srcname, dest_dir)
				conf.Sendmail(fmt.Sprintf("%s downloaded!", torrent_name), msg)
				log.Println(msg)
			}
		}

	}

}

func get_glob(ext string) []string {
	glob_output, err := filepath.Glob(
		fmt.Sprintf("%s/%s/*.%s",
			os.Getenv("TR_TORRENT_DIR"),
			os.Getenv("TR_TORRENT_NAME"),
			ext))

	if err != nil {
		log.Printf("filepath.Glob(): %v", err)
		os.Exit(1)
	}

	return glob_output
}
