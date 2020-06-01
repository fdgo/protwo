package elementary

import (
	"github.com/wangmhgo/go-project/gdun/common"
	"net/http"
	"strconv"
)

//----------------------------------------------------------------------------

func (sv *Server) onHttpGetIssues(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForUser(w, r)
	if err != nil {
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -1, common.S_INVALID_CLASS, "")
		return
	}

	result, err := sv.is.Get(classID, session)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpAddIssue(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForStudent(w, r)
	if err != nil {
		return
	}

	// classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	// if err != nil {
	// 	sv.Send(w, r, -1, common.S_INVALID_CLASS, "")
	// 	return
	// }

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, common.S_INVALID_MEETING, "")
		return
	}

	t, err := strconv.Atoi(r.FormValue(common.FIELD_TYPE))
	if err != nil {
		sv.Send(w, r, -2, common.S_INVALID_TYPE, "")
		return
	}

	key := common.Prune(r.FormValue(common.FIELD_KEY))
	if len(key) == 0 {
		sv.Send(w, r, -3, common.S_INVALID_KEY, "")
		return
	}

	subKey, err := strconv.Atoi(r.FormValue(common.FIELD_SUB_KEY))
	if err != nil {
		sv.Send(w, r, -4, common.S_INVALID_SUB_KEY, "")
		return
	}

	body := common.Prune(r.FormValue(common.FIELD_QUESTION))
	if len(body) == 0 {
		sv.Send(w, r, -5, common.S_INVALID_QUESTION, "")
		return
	}

	issueID, err := sv.is.Ask(meetingID, t, key, subKey, body, session)
	if err != nil {
		sv.Send(w, r, -6, err.Error(), "")
		return
	}

	result := `"` + common.FIELD_ISSUE + `":` + strconv.Itoa(issueID)
	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpAnswerIssue(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForTeacher(w, r)
	if err != nil {
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -1, common.S_INVALID_CLASS, "")
		return
	}

	issueID, err := strconv.Atoi(r.FormValue(common.FIELD_ISSUE))
	if err != nil {
		sv.Send(w, r, -2, common.S_INVALID_ISSUE, "")
		return
	}

	body := common.Prune(r.FormValue(common.FIELD_ANSWER))
	if len(body) == 0 {
		sv.Send(w, r, -3, common.S_INVALID_ANSWER, "")
	}

	err = sv.is.Answer(classID, issueID, body, session)
	if err != nil {
		sv.Send(w, r, -4, err.Error(), "")
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpChangeIssueAnswer(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForTeacher(w, r)
	if err != nil {
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -1, common.S_INVALID_CLASS, "")
		return
	}

	issueID, err := strconv.Atoi(r.FormValue(common.FIELD_ISSUE))
	if err != nil {
		sv.Send(w, r, -2, common.S_INVALID_ISSUE, "")
		return
	}

	body := common.Prune(r.FormValue(common.FIELD_ANSWER))
	if len(body) == 0 {
		sv.Send(w, r, -3, common.S_INVALID_ANSWER, "")
	}

	err = sv.is.Change(classID, issueID, body, session)
	if err != nil {
		sv.Send(w, r, -4, err.Error(), "")
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpGetIssueResource(w http.ResponseWriter, r *http.Request) {
	if !sv.isLanIP(r.RemoteAddr) {
		return
	}

	err := r.ParseForm()
	if err != nil {
		sv.Send(w, r, -1, "Invalid form.", "")
		return
	}

	groupID, err := strconv.Atoi(r.FormValue(common.FIELD_GROUP))
	if err != nil {
		sv.Send(w, r, -2, "Invalid group ID.", "")
		return
	}

	issueID, err := strconv.Atoi(r.FormValue(common.FIELD_ISSUE))
	if err != nil {
		sv.Send(w, r, -3, "Invalid issue ID.", "")
		return
	}

	// result, err := sv.cache.GetKey(sv.is.GetIssueResourceKey(groupID, issueID))
	result, err := sv.is.GetIssueResource(groupID, issueID)
	if err != nil {
		sv.Send(w, r, -4, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpGetIssueQuestion(w http.ResponseWriter, r *http.Request) {
	if !sv.isLanIP(r.RemoteAddr) {
		return
	}

	err := r.ParseForm()
	if err != nil {
		sv.Send(w, r, -1, "Invalid form.", "")
		return
	}

	groupID, err := strconv.Atoi(r.FormValue(common.FIELD_GROUP))
	if err != nil {
		sv.Send(w, r, -2, "Invalid group ID.", "")
		return
	}

	issueID, err := strconv.Atoi(r.FormValue(common.FIELD_ISSUE))
	if err != nil {
		sv.Send(w, r, -3, "Invalid issue ID.", "")
		return
	}

	// result, err := sv.cache.GetKey(sv.is.GetIssueResourceKey(groupID, issueID))
	result, err := sv.is.GetIssueQuestion(groupID, issueID)
	if err != nil {
		sv.Send(w, r, -4, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", `"`+common.FIELD_BODY+`":"`+common.ReplaceForJSON(result)+`"`)
}

//----------------------------------------------------------------------------
