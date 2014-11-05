package plugins

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

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

func (s openSubtitlesSearcher) hashChunk(reader io.Reader) (uint64, error) {
	// Read all int64s
	int64Buffer := make([]uint64, sixtyFourKiloBytes/int64InBytes)
	err := binary.Read(reader, binary.LittleEndian, &int64Buffer)
	if err != nil {
		return 0, err
	}

	// Sum em up
	var sum uint64
	for _, n := range int64Buffer {
		sum += n
	}

	return sum, nil
}

// HashFile hashed a given file by summing up it's size in bytes, and the checksum
// of the first and last 64K, even if they overlap.
// Reference: http://trac.opensubtitles.org/projects/opensubtitles/wiki/HashSourceCodes
func (s openSubtitlesSearcher) hashFile(filepath string) (string, int64, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", 0, err
	}

	// File size is an int64 - important!
	fileSize := fileInfo.Size()
	if fileSize < sixtyFourKiloBytes {
		return "", 0, ErrFileSizeTooSmall
	}

	// Read the first 64K
	fileStartReader := io.LimitReader(file, sixtyFourKiloBytes)
	head, err := s.hashChunk(fileStartReader)
	if err != nil {
		return "", 0, err
	}

	// Seek to and read the last 64K
	_, err = file.Seek(-sixtyFourKiloBytes, os.SEEK_END)
	if err != nil {
		return "", 0, err
	}

	tail, err := s.hashChunk(file)
	if err != nil {
		return "", 0, err
	}

	return fmt.Sprintf("%x", uint64(fileSize)+head+tail), fileSize, nil
}

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

	hash, size, err := s.hashFile(file)
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
