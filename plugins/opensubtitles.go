package plugins

import (
	"errors"
	"fmt"
	"log"

	"github.com/kolo/xmlrpc"
)

const sixtyFourKiloBytes = 64 * 1024
const int64InBytes = 64 / 8

var (
	// ErrFileSizeTooSmall is the error that is thrown when the file is too short
	ErrFileSizeTooSmall = errors.New("The file is too short to be hashed (< 64K).")
)

var (
	openSubtitlesSource = SubtitleSource{Name: "OpenSubtitles.org", Impl: openSubtitlesSearcher{}}
)

func init() {
	db := GetSubtitleDB()
	db.addSource(openSubtitlesSource)
}

type openSubtitlesSearcher struct{}

func (s openSubtitlesSearcher) GetSubtitle(file, language string) ([]SubtitleRef, error) {
	client, err := xmlrpc.NewClient("http://api.opensubtitles.org/xml-rpc", nil)
	if err != nil {
		return nil, err
	}

	loginRequest := []interface{}{"", "", "en", "OSTestUserAgent"}
	var loginResponse struct {
		Token   string  `xmlrpc:"token"`
		Status  string  `xmlrpc:"status"`
		Seconds float32 `xmlrpc:"seconds"`
	}

	err = client.Call("LogIn", loginRequest, &loginResponse)
	if err != nil {
		return nil, err
	}

	if loginResponse.Status != "200 OK" {
		return nil, fmt.Errorf("Bad rc from login call to opensubtitles: %s", loginResponse.Status)
	}

	hash, size, err := movieHashFile(file)
	if err != nil {
		return nil, err
	}

	searchRequest := []interface{}{
		loginResponse.Token,
		[]struct {
			MovieByteSize string `xmlrpc:"moviebytesize"`
			MovieHash     string `xmlrpc:"moviehash"`
			Language      string `xmlrpc:"sublanguageid"`
		}{{fmt.Sprintf("%d", size), hash, language}}}
	// SubFileName, SubHash, MovieNameEng, SubDownloadLink, SubtitlesLink
	var searchResponse struct {
		Status    string `xmlrpc:"status"`
		Subtitles []struct {
			FileName  string `xmlrpc:"SubFileName"`
			Hash      string `xmlrpc:"SubHash"`
			Format    string `xmlrpc:"SubFormat"`
			MovieName string `xmlrpc:"MovieName"`
			URL       string `xmlrpc:"SubDownloadLink"`
			Page      string `xmlrpc:"SubtitlesLink"`
		} `xmlrpc:"data"`
	}

	err = client.Call("SearchSubtitles", searchRequest, &searchResponse)
	if err != nil {
		return nil, err
	}

	var subs []SubtitleRef
	for _, sub := range searchResponse.Subtitles {
		subs = append(subs, SubtitleRef{
			FileName: sub.FileName,
			URL:      sub.URL,
			Source:   &openSubtitlesSource,
		})
	}

	err = client.Call("LogOut", loginResponse.Token, nil)
	if err != nil {
		log.Printf("LogOut from opensubtitles failed. Reason: %s\n", err)
	}

	return subs, nil
}
