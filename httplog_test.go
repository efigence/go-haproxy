package haproxy

import (
	"bufio"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
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
	if err != nil {
		t.Errorf("cant decode log: %s", err)
	}
	Convey("Log - Error", t, func() {
		_, err := DecodeHTTPLog("23ej87thfdsg623gtr")
		So(err, ShouldNotEqual, nil)
	})
	Convey("Log - parseError - too big int", t, func() {
		_, err := DecodeHTTPLog(`<158>Jul 23 13:49:13 haproxy[1234]: 83.3.255.169:61059 [23/Jul/2015:13:49:11.933] front1_foobar~ backend_foobar-ssl/app3-backend 1294/0/1/52/1348 1000000 1140 - - --VN 1637/7/5/6/0 0/0 "POST /query/q/Sql HTTP/1.1"`)
		So(err, ShouldNotEqual, nil)
	})
	Convey("Log - POST with SSL", t, func() {
		Convey("TS", func() {
			So(out.TS, ShouldEqual, uint64(1437659351933000000))
		})
		Convey("Pid", func() {
			So(out.PID, ShouldEqual, int(11446))
		})
		Convey("ClientIP", func() {
			So(out.ClientIP, ShouldEqual, "83.3.255.169")
		})
		Convey("ClientPort", func() {
			So(out.ClientPort, ShouldEqual, uint16(61059))
		})
		Convey("SSL", func() {
			So(out.ClientSSL, ShouldEqual, true)
		})
		Convey("FrontendName", func() {
			So(out.FrontendName, ShouldEqual, "front1_foobar")
		})
		Convey("BackendName", func() {
			So(out.BackendName, ShouldEqual, "backend_foobar-ssl")
		})
		Convey("ServerName", func() {
			So(out.ServerName, ShouldEqual, "app3-backend")
		})
		Convey("StatusCode", func() {
			So(out.StatusCode, ShouldEqual, int16(200))
		})
		Convey("BytesRead", func() {
			So(out.BytesRead, ShouldEqual, uint64(1140))
		})
		Convey("RequestPath", func() {
			So(out.RequestPath, ShouldEqual, "/query/q/Sql")
		})
		Convey("RequestMethod", func() {
			So(out.RequestMethod, ShouldEqual, "POST")
		})
		Convey("HTTPVersion", func() {
			So(out.HTTPVersion, ShouldEqual, "HTTP/1.1")
		})
		Convey("BadReq", func() {
			So(out.BadReq, ShouldEqual, false)
		})
		Convey("Truncated", func() {
			So(out.Truncated, ShouldEqual, false)
		})
	})
}

func TestBadReq(t *testing.T) {
	s := `<158>Jul 23 13:49:11 haproxy[11446]: 83.7.1.151:52174 [23/Jul/2015:13:49:06.525] front_tst-static front_tst-static/<NOSRV> -1/-1/-1/-1/5000 400 187 - - CR-- 1615/1130/0/0/0 0/0 "<BADREQ>"`
	out, err := DecodeHTTPLog(s)
	Convey("Log - Bad request", t, func() {
		Convey("Should parse", func() {
			So(err, ShouldEqual, nil)
		})
		Convey("BadReq", func() {
			So(out.BadReq, ShouldEqual, true)
		})
		Convey("StatusCode", func() {
			So(out.StatusCode, ShouldBeGreaterThan, 0)
		})

	})
}

func TestInvalidReq(t *testing.T) {
	s := `<158>Jul 23 13:49:13 haproxy[11446]: 83.3.255.169:61059 [23/Jul/2015:13:49:11.933] front1_foobar~ backend_foobar-ssl/app3-backend 1294/0/1/52/1348 200 1140 - - --VN 1637/7/5/6/0 0/0 "POST  HTTP/1.1"`
	_, err := DecodeHTTPLog(s)
	Convey("Bad input format", t, func() {
		So(err, ShouldNotEqual, nil)
	})
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
		Convey(tName+":"+s, t, func() {
			Convey(tName+" parsing", func() {
				So(err, ShouldEqual, nil)
			})
			Convey(tName+" TS", func() {
				So(out.TS, ShouldBeGreaterThan, 1437153662000)
			})
			Convey(tName+" StatusCode", func() {
				So(out.StatusCode, ShouldNotEqual, 0)
			})
			Convey(tName+" Path", func() {
				if strings.Contains(s, "<BADREQ>") {
					So(out.RequestPath, ShouldContainSubstring, "BADREQ")
				} else {
					So(out.RequestPath, ShouldContainSubstring, "/")
				}
			})
			Convey(tName+" HTTPVersion", func() {
				So(out.HTTPVersion, ShouldContainSubstring, "HTTP")
			})
		})

	}
}

