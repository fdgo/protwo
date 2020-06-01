package booking

import (
	"bytes"
	"compress/gzip"
	"github.com/wangmhgo/go-project/gdun/common"
	"github.com/wangmhgo/go-project/gdun/log"
	"github.com/wangmhgo/go-project/gdun/service"
	"net/http"
	"strconv"
	"strings"
)

//----------------------------------------------------------------------------

type Server struct {
	db        *common.Database
	cache     *common.Cache
	ss        *service.SessionService
	urlPrefix string
	accessLog *log.Logger
}

func NewServer(dataSourceName string, cacheServers string) (*Server, error) {
	db, err := common.NewDatabase(dataSourceName, 10)
	if err != nil {
		return nil, err
	}

	cache, err := common.NewCache(cacheServers)
	if err != nil {
		return nil, err
	}

	sv := new(Server)
	sv.db = db
	sv.cache = cache
	sv.urlPrefix = ""

	sv.accessLog = log.GetLogger("log", "access", log.LEVEL_INFO)

	sv.ss = service.NewSessionService(cache, sv.accessLog)
	sv.registerHttpHandles()

	return sv, nil
}

//----------------------------------------------------------------------------

func (sv *Server) createLogLine(r *http.Request) string {
	s := r.RemoteAddr + " " + r.Host + r.RequestURI + " (" + r.UserAgent() + ") (" + r.Referer() + ")"
	if r.TLS != nil {
		s += " TLS"
	}

	return s
}

//----------------------------------------------------------------------------

func (sv *Server) Send(w http.ResponseWriter, r *http.Request, status int, info string, result string) (int, error) {
	if sv.accessLog != nil {
		s := r.RemoteAddr + " " + r.Host + r.RequestURI
		if status == 0 {
			go sv.accessLog.Info(s)
		} else {
			go sv.accessLog.Warning(s + " " + strconv.Itoa(status) + " (" + info + ") (" + r.UserAgent() + ") (" + r.Referer() + ")")
		}
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	body := common.PlainJSONMessage(status, info, result)

	if len(body) > 10240 {
		ae := r.Header.Get("Accept-Encoding")
		if strings.Index(ae, "gzip") >= 0 {
			var b bytes.Buffer
			gz := gzip.NewWriter(&b)

			if _, err := gz.Write(body); err == nil {
				if err := gz.Flush(); err == nil {
					if err := gz.Close(); err == nil {
						w.Header().Set("Content-Encoding", "gzip")
						return w.Write(b.Bytes())
					}
				}
			}
		}
	}

	return w.Write(body)
}

//----------------------------------------------------------------------------

func (sv *Server) registerHttpHandles() {
	http.HandleFunc(sv.urlPrefix+"/test", sv.onHttpTest)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpTest(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForStudent(w, r)
	if err != nil {
		return
	}

	id, err := strconv.Atoi(r.FormValue(common.FIELD_ID))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	if session.UserID != id {
		sv.Send(w, r, -2, "Error hint.", "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------
