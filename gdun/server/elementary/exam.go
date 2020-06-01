package elementary

import (
	"github.com/wangmhgo/go-project/gdun/common"
	"net/http"
	"strconv"
	"strings"
)

//----------------------------------------------------------------------------
// For meeting service.

func (sv *Server) onHttpAddExamToMeeting(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForTeacher(w, r)
	if err != nil {
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	examID, err := strconv.Atoi(r.FormValue(common.FIELD_EXAM))
	if err != nil {
		examID = 0
	}

	gdExamID, err := strconv.Atoi(r.FormValue(common.FIELD_GAODUN_EXAM_ID))
	if err != nil {
		gdExamID = 0
	}
	if (examID == 0) && (gdExamID == 0) {
		sv.Send(w, r, -2, "Invalid exam ID or Gd exam ID.", "")
		return
	}

	name := r.FormValue(common.FIELD_NAME)

	startTime, err := strconv.Atoi(r.FormValue(common.FIELD_START_TIME))
	if err != nil {
		startTime = 0
	}
	duration, err := strconv.Atoi(r.FormValue(common.FIELD_DURATION))
	if err != nil {
		duration = 0
	}

	preparation, err := strconv.Atoi(r.FormValue(common.FIELD_PREPARATION))
	if err != nil {
		preparation = 1
	}

	s := r.FormValue(common.FIELD_NECESSARY)
	isNecessary := (len(s) == 0) || (s == "1")

	groupID, err := strconv.Atoi(r.FormValue(common.FIELD_GROUP))
	if err != nil {
		groupID = 0
	}

	err = sv.ms.AddExam(meetingID, examID, gdExamID, name, startTime, duration, preparation, isNecessary, groupID, session)
	if err != nil {
		sv.Send(w, r, -7, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

func (sv *Server) onHttpResyncExam(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForTeacher(w, r)
	if err != nil {
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	examID, err := strconv.Atoi(r.FormValue(common.FIELD_EXAM))
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	err = sv.ms.ResyncExam(meetingID, examID, session)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpDeleteExamFromMeeting(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	examID, err := strconv.Atoi(r.FormValue(common.FIELD_EXAM))
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	err = sv.ms.DeleteExam(meetingID, examID, session)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpAnswerMeetingExam(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForStudent(w, r)
	if err != nil {
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	examID, err := strconv.Atoi(r.FormValue(common.FIELD_EXAM))
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	answer := common.Prune(r.FormValue(common.FIELD_ANSWER))

	err = sv.ms.AnswerExam(meetingID, examID, answer, session)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpAnswerMeetingExamQuestion(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForStudent(w, r)
	if err != nil {
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	examID, err := strconv.Atoi(r.FormValue(common.FIELD_EXAM))
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	questionID, err := strconv.Atoi(r.FormValue(common.FIELD_QUESTION))
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	answer := strings.TrimSpace(r.FormValue(common.FIELD_ANSWER))
	if len(answer) == 0 {
		sv.Send(w, r, -4, "Empty question answer.", "")
		return
	}

	err = sv.ms.AnswerExamQuestion(meetingID, examID, questionID, answer, session)
	if err != nil {
		sv.Send(w, r, -5, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpGetMeetingExamAnswer(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForUser(w, r)
	if err != nil {
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, "Invalid meeting ID.", "")
		return
	}

	examID, err := strconv.Atoi(r.FormValue(common.FIELD_EXAM))
	if err != nil {
		sv.Send(w, r, -2, "Invalid exam ID.", "")
		return
	}

	result := ""
	if session.IsStudent() {
		result, err = sv.ms.GetMyExamResult(meetingID, examID, session)
		if err != nil {
			sv.Send(w, r, -3, err.Error(), "")
			return
		}
	} else {
		result, err = sv.ms.GetExamResults(meetingID, examID, session)
		if err != nil {
			sv.Send(w, r, -4, err.Error(), "")
			return
		}
	}

	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpAuhtorizeMeetingExam(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForUser(w, r)
	if err != nil {
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	examID, err := strconv.Atoi(r.FormValue(common.FIELD_EXAM))
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	ei, err := sv.ms.AuthorizeExam(meetingID, examID, session)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	result := `"` + common.FIELD_KEY + `":"` + common.UnescapeForJSON(ei.Key) + `","` + common.FIELD_IV + `":"` + common.UnescapeForJSON(ei.IV) + `"`
	if session.IsTeacher() {
		result += `,"` + common.FIELD_ANSWER + `":"` + common.UnescapeForJSON(ei.Answer) + `"`
	}
	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpAddQuestionToCollection(w http.ResponseWriter, r *http.Request) {
	_, err := sv.ss.CheckHttpSessionForStudent(w, r)
	if err != nil {
		return
	}

	// classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	// if err != nil {
	// 	sv.Send(w, r, -1, common.S_INVALID_CLASS, "")
	// 	return
	// }

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpDeleteQuestionFromCollection(w http.ResponseWriter, r *http.Request) {
	_, err := sv.ss.CheckHttpSessionForStudent(w, r)
	if err != nil {
		return
	}

	// classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	// if err != nil {
	// 	sv.Send(w, r, -1, common.S_INVALID_CLASS, "")
	// 	return
	// }

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------
