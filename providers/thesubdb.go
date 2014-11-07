package providers

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type theSubDbProvider struct {
	UserAgent string
	Server    string
}

func init() {
	db := GetSubtitleDB()
	db.addSource(theSubDbProvider{
		UserAgent: "SubDB/1.0 (GoSub/0.1; http://github.com/Rubyss/gosub)",
		Server:    "api.thesubdb.com",
	})
}

func (s theSubDbProvider) hashFile(filePath string) (string, error) {
	var readSize int64 = 64 * 1024
	hash := md5.New()
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Read the first 64k
	_, err = io.CopyN(hash, file, readSize)
	if err != nil {
		return "", err
	}

	_, err = file.Seek(-readSize, os.SEEK_END)
	if err != nil {
		return "", err
	}

	// Read the last 64k
	_, err = io.CopyN(hash, file, readSize)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// Name returns the name of this provider
func (s theSubDbProvider) Name() string {
	return "TheSubDB.com"
}

// GetSubtitle accepts a filepath and a language, and searches for subtitles
func (s theSubDbProvider) GetSubtitle(filePath, language string) ([]Subtitle, error) {
	var subs []Subtitle
	hash, err := s.hashFile(filePath)
	if err != nil {
		return subs, err
	}

	params := url.Values{}
	params.Set("action", "search")
	params.Set("hash", hash)

	url := url.URL{
		Scheme:   "http",
		Host:     s.Server,
		RawQuery: params.Encode(),
	}

	resp, err := http.Get(url.String())
	if err != nil {
		return subs, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusBadRequest:
		log.Printf("Error while searching %s: 400 Bad Request\n", s.Name())
	case http.StatusOK:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return subs, err
		}

		bodyString := string(body)
		if strings.Contains(bodyString, language) {
			params.Set("action", "download")
			params.Set("language", language)
			url.RawQuery = params.Encode()
			subs = append(subs, Subtitle{Format: "srt", URL: url.String(), Source: s})
		}
	}

	return subs, nil
}

// Download returns the path of the downloaded subtitle
func (s theSubDbProvider) Download(subtitle Subtitle, filePath string) (string, error) {
	resp, err := http.Get(subtitle.URL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	subtitlePath := createSubtitlePath(filePath, subtitle.Format)
	file, err := os.OpenFile(subtitlePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return "", err
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)

	return subtitlePath, nil
}
