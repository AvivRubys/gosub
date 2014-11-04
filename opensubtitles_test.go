package gosub

import (
	"fmt"
	"testing"
)

func test(t *testing.T, filename, expectedHash string) {
	hash, _, err := hashFile(filename)
	if err != nil {
		t.Fatalf("Somethings gone fucked! Error:\n%s\n", err)
	}

	if hash != expectedHash {
		t.Fatalf("\nExpected hash:  %s\nGot hash:  %s\n", expectedHash, hash)
	}
}

func TestSimple(t *testing.T) {
	test(t, "breakdance.avi", "8e245d9679d31e12")
}

func TestLarge(t *testing.T) {
	test(t, "dummy.bin", "61f7751fc2a72bfb")
}

func TestRequest(t *testing.T) {
	err := GetSubtitle("C:/Users/Rubys/Videos/Doctor Who/Season 7/doctor.who.2005.s07e01.bdrip.xvid-haggis.avi")
	if err != nil {
		fmt.Printf("ERR:\n%s\n", err)
	}
}
