package haproxy



// StatsByID defines order of fields when asking HAProxy for csv
// ORDER MATTERS, it is used to decode csv without having to parse header on each request
// it is also used by GenerateStatusId() to fill in StatsInfo
// only reason to modify it is if upstream haproxy adds new stats, however you should consider
// making a pull request instead.
// stats description under
// https://www.haproxy.org/download/1.7/doc/management.txt
// new ones are always added at the end
var StatsById = []string{
	"pxname",
	"svname",
	"qcur",
	"qmax",
	"scur",
	"smax",
	"slim",
	"stot",
	"bin",
	"bout",
	"dreq",
	"dresp",
	"ereq",
	"econ",
	"eresp",
	"wretr",
	"wredis",
	"status",
	"weight",
	"act",
	"bck",
	"chkfail",
	"chkdown",
	"lastchg",
	"downtime",
	"qlimit",
	"pid",
	"iid",
	"sid",
	"throttle",
	"lbtot",
	"tracked",
	"type",
	"rate",
	"rate_lim",
	"rate_max",
	"check_status",
	"check_code",
	"check_duration",
	"hrsp_1xx",
	"hrsp_2xx",
	"hrsp_3xx",
	"hrsp_4xx",
	"hrsp_5xx",
	"hrsp_other",
	"hanafail",
	"req_rate",
	"req_rate_max",
	"req_tot",
	"cli_abrt",
	"srv_abrt",
	"comp_in",
	"comp_out",
	"comp_byp",
	"comp_rsp",
	"lastsess",
	"last_chk",
	"last_agt",
	"qtime",
	"ctime",
	"rtime",
	"ttime",
	"agent_status",
	"agent_code",
	"agent_duration",
	"check_desc",
	"agent_desc",
	"check_rise",
	"check_fall",
	"check_health",
	"agent_rise",
	"agent_fall",
	"agent_health",
	"addr",
	"cookie",
	"mode",
	"algo",
	"conn_rate",
	"conn_rate_max",
	"conn_tot",
	"intercepted",
	"dcon",
	"dses",
}


// DataPoint describes value of stat and its metadata like which process/server it came from and which instance (haproxy subprocess when using nbproc > 1) it came from.
// It is used because some stats need different merge function for sub-processes and for different servers.
// For example max queue size should be summed over same machine, but not everyone will want it summed over whole cluster, preferring per-machine maximum.
// On top of it, golang type system is too poor to deal with values that can be both strings and numbers and I dont want to have interface{} everywhere
type DataPoint struct {
	Server string
	Instance int
	Value float64
	StringValue string
}

// function prototype to merge stats from different HAProxies.
// in general it should ignore negative values as that is the haproxy way of indicating no data/does not apply

type MergeFunction interface {
	Merge(arg ...DataPoint) DataPoint
}

type MergeSum struct {
}

func (m *MergeSum) Merge(arg ...DataPoint) (out DataPoint) {
	for _, point := range arg {
		if point.Value > 0 {
			out.Value = out.Value + point.Value
		}
	}
	return out
}

type MergeAvg struct {
}

func (m *MergeAvg) Merge(arg ...DataPoint) (out DataPoint) {
	var acc float64
	var count int
	for _, point := range arg {
		if !(point.Value < 0) {
			acc = acc + point.Value
			count = count + 1
		}
	}
	out.Value =  acc / float64(count)
	return out
}

type MergeMax struct {
	
}

func (m *MergeMax) Merge(arg ...DataPoint) (out DataPoint) {
	out = arg[0]
	for _, point := range arg {
		if point.Value > out.Value { out.Value = point.Value }
	}
	return out
}

type MergeMin struct {
	
}
// gets minimal Value while ignoreing -1 (haproxy speech for nodata)
func (m *MergeMin) Merge(arg ...DataPoint) (out DataPoint) {
	out = arg[0]
	for _, point  := range arg {
		// drop -1, that in haproxy speak means no data
		if point.Value < 0 {continue}
		if point.Value < out.Value { out.Value = point.Value }
	}
	return out
}

