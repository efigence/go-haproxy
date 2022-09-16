[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://pkg.go.dev/github.com/efigence/go-haproxy)

# HAProxy utilities and helpers for Golang

## parsing HTTP log from syslog format (WiP)

Most of the fields are decoded, log accepts format of `log  127.0.0.1:50514   local3 debug`.
The syslog facility is ignored but the app name must be `haproxy`

That's how basic ingestor can look like:
```go
func (i *Ingest) ingestor(conn *net.UDPConn, ch chan haproxy.HTTPRequest) {
	buf := make([]byte, 65535)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		log_str := string(buf[0:n])
		if err != nil {
			i.l.Errorf("Error: %s", err)
		}
		if strings.Contains(log_str, " SSL handshake") {
			continue
		}
		req, err := haproxy.DecodeHTTPLog(log_str)
		if err != nil {
			continue
		}
		ch <- req
	}
}	
```


### Quirks

#### Request/response variable capture

This is how `capture request ` only block looks like: `{}`

This is how `capture response ` only block looks like: `{}`

This is how config with both looks like in log: `{} {}`

This code is not and will not guess on that, it just dumps the whole into `CapturedHeaders` string

#### Handling truncated requests

The usual reason is the last part, HTTP request part, not fitting the packet.

It will be parsed on best effort basis and will have `Truncated=true` set in the structure

#### Time handling

HAProxy sends logs in local time, so we decode it in local time. 
There is a global variable `HaproxyLogTimezone` that will set the zone used to decode time.

### Testing

add your local log lines to `t-data/haproxy_log_local` and they will be used instead of ones in repo (that file is in gitignore)

to generate it just run haproxy with `log        127.0.0.1:50514   local3 debug` in global section and record it with `nc -l -u 50514 > /tmp/log`

