package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"bitbucket.org/kardianos/osext"
	"github.com/Rubyss/gosub/providers"
)

const logFileName = "gosub.log"

var (
	language = flag.String("language", "en", "")
	help     = flag.Bool("h", false, "Help")
)

func main() {
	flag.Parse()
	if *help {
		usage := `Usage: gosub (flags) [FILES/DIRECTORIES]
  -language="eng": Language to search with
  -h: Display this help message`
		fmt.Println(usage)
		os.Exit(0)
	}

	// Set up proper logging
	dir, _ := osext.ExecutableFolder()
	logfile, err := os.OpenFile(dir+logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Error opening logfile %s (%v)", dir+logFileName, err)
	}
	defer logfile.Close()

	log.SetOutput(io.MultiWriter(logfile, os.Stderr))

	// mkv isn't listed in windows mime types, for some reason.
	mime.AddExtensionType(".mkv", "video/x-matroska")

	// Arguments we have been given may be files or directories,
	// filter them for video files
	var files []string
	if args := flag.Args(); len(args) >= 1 {
		for _, arg := range args {
			filepath.Walk(arg, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					log.Printf("Error in walking directory %s\n", err)
					return nil
				}

				fileInfo, err := os.Stat(path)
				if err != nil {
					log.Printf("Error in opening path %s\n", path)
					return nil
				}

				if fileInfo.IsDir() {
					return nil
				}

				mimeType := mime.TypeByExtension(filepath.Ext(path))
				if strings.Contains(mimeType, "video") {
					files = append(files, path)
				} else {
					log.Printf("Ignoring %s (not a video file)\n", filepath.Base(path))
				}

				return nil
			})
		}
	}

	db := providers.GetSubtitleDB()
	db.GetAll(files, *language)
}
