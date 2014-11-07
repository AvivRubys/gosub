package main

import (
	"flag"
	"log"

	"github.com/Rubyss/gosub/plugins"
)

var (
	language   = flag.String("language", "eng", "Language to search with")
	file       = flag.String("file", "", "The file to search subtitles for")
	autoDecide = flag.Bool("auto", true, "Automatically select the best subtitle")
)

func main() {
	// Go over all plugins, search, if there's more than one result - ask the user
	//  what he wants unless there's a STFU flag, and then select automatically.

	flag.Parse()
	db := plugins.GetSubtitleDB()
	subs, err := db.SearchAll(*file, *language)
	if err != nil {
		log.Fatalf("Error while searching!\n%s\n", err)
	}

	if len(subs) == 0 {
		log.Fatalf("No subtitles found.\n")
	}

	var selectedSub *plugins.SubtitleRef

	if *autoDecide {
		selectedSub = &subs[0]
		for _, sub := range subs {
			if sub.Downloads > selectedSub.Downloads {
				selectedSub = &sub
			}
		}
	} else {
		// Let the user select
	}

	subPath, err := selectedSub.Source.Impl.Download(*selectedSub, *file)
	if err != nil {
		log.Fatalln("Error in downloading the subtitle.")
	}

	log.Printf("Got %s from %s.\n", subPath, selectedSub.Source.Name)
}
