package main

import (
	"fmt"
	"log"

	"github.com/Rubyss/gosub/plugins"
)

func main() {
	// Go over all plugins, search, if there's more than one result - ask the user
	//  what he wants unless there's a STFU flag, and then select automatically.

	db := plugins.GetSubtitleDB()
	subs, err := db.SearchAll("C:/Users/Rubys/Videos/Doctor Who/Season 7/doctor.who.2005.s07e01.bdrip.xvid-haggis.avi", "eng")
	if err != nil {
		log.Fatalf("Error while searching!\n%s\n", err)
	}

	for _, sub := range subs {
		fmt.Printf("%s - %s\n", sub.Source.Name, sub.URL)
	}
}
