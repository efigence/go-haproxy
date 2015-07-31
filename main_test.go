package haproxy

import (
	"os"
	"os/exec"
	"testing"
	//	"fmt"
	"errors"
	"time"
)

var haproxy *exec.Cmd
var socketFile = "tmp/haproxy.sock"

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func runTestHaproxy() error {
	if haproxy != nil {
		haproxy.Process.Kill()
	}
	// just in case
	if _, err := os.Stat(socketFile); err == nil {
		os.Remove(socketFile)
	}
	haproxy = exec.Command("haproxy", "-f", "t-data/haproxy.conf")
	runerr := haproxy.Start()
	time.Sleep(100 * time.Millisecond)
	if _, err := os.Stat(socketFile); err == nil {
		return runerr
	}
	time.Sleep(1000 * time.Millisecond)
	if _, err := os.Stat(socketFile); err == nil {
		return runerr
	}
	// is that a rPi I spy ?
	time.Sleep(10000 * time.Millisecond)
	if _, err := os.Stat(socketFile); err == nil {
		return runerr
	}
	return errors.New("tried to start haproxy -f t-data/haproxy.conf but socket still does not exist!")

}

func stopTestHaproxy() {
	if haproxy == nil {
		time.Sleep(100 * time.Millisecond)
	}
	if haproxy != nil {
		haproxy.Process.Kill()
		os.Remove(socketFile)
	} else {
		return
	}
}
