// +build !test

package haproxy
import (
	"net"
)


// HAProxy socket interface
type Conn struct {
	socketPath string
	conn net.UnixConn
}


func NewConnection(path string) Conn {
	var c Conn
	return c
}

func (t *Conn) Connect() {
}

func (t *Conn) AddACL(acl string, pattern string) error {
	var err error
	return err
}
