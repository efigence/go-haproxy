package haproxy

import (
	"time"
	"regexp"
	"strconv"
	"errors"
)

const haproxyTimeFormat = "02/Jan/2006:15:04:05.000"
var haproxyRegex      = regexp.MustCompile(`.*haproxy\[(\d+)]: (.+?):(\d+) \[(.+?)\] (.+?)([\~]) (.+?)\/(.+?) ([\-0-9]+)\/([\-0-9]+)\/([\-0-9]+)\/([\-0-9]+)\/([\-0-9]+) (\d+) (\d+)`)

// HAProxy http log format
// https://cbonte.github.io/haproxy-dconv/configuration-1.5.html#8.2.3
type HTTPRequest struct {
	TS                      int64             `json:"ts_ms"`
	PID                     int               `json:"pid"` // necessary to distinguish conns hitting different processes
	ClientIP                string            `json:"client_ip"`
	ClientPort              uint16            `json:"client_port"`
	ClientSSL               bool              `json:"client_ssl"`
	FrontendName            string            `json:"frontend_name"`
	BackendName             string            `json:"backend_name"`
	ServerName              string            `json:"server_name"`
	StatusCode              uint16            `json:"status_code"`
	BytesRead               uint64            `json:"bytes_read"`
	CapturedRequestCookie   map[string]string `json:"captured_request_cookie"`
	CapturedResponseCookie  map[string]string `json:"captured_response_cookie"`
	CapturedRequestHeaders  map[string]string `json:"captured_request_headers"`
	CapturedResponseHeaders map[string]string `json:"captured_response_headers"`
	// timings
	// aborted connections are marked via -1 by haproxy
	RequestHeaderDurationMs  int `json:"request_header_duration_ms"`  // Tq
	QueueDurationMs          int `json:"queue_duration_ms"`           // Tw
	ServerConnDurationMs     int `json:"server_conn_duration_ms"`     // Tc
	ResponseHeaderDurationMs int `json:"response_header_duration_ms"` // Tr
	TotalDurationMs          int `json:"total_duration_ms"`           // Tt

	HTTPRequest string `json:"http_request"`
	HTTPMethod  string `json:"http_method"`

	// Connection state
	TerminationReason      rune `json:"termination_reason"`
	SessionCloseState      rune `json:"session_close_state"`
	ClientPersistenceState rune `json:"client_persistence_state"`
	PersistenceCookieState rune `json:"persistence_cookie"`

	// conn count stats (per-pid
	TotalConn    uint `json:"total_conn"`
	FrontendConn uint `json:"frontend_conn"`
	BackendConn  uint `json:"backend_conn"`
	ServerConn   uint `json:"server_conn"`
	Retries      uint `json:"retries"`
	ServerQueue  uint `json:"server_queue"`
	BackendQueue uint `json:"backend_queue"`
}

// Decode haproxy UDP sender string into http request

func DecodeHTTPLog(s string) (HTTPRequest, error) {
	var r HTTPRequest
	var err error
	matches := haproxyRegex.FindStringSubmatch(s)
	if len(matches) < 8 {
		return r, errors.New("input not matching regex")
	}
	r.PID, err = strconv.Atoi(matches[1])
	r.ClientIP = matches[2]
	
	ui16_cp, err :=  strconv.ParseUint(matches[3],10,16)
	if err != nil { return r, err }
	r.ClientPort = uint16(ui16_cp)
	
	ts, err := decodeTs(matches[4])
	if err != nil { return r, err }
	r.TS = ts.UnixNano()

	r.FrontendName = matches[5]

	if matches[6] == `~` {
		r.ClientSSL = true
	}
	

	r.BackendName = matches[7]

	r.ServerName = matches[8]
	
	r.RequestHeaderDurationMs , err =  strconv.Atoi(matches[9])
	if err != nil { return r, err }

	r.QueueDurationMs , err =  strconv.Atoi(matches[10])
	if err != nil { return r, err }

	r.ServerConnDurationMs , err =  strconv.Atoi(matches[11])
	if err != nil { return r, err }

	r.ResponseHeaderDurationMs , err =  strconv.Atoi(matches[12])
	if err != nil { return r, err }

	r.TotalDurationMs , err =  strconv.Atoi(matches[13])

	ui16_sc, err :=  strconv.ParseUint(matches[14],10,16)
	if err != nil { return r, err }
	r.StatusCode = uint16(ui16_sc)

	ui64_br, err :=  strconv.ParseUint(matches[15],10,64)
	if err != nil { return r, err }
	r.BytesRead = uint64(ui64_br)

	return r, err
}

func decodeTs(s string) (ts time.Time, err error) {
	ts, err = time.Parse(haproxyTimeFormat, s)
	return ts, err
}
