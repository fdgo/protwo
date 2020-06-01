package elementary

import (
	"github.com/wangmhgo/go-project/gdun/common"
	"net/http"
	"strconv"
)

//----------------------------------------------------------------------------

func (sv *Server) onHttpAddTag(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	name := common.Prune(r.FormValue(common.FIELD_NAME))
	if len(name) == 0 {
		sv.Send(w, r, -1, common.S_INVALID_NAME, "")
		return
	}

	groupID := session.GroupID
	if session.IsSystem() {
		groupID, err = strconv.Atoi(r.FormValue(common.FIELD_GROUP))
		if err != nil {
			sv.Send(w, r, -2, common.S_INVALID_GROUP, "")
			return
		}
	}

	id, status, err := sv.ts.AddTag(name, groupID, session)
	if err != nil {
		sv.Send(w, r, status-2, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", `"`+common.FIELD_ID+`":`+strconv.Itoa(id))
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpChangeTag(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	id, err := strconv.Atoi(r.FormValue(common.FIELD_ID))
	if err != nil {
		sv.Send(w, r, -1, common.S_INVALID_ID, "")
		return
	}

	name := common.Prune(r.FormValue(common.FIELD_NAME))
	if len(name) == 0 {
		sv.Send(w, r, -2, common.S_INVALID_NAME, "")
		return
	}

	groupID := session.GroupID
	if session.IsSystem() {
		groupID, err = strconv.Atoi(r.FormValue(common.FIELD_GROUP))
		if err != nil {
			sv.Send(w, r, -3, common.S_INVALID_GROUP, "")
			return
		}
	}

	status, err := sv.ts.ChangeTag(id, name, groupID, session)
	if err != nil {
		sv.Send(w, r, status-3, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpQueryTag(w http.ResponseWriter, r *http.Request) {
	_, err := sv.ss.CheckHttpSessionForUser(w, r)
	if err != nil {
		return
	}

	groupID, err := strconv.Atoi(r.FormValue(common.FIELD_GROUP))
	if err != nil {
		sv.Send(w, r, -1, common.S_INVALID_GROUP, "")
		return
	}

	result, status, err := sv.ts.QueryTags(groupID)
	if err != nil {
		sv.Send(w, r, status-1, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------
