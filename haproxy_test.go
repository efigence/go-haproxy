package haproxy

import (
	"testing"
	"fmt"
	"os"
)

var testStrings[] string


func TestMain(m *testing.M) {
	var err error
	testStrings, err = readLines("t-data/haproxy_udp")
	if err != nil {
		fmt.Printf("Can't load haproxy logs")
		os.Exit(255)
	}
	_ = testStrings
	os.Exit(m.Run())
}


func TestTS(t *testing.T) {
	var err error
	ts, err := decodeTs("23/Jul/2015:13:49:11.933")
	if err != nil {
		t.Errorf("cant decode timestamp: %s", err)
	}
	if ts.UnixNano() != 1437659351933000000 {
		t.Errorf("Timestamp decoded incorrectly")
	}
	ts, err = decodeTs("03/Jul/2015:13:49:11.933")
	fmt.Printf("%s",testStrings)
	if err != nil {
		t.Errorf("cant decode timestamp: %s", err)
	}
	if ts.UnixNano() != 1435931351933000000 {
		t.Errorf("Timestamp decoded incorrectly")
	}
}
