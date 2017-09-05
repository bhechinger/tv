package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {

	// To set a key/value pair, use `os.Setenv`. To get a
	// value for a key, use `os.Getenv`. This will return
	// an empty string if the key isn't present in the
	// environment.

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

	output := fmt.Sprintf("Environment:\n" +
		"\tTR_APP_VERSION: %s\n" +
		"\tTR_TIME_LOCALTIME: %s\n" +
		"\tTR_TORRENT_DIR: %s\n" +
		"\tTR_TORRENT_HASH: %s\n" +
		"\tTR_TORRENT_ID: %s\n" +
		"\tTR_TORRENT_NAME: %s\n\n",
		app_version,
		time_localtime,
		torrent_dir,
		torrent_hash,
		torrent_id,
		torrent_name)

	if _, err = f.WriteString(output); err != nil {
		fmt.Printf("Error writing string: %s", err)
		panic(err)
	}

	var exist_output string
	var glob_output []string

	if glob_output, err = filepath.Glob(fmt.Sprintf("%s/%s/*.rar", torrent_dir, torrent_name)); err != nil {
		if _, err = f.WriteString(fmt.Sprintf("Error globbing filename: %s", err)); err != nil{
			fmt.Printf("Error writing string: %s", err)
		}
		panic(err)
	}

	if _, err = f.WriteString(fmt.Sprintf("glob: %+v", glob_output)); err != nil{
		fmt.Printf("Error writing string: %s", err)
		panic(err)
	}

	if _, err = os.Stat(glob_output[0]); err != nil {
		if os.IsNotExist(err) {
			exist_output = "\tFile doesn't exist\n"
		} else {
			exist_output = "\tFile exists\n"
		}
	}

	if _, err = f.WriteString(exist_output); err != nil{
		fmt.Printf("Error writing string: %s", err)
		panic(err)
	}
}
