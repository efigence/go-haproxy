package haproxy


// 4 bits
// TYPED-DATA    : <TYPE:4 bits><FLAGS:4 bits><DATA>
//
// Variable-length integer (varint) are encoded using Peers encoding:
//
//
//        0  <= X < 240        : 1 byte  (7.875 bits)  [ XXXX XXXX ]
//       240 <= X < 2288       : 2 bytes (11 bits)     [ 1111 XXXX ] [ 0XXX XXXX ]
//      2288 <= X < 264432     : 3 bytes (18 bits)     [ 1111 XXXX ] [ 1XXX XXXX ]   [ 0XXX XXXX ]
//    264432 <= X < 33818864   : 4 bytes (25 bits)     [ 1111 XXXX ] [ 1XXX XXXX ]*2 [ 0XXX XXXX ]
//  33818864 <= X < 4328786160 : 5 bytes (32 bits)     [ 1111 XXXX ] [ 1XXX XXXX ]*3 [ 0XXX XXXX ]
const (
	// TYPED-DATA    : <TYPE:4 bits><FLAGS:4 bits><DATA>
    DataTypeNULL   = 0 //<0>
    DataTypeBOOL   = 1 //<1+FLAG>
    DataTypeINT32  = 2 //<2><VALUE:varint>
    DataTypeUINT32 = 3 //<3><VALUE:varint>
    DataTypeINT64  = 4 //<4><VALUE:varint>
    DataTypeUNIT64 = 5 //<5><VALUE:varint>
    DataTypeIPV4   = 6 //<6><STRUCT IN_ADDR:4 bytes>
    DataTypeIPV6   = 7 //<7><STRUCT IN_ADDR6:16 bytes>
    DataTypeSTRING = 8 //<8><LENGTH:varint><BYTES>
    DataTypeBINARY = 9 //<9><LENGTH:varint><BYTES>
    DataTypeRSVD1  = 10
    DataTypeRSVD2  = 11
    DataTypeRSVD3  = 12
    DataTypeRSVD4  = 13
    DataTypeRSVD5  = 14
    DataTypeRSVD6  = 15
)

const (
	FrameHaproxyHello      = 1   // Sent by HAProxy when it opens a connection on an agent.
	FrameHaproxyDisconnect = 2   // Sent by HAProxy when it want to close the connection or in reply to an AGENT-DISCONNECT frame
	FrameNotify            = 3   // Sent by HAProxy to pass information to an agent
	FrameAgentHello        = 101 // Reply to a HAPROXY-HELLO frame, when the connection is established
    FrameAgentDDisconnect  = 102 // Sent by an agent just before closing the connection
	FrameAck               = 103 // Sent to acknowledge a NOTIFY frame
)
type Frame struct {
	Length uint32
    FrameType uint8
    Frame struct {
		Metadata uint64 `type:"varint"`
	}
}
type FrameMetadata struct {
	Flags    [4]byte
	StreamId uint64 `type:"varint"`
	FrameId  uint64 `type:"varint"`
}

