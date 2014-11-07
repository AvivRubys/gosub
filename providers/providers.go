package providers

import (
	"log"
	"sync"
)

// SubtitleDB is the interface through which we search for subtitles
type SubtitleDB interface {
	addSource(SubtitleProvider)
	SearchAll(filePath, language string) ([]Subtitle, error)
}

type providerDB struct {
	providers []SubtitleProvider
}

func (db *providerDB) addSource(s SubtitleProvider) {
	db.providers = append(db.providers, s)
}

func (db *providerDB) SearchAll(fileName, language string) ([]Subtitle, error) {
	var result []Subtitle
	var wg sync.WaitGroup
	subReceiver := make(chan []Subtitle)

	// Launch a goroutine for each provider, send results into subReceiver channel
	for _, provider := range db.providers {
		wg.Add(1)
		go func(p SubtitleProvider) {
			defer wg.Done()
			subs, err := p.GetSubtitles(fileName, language)
			if err != nil {
				log.Printf("ERR When getting subtitles from %s: %s", p.Name(), err)
			} else {
				subReceiver <- subs
			}
		}(provider)
	}

	// When work is done, close the channel, leaving up the loop below
	go func() {
		wg.Wait()
		close(subReceiver)
	}()

	for subs := range subReceiver {
		result = append(result, subs...)
	}

	return result, nil
}

var searchers = providerDB{make([]SubtitleProvider, 0)}

// GetSubtitleDB returns a SubtitleDB through which we can search for subtitles
func GetSubtitleDB() SubtitleDB {
	return &searchers
}