type StatDescription struct {
	Id          int
	Name        string
	Description string
	MergeType   MergeFunction
	// Where given stat can live: L(Listener), F(Frontend), B(Backend), S(Server)
	Type string
}

type StatsInfo map[string]*StatDescription

var StatusInfo = StatsInfo{
	"pxname":         {Type: `LFBS`, Description: `proxy name`},
	"svname":         {Type: `LFBS`, Description: `service name (FRONTEND for frontend, BACKEND for backend, any name for server/listener)`},
	"qcur":           {Type: `..BS`, Description: `current queued requests. For the backend this reports the number queued without a server assigned.`},
	"qmax":           {Type: `..BS`, Description: `max Value of qcur`},
	"scur":           {Type: `LFBS`, Description: `current sessions`},
	"smax":           {Type: `LFBS`, Description: `max sessions`},
	"slim":           {Type: `LFBS`, Description: `configured session limit`},
	"stot":           {Type: `LFBS`, Description: `cumulative number of sessions`},
	"bin":            {Type: `LFBS`, Description: `bytes in`},
	"bout":           {Type: `LFBS`, Description: `bytes out`},
	"dreq":           {Type: `LFB.`, Description: `requests denied because of security concerns. \n- For tcp this is because of a matched tcp-request content rule. \n- For http this is because of a matched http-request or tarpit rule.`},
	"dresp":          {Type: `LFBS`, Description: `responses denied because of security concerns. \n- For http this is because of a matched http-request rule, or   "option checkcache".`},
	"ereq":           {Type: `LF..`, Description: `request errors. Some of the possible causes are: \n- early termination from the client, before the request has been sent. \n- read error from the client \n- client timeout \n- client closed connection \n- various bad requests from the client. \n- request was tarpitted.`},
	"econ":           {Type: `..BS`, Description: `number of requests that encountered an error trying to connect to a backend server. The backend stat is the sum of the stat for all servers of that backend, plus any connection errors not associated with a particular server (such as the backend having no active servers).`},
	"eresp":          {Type: `..BS`, Description: `response errors. srv_abrt will be counted here also. Some other errors are: \n- write error on the client socket (won't be counted for the server stat) \n- failure applying filters to the response.`},
	"wretr":          {Type: `..BS`, Description: `number of times a connection to a server was retried.`},
	"wredis":         {Type: `..BS`, Description: `number of times a request was redispatched to another server. The server Value counts the number of times that server was switched away from.`},
	"status":         {Type: `LFBS`, Description: `status (UP/DOWN/NOLB/MAINT/MAINT(via)/MAINT(resolution)...)`},
	"weight":         {Type: `..BS`, Description: `total weight (backend), server weight (server)`},
	"act":            {Type: `..BS`, Description: `number of active servers (backend), server is active (server)`},
	"bck":            {Type: `..BS`, Description: `number of backup servers (backend), server is backup (server)`},
	"chkfail":        {Type: `...S`, Description: `number of failed checks. (Only counts checks failed when the server is up.)`},
	"chkdown":        {Type: `..BS`, Description: `number of UP->DOWN transitions. The backend counter counts transitions to the whole backend being down, rather than the sum of the counters for each server.`},
	"lastchg":        {Type: `..BS`, Description: `number of seconds since the last UP<->DOWN transition`},
	"downtime":       {Type: `..BS`, Description: `total downtime (in seconds). The Value for the backend is the downtime for the whole backend, not the sum of the server downtime.`},
	"qlimit":         {Type: `...S`, Description: `configured maxqueue for the server, or nothing in the Value is 0 (default, meaning no limit)`},
	"pid":            {Type: `LFBS`, Description: `process id (0 for first instance, 1 for second, ...)`},
	"iid":            {Type: `LFBS`, Description: `unique proxy id`},
	"sid":            {Type: `L..S`, Description: `server id (unique inside a proxy)`},
	"throttle":       {Type: `...S`, Description: `current throttle percentage for the server, when slowstart is active, or no value if not in slowstart.`},
	"lbtot":          {Type: `..BS`, Description: `total number of times a server was selected, either for new sessions, or when re-dispatching. The server counter is the number of times that server was selected.`},
	"tracked":        {Type: `...S`, Description: `id of proxy/server if tracking is enabled.`},
	"type":           {Type: `LFBS`, Description: `(0=frontend, 1=backend, 2=server, 3=socket/listener)`},
	"rate":           {Type: `.FBS`, Description: `number of sessions per second over last elapsed second`},
	"rate_lim":       {Type: `.F..`, Description: `configured limit on new sessions per second`},
	"rate_max":       {Type: `.FBS`, Description: `max number of new sessions per second`},
	"check_status":   {Type: `...S`, Description: `status of last health check, one of:    UN -> unknown    IN -> initializing    SOCKERR -> socket error    L4OK    -> check passed on layer 4, no upper layers testing enabled    L4TOUT  -> layer 1-4 timeout    L4CON   -> layer 1-4 connection problem, for examp       "Connection refused" (tcp rst) or "No route to host" (icmp)    L6OK    -> check passed on layer 6    L6TOUT  -> layer 6 (SSL) timeout    L6RSP   -> layer 6 invalid response \n- protocol error    L7OK    -> check passed on layer 7    L7OKC   -> check conditionally passed on layer 7, for example 404 wi       disable-on-404    L7TOUT  -> layer 7 (HTTP/SMTP) timeout    L7RSP   -> layer 7 invalid response \n- protocol error    L7STS   -> layer 7 response error, for example HTTP 5xx`},
	"check_code":     {Type: `...S`, Description: `layer5-7 code, if available`},
	"check_duration": {Type: `...S`, Description: `time in ms took to finish last health check`},
	"hrsp_1xx":       {Type: `.FBS`, Description: `http responses with 1xx code`},
	"hrsp_2xx":       {Type: `.FBS`, Description: `http responses with 2xx code`},
	"hrsp_3xx":       {Type: `.FBS`, Description: `http responses with 3xx code`},
	"hrsp_4xx":       {Type: `.FBS`, Description: `http responses with 4xx code`},
	"hrsp_5xx":       {Type: `.FBS`, Description: `http responses with 5xx code`},
	"hrsp_other":     {Type: `.FBS`, Description: `http responses with other codes (protocol error)`},
	"hanafail":       {Type: `...S`, Description: `failed health checks details`},
	"req_rate":       {Type: `.F..`, Description: `HTTP requests per second over last elapsed second`},
	"req_rate_max":   {Type: `.F..`, Description: `max number of HTTP requests per second observed`},
	"req_tot":        {Type: `.FB.`, Description: `total number of HTTP requests received`},
	"cli_abrt":       {Type: `..BS`, Description: `number of data transfers aborted by the client`},
	"srv_abrt":       {Type: `..BS`, Description: `number of data transfers aborted by the server (inc. in eresp)`},
	"comp_in":        {Type: `.FB.`, Description: `number of HTTP response bytes fed to the compressor`},
	"comp_out":       {Type: `.FB.`, Description: `number of HTTP response bytes emitted by the compressor`},
	"comp_byp":       {Type: `.FB.`, Description: `number of bytes that bypassed the HTTP compressor (CPU/BW limit)`},
	"comp_rsp":       {Type: `.FB.`, Description: `number of HTTP responses that were compressed`},
	"lastsess":       {Type: `..BS`, Description: `number of seconds since last session assigned to server/backend`},
	"last_chk":       {Type: `...S`, Description: `last health check contents or textual error`},
	"last_agt":       {Type: `...S`, Description: `last agent check contents or textual error`},
	"qtime":          {Type: `..BS`, Description: `the average queue time in ms over the 1024 last requests`},
	"ctime":          {Type: `..BS`, Description: `the average connect time in ms over the 1024 last requests`},
	"rtime":          {Type: `..BS`, Description: `the average response time in ms over the 1024 last requests (0 for TCP)`},
	"ttime":          {Type: `..BS`, Description: `the average total session time in ms over the 1024 last requests`},
	"agent_status":   {Type: `...S`, Description: `status of last agent check, one of:    UN -> unknown    IN -> initializing    SOCKERR -> socket error    L4OK    -> check passed on layer 4, no upper layers testing enabled    L4TOUT  -> layer 1-4 timeout    L4CON   -> layer 1-4 connection problem, for examp       "Connection refused" (tcp rst) or "No route to host" (icmp)    L7OK    -> agent reported "up"    L7STS   -> agent reported "fail", "stop", or "down"`},
	"agent_code":     {Type: `...S`, Description: `numeric code reported by agent if any (unused for now)`},
	"agent_duration": {Type: `...S`, Description: `time in ms taken to finish last check`},
	"check_desc":     {Type: `...S`, Description: `short human-readable description of check_status`},
	"agent_desc":     {Type: `...S`, Description: `short human-readable description of agent_status`},
	"check_rise":     {Type: `...S`, Description: `server's "rise" parameter used by checks`},
	"check_fall":     {Type: `...S`, Description: `server's "fall" parameter used by checks`},
	"check_health":   {Type: `...S`, Description: `server's health check value between 0 and rise+fall-1`},
	"agent_rise":     {Type: `...S`, Description: `agent's "rise" parameter, normally 1`},
	"agent_fall":     {Type: `...S`, Description: `agent's "fall" parameter, normally 1`},
	"agent_health":   {Type: `...S`, Description: `agent's health parameter, between 0 and rise+fall-1`},
	"addr":           {Type: `L..S`, Description: `address:port or "unix". IPv6 has brackets around the address.`},
	"cookie":         {Type: `..BS`, Description: `server's cookie value or backend's cookie name`},
	"mode":           {Type: `LFBS`, Description: `proxy mode (tcp, http, health, unknown)`},
	"algo":           {Type: `..B.`, Description: `load balancing algorithm`},
	"conn_rate":      {Type: `.F..`, Description: `number of connections over the last elapsed second`},
	"conn_rate_max":  {Type: `.F..`, Description: `highest known conn_rate`},
	"conn_tot":       {Type: `.F..`, Description: `cumulative number of connections`},
	"intercepted":    {Type: `.FB.`, Description: `cum. number of intercepted requests (monitor, stats)`},
	"dcon":           {Type: `LF..`, Description: `requests denied by "tcp-request connection" rules`},
	"dses":           {Type: `LF..`, Description: `requests denied by "tcp-request session" rules`},
}

// stats that should be summed up when aggregating multiple backends
var statSum = []string{
	"dcon",
	"dses",
	"qmax",
	"smax",
	"rate_max",
	"req_rate_max",

}
// stuff that should be averaged out between multiple instances
var statMax = []string{
	"scur",
	"smax",
}

// stuff that should be minimal value of all (like "last change")
var statMin = []string{
	"lastch",
	"lastsess",
}


const Frontend = "F"
const Backend = "B"
const Listener = "L"
const Server = "S"


func init() {
	// sadly there is no way to do it non-runtime without duplication.
	GenerateStatusInfo()
}

// update status info and merge function (which stats should be averaged or summed). Normally called automatically from init.
func GenerateStatusInfo() {

	for k, v := range StatsById {
		StatusInfo[v].Name = v
		StatusInfo[v].Id = k
	}
	for _, v := range statSum {
		StatusInfo[v].MergeType = &MergeSum{}
	}
	for _, v := range statMax {
		StatusInfo[v].MergeType = &MergeMax{}
	}
}
