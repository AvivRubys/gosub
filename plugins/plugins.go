package plugins

import "log"

// Consts
const (
	userAgent = "periscope" // gosub v0.1
)

// Basic structs and interfaces of the plugins module

// SubtitleRef is a reference to a subtitle, i.e. it wasn't downloaded yet, just found.
type SubtitleRef struct {
	FileName string
	URL      string
	Source   *SubtitleSource
}

// SubtitleSearcher is the interface a search plugin needs to implement
type SubtitleSearcher interface {
	GetSubtitle(filePath, language string) ([]SubtitleRef, error)
}

// SubtitleSource represents a search plugin
type SubtitleSource struct {
	Name string
	Impl SubtitleSearcher
}

// Plugin registry mechanism

// SubtitleDB is the interface the outer world uses to access plugins
type SubtitleDB interface {
	addSource(SubtitleSource)
	SearchAll(filePath, language string) ([]SubtitleRef, error)
}

type pluginsDB struct {
	plugins []SubtitleSource
}

func (db *pluginsDB) addSource(s SubtitleSource) {
	db.plugins = append(db.plugins, s)
}

func (db *pluginsDB) SearchAll(fileName, language string) ([]SubtitleRef, error) {
	var subs []SubtitleRef
	for _, src := range db.plugins {
		srcSubs, err := src.Impl.GetSubtitle(fileName, language)
		if err != nil {
			log.Printf("ERR When getting subtitles from %s: %s", src.Name, err)
		} else {
			subs = append(subs, srcSubs...)
		}
	}

	return subs, nil
}

var searchers = pluginsDB{make([]SubtitleSource, 0)}

// GetSubtitleDB returns a SubtitleDB through which we can search for subtitles
func GetSubtitleDB() SubtitleDB {
	return &searchers
}
