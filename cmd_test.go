package haproxy

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var testConn Conn

func TestACL(t *testing.T) {
	// fixme generate tmpfile unix socket
	err := runTestHaproxy()
	defer stopTestHaproxy()
	t.Run("Start without module", func(t *testing.T) {
		c := &Conn{}
		// TODO fix panics
		assert.Error(t, c.AddACL("asd", "/asd"))
		assert.Panics(t, func() { c.DeleteACL("asd", "1") })
		_, err := c.GetACL("asd")
		assert.Error(t, err)
		assert.Panics(t, func() { c.ClearACL("asd") })
		_, err = c.ListACL()
		assert.Error(t, err)
		_, err = c.ListACLFiles()
		assert.Error(t, err)
	})
	t.Run("Start test haproxy", func(t *testing.T) {
		assert.NoError(t, err)
	})
	c := New("tmp/haproxy.sock")
	t.Run("List ACL", func(t *testing.T) {
		out, err := c.RunCmd("show acl")
		assert.NoError(t, err)
		assert.Contains(t, out[0], "# id") // header
	})
	t.Run("List all ACL", func(t *testing.T) {
		out, err := c.RunCmd("show acl")
		assert.NoError(t, err)
		assert.Contains(t, out[1], `blacklist.lst`)
	})
	t.Run("List ACL entries", func(t *testing.T) {
		out, err := c.GetACL("t-data/blacklist.lst")
		assert.NoError(t, err)
		assert.NotEqual(t, out["/from/file"], nil)
	})
	t.Run("Add to existing fileACL", func(t *testing.T) {
		out, err := c.GetACL("t-data/blacklist.lst")
		assert.Equal(t, out["/bad/test1"], "")
		err = c.AddACL("t-data/blacklist.lst", "/bad/test1")
		assert.NoError(t, err)
		out, err = c.GetACL("t-data/blacklist.lst")
		assert.NotEqual(t, out["/bad/test1"], "")
	})
	t.Run("Add ACL via id", func(t *testing.T) {
		err = c.AddACL("#1", "/bad/test2")
		assert.NoError(t, err)
		out, err := c.GetACL("#1")
		assert.NoError(t, err)
		assert.NotEqual(t, out["/bad/test2"], "")
	})
	t.Run("Add ACL wrong data", func(t *testing.T) {
		err = c.AddACL("!@%#&^$((()#)&^%", "@(#*&!^#@$(^!@)*#^!@")
		assert.Error(t, err)
	})

	t.Run("Delete ACL", func(t *testing.T) {
		t.Run("Delete existing acl", func(t *testing.T) {
			err = c.AddACL("t-data/blacklist.lst", "/bad/test1")
			require.NoError(t, err)
			err = c.DeleteACL("t-data/blacklist.lst", "/bad/test1")
			require.NoError(t, err)
			out, _ := c.GetACL("t-data/blacklist.lst")
			require.Equal(t, out["/bad/test1"], "")
		})
		t.Run("Delete nonexisting acl", func(t *testing.T) {
			err = c.DeleteACL("t-data/blacklist.lst", "/bad/test1/nothing")
			assert.Error(t, err)
		})
		_ = err
	})
	t.Run("Delete with empty ID", func(t *testing.T) {
		err = c.DeleteACL("t-data/blacklist.lst", "  \t")
		assert.Error(t, err)
	})
	t.Run("Clear ACL", func(t *testing.T) {
		t.Run("Clear existing file ACL", func(t *testing.T) {
			err := c.ClearACL("t-data/blacklist.lst")
			assert.NoError(t, err)
		})
		t.Run("Clear ACL by ID", func(t *testing.T) {
			err := c.ClearACL("#1")
			assert.NoError(t, err)
		})
		t.Run("Clear nonexisting ACL", func(t *testing.T) {
			err := c.ClearACL("1")
			assert.Error(t, err)
		})
	})
	t.Run("List all ACLs", func(t *testing.T) {
		out, err := c.ListACL()
		fmt.Printf("%+v", out)
		t.Run("File acl", func(t *testing.T) {
			assert.NoError(t, err)
			assert.Equal(t, 0, out[0].ID)
			assert.Equal(t, "file", out[0].Type)
			assert.Equal(t, "t-data/blacklist.lst", out[0].SourceFile)
		})
		t.Run("Inline acl", func(t *testing.T) {
			assert.NoError(t, err)
			assert.Equal(t, 1, out[1].ID)
			assert.Equal(t, "path_beg", out[1].Type)
			assert.Equal(t, 18, out[1].Line)
		})
		t.Run("ACL by file name", func(t *testing.T) {
			out, err := c.ListACLFiles()
			assert.NoError(t, err)
			assert.EqualValues(t, out["t-data/blacklist.lst"].ID, 0)
		})
	})

	defer stopTestHaproxy()
}
