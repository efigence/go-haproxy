package haproxy

import (
	"bufio"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"os"
	"strings"
	"testing"
)

var testStrings []string

func TestTS(t *testing.T) {
	var err error
	ts, err := decodeTs("23/Jul/2015:13:49:11.933")
	if err != nil {
		t.Errorf("cant decode timestamp: %s", err)
	}
	if ts.UnixNano() != 1437659351933000000 {
		t.Errorf("Timestamp decoded incorrectly")
	}
	ts, err = decodeTs("03/Jul/2015:13:49:11.933")
	if err != nil {
		t.Errorf("cant decode timestamp: %s", err)
	}
	if ts.UnixNano() != 1435931351933000000 {
		t.Errorf("Timestamp decoded incorrectly")
	}
}

func TestLogParsing(t *testing.T) {
	s := `<158>Jul 23 13:49:13 haproxy[11446]: 83.3.255.169:61059 [23/Jul/2015:13:49:11.933] front1_foobar~ backend_foobar-ssl/app3-backend 1294/0/1/52/1348 200 1140 - - --VN 1637/7/5/6/0 0/0 "POST /query/q/Sql HTTP/1.1"`
	out, err := DecodeHTTPLog(s)
	require.NoError(t, err)
	t.Run("Log - Error", func(t *testing.T) {
		_, err := DecodeHTTPLog("23ej87thfdsg623gtr")
		require.Error(t, err)
	})
	t.Run("Log - parseError - too big int", func(t *testing.T) {
		_, err := DecodeHTTPLog(`<158>Jul 23 13:49:13 haproxy[1234]: 83.3.255.169:61059 [23/Jul/2015:13:49:11.933] front1_foobar~ backend_foobar-ssl/app3-backend 1294/0/1/52/1348 1000000 1140 - - --VN 1637/7/5/6/0 0/0 "POST /query/q/Sql HTTP/1.1"`)
		require.Error(t, err)
	})
	t.Run("Log - POST with SSL", func(t *testing.T) {
		a := assert.New(t)
		a.EqualValues(out.TS, 1437659351933000000, "ts")
		a.EqualValues(out.PID, 11446, "pid")
		a.Equal(out.ClientIP, "83.3.255.169", "ClientIP")
		a.EqualValues(out.ClientPort, 61059, "ClientPort")
		a.Equal(out.ClientSSL, true, "SSL")
		a.Equal(out.FrontendName, "front1_foobar", "FrontendName")
		a.Equal(out.BackendName, "backend_foobar-ssl", "BackendName")
		a.Equal(out.ServerName, "app3-backend", "ServerName")
		a.EqualValues(out.StatusCode, http.StatusOK, "StatusCode")
		a.EqualValues(out.BytesRead, 1140, "BytesRead")
		a.Equal(out.RequestPath, "/query/q/Sql", "Request Path")
		a.Equal(out.RequestMethod, "POST", "RequestMethod")
		a.Equal(out.HTTPVersion, "HTTP/1.1", "HTTPVersion")
		a.Equal(out.BadReq, false, "BadReq")
		a.Equal(out.Truncated, false, "Truncated")
	})
}

func TestBadReq(t *testing.T) {
	s := `<158>Jul 23 13:49:11 haproxy[11446]: 83.7.1.151:52174 [23/Jul/2015:13:49:06.525] front_tst-static front_tst-static/<NOSRV> -1/-1/-1/-1/5000 400 187 - - CR-- 1615/1130/0/0/0 0/0 "<BADREQ>"`
	out, err := DecodeHTTPLog(s)
	t.Run("1", func(t *testing.T) {
		assert.NoError(t, err)
		assert.True(t, out.BadReq)
		assert.Greater(t, out.StatusCode, int16(0))
		assert.EqualValues(t, TerminationClientAbort, out.TerminationReason)
		assert.EqualValues(t, SessionCloseRequest, out.SessionCloseState)
		assert.Equal(t, "<NOSRV>", out.ServerName, "%+v", out)

	})
}

func TestInvalidReq(t *testing.T) {
	s := `<158>Jul 23 13:49:13 haproxy[11446]: 83.3.255.169:61059 [23/Jul/2015:13:49:11.933] front1_foobar~ backend_foobar-ssl/app3-backend 1294/0/1/52/1348 200 1140 - - --VN 1637/7/5/6/0 0/0 "POST  HTTP/1.1"`
	_, err := DecodeHTTPLog(s)
	assert.Error(t, err)
}

