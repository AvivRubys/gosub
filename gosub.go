package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Rubyss/gosub/providers"
)

var (
	language   = flag.String("language", "eng", "")
	autoDecide = flag.Bool("auto", true, "")
	help       = flag.Bool("h", false, "Help")
	file       string
)

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
	if args := flag.Args(); len(args) == 1 {
		file = args[0]
	}

	db := providers.GetSubtitleDB()
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
		for _, sub := range subs {
			if sub.Downloads > selectedSub.Downloads {
				selectedSub = &sub
			}
		}
	} else {
		// Let the user select
	}

	subPath, err := selectedSub.Source.Download(*selectedSub, file)
	if err != nil {
		log.Fatalf("Error in downloading the subtitle. \n%s\n", err)
	}

	log.Printf("Got %s from %s.\n", subPath, selectedSub.Source.Name)
}
