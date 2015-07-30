// +build !test

package haproxy

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"errors"
)

// HAProxy socket interface
type Conn struct {
	socketPath string
}


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
	out, err := c.RunCmd(fmt.Sprintf("add acl %s %s\n",acl, pattern))
	if err != nil {
		return err
	}
	if out[0] != "" {
		return errors.New(fmt.Sprintf("error: %s", out[0]))
	}
	return err
}

// Delete entry from ACL
// ID is value of map returned by GetACL

func (c *Conn) DeleteACL(acl string, id string) (error) {
	var err error
	if  strings.ContainsAny(id, " \t\n") || id == "" {
		return errors.New("id should not contain whitespaces or be empty as that would remove every ACL, use ClearACL for that")
	}
	out, err := c.RunCmd(fmt.Sprintf("del acl %s %s\n",acl, id))
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
	out, err := c.RunCmd(fmt.Sprintf("show acl %s",acl))
	acls := make(map[string]string)
	for _, line := range out {
		parts := strings.Split(line, " ")
		if len(parts) > 1 {
			acls[parts[1]] = parts[0]
		}
	}
	return acls, err
}
func (c *Conn) ListACL(acl string) (error) {
	var err error
	return err
}

// Clear all entries in ACL
func (c *Conn) ClearACL(acl string) (error) {
	var err error
	out, err := c.RunCmd(fmt.Sprintf("clear acl %s",acl))
	if out[0] != "" {
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
