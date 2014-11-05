package plugins

import (
	"fmt"
	"log"

	"github.com/kolo/xmlrpc"
)

var (
	openSubtitlesSource = SubtitleSource{Name: "OpenSubtitles.org", Impl: openSubtitlesSearcher{}}
)

func init() {
	db := GetSubtitleDB()
	db.addSource(openSubtitlesSource)
}

type openSubtitlesSearcher struct{}

func (s openSubtitlesSearcher) login(client *xmlrpc.Client, username, password, language, useragent string) (string, error) {
	request := []interface{}{username, password, language, useragent}
	var response struct {
		Token   string  `xmlrpc:"token"`
		Status  string  `xmlrpc:"status"`
		Seconds float32 `xmlrpc:"seconds"`
	}

	err := client.Call("LogIn", request, &response)
	if err != nil {
		return "", err
	}

	if response.Status != "200 OK" {
		return "", fmt.Errorf("Bad rc from login call to opensubtitles: %s", response.Status)
	}

	return response.Token, nil
}

func (s openSubtitlesSearcher) searchSubtitles(client *xmlrpc.Client, token, hash, language string, size int64) ([]SubtitleRef, error) {
	request := []interface{}{
		token,
		[]struct {
			MovieByteSize string `xmlrpc:"moviebytesize"`
			MovieHash     string `xmlrpc:"moviehash"`
			Language      string `xmlrpc:"sublanguageid"`
		}{{fmt.Sprintf("%d", size), hash, language}}}

	var response struct {
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

	err := client.Call("SearchSubtitles", request, &response)
	if err != nil {
		return nil, err
	}

	var subs []SubtitleRef
	for _, sub := range response.Subtitles {
		subs = append(subs, SubtitleRef{
			FileName: sub.FileName,
			URL:      sub.URL,
			Source:   &openSubtitlesSource,
		})
	}

	return subs, nil
}

func (s openSubtitlesSearcher) GetSubtitle(file, language string) ([]SubtitleRef, error) {
	client, err := xmlrpc.NewClient("http://api.opensubtitles.org/xml-rpc", nil)
	if err != nil {
		return nil, err
	}

	token, err := s.login(client, "", "", language, userAgent)
	if err != nil {
		return nil, err
	}

	hash, size, err := movieHashFile(file)
	if err != nil {
		return nil, err
	}

	subs, err := s.searchSubtitles(client, token, hash, language, size)

	err = client.Call("LogOut", token, nil)
	if err != nil {
		log.Printf("LogOut from opensubtitles failed. Reason: %s\n", err)
	}

	return subs, nil
}
