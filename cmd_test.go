package haproxy

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

var testConn Conn

func TestACL(t *testing.T) {
	// fixme generate tmpfile unix socket
	c := NewConnection("/tmp/file")
	Convey("Add ACL", t, func() {
		err := c.AddACL("test", "acl")
		So(err, ShouldEqual, nil)
	})
}
