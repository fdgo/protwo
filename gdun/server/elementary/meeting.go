package elementary

import (
	"fmt"
	"github.com/wangmhgo/go-project/gdun/common"
	// "gitlab.hfjy.com/gdun/server/live2"
	"net/http"
	"strconv"
)

//----------------------------------------------------------------------------

func (sv *Server) onHttpGetMeeting(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForUser(w, r)
	if err != nil {
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, common.S_INVALID_MEETING, "")
		return
	}

	mi, err := sv.ms.GetMeeting(meetingID, session, true)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", mi.ToJSON(session.IsTeacherOrAbove()))
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpAddMeeting(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	name := r.FormValue(common.FIELD_NAME)
	if len(name) == 0 {
		sv.Send(w, r, -1, "Empty meeting name.", "")
		return
	}

	subjects := common.StringToIntArray(r.FormValue(common.FIELD_SUBJECT))

	section, err := strconv.Atoi(r.FormValue(common.FIELD_SECTION))
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	startTime, err := strconv.Atoi(r.FormValue(common.FIELD_START_TIME))
	if err != nil {
		sv.Send(w, r, -4, err.Error(), "")
		return
	}

	duration, err := strconv.Atoi(r.FormValue(common.FIELD_DURATION))
	if err != nil {
		sv.Send(w, r, -5, err.Error(), "")
		return
	}

	t, err := strconv.Atoi(r.FormValue(common.FIELD_TYPE))
	if err != nil {
		t = 0
	}
	tData := r.FormValue(common.FIELD_DATA)

	_, err = sv.cs.AddMeeting(name, subjects, section, startTime, duration, classID, t, tData, session, false)
	if err != nil {
		sv.Send(w, r, -6, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpChangeMeeting(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	//----------------------------------------------------

	// Get class ID.
	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -1, common.S_INVALID_CLASS, "")
		return
	}

	// Check authority.
	ci, err := sv.cs.GetClass(classID, session)
	if err != nil {
		sv.Send(w, r, -2, common.S_NO_AUTHORITY, "")
		return
	}

	// Get meeting ID.
	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -3, common.S_INVALID_MEETING, "")
		return
	}

	// Check whether the meeting resides in the class.
	okay := false
	for i := 0; i < len(ci.Meetings); i++ {
		if ci.Meetings[i] == meetingID {
			okay = true
			break
		}
	}
	if !okay {
		sv.Send(w, r, -4, common.S_NO_AUTHORITY, "")
		return
	}

	//----------------------------------------------------

	name := common.Prune(r.FormValue(common.FIELD_NAME))
	if len(name) == 0 {
		sv.Send(w, r, -5, common.S_INVALID_NAME, "")
		return
	}

	subjects := common.StringToIntArray(r.FormValue(common.FIELD_SUBJECT))

	section, err := strconv.Atoi(r.FormValue(common.FIELD_SECTION))
	if err != nil {
		sv.Send(w, r, -6, err.Error(), "")
		return
	}

	startTime, err := strconv.Atoi(r.FormValue(common.FIELD_START_TIME))
	if err != nil {
		sv.Send(w, r, -7, err.Error(), "")
		return
	}

	duration, err := strconv.Atoi(r.FormValue(common.FIELD_DURATION))
	if err != nil {
		sv.Send(w, r, -8, err.Error(), "")
		return
	}

	t, err := strconv.Atoi(r.FormValue(common.FIELD_TYPE))
	if err != nil {
		t = 0
	}
	tData := common.Prune(r.FormValue(common.FIELD_DATA))

	status, err := sv.ms.ChangeMeeting(meetingID, name, subjects, section, startTime, duration, t, tData, session)
	if err != nil {
		sv.Send(w, r, status-8, err.Error(), "")
		return
	}

	if status == 1 {
		if err = sv.cs.UpdateMeetingTime(classID, session); err != nil {
			// TODO:
		}
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpSetMeetingAlly(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, common.S_INVALID_MEETING, "")
		return
	}

	ally, err := strconv.Atoi(r.FormValue(common.FIELD_ALLY))
	if err != nil {
		sv.Send(w, r, -2, common.S_INVALID_MEETING, "")
		return
	}

	err = sv.ms.ChangeMeetingAlly(meetingID, ally, session)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpChangeMeetingName(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	name := r.FormValue(common.FIELD_NAME)
	if len(name) == 0 {
		sv.Send(w, r, -2, "Empty meeting name.", "")
		return
	}

	err = sv.ms.ChangeMeetingName(meetingID, name, session)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpChangeMeetingSubject(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	subjects := common.StringToIntArray(r.FormValue(common.FIELD_SUBJECT))

	err = sv.ms.ChangeMeetingSubjects(meetingID, subjects, session)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpChangeMeetingTime(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		classID = 0
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	startTime, err := strconv.Atoi(r.FormValue(common.FIELD_START_TIME))
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	duration, err := strconv.Atoi(r.FormValue(common.FIELD_DURATION))
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	err = sv.ms.ChangeMeetingTime(meetingID, startTime, duration, session)
	if err != nil {
		sv.Send(w, r, -4, err.Error(), "")
		return
	}

	if classID > 0 {
		if err = sv.cs.UpdateMeetingTime(classID, session); err != nil {
			sv.Send(w, r, -5, err.Error(), "")
			return
		}
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpChangeMeetingSection(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, "Invalid meeting ID.", "")
		return
	}

	section, err := strconv.Atoi(r.FormValue(common.FIELD_SECTION))
	if err != nil {
		sv.Send(w, r, -2, "Invalid section.", "")
		return
	}

	err = sv.ms.ChangeMeetingSection(meetingID, section, session)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpSetMeetingConfig(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	if sv.cache == nil {
		sv.Send(w, r, -1, common.S_NO_SERVICE, "")
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -2, common.S_INVALID_MEETING, "")
		return
	}

	cfg := common.Prune(r.FormValue(common.FIELD_CONFIG))
	if len(cfg) == 0 {
		sv.Send(w, r, -3, "Invalid configuration.", "")
		return
	}

	_, err = sv.ms.GetMeeting(meetingID, session, true)
	if err != nil {
		sv.Send(w, r, -4, common.S_NO_AUTHORITY, "")
		return
	}

	err = sv.cache.SetField(common.KEY_PREFIX_MEETING+strconv.Itoa(meetingID), common.FIELD_CONFIG, common.Escape(cfg))
	if err != nil {
		sv.Send(w, r, -5, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

func (sv *Server) onHttpGetMeetingConfig(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		sv.Send(w, r, -1, "Invalid form.", "")
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -2, "Invalid form.", "")
		return
	}

	if sv.cache == nil {
		sv.Send(w, r, -3, common.S_NO_SERVICE, "")
		return
	}

	cfg, err := sv.cache.GetField(common.KEY_PREFIX_MEETING+strconv.Itoa(meetingID), common.FIELD_CONFIG)
	if err != nil {
		sv.Send(w, r, -4, err.Error(), "")
		return
	}

	result := `"` + common.FIELD_CONFIG + `":"` + common.UnescapeForJSON(cfg) + `"`
	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpChangeMeetingType(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, "Invalid meeting ID.", "")
		return
	}

	t, err := strconv.Atoi(r.FormValue(common.FIELD_TYPE))
	if err != nil {
		sv.Send(w, r, -2, "Invalid type.", "")
		return
	}

	data := r.FormValue(common.FIELD_DATA)

	err = sv.ms.ChangeMeetingType(meetingID, t, data, session)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpDeleteMeetingFromClass(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -1, "Invalid class ID.", "")
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -2, "Invalid meeting ID.", "")
		return
	}

	isPermanent := false
	if n, err := strconv.Atoi(r.FormValue(common.FIELD_PERMANENT)); err == nil {
		if n == 1 {
			isPermanent = true
		}
	}

	if isPermanent {
		err = sv.cs.DeleteMeetingPermanently(meetingID, classID, session)
		if err != nil {
			sv.Send(w, r, -3, err.Error(), "")
			return
		}
	} else {
		err = sv.cs.DeleteMeeting(meetingID, classID, session)
		if err != nil {
			sv.Send(w, r, -4, err.Error(), "")
			return
		}
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpEndMeeting(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForTeacher(w, r)
	if err != nil {
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -1, common.S_INVALID_CLASS, "")
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -2, common.S_INVALID_MEETING, "")
		return
	}

	// if classID > 0 {
	err = sv.cs.EndMeeting(classID, meetingID, session)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}
	// } else {
	// 	err = sv.ms.EndMeeting(meetingID, session)
	// 	if err != nil {
	// 		sv.Send(w, r, -4, err.Error(), "")
	// 		return
	// 	}
	// }

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpCopyMeeting(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, "Invalid meeting ID.", "")
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -2, "Invalid class ID.", "")
		return
	}

	if err = sv.cs.CopyMeetingTo(meetingID, classID, session); err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpJoinMeeting(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForUser(w, r)
	if err != nil {
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	result, err := sv.ms.JoinMeeting(meetingID, session)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpLeaveMeeting(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForStudent(w, r)
	if err != nil {
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	cancel, err := strconv.Atoi(r.FormValue(common.FIELD_CANCEL))
	if err != nil {
		cancel = 0
	}
	if cancel > 0 {
		cancel = 1
	}

	err = sv.ms.LeaveMeeting(meetingID, cancel, session)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpScoreMeeting(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForStudent(w, r)
	if err != nil {
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, "Invalid meeting ID.", "")
		return
	}

	arr := common.StringToIntArray(r.FormValue(common.FIELD_SCORE))
	if len(arr) == 0 {
		sv.Send(w, r, -2, "Invalid scores.", "")
		return
	}

	feedback := common.Prune(r.FormValue(common.FIELD_FEEDBACK))

	err = sv.ms.ScoreMeeting(meetingID, arr, feedback, session)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpGetMeetingFeedback(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForTeacher(w, r)
	if err != nil {
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, common.S_INVALID_MEETING, "")
		return
	}

	result, err := sv.ms.GetMeetingFeedback(meetingID, session)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpFinishMeeting(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForStudent(w, r)
	if err != nil {
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, common.S_INVALID_MEETING, "")
		return
	}

	if err := sv.ms.SetMeetingProgress(meetingID, 1, session); err != nil {
		// TODO:
	}

	if sv.cache != nil {
		go (func() {
			result := `"` + common.FIELD_IS_TEACHER + `":false,"` + common.FIELD_NAME + `":"` + common.UnescapeForJSON(session.Nickname) + `"`
			msg := common.PlainCommandJSONMessage(0, "", common.COMMAND_USER, common.COMMAND_USER_JOINED, session.UserID, result)
			gossipMsg := fmt.Sprintf("%08x%08x%08x0%s", 0, meetingID, common.ATTENDEE_GROUP_TEACHER, string(msg))

			if err := sv.cache.Publish(common.CHANNEL_GOSSIP, gossipMsg); err != nil {
				// TODO:
				fmt.Println("onHttpFinishMeeting() " + err.Error())
			}

			// if lSv := live2.GetCurrentServer(); lSv != nil {

			// 	fmt.Println("lSv != nil " + strconv.Itoa(session.UserID))

			// 	lSv.Broadcast(meetingID, common.ATTENDEE_GROUP_TEACHER, string(msg))

			// 	lSv.GetOrCreateMeeting(meetingID, name)
			// } else {
			// 	fmt.Println("lSv == nil " + strconv.Itoa(session.UserID))
			// }
		})()
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpQueryMeeting(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForUser(w, r)
	if err != nil {
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	result, status, err := sv.cs.GetMeetings(classID, session)
	if err != nil {
		sv.Send(w, r, status-1, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", result.ToJSON(session.IsTeacherOrAbove()))
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpQueryMeetingProgress(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForUser(w, r)
	if err != nil {
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, "Invalid meeting ID.", "")
		return
	}

	if session.IsStudent() {
		// result, err := sv.ms.GetCachedUserProgress(meetingID, session.UserID, true)
		umpi, err := sv.ms.GetUserProgress(meetingID, session.UserID)
		if err != nil {
			sv.Send(w, r, -2, err.Error(), "")
			return
		}

		sv.Send(w, r, 0, "", umpi.ToJSON(true))
	} else {
		// mi, err := sv.ms.GetMeeting(meetingID, session, true)
		// if err != nil {
		// 	sv.Send(w, r, -3, err.Error(), "")
		// 	return
		// }

		// ci, err := sv.cs.GetClass(mi.ClassID, session)
		// if err != nil {
		// 	sv.Send(w, r, -4, err.Error(), "")
		// 	return
		// }

		umpi, err := sv.ms.GetMeetingProgresses(meetingID, session)
		// result, err := sv.ms.GetCachedUserProgressesByMeeting(meetingID, ci.Students, true)
		if err != nil {
			sv.Send(w, r, -5, err.Error(), "")
			return
		}

		sv.Send(w, r, 0, "", umpi.ToJSON(true))
	}
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpArrangeMeetingResources(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForTeacher(w, r)
	if err != nil {
		return
	}

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

	s := common.Prune(r.FormValue(common.FIELD_LIST))
	arr := common.StringToStringArray(s)
	if len(arr) == 0 {
		sv.Send(w, r, -3, "Empty resource ID list.", "")
		return
	}

	err = sv.ms.ArrangeResource(meetingID, t, arr, session)
	if err != nil {
		sv.Send(w, r, -4, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpSyncMeeting(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForTeacher(w, r)
	if err != nil {
		return
	}

	from, err := strconv.Atoi(r.FormValue(common.FIELD_FROM))
	if err != nil {
		sv.Send(w, r, -1, "Invalid source meeting ID.", "")
		return
	}

	to := common.StringToIntArray(common.Unescape(r.FormValue(common.FIELD_TO)))
	if len(to) == 0 {
		sv.Send(w, r, -2, "Invalid destination meeting ID.", "")
		return
	}

	data := common.Prune(r.FormValue(common.FIELD_DATA))
	if len(data) != 12 {
		sv.Send(w, r, -3, "Invalid synchronization content.", "")
		return
	}

	err = sv.ms.SyncMeeting(from, to, data, session)
	if err != nil {
		sv.Send(w, r, -4, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpQueryMeetingViaGdCourseID(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForStudent(w, r)
	if err != nil {
		return
	}

	gdCourseID, err := strconv.Atoi(r.FormValue(common.FIELD_GAODUN_COURSE_ID))
	if err != nil {
		sv.Send(w, r, -1, common.S_INVALID_CLASS_ID, "")
		return
	}

	classID, err := sv.cs.GetClassIDViaGdCourseID(gdCourseID)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	// ci, err := sv.cs.GetClass(classID, session)
	// if err != nil {
	// 	sv.Send(w, r, -3, err.Error(), result)
	// }

	mia, status, err := sv.cs.GetMeetings(classID, session)
	if err != nil {
		sv.Send(w, r, status-2, err.Error(), "")
		return
	}

	result := `"` + common.FIELD_USER + `":` + strconv.Itoa(session.UserID) + `,` +
		`"` + common.FIELD_NICKNAME + `":"` + common.UnescapeForJSON(session.Nickname) + `",` +
		`"` + common.FIELD_CLASS + `":` + strconv.Itoa(classID) + `,` +
		mia.ToJSON(false)

	// fmt.Println(result)
	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------

// func (sv *Server) onHttpMeetingAddTeacher(w http.ResponseWriter, r *http.Request) {
// 	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
// 	if err != nil {
// 		return
// 	}

// 	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
// 	if err != nil {
// 		sv.Send(w, r, -1, common.S_INVALID_MEETING, "")
// 		return
// 	}

// 	userID, err := strconv.Atoi(r.FormValue(common.FIELD_USER))
// 	if err != nil {
// 		sv.Send(w, r, -2, common.S_INVALID_USER, "")
// 		return
// 	}

// 	if err = sv.ms.ChangeTeacher(meetingID, userID, true, session); err != nil {
// 		sv.Send(w, r, -3, err.Error(), "")
// 		return
// 	}

// 	sv.Send(w, r, 0, "", "")
// }

//----------------------------------------------------------------------------

// func (sv *Server) onHttpMeetingDeleteTeacher(w http.ResponseWriter, r *http.Request) {
// 	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
// 	if err != nil {
// 		return
// 	}

// 	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
// 	if err != nil {
// 		sv.Send(w, r, -1, common.S_INVALID_MEETING, "")
// 		return
// 	}

// 	userID, err := strconv.Atoi(r.FormValue(common.FIELD_USER))
// 	if err != nil {
// 		sv.Send(w, r, -2, common.S_INVALID_USER, "")
// 		return
// 	}

// 	if err = sv.ms.ChangeTeacher(meetingID, userID, false, session); err != nil {
// 		sv.Send(w, r, -3, err.Error(), "")
// 		return
// 	}

// 	sv.Send(w, r, 0, "", "")
// }

//----------------------------------------------------------------------------

// func (sv *Server) onHttpMeetingNotifyStudent(w http.ResponseWriter, r *http.Request) {
// session, err := sv.ss.CheckHttpSessionForTeacher(w, r)
// if err != nil {
// 	return
// }

// meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
// if err != nil {
// 	sv.Send(w, r, -1, common.S_INVALID_MEETING, "")
// 	return
// }

// TODO:

// sv.Send(w, r, 0, "", "")
// }

//----------------------------------------------------------------------------
