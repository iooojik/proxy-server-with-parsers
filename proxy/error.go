package proxy

import (
	"github.com/iooojik/go-logger"
	"github.com/pkg/errors"
	"io"
	"net"
	"os"
	"syscall"
)

// Кастомные ошибки
var (
	ErrPanic                       = logger.MakeError(nil, errors.New("panic"))
	ErrResponseWrite               = logger.MakeError(nil, errors.New("response write"))
	ErrRequestRead                 = logger.MakeError(nil, errors.New("request read"))
	ErrRemoteConnect               = logger.MakeError(nil, errors.New("remote connect"))
	ErrNotSupportHijacking         = logger.MakeError(nil, errors.New("hijacking not supported"))
	ErrTLSSignHost                 = logger.MakeError(nil, errors.New("TLS sign host"))
	ErrTLSHandshake                = logger.MakeError(nil, errors.New("TLS handshake"))
	ErrAbsURLAfterCONNECT          = logger.MakeError(nil, errors.New("absolute URL after CONNECT"))
	ErrRoundTrip                   = logger.MakeError(nil, errors.New("round trip"))
	ErrUnsupportedTransferEncoding = logger.MakeError(nil, errors.New("unsupported transfer encoding"))
	ErrNotSupportHTTPVer           = logger.MakeError(nil, errors.New("http version not supported"))
)

func isConnectionClosed(err error) bool {
	if err == nil {
		return false
	}
	if err == io.EOF {
		return true
	}
	i := 0
	var newerr = &err
	for opError, ok := (*newerr).(*net.OpError); ok && i < 10; {
		i++
		newerr = &opError.Err
		if syscallError, ok := (*newerr).(*os.SyscallError); ok {
			if syscallError.Err == syscall.EPIPE || syscallError.Err == syscall.ECONNRESET || syscallError.Err == syscall.EPROTOTYPE {
				return true
			}
		}
	}
	return false
}
