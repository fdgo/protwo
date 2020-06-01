package elementary

import (
	"github.com/wangmhgo/go-project/gdun/common"
	"net/http"
	"strings"
)

//----------------------------------------------------------------------------

func (sv *Server) onHttpGetInternalVideo(w http.ResponseWriter, r *http.Request) {
	// Check requirements.
	if sv.cache == nil {
		sv.Send(w, r, -1, common.S_NO_SERVICE, "")
		return
	}

	// Check client IP address.
	ip := common.Prune((strings.Split(r.RemoteAddr, ":"))[0])
	okay, err := sv.cache.FieldExist("bokecc:internal:ip", ip)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}
	if !okay {
		w.WriteHeader(404)
		return
	}

	m, err := sv.cache.GetAllFields("vod:internal:id")
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	result := `"` + common.FIELD_LIST + `":[`
	first := true
	for id, remark := range m {
		if first {
			first = false
		} else {
			result += `,`
		}

		result += `{"` + common.FIELD_ID + `":"` + common.UnescapeForJSON(id) + `","` + common.FIELD_REMARK + `":"` + common.UnescapeForJSON(remark) + `"}`
	}
	result += `]`

	sv.Send(w, r, 0, "", result)
}

func (sv *Server) onHttpAddInternalVideo(w http.ResponseWriter, r *http.Request) {
	// Check requirements.
	if sv.cache == nil {
		sv.Send(w, r, -1, common.S_NO_SERVICE, "")
		return
	}

	// Check client IP address.
	ip := common.Prune((strings.Split(r.RemoteAddr, ":"))[0])
	okay, err := sv.cache.FieldExist("bokecc:internal:ip", ip)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}
	if !okay {
		w.WriteHeader(404)
		return
	}

	// Get parameters.
	err = r.ParseForm()
	if err != nil {
		sv.Send(w, r, -3, "Invalid HTTP form.", "")
		return
	}
	id := common.Prune(r.FormValue(common.FIELD_ID))
	if len(id) == 0 {
		sv.Send(w, r, -4, "Invalid video ID.", "")
		return
	}
	remark := common.Prune(r.FormValue(common.FIELD_REMARK))
	if len(remark) == 0 {
		sv.Send(w, r, -5, "Invalid remark.", "")
		return
	}

	err = sv.cache.SetField("vod:internal:id", id, common.Escape(remark))
	if err != nil {
		sv.Send(w, r, -6, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

func (sv *Server) onHttpDeleteInternalVideo(w http.ResponseWriter, r *http.Request) {
	// Check requirements.
	if sv.cache == nil {
		sv.Send(w, r, -1, common.S_NO_SERVICE, "")
		return
	}

	// Check client IP address.
	ip := common.Prune((strings.Split(r.RemoteAddr, ":"))[0])
	okay, err := sv.cache.FieldExist("bokecc:internal:ip", ip)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}
	if !okay {
		w.WriteHeader(404)
		return
	}

	// Get parameters.
	err = r.ParseForm()
	if err != nil {
		sv.Send(w, r, -3, "Invalid HTTP form.", "")
		return
	}
	id := common.Prune(r.FormValue(common.FIELD_ID))
	if len(ip) == 0 {
		sv.Send(w, r, -4, "Invalid video ID.", "")
		return
	}

	err = sv.cache.DelField("vod:internal:id", id)
	if err != nil {
		sv.Send(w, r, -5, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpGetInternalIP(w http.ResponseWriter, r *http.Request) {
	// Check requirements.
	if sv.cache == nil {
		sv.Send(w, r, -1, common.S_NO_SERVICE, "")
		return
	}

	// Get client IP address.
	ip := common.Prune((strings.Split(r.RemoteAddr, ":"))[0])

	m, err := sv.cache.GetAllFields("bokecc:internal:ip")
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	// Check authority.
	okay := sv.isLanIP(ip)
	if !okay {
		_, okay = m[ip]
	}
	if !okay {
		w.WriteHeader(404)
		return
	}

	result := `"` + common.FIELD_LIST + `":[`
	first := true
	for ip, remark := range m {
		if first {
			first = false
		} else {
			result += `,`
		}

		result += `{"` + common.FIELD_IP + `":"` + ip + `","` + common.FIELD_REMARK + `":"` + common.UnescapeForJSON(remark) + `"}`
	}
	result += `]`

	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpAddInternalIP(w http.ResponseWriter, r *http.Request) {
	// Check requirements.
	if sv.cache == nil {
		sv.Send(w, r, -1, common.S_NO_SERVICE, "")
		return
	}

	// Check client IP address.
	ip := common.Prune((strings.Split(r.RemoteAddr, ":"))[0])
	okay, err := sv.cache.FieldExist("bokecc:internal:ip", ip)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}
	if !okay {
		w.WriteHeader(404)
		return
	}

	// Get parameters.
	err = r.ParseForm()
	if err != nil {
		sv.Send(w, r, -3, "Invalid HTTP form.", "")
		return
	}
	ip = common.Prune(r.FormValue(common.FIELD_IP))
	if len(ip) == 0 {
		sv.Send(w, r, -4, "Invalid IP.", "")
		return
	}
	remark := common.Prune(r.FormValue(common.FIELD_REMARK))
	if len(remark) == 0 {
		sv.Send(w, r, -5, "Invalid remark.", "")
		return
	}

	err = sv.cache.SetField("bokecc:internal:ip", ip, common.Escape(remark))
	if err != nil {
		sv.Send(w, r, -6, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpDeleteInternalIP(w http.ResponseWriter, r *http.Request) {
	// Check requirements.
	if sv.cache == nil {
		sv.Send(w, r, -1, common.S_NO_SERVICE, "")
		return
	}

	// Check client IP address.
	ip := common.Prune((strings.Split(r.RemoteAddr, ":"))[0])
	okay, err := sv.cache.FieldExist("bokecc:internal:ip", ip)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}
	if !okay {
		w.WriteHeader(404)
		return
	}

	// Get parameters.
	err = r.ParseForm()
	if err != nil {
		sv.Send(w, r, -3, "Invalid HTTP form.", "")
		return
	}
	ip = common.Prune(r.FormValue(common.FIELD_IP))
	if len(ip) == 0 {
		sv.Send(w, r, -4, "Invalid IP.", "")
		return
	}

	err = sv.cache.DelField("bokecc:internal:ip", ip)
	if err != nil {
		sv.Send(w, r, -5, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------
