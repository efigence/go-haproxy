package haproxy

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

var testConn Conn

func TestACL(t *testing.T) {
	// fixme generate tmpfile unix socket
	err := runTestHaproxy()
	defer stopTestHaproxy()
	Convey("Start test haproxy", t, func() {
		So(err, ShouldEqual, nil)
	})
	c := NewConnection("tmp/haproxy.sock")
	Convey("List ACL", t, func() {
		out, err := c.RunCmd("show acl")
		So(err, ShouldEqual, nil)
		So(out[0], ShouldContainSubstring, "# id") // header
	})
	Convey("List all ACL", t, func() {
		out, err := c.RunCmd("show acl")
		So(err, ShouldEqual, nil)
		So(out[1], ShouldContainSubstring, `blacklist.lst`)
	})
	Convey("List ACL entries", t, func() {
		out, err := c.GetACL("t-data/blacklist.lst")
		So(err, ShouldEqual, nil)
		So(out[0], ShouldContainSubstring, "/from/file")
	})
	Convey("Add ACL", t, func() {
		err := c.AddACL("t-data/blacklist.lst", "/bad/test1")
		So(err, ShouldEqual, nil)
		out, err := c.GetACL("t-data/blacklist.lst")
		So(out[1], ShouldContainSubstring, "/bad/test1")
	})

	defer stopTestHaproxy()
}
