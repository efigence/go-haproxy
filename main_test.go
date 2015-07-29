package haproxy

import (
	"testing"
	"os"
	"os/exec"
//	"fmt"
	"time"
)
var haproxy *exec.Cmd
func TestMain(m *testing.M) {
	go runTestHaproxy()
	stopTestHaproxy()
	os.Exit(m.Run())
}

func runTestHaproxy() {
	if haproxy != nil {
		haproxy.Process.Kill()
	}
	haproxy = exec.Command("haproxy", "-f", "t-data/haproxy.conf")
	haproxy.Start()
}
 
func stopTestHaproxy() {
	if haproxy == nil {
		time.Sleep(100 * time.Millisecond)
	}
	if haproxy != nil {
		haproxy.Process.Kill()
	} else {
		return
	}
}
