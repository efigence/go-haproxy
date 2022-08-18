package haproxy

// https://cbonte.github.io/haproxy-dconv/2.0/configuration.html#8.5
const (
	TerminationClientAbort     = `C`
	TerminationServerAbort     = `S`
	TerminationDeny            = `P`
	TerminationLocal           = `L`
	TerminationExhausted       = `R`
	TerminationInternalError   = `I`
	TerminationServerGoingDown = `D`
	TerminationActiveUp        = `U`
	TerminationAdmin           = `K`
	TerminationClientWait      = `c`
	TerminationServerWait      = `s`
	TerminationNone            = `-`
)
const (
	SessionCloseRequest    = `R`
	SessionCloseQueue      = `Q`
	SessionCloseConnection = `C`
	SessionCloseHeaders    = `H`
	SessionCloseData       = `D`
	SessionCloseLast       = `L`
	SessionCloseTarpit     = `T`
	SessionCloseNone       = `-`
)
