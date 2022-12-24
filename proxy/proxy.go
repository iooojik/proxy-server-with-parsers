package proxy

import (
	"crypto/tls"
	"fmt"
	"github.com/iooojik-dev/proxy/provider"
	"github.com/iooojik/go-logger"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"sync/atomic"
)

// Proxy defines parameters for running an HTTP  It implements
// http.Handler interface for ListenAndServe function. If you need, you must
// set Proxy struct before handling requests.
type Proxy struct {
	// Session number of last proxy request.
	SessionNo int64

	// Certificate key pair.
	Ca tls.Certificate

	// User data to use free.
	UserData interface{}

	// Error callback.
	OnError func(ctx *Context, where string, err error, opErr error)

	// Accept callback. It greets proxy request like ServeHTTP function of
	// http.Handler.
	// If it returns true, stops processing proxy request.
	OnAccept func(ctx *Context, w http.ResponseWriter, r *http.Request) bool

	// Auth callback. If you need authentication, set this callback.
	// If it returns true, authentication succeeded.
	OnAuth func(ctx *Context, authType string, user string, pass string) bool

	// Connect callback. It sets connect action and new host.
	// If len(newhost) > 0, host changes.
	OnConnect func(ctx *Context, host string) (ConnectAction ConnectAction,
		newHost string)

	// Request callback. It greets remote request.
	// If it returns non-nil response, stops processing remote request.
	OnRequest func(ctx *Context, req *http.Request) (resp *http.Response)

	// Response callback. It greets remote response.
	// Remote response sends after this callback.
	OnResponse func(ctx *Context, req *http.Request, resp *http.Response)

	// If ConnectAction is ConnectMitm, it sets chunked to Transfer-Encoding.
	// By default, true.
	MitmChunked bool

	// HTTP Authentication type. If it's not specified (""), uses "Basic".
	// By default, "".
	AuthType string

	signer *CaSigner
}

func OnError(_ *Context, where string, err error, opErr error) {
	// Log errors.
	logger.LogError(errors.New(fmt.Sprintf("%s: %s [%s]", where, err, opErr)))
}

func OnAccept(_ *Context, w http.ResponseWriter,
	r *http.Request) bool {
	// Handle local request has path "/info"
	if r.Method == http.MethodGet && !r.URL.IsAbs() && r.URL.Path == "/info" {
		_, err := w.Write([]byte("This is go-"))
		if err != nil {
			logger.LogError(err)
			return false
		}
		return true
	}
	return false
}

func OnAuth(_ *Context, _ string, user string, pass string) bool {
	// Auth test user. todo
	if user == "test" && pass == "1234" {
		return true
	}
	return false
}

func OnConnect(_ *Context, host string) (ConnectAction ConnectAction, newHost string) {
	// Apply "Man in the Middle" to all ssl connections. Never change host.
	return ConnectMitm, host
}

func OnRequest(_ *Context, req *http.Request) (resp *http.Response) {
	// Log proxying requests.
	logger.LogPositive(fmt.Sprintf("Proxy: %s %s", req.Method, req.URL.String()))
	return
}

func OnResponse(ctx *Context, req *http.Request, resp *http.Response) {
	// Add header "Via: go-httpproxy".
	resp.Header.Add("Via", "go-httpproxy")
}

// readCertFiles получает данные о сертфикате и ключе из файлов server.crt и server.key
func readCertFiles() ([]byte, []byte, error) {
	logger.LogInfo("initializing certificate and key")
	certContent, errCrt := ioutil.ReadFile("server.crt")
	if errCrt != nil {
		return nil, nil, errCrt
	}
	keyContent, errKey := ioutil.ReadFile("server.key")
	if errKey != nil {
		return nil, nil, errKey
	}
	logger.LogPositive("certificate and key successfully initialized")
	return certContent, keyContent, nil
}

// RunHttpsProxy создает новый объект прокси-сервера
func RunHttpsProxy() (*Proxy, error) {
	logger.LogInfo("creating proxy server")
	prx := &Proxy{
		MitmChunked: true,
		signer:      NewCaSignerCache(1024),
	}
	prx.OnError = OnError
	prx.OnAccept = OnAccept
	prx.OnAuth = OnAuth
	prx.OnConnect = OnConnect
	prx.OnRequest = OnRequest
	prx.OnResponse = OnResponse
	prx.signer.Ca = &prx.Ca
	var err error
	cert, key, err := readCertFiles()
	if err != nil {
		return nil, err
	}
	prx.Ca, err = tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}
	logger.LogPositive("proxy server was created")
	go func() {
		log.Fatal(http.ListenAndServe(":80", prx))
	}()
	log.Fatal(http.ListenAndServeTLS(":443", "server.crt", "server.key", prx))
	return prx, nil
}

// ServeHTTP implements http.Handler
func (prx *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p, err := provider.NewProxyProvider()
	if err != nil {
		logger.LogError(err)
		return
	}
	ctx := &Context{Prx: prx, SessionNo: atomic.AddInt64(&prx.SessionNo, 1), proxyProvider: p}
	defer func() {
		rec := recover()
		if rec != nil {
			if err, ok := rec.(error); ok && prx.OnError != nil {
				prx.OnError(ctx, "ServeHTTP", ErrPanic, err)
			}
			panic(rec)
		}
	}()
	// проверка подключения (проверка версии протокола, выполнение кастомного doAccept...)
	if ctx.doAccept(w, r) {
		return
	}
	// авторизация запроса
	if ctx.doAuth(w, r) {
		return
	}
	// удаляем из запроса прокси-заголовки
	r.Header.Del("Proxy-Connection")
	r.Header.Del("Proxy-Authenticate")
	r.Header.Del("Proxy-Authorization")

	// выполняем подключение
	if b := ctx.doConnect(w, r); b {
		return
	}

	for {
		var w2 = w
		var r2 = r
		var cyclic = false
		switch ctx.ConnectAction {
		case ConnectMitm:
			if prx.MitmChunked {
				cyclic = true
			}
			w2, r2 = ctx.doMitm()
		}
		if w2 == nil || r2 == nil {
			break
		}
		//r.Header.Del("Accept-Encoding")
		//r.Header.Del("Connection")
		ctx.SubSessionNo++
		if b, err := ctx.doRequest(w2, r2); err != nil {
			logger.LogError(err)
		} else {
			if b {
				if !cyclic {
					break
				} else {
					continue
				}
			}
		}
		if err := ctx.doResponse(w2, r2); err != nil || !cyclic {
			if err != nil {
				logger.LogError(err)
			}
		}
	}

	if ctx.hijTLSConn != nil {
		err := ctx.hijTLSConn.Close()
		if err != nil {
			logger.LogError(err)
		}
	}
}
