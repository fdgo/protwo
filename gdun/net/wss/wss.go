package wss

import (
	"fmt"
	"github.com/wangmhgo/go-project/gdun/log"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
)

//----------------------------------------------------------------------------

type HttpResource struct {
	Type    string
	Content []byte
}

type WebSocketServer struct {
	addr          string
	tlsAddr       string
	hostname      string
	checkHostname bool
	root          string
	resHandlers   map[string]func(http.ResponseWriter, *http.Request)
	cache         map[string]*HttpResource
	cacheLock     *sync.RWMutex
	logDir        string
	missingLog    *log.Logger
}

//----------------------------------------------------------------------------

func NewWebSocketServer(addr string, tlsAddr string, hostname string, root string, cached bool, logDir string) *WebSocketServer {
	sv := new(WebSocketServer)
	sv.root = root
	if len(sv.root) == 0 {
		sv.root = "htdoc"
	}
	sv.logDir = logDir
	if len(sv.logDir) == 0 {
		sv.logDir = "log"
	}

	sv.addr = addr
	if len(sv.addr) == 0 {
		sv.addr = ":80"
	}
	sv.tlsAddr = tlsAddr
	if len(sv.tlsAddr) == 0 {
		sv.tlsAddr = ":443"
	}
	sv.hostname = hostname
	if len(sv.hostname) > 0 {
		sv.checkHostname = true
	} else {
		sv.checkHostname = false
	}

	sv.resHandlers = make(map[string]func(http.ResponseWriter, *http.Request))

	if cached {
		sv.cache = make(map[string]*HttpResource)
		sv.cacheLock = new(sync.RWMutex)
	} else {
		sv.cache = nil
		sv.cacheLock = nil
	}

	sv.missingLog = log.GetLogger(sv.logDir, "missing", log.LEVEL_INFO)

	return sv
}

func NewWebSocketServerA(cfg *Config) (*WebSocketServer, error) {
	sv := new(WebSocketServer)

	sv.addr = cfg.Addr
	sv.tlsAddr = cfg.TLSAddr

	if cfg.Cached {
		sv.cache = make(map[string]*HttpResource)
		sv.cacheLock = new(sync.RWMutex)
	} else {
		sv.cache = nil
		sv.cacheLock = nil
	}

	sv.hostname = cfg.Hostname
	if len(sv.hostname) > 0 {
		sv.checkHostname = true
	} else {
		sv.checkHostname = false
	}

	sv.root = cfg.HtDocDir
	sv.resHandlers = make(map[string]func(http.ResponseWriter, *http.Request))

	sv.logDir = cfg.LogDir
	sv.missingLog = log.GetLogger(sv.logDir, "missing", log.LEVEL_INFO)

	return sv, nil
}

//----------------------------------------------------------------------------

func (sv *WebSocketServer) HandleHttp(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc(pattern, handler)
}

//----------------------------------------------------------------------------

func (sv *WebSocketServer) HandleWebSocket(pattern string, onOpen func(*websocket.Conn), onRead func(int, []byte, *websocket.Conn) bool, onError func(error, *websocket.Conn) bool) bool {
	if sv == nil {
		return false
	}

	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}

		// Upgrade this HTTP connection.
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()

		onOpen(c)

		for {
			// Read a message.
			mt, message, err := c.ReadMessage()
			if err != nil {
				if !onError(err, c) {
					break
				}
			}

			if !onRead(mt, message, c) {
				break
			}
		}
	})

	return true
}

//----------------------------------------------------------------------------

func (sv *WebSocketServer) HandleResource(suffix string, handler func(http.ResponseWriter, *http.Request)) {
	sv.resHandlers[suffix] = handler
}

//----------------------------------------------------------------------------

func (sv *WebSocketServer) Start() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			sv.go404(w, r)
			return
		}

		host := strings.Split(r.Host, ":")
		if sv.checkHostname {
			if host[0] != sv.hostname {
				sv.go404(w, r)
				return
			}
		}

		// Get the request URI.
		uri := strings.Split(r.RequestURI, "?")
		if strings.HasSuffix(uri[0], "/") {
			uri[0] += "index.html"
		}

		//------------------------------------------------

		// Get it from cache.
		if sv.cache != nil {
			sv.cacheLock.RLock()
			hr, okay := sv.cache[uri[0]]
			sv.cacheLock.RUnlock()

			if okay {
				w.Header().Set("Content-type", hr.Type)
				w.Write(hr.Content)
				return
			}
		}

		//------------------------------------------------

		// Get its suffix.
		n := len(uri[0])
		i := n - 1
		for i >= 0 {
			if uri[0][i] == '.' {
				break
			} else {
				i--
			}
		}
		if i < 0 {
			sv.go404(w, r)
			return
		}

		// Compute content type.
		contentType := ""
		suffix := uri[0][i+1 : n]
		switch suffix {
		case "html":
			contentType = "text/html"
		case "xml":
			contentType = "text/xml"
		case "js":
			contentType = "application/x-javascript"
		case "css":
			contentType = "text/css"
		case "png":
			contentType = "image/png"
		case "ico":
			contentType = "image/x-icon"
		case "gif":
			contentType = "image/gif"
		case "swf":
			contentType = "application/x-shockwave-flash"
		case "ttf":
			contentType = "application/vnd.ms-fontobject"
		case "txt":
			contentType = "text/plain"
		case "woff":
			contentType = "application/font-woff"
		case "woff2":
			contentType = "application/font-woff"
		case "zip":
			contentType = "application/zip"
		default:
			handler, okay := sv.resHandlers[suffix]
			if okay {
				handler(w, r)
			} else {
				sv.go404(w, r)
			}
			return
		}

		//------------------------------------------------

		// Get its content via file system.
		fin, err := os.Open(sv.root + uri[0])
		if err != nil {
			sv.go404(w, r)
			return
		}
		defer fin.Close()

		// Load its content.
		buf, err := ioutil.ReadAll(fin)
		if err != nil {
			sv.go404(w, r)
			return
		}

		//------------------------------------------------

		// Record it.
		if sv.cache != nil {
			hr := new(HttpResource)
			hr.Content = buf
			hr.Type = contentType

			sv.cacheLock.Lock()
			sv.cache[uri[0]] = hr
			sv.cacheLock.Unlock()
		}

		w.Header().Set("Content-type", contentType)
		w.Write(buf)
	})

	go (func() {
		fi, err := os.Stat("ca/key")
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		if !fi.Mode().IsRegular() {
			return
		}

		fi, err = os.Stat("ca/crt")
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		if !fi.Mode().IsRegular() {
			return
		}

		if err = http.ListenAndServeTLS(sv.tlsAddr, "ca/crt", "ca/key", nil); err != nil {
			fmt.Println(err.Error())
			return
		}
	})()

	if err := http.ListenAndServe(sv.addr, nil); err != nil {
		fmt.Println(err.Error())
	}
}

//----------------------------------------------------------------------------

func (sv *WebSocketServer) go404(w http.ResponseWriter, r *http.Request) {
	if sv.missingLog != nil {
		go sv.missingLog.Info(r.RemoteAddr + " (" + r.Host + ")" + r.RequestURI)
	}

	w.WriteHeader(404)
}

//----------------------------------------------------------------------------
