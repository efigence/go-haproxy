// +build !test

package haproxy

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

// HAProxy socket interface
type Conn struct {
	socketPath string
}

func NewConnection(path string) Conn {
	var c Conn
	c.socketPath = path
	return c
}

func (c *Conn) AddACL(acl string, pattern string) error {
	var err error
	_, err = c.RunCmd(fmt.Sprintf("add acl %s %s\n",acl, pattern))
	return err
}
func (c *Conn) DelACL(acl string, id string) error {
	var err error
	return err
}
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
