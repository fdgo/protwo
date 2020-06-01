package elementary

import (
	"github.com/wangmhgo/go-project/gdun/common"
	"net/http"
	"strconv"
)

//----------------------------------------------------------------------------

func (sv *Server) onHttpAddGroup(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForSystem(w, r)
	if err != nil {
		return
	}

	err = sv.gs.AddGroup(r.FormValue(common.FIELD_NAME), session)
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpDeleteGroup(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForSystem(w, r)
	if err != nil {
		return
	}

	groupID, err := strconv.Atoi(r.FormValue(common.FIELD_ID))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	err = sv.gs.DeleteGroup(groupID, session)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpQueryGroup(w http.ResponseWriter, r *http.Request) {
	_, err := sv.ss.CheckHttpSessionForUser(w, r)
	if err != nil {
		return
	}

	result, err := sv.gs.QueryGroup()
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------
