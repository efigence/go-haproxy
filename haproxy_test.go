package haproxy

import (
	"testing"
	"fmt"
	"os"
	. "github.com/smartystreets/goconvey/convey"
)

var testStrings[] string


func TestMain(m *testing.M) {
	var err error
	testStrings, err = readLines("t-data/haproxy_udp")
	if err != nil {
		fmt.Printf("Can't load haproxy logs")
		os.Exit(255)
	}
	_ = testStrings
	os.Exit(m.Run())
}


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
	decoded, err := DecodeHTTPLog(s)
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
			So(decoded.TS, ShouldEqual, uint64(1437659351933000000))
		})
		Convey("Pid", func() {
			 So(decoded.PID, ShouldEqual, int(11446))
		 })
		Convey("ClientIP", func() {
			 So(decoded.ClientIP, ShouldEqual, "83.3.255.169")
		 })
		Convey("ClientPort", func() {
			So(decoded.ClientPort, ShouldEqual, uint16(61059))
		})
		Convey("SSL", func() {
			 So(decoded.ClientSSL, ShouldEqual, true)
		})
		Convey("FrontendName", func() {
			So(decoded.FrontendName, ShouldEqual, "front1_foobar")
		})
		Convey("BackendName", func() {
			So(decoded.BackendName, ShouldEqual, "backend_foobar-ssl")
		})
		Convey("ServerName", func() {
			So(decoded.ServerName, ShouldEqual, "app3-backend")
		})
		Convey("StatusCode", func() {
			So(decoded.StatusCode, ShouldEqual, uint16(200))
		})
		Convey("BytesRead", func() {
			So(decoded.BytesRead, ShouldEqual, uint64(1140))
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