func TestBulkLog(t *testing.T) {
	// use if you have some local data you want to test it with
	local_log_file := "t-data/haproxy_log_local"
	log_file := "t-data/haproxy_log"
	if _, err := os.Stat(local_log_file); err == nil {
		log_file = local_log_file
	}
	f, err := os.Open(log_file)
	if err != nil {
		t.Errorf("Cant open log file %s:%s", log_file, err)

	}
	scanner := bufio.NewScanner(f)
	i := 0
	for scanner.Scan() {
		i++
		s := scanner.Text()
		tName := fmt.Sprintf("Batch: Line %d", int(i))
		out, err := DecodeHTTPLog(s)
		t.Run(tName+":"+s, func(t *testing.T) {
			assert.NoError(t, err)
			assert.Greater(t, out.TS, int64(1437153662000))
			assert.NotEqual(t, out.StatusCode, 0)
			if strings.Contains(s, "<BADREQ>") {
				assert.Contains(t, out.RequestPath, "BADREQ")
			} else {
				assert.Contains(t, out.RequestPath, "/")
			}
			assert.Contains(t, out.HTTPVersion, "HTTP")

		})
	}
}

func TestTruncatedReq(t *testing.T) {
	s := `<158>Jul 23 13:49:11 haproxy[12345]: 11174.211.190:10165 [23/Jul/2015:13:49:10.989] front1 backend-static/bl3-varnish 432/0/0/0/432 200 11429 - - ---- 1590/1125/3/2/0 0/0 "GET /gfx/11/11/11/test/111111111111111111111//S%C3%83%C6%92%C3%86%E2%80%99%C3%83%E2%80%A0%C3%A2%E2%82%AC%E2%84%A2%C3%83%C6%92%C3%A2%E2%82%AC%C2%A0%C3%83%C2%A2%C3%A2%E2%80%9A%C2%AC%C3%A2%E2%80%9E%C2%A2%C3%83%C6%92%C3%86%E2%80%99%C3%83%C2%A2%C3%A2%E2%80%9A%C2%AC%C3%82%C2%A0%C3%83%C6%92%C3%82%C2%A2%C3%83%C2%A2%C3%A2%E2%82%AC%C5%A1%C3%82%C2%AC%C3%83%C2%A2%C3%A2%E2%82%AC%C5%BE%C3%82%C2%A2%C3%83%C6%92%C3%86%E2%80%99%C3%83%E2%80%A0%C3%A2%E2%82%AC%E2%84%A2%C3%83%C6%92%C3%A2%E2%82%AC%C5%A1%C3%83%E2%80%9A%C3%82%C2%A2%C3%83%C6%92%C3%86%E2%80%99%C3%83%E2%80%9A%C3%82%C2%A2%C3%83%C6%92%C3%82%C2%A2%C3%83%C2%A2%C3%A2%E2%82%AC%C5%A1%C3%82%C2%AC%C3%83%E2%80%A6%C3%82%C2%A1%C3%83%C6%92%C3%A2%E2%82%AC%C5%A1%C3%83%E2%80%9A%C3%82%C2%AC%C3%83%C6%92%C3%86%E2%80%99%C3%83%C2%A2%C3%A2%E2%80%9A%C2%AC%C3%82%C2%A6%C3%83%C6%92%C3%A2%E2%82%AC%C5%A1%C3%83%E2%80%9A%C3%82%C2%BE%C3%83%C6%92%C3%86`
	out, err := DecodeHTTPLog(s)
	t.Run("1", func(t *testing.T) {
		a := assert.New(t)
		a.Equal(err, nil)
		a.Equal(out.Truncated, true)
		a.Greater(out.StatusCode, int16(0))
		a.Contains(out.RequestPath, "11/test/11")
		a.Contains(out.HTTPVersion, "HTTP")
	})
}

func TestEolGarbage(t *testing.T) {
	s := `<158>Jul 23 13:49:13 haproxy[11446]: 83.3.255.169:61059 [23/Jul/2015:13:49:11.933] front1_foobar~ backend_foobar-ssl/app3-backend 1294/0/1/52/1348 200 1140 - - --VN 1637/7/5/6/0 0/0 "POST /query/q/Sql HTTP/1.1"  
`
	_, err := DecodeHTTPLog(s)
	assert.NoError(t, err)
}

func BenchmarkParser(b *testing.B) {
	s := `<158>Jul 23 13:49:13 haproxy[11446]: 83.3.255.169:61059 [23/Jul/2015:13:49:11.933] front1_foobar~ backend_foobar-ssl/app3-backend 1294/0/1/52/1348 200 1140 - - --VN 1637/7/5/6/0 0/0 "POST /query/q/Sql HTTP/1.1"`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DecodeHTTPLog(s)
	}
}
