//go:build !test
// +build !test

package haproxy

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

type ACL struct {
	// haproxy-assigned ID
	ID int `json:"id"`
	// path of external file that is source of this ACL entries
	SourceFile string `json:"source_file"`
	// type of ACL if it is from config or "file" if it sourced from file (haproxy doesnt return type here)
	Type string `json:"type"`
	// Line of inline ACL
	Line int `json:"line"`
}

// HAProxy socket interface
type Conn struct {
	socketPath string
}

// pattern for ACLs originating from config files
var inlineACLRegex = regexp.MustCompile(`^(\d+) \((.*)\) acl '(.+)' file '(.*)' line (\d+)`)

// pattern for ACLs included via -f option
var fileACLRegex = regexp.MustCompile(`^(\d+) \((.*)\) pattern loaded from file '(.*)' used by`)

// Setup new connection
// accepts path to haproxy unix socket
func New(path string) Conn {
	var c Conn
	c.socketPath = path
	return c
}

// Add new entry to acl
// ACL name is either file path ( if haproxy config uses -f option to load acls from file ) or ACL id prepended with hash
func (c *Conn) AddACL(acl string, pattern string) error {
	var err error
	out, err := c.RunCmd(fmt.Sprintf("add acl %s %s\n", acl, pattern))
	if err != nil {
		return err
	}
	if out[0] != "" && out[0] != "Done." {
		return errors.New(fmt.Sprintf("error: %s", out[0]))
	}
	return err
}

// Delete entry from ACL
// ID is value of map returned by GetACL

func (c *Conn) DeleteACL(acl string, id string) error {
	var err error
	if strings.ContainsAny(id, " \t\n") || id == "" {
		return errors.New("id should not contain whitespaces or be empty as that would remove every ACL, use ClearACL for that")
	}
	out, err := c.RunCmd(fmt.Sprintf("del acl %s %s\n", acl, id))
	if strings.Contains(out[0], "Key not found") {
		return errors.New("Key not found")
	}
	return err
}

// Get map of all entries in ACL
// map is value => ID for easy lookup
// so deleting acls is just
//     err := ha.DeleteACL( acls["/test/acl"] )

func (c *Conn) GetACL(acl string) (map[string]string, error) {
	var err error
	out, err := c.RunCmd(fmt.Sprintf("show acl %s", acl))
	acls := make(map[string]string)
	for _, line := range out {
		parts := strings.Split(line, " ")
		if len(parts) > 1 {
			acls[parts[1]] = parts[0]
		}
	}
	return acls, err
}

// List all ACLs that haproxy currenty uses
func (c *Conn) ListACL() ([]ACL, error) {
	var err error
	var acl []ACL
	out, err := c.RunCmd("show acl")

	for _, line := range out {
		var a ACL
		matches := inlineACLRegex.FindStringSubmatch(line)
		if len(matches) > 2 {
			a.ID, _ = strconv.Atoi(matches[1])
			a.Type = matches[3]
			a.Line, _ = strconv.Atoi(matches[5])
			acl = append(acl, a)
		} else {
			matches = fileACLRegex.FindStringSubmatch(line)
			if len(matches) > 2 {
				a.ID, _ = strconv.Atoi(matches[1])
				a.SourceFile = matches[2]
				a.Type = "file"
				acl = append(acl, a)
			} else {
				continue
			}
		}
	}

	return acl, err
}

// Return map with  all external files that are used as ACL entry source with first occurence of ACL as a value
func (c *Conn) ListACLFiles() (map[string]ACL, error) {
	var err error
	acl_list, err := c.ListACL()
	out := make(map[string]ACL)
	for _, acl := range acl_list {
		if acl.Type == "file" {
			if _, ok := out[acl.SourceFile]; ok {
				// we dont want to overwrite existing entries as we are interested only in first occurence of acl
				continue
			} else {
				out[acl.SourceFile] = acl
			}
		}
	}

	return out, err
}

// Clear all entries in ACL
func (c *Conn) ClearACL(acl string) error {
	var err error
	out, err := c.RunCmd(fmt.Sprintf("clear acl %s", acl))
	if out[0] != "" && out[0] != "Done." {
		return errors.New(fmt.Sprintf("error: %+v", out))
	}
	return err
}

// Run arbitrary haproxy command and return output
func (c *Conn) RunCmd(cmd string) ([]string, error) {
	conn, err := net.Dial("unix", c.socketPath)
	var out []string
	if err != nil {
		return out, err
	}
	fmt.Fprintf(conn, "%s\n", cmd)
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		out = append(out, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return out, err
	}
	return out, err
}
