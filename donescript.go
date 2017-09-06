package main

import (
	"fmt"
	unarr "github.com/gen2brain/go-unarr"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

func main() {

	f, err := os.OpenFile("/tmp/TR_OUT.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Printf("Error opening file: %s", err)
		panic(err)
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
		_, err = f.WriteString(fmt.Sprintf("uncompressing '%+v' to '%s'\n", glob[0], dest_dir))
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
				_, err = f.WriteString(fmt.Sprintf("copying '%+v' to '%s'\n", srcname, dest_dir))
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
