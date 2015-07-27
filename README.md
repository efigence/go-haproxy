[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/efigence/go-haproxy)

Various HAProxy-related helpers. So far:

* parsing HTTP log from syslog format (WiP)



## Testing

add your local log lines to `t-data/haproxy_log_local` and they will be used instead of ones in repo (that file is in gitignore)

to generate it just run haproxy with `log        127.0.0.1:50514   local3 debug` in global section and record it with `nc -l -u 50514 > /tmp/log`

