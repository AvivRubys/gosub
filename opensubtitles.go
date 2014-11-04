package gosub

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

const sixtyFourKiloBytes = 64 * 1024
const int64InBytes = 64 / 8

var (
	// ErrFileSizeTooSmall is the error that is thrown when the file is too short
	ErrFileSizeTooSmall = errors.New("The file is too short to be hashed (< 64K).")
)

func hashChunk(reader io.Reader) (uint64, error) {
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
func hashFile(filepath string) (string, int64, error) {
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
	head, err := hashChunk(fileStartReader)
	if err != nil {
		return "", 0, err
	}

	// Seek to and read the last 64K
	_, err = file.Seek(-sixtyFourKiloBytes, os.SEEK_END)
	if err != nil {
		return "", 0, err
	}

	tail, err := hashChunk(file)
	if err != nil {
		return "", 0, err
	}

	return fmt.Sprintf("%x", uint64(fileSize)+head+tail), fileSize, nil
}
