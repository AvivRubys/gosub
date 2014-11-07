package providers

import "log"

// Consts
const (
	userAgent = "OSTestUserAgent" // gosub v0.1
)

// Basic structs and interfaces of the providers module

// Subtitle is a reference to a subtitle, i.e. it wasn't downloaded yet, just found.
type Subtitle struct {
	FileName  string
	Hash      string
	Format    string
	Downloads int
	URL       string
	Source    SubtitleProvider
}

// SubtitleProvider is the interface a provider needs to implement
type SubtitleProvider interface {
	// Name returns the name of this provider
	Name() string

	// GetSubtitle accepts a filepath and a language, and searches for subtitles
	GetSubtitle(filePath, language string) ([]Subtitle, error)

	// Download returns the path of the downloaded subtitle
	Download(subtitle Subtitle, filePath string) (string, error)
}

// Provider registry mechanism

// SubtitleDB is the interface the outer world uses to access providers
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
		srcSubs, err := src.GetSubtitle(fileName, language)
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
