package providers

import "log"

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
	var subs []Subtitle
	for _, src := range db.providers {
		srcSubs, err := src.GetSubtitles(fileName, language)
		if err != nil {
			log.Printf("ERR When getting subtitles from %s: %s", src.Name(), err)
		} else {
			subs = append(subs, srcSubs...)
		}
	}

	return subs, nil
}

var searchers = providerDB{make([]SubtitleProvider, 0)}

// GetSubtitleDB returns a SubtitleDB through which we can search for subtitles
func GetSubtitleDB() SubtitleDB {
	return &searchers
}
