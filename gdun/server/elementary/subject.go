package elementary

import (
	"github.com/wangmhgo/go-project/gdun/common"
	"net/http"
	"strconv"
)

//----------------------------------------------------------------------------

func (sv *Server) onHttpAddSubject(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForSystem(w, r)
	if err != nil {
		return
	}

	name := common.Prune(r.FormValue(common.FIELD_NAME))
	if len(name) == 0 {
		sv.Send(w, r, -1, common.S_INVALID_NAME, "")
		return
	}

	id, err := sv.sbjs.AddSubject(name, session)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	result := `"` + common.FIELD_ID + `":` + strconv.Itoa(id)
	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpChangeSubject(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForSystem(w, r)
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

	err = sv.sbjs.ChangeSubject(id, name, session)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpQuerySubject(w http.ResponseWriter, r *http.Request) {
	_, err := sv.ss.CheckHttpSessionForSystem(w, r)
	if err != nil {
		return
	}

	result, err := sv.sbjs.GetSubject()
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpAddSubjectToGroup(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	groupID := session.GroupID
	if session.IsSystem() {
		groupID, err = strconv.Atoi(r.FormValue(common.FIELD_GROUP))
		if err != nil {
			sv.Send(w, r, -1, common.S_INVALID_GROUP, "")
			return
		}
	}

	subjectID, err := strconv.Atoi(r.FormValue(common.FIELD_ID))
	if err != nil {
		sv.Send(w, r, -2, common.S_INVALID_ID, "")
		return
	}

	err = sv.sbjs.ChangeSubjectList(groupID, subjectID, true, session)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpDeleteSubjectFromGroup(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	groupID := session.GroupID
	if session.IsSystem() {
		groupID, err = strconv.Atoi(r.FormValue(common.FIELD_GROUP))
		if err != nil {
			sv.Send(w, r, -1, common.S_INVALID_GROUP, "")
			return
		}
	}

	subjectID, err := strconv.Atoi(r.FormValue(common.FIELD_ID))
	if err != nil {
		sv.Send(w, r, -2, common.S_INVALID_ID, "")
		return
	}

	err = sv.sbjs.ChangeSubjectList(groupID, subjectID, false, session)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpQuerySubjectForGroup(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	groupID := session.GroupID
	if session.IsSystem() {
		groupID, err = strconv.Atoi(r.FormValue(common.FIELD_GROUP))
		if err != nil {
			sv.Send(w, r, -1, common.S_INVALID_GROUP, "")
			return
		}
	}

	result, err := sv.sbjs.GetSubjectList(groupID)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------
