package proxy

// ConnectAction specifies action of after the CONNECT.
type ConnectAction int

// Constants of ConnectAction type.
const (
	// ConnectNone specifies that proxy request is not CONNECT.
	// If it returned in OnConnect, proxy connection closes immediately.
	ConnectNone = ConnectAction(iota)

	// ConnectProxy specifies directly socket proxy after the CONNECT.
	ConnectProxy

	// ConnectMitm specifies proxy "Man in the Middle" style attack
	// after the CONNECT.
	ConnectMitm
)
