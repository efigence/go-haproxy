package haproxy

import (
	"fmt"
)

func ExampleAddACL() {
	// Initialize
	ha := New("/var/run/haproxy.sock")

	// Get ACL entries from file in config (via -f in haproxy cfg)
	acls, _ := ha.GetACL("inc/blacklist.lst")
	
	// Check if it exists and add
	if acls["/bad/path"] == "" {
		ha.AddACL("inc/blacklist.lst", acls["/bad/path"])
	}
}

func ExampleDeleteACL() {
	// Initialize
	ha := New("/var/run/haproxy.sock")

	// Get ACL entries from file in config (via -f in haproxy cfg)
	acls, _ := ha.GetACL("inc/blacklist.lst")
	
	// Check if it exists and delete
	if acls["/bad/path"] != "" {
		ha.DeleteACL("inc/blacklist.lst", acls["/bad/path"])
	}
}

func ExampleGetACL() {
	// Initialize
	ha := New("/var/run/haproxy.sock")

	// Get ACL entries from file in config (via -f in haproxy cfg)
	acls, _ := ha.GetACL("inc/blacklist.lst")
	
	for k, v := range acls {
        fmt.Println(k, ":\t", v)
    }
	// Output:
	// /bad/path:    0x121e940
}

func ExampleClearACL() {
	// Initialize
	ha := New("/var/run/haproxy.sock")

	// clear all entries in ACL
	_ = ha.ClearACL("inc/blacklist.lst")
}

 
