package providers

import (
	"log"
	"path/filepath"
	"sync"
)

// SubtitleDB is the interface through which we search for subtitles
type SubtitleDB interface {
	addSource(SubtitleProvider)
	Get(path, language string) error
	GetAll(paths []string, language string)
}

type providerDB struct {
	providers []SubtitleProvider
}

func (db *providerDB) addSource(s SubtitleProvider) {
	db.providers = append(db.providers, s)
}

func (db *providerDB) Get(path, language string) error {
	var result []Subtitle
	var wg sync.WaitGroup
	subReceiver := make(chan []Subtitle)
	defer recoverFromSearchError(path)

	// Launch a goroutine for each provider, send results into subReceiver channel
	for _, provider := range db.providers {
		wg.Add(1)
		go func(p SubtitleProvider) {
			defer wg.Done()
			subs, err := p.GetSubtitles(path, language)
			if err != nil {
				log.Printf("Error when getting subtitles from %s: %s", p.Name(), err)
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

	// Take the first sub, since we currently can't select intelligently
	selectedSub := result[0]
	subPath, err := selectedSub.Source.Download(selectedSub, path)
	if err != nil {
		log.Printf("Error in downloading subtitle. \n%s\n", err)
	}

	log.Printf("Got \"%s\" from %s.\n", filepath.Base(subPath), selectedSub.Source.Name())

	return nil
}

func (db *providerDB) GetAll(paths []string, language string) {
	var wg sync.WaitGroup
	for _, path := range paths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			err := db.Get(path, language)
			if err != nil {
				log.Printf("Error getting subtitles for: \"%s\"\n", filepath.Base(path))
			}
		}(path)
	}

	wg.Wait()
}

func recoverFromSearchError(file string) {
	if e := recover(); e != nil {
		log.Printf("Error in search for subtitle %s\n. %s\n", file, e)
	}
}

var searchers = providerDB{make([]SubtitleProvider, 0)}

// GetSubtitleDB returns a SubtitleDB through which we can search for subtitles
func GetSubtitleDB() SubtitleDB {
	return &searchers
}