func TestTruncatedReq(t *testing.T) {
	s := `<158>Jul 23 13:49:11 haproxy[12345]: 11174.211.190:10165 [23/Jul/2015:13:49:10.989] front1 backend-static/bl3-varnish 432/0/0/0/432 200 11429 - - ---- 1590/1125/3/2/0 0/0 "GET /gfx/11/11/11/test/111111111111111111111//S%C3%83%C6%92%C3%86%E2%80%99%C3%83%E2%80%A0%C3%A2%E2%82%AC%E2%84%A2%C3%83%C6%92%C3%A2%E2%82%AC%C2%A0%C3%83%C2%A2%C3%A2%E2%80%9A%C2%AC%C3%A2%E2%80%9E%C2%A2%C3%83%C6%92%C3%86%E2%80%99%C3%83%C2%A2%C3%A2%E2%80%9A%C2%AC%C3%82%C2%A0%C3%83%C6%92%C3%82%C2%A2%C3%83%C2%A2%C3%A2%E2%82%AC%C5%A1%C3%82%C2%AC%C3%83%C2%A2%C3%A2%E2%82%AC%C5%BE%C3%82%C2%A2%C3%83%C6%92%C3%86%E2%80%99%C3%83%E2%80%A0%C3%A2%E2%82%AC%E2%84%A2%C3%83%C6%92%C3%A2%E2%82%AC%C5%A1%C3%83%E2%80%9A%C3%82%C2%A2%C3%83%C6%92%C3%86%E2%80%99%C3%83%E2%80%9A%C3%82%C2%A2%C3%83%C6%92%C3%82%C2%A2%C3%83%C2%A2%C3%A2%E2%82%AC%C5%A1%C3%82%C2%AC%C3%83%E2%80%A6%C3%82%C2%A1%C3%83%C6%92%C3%A2%E2%82%AC%C5%A1%C3%83%E2%80%9A%C3%82%C2%AC%C3%83%C6%92%C3%86%E2%80%99%C3%83%C2%A2%C3%A2%E2%80%9A%C2%AC%C3%82%C2%A6%C3%83%C6%92%C3%A2%E2%82%AC%C5%A1%C3%83%E2%80%9A%C3%82%C2%BE%C3%83%C6%92%C3%86`
	out, err := DecodeHTTPLog(s)
	Convey("Log - truncated request", t, func() {
		Convey("Should parse", func() {
			So(err, ShouldEqual, nil)
		})
		Convey("Truncated", func() {
			So(out.Truncated, ShouldEqual, true)
		})
		Convey("StatusCode", func() {
			So(out.StatusCode, ShouldBeGreaterThan, 0)
		})
		Convey("RequestPath", func() {
			So(out.RequestPath, ShouldContainSubstring, "11/test/11")
		})
		Convey("HTTPVersion"+" HTTPVersion", func() {
			So(out.HTTPVersion, ShouldContainSubstring, "HTTP")
		})
	})
}

func TestEolGarbage(t *testing.T) {

	s := `<158>Jul 23 13:49:13 haproxy[11446]: 83.3.255.169:61059 [23/Jul/2015:13:49:11.933] front1_foobar~ backend_foobar-ssl/app3-backend 1294/0/1/52/1348 200 1140 - - --VN 1637/7/5/6/0 0/0 "POST /query/q/Sql HTTP/1.1"  
`
	_, err := DecodeHTTPLog(s)
	Convey("Log - eol whitespace/newline", t, func() {
		Convey("Should parse", func() {
			So(err, ShouldEqual, nil)
		})
	})
}

func BenchmarkParser(b *testing.B) {
	s := `<158>Jul 23 13:49:13 haproxy[11446]: 83.3.255.169:61059 [23/Jul/2015:13:49:11.933] front1_foobar~ backend_foobar-ssl/app3-backend 1294/0/1/52/1348 200 1140 - - --VN 1637/7/5/6/0 0/0 "POST /query/q/Sql HTTP/1.1"`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DecodeHTTPLog(s)
	}
}
