package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/Rubyss/gosub/providers"
)

var (
	language   = flag.String("language", "en", "")
	autoDecide = flag.Bool("auto", true, "")
	help       = flag.Bool("h", false, "Help")
	files      []string
)

func getSubtitle(db providers.SubtitleDB, file string) {
	subs, err := db.SearchAll(file, *language)
	if err != nil {
		log.Fatalf("Error while searching!\n%s\n", err)
	}

	if len(subs) == 0 {
		log.Fatalf("No subtitles found.\n")
	}

	var selectedSub *providers.Subtitle

	if *autoDecide || len(subs) == 1 {
		selectedSub = &subs[0]
	} else {
		// Let the user select
	}

	subPath, err := selectedSub.Source.Download(*selectedSub, file)
	if err != nil {
		log.Printf("Error in downloading subtitle. \n%s\n", err)
	}

	log.Printf("Got %s from %s.\n", subPath, selectedSub.Source.Name())
}

func isDirectory(file string) (bool, error) {
	fileInfo, err := os.Stat(file)
	return fileInfo.IsDir(), err
}

func main() {
	// Go over all providers, search, if there's more than one result - ask the user
	//  what he wants unless there's a STFU flag, and then select automatically.
	flag.Parse()
	if *help {
		usage := `Usage: gosub (flags) [FILE]
  -auto=true: Automatically select the best subtitle
  -language="eng": Language to search with
  -h: Display this help message`
		fmt.Println(usage)
		os.Exit(0)
	}
	// Reading the last argument as the file
	if args := flag.Args(); len(args) >= 1 {
		for _, arg := range args {
			dir, err := isDirectory(arg)
			if err != nil {
				fmt.Errorf("Unknown argument: %s", arg)
				continue
			}

			if dir {
				dirContents, err := ioutil.ReadDir(arg)
				if err != nil {
					continue
				}

				for _, fileInfo := range dirContents {
					file := path.Join(arg, fileInfo.Name())
					files = append(files, file)
				}
			} else {
				files = append(files, arg)
			}
		}
	} else {
		log.Fatalln("No files given to search for.")
	}

	db := providers.GetSubtitleDB()
	for _, file := range files {
		fileName := path.Base(file)
		log.Printf("%s:", fileName)
		getSubtitle(db, file)
	}
}
