package providers

// Basic structs and interfaces of the providers module

// Subtitle is a reference to a subtitle, i.e. it wasn't downloaded yet, just found.
type Subtitle struct {
	FileName string
	Hash     string
	Format   string
	URL      string
	Source   SubtitleProvider
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
