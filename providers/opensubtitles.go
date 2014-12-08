package providers

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/kolo/xmlrpc"
)

func init() {
	db := GetSubtitleDB()
	db.addSource(openSubtitlesProvider{
		UserAgent: "periscope",
		Server:    "http://api.opensubtitles.org/xml-rpc",
		LanguageMap: map[string]string{
			"en": "eng", "fr": "fre", "hu": "hun", "cs": "cze",
			"pl": "pol", "sk": "slo", "pt": "por", "pt-br": "pob",
			"es": "spa", "el": "ell", "ar": "ara", "sq": "alb",
			"hy": "arm", "ay": "ass", "bs": "bos", "bg": "bul",
			"ca": "cat", "zh": "chi", "hr": "hrv", "da": "dan",
			"nl": "dut", "eo": "epo", "et": "est", "fi": "fin",
			"gl": "glg", "ka": "geo", "de": "ger", "he": "heb",
			"hi": "hin", "is": "ice", "id": "ind", "it": "ita",
			"ja": "jpn", "kk": "kaz", "ko": "kor", "lv": "lav",
			"lt": "lit", "lb": "ltz", "mk": "mac", "ms": "may",
			"no": "nor", "oc": "oci", "fa": "per", "ro": "rum",
			"ru": "rus", "sr": "scc", "sl": "slv", "sv": "swe",
			"th": "tha", "tr": "tur", "uk": "ukr", "vi": "vie",
		},
	})
}

type openSubtitlesProvider struct {
	UserAgent   string
	Server      string
	LanguageMap map[string]string
}

func (s openSubtitlesProvider) login(client *xmlrpc.Client, username, password, language, useragent string) (string, error) {
	request := []interface{}{username, password, s.LanguageMap[language], useragent}
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

func (s openSubtitlesProvider) searchSubtitles(client *xmlrpc.Client, token, hash, language string, size int64) ([]Subtitle, error) {
	request := []interface{}{
		token,
		[]struct {
			MovieByteSize string `xmlrpc:"moviebytesize"`
			MovieHash     string `xmlrpc:"moviehash"`
			Language      string `xmlrpc:"sublanguageid"`
		}{{fmt.Sprintf("%d", size), hash, s.LanguageMap[language]}}}

	var response struct {
		Status    string `xmlrpc:"status"`
		Subtitles []struct {
			FileName  string `xmlrpc:"SubFileName"`
			Hash      string `xmlrpc:"SubHash"`
			Format    string `xmlrpc:"SubFormat"`
			MovieName string `xmlrpc:"MovieName"`
			Downloads string `xmlrpc:"SubDownloadsCnt"`
			URL       string `xmlrpc:"SubDownloadLink"`
			Page      string `xmlrpc:"SubtitlesLink"`
		} `xmlrpc:"data"`
	}

	err := client.Call("SearchSubtitles", request, &response)
	if err != nil {
		return nil, err
	}

	var subs []Subtitle
	for _, sub := range response.Subtitles {
		/*downloadsInt, err := strconv.Atoi(sub.Downloads)
		if err != nil {
			downloadsInt = -1
		}*/

		subs = append(subs, Subtitle{
			FileName: sub.FileName,
			Hash:     sub.Hash,
			Format:   sub.Format,
			//Downloads: downloadsInt,
			URL:    sub.URL,
			Source: s,
		})
	}

	return subs, nil
}

func (s openSubtitlesProvider) Name() string {
	return "OpenSubtitles.org"
}

func (s openSubtitlesProvider) GetSubtitles(file, language string) ([]Subtitle, error) {
	client, err := xmlrpc.NewClient(s.Server, nil)
	if err != nil {
		return nil, err
	}

	token, err := s.login(client, "", "", language, s.UserAgent)
	if err != nil {
		return nil, err
	}

	hash, size, err := movieHashFile(file)
	if err != nil {
		return nil, err
	}

	subs, err := s.searchSubtitles(client, token, hash, language, size)
	if err != nil {
		return nil, err
	}

	err = client.Call("LogOut", token, nil)
	if err != nil {
		log.Printf("ERR LogOut from opensubtitles failed. Reason: %s\n", err)
	}

	return subs, nil
}

func (s openSubtitlesProvider) Download(subtitle Subtitle, filePath string) (string, error) {
	resp, err := http.Get(subtitle.URL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	// Substrings the extension out, and adds in the new one
	subtitlePath := createSubtitlePath(filePath, subtitle.Format)
	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return "", err
	}

	// All these flags mean: open for write only, create it if it doesnt exists but if it does - empty it
	file, err := os.OpenFile(subtitlePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return "", err
	}

	defer file.Close()

	_, err = io.Copy(file, reader)
	if err != nil {
		return "", err
	}

	return subtitlePath, nil
}
