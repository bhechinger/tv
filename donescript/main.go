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

	appVersion := os.Getenv("TR_APP_VERSION")
	timeLocaltime := os.Getenv("TR_TIME_LOCALTIME")
	torrentDir := os.Getenv("TR_TORRENT_DIR")
	torrentHash := os.Getenv("TR_TORRENT_HASH")
	torrentId := os.Getenv("TR_TORRENT_ID")
	torrentName := os.Getenv("TR_TORRENT_NAME")

	output := fmt.Sprintf("Environment:\n"+
		"\tTR_APP_VERSION: %s\n"+
		"\tTR_TIME_LOCALTIME: %s\n"+
		"\tTR_TORRENT_DIR: %s\n"+
		"\tTR_TORRENT_HASH: %s\n"+
		"\tTR_TORRENT_ID: %s\n"+
		"\tTR_TORRENT_NAME: %s\n\n",
		appVersion,
		timeLocaltime,
		torrentDir,
		torrentHash,
		torrentId,
		torrentName)

	log.Println(output)

	var glob []string

	glob = getGlob("rar")
	if len(glob) > 0 {
		destDir := conf.GetDestination(filepath.Base(glob[0]))

		//binary, lookErr := exec.LookPath("unrar")
		//if lookErr != nil {
		//	panic(lookErr)
		//}

		unrar := exec.Command("unrar", "e", "-y", glob[0], destDir)
		_, err := unrar.Output()
		if err != nil {
			panic(err)
		}

		msg := fmt.Sprintf("uncompressing '%+v' to '%s'\n", glob[0], destDir)
		conf.Sendmail(fmt.Sprintf("%s downloaded!", torrentName), msg)
		log.Println(msg)

		// We're done
		os.Exit(0)
	}

	extList := [...]string{"mkv", "avi", "mpg", "mp4"}

	for _, ext := range extList {
		glob = getGlob(ext)
		if len(glob) > 0 {
			for _, srcname := range glob {
				destDir := conf.GetDestination(filepath.Base(srcname))

				in, err := os.Open(srcname)
				if err != nil {
					log.Printf("os.Open(): %v", err)
					os.Exit(5)
				}
				defer in.Close()

				destFile := fmt.Sprintf("%s/%s", destDir, filepath.Base(srcname))

				out, err := os.Create(destFile)
				if err != nil {
					log.Printf("os.Create(): %v", err)
					os.Exit(6)
				}
				defer out.Close()

				_, err = io.Copy(out, in)
				if err != nil {
					log.Printf("os.Copy(): %v", err)
					os.Exit(7)
				}

				if err := os.Chmod(destFile, 0644); err != nil {
					log.Printf("os.Chmod(): %v", err)
					os.Exit(8)
				}

				msg := fmt.Sprintf("copying '%+v' to '%s'\n", srcname, destDir)
				conf.Sendmail(fmt.Sprintf("%s downloaded!", torrentName), msg)
				log.Println(msg)
			}
		}

	}

}

func getGlob(ext string) []string {
	globOutput, err := filepath.Glob(
		fmt.Sprintf("%s/%s/*.%s",
			os.Getenv("TR_TORRENT_DIR"),
			os.Getenv("TR_TORRENT_NAME"),
			ext))

	if err != nil {
		log.Printf("filepath.Glob(): %v", err)
		os.Exit(9)
	}

	return globOutput
}
