package haproxy

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"fmt"
)

var testConn Conn

func TestACL(t *testing.T) {
	// fixme generate tmpfile unix socket
	err := runTestHaproxy()
	defer stopTestHaproxy()
	Convey("Start test haproxy", t, func() {
		So(err, ShouldEqual, nil)
	})
	c := New("tmp/haproxy.sock")
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
		So(out["/from/file"], ShouldNotEqual, nil)
	})
	Convey("Add to existing fileACL", t, func() {
		out, err := c.GetACL("t-data/blacklist.lst")
		So(out["/bad/test1"], ShouldEqual, "")
		err = c.AddACL("t-data/blacklist.lst", "/bad/test1")
		So(err, ShouldEqual, nil)
		out, err = c.GetACL("t-data/blacklist.lst")
		So(out["/bad/test1"], ShouldNotEqual, "")
	})
	Convey("Add ACL via id", t, func() {
		err = c.AddACL("#1", "/bad/test2")
		So(err, ShouldEqual, nil)
		out, err := c.GetACL("#1")
		So(err, ShouldEqual, nil)
		So(out["/bad/test2"], ShouldNotEqual, "")
	})
	
	Convey("Delete ACL", t, func() {
		Convey("Delete existing acl", func() {
			_ = c.AddACL("t-data/blacklist.lst", "/bad/test1")
			err = c.DeleteACL("t-data/blacklist.lst", "/bad/test1")
			So(err, ShouldEqual, nil)
			out, _ := c.GetACL("t-data/blacklist.lst")
			So(out["/bad/test1"], ShouldEqual, "")
		})
		Convey("Delete nonexisting acl", func() {
			err = c.DeleteACL("t-data/blacklist.lst", "/bad/test1/nothing")
			So(err, ShouldNotEqual, nil)
		})
		_ = err
	})
	Convey("Delete with empty ID", t, func() {
		err = c.DeleteACL("t-data/blacklist.lst", "  \t")
		So(err, ShouldNotEqual, nil)
	})
	Convey("Clear ACL", t, func() {
		Convey("Clear existing file ACL", func() {
			err := c.ClearACL("t-data/blacklist.lst")
			So(err, ShouldEqual, nil)			
		})
		Convey("Clear ACL by ID", func() {
			err := c.ClearACL("#1")
			So(err, ShouldEqual, nil)			
		})
		Convey("Clear nonexisting ACL", func() {
			err := c.ClearACL("1")
			So(err, ShouldNotEqual, nil)			
		})

	})	

	defer stopTestHaproxy()
}

