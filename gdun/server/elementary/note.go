package elementary

import (
	"github.com/wangmhgo/go-project/gdun/common"
	"net/http"
	"strconv"
	"strings"
)

//----------------------------------------------------------------------------

func (sv *Server) onHttpAddNote(w http.ResponseWriter, r *http.Request) {
	// Check authority.
	session, err := sv.ss.CheckHttpSessionForStudent(w, r)
	if err != nil {
		return
	}

	// Get inputs.

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

	t, err := strconv.Atoi(r.FormValue(common.FIELD_TYPE))
	if err != nil {
		sv.Send(w, r, -3, "Invalid note type.", "")
		return
	}

	key := strings.TrimSpace(r.FormValue(common.FIELD_KEY))
	if len(key) == 0 {
		sv.Send(w, r, -4, "Empty note key.", "")
		return
	}

	subKey, err := strconv.Atoi(r.FormValue(common.FIELD_SUB_KEY))
	if err != nil {
		sv.Send(w, r, -5, "Invalid sub key of the note.", "")
		return
	}

	body := strings.TrimSpace(r.FormValue(common.FIELD_BODY))
	if len(body) == 0 {
		sv.Send(w, r, -6, "Empty note body.", "")
		return
	}

	// Check authority.

	ci, err := sv.cs.GetClass(classID, session)
	if err != nil {
		sv.Send(w, r, -7, common.S_NO_AUTHORITY, "")
		return
	}

	if t == common.TYPE_FOR_MEETING {
		existing := false
		for i := 0; i < len(ci.Meetings); i++ {
			if strconv.Itoa(ci.Meetings[i]) == key {
				existing = true
				break
			}
		}
		if !existing {
			sv.Send(w, r, -8, common.S_NO_AUTHORITY, "")
			return
		}
	} else {
		mi, err := sv.ms.GetMeeting(meetingID, session, false)
		if err != nil {
			return
		}

		var arr []string
		switch t {
		case common.TYPE_FOR_COURSEWARE:
			arr = mi.Coursewares
		case common.TYPE_FOR_EXAM:
			arr = mi.Exams
		case common.TYPE_FOR_VIDEO:
			arr = mi.Videos
		default:
			sv.Send(w, r, -9, common.S_NO_AUTHORITY, "")
			return
		}

		existing := false
		target := key + ":"
		for i := 0; i < len(arr); i++ {
			if strings.HasPrefix(arr[i], target) {
				existing = true
				break
			}
			// fmt.Println("(" + arr[i] + "):(" + key + ")")
			// if arr[i] == key {
			// existing = true
			// break
			// }
		}
		if !existing {
			sv.Send(w, r, -10, common.S_NO_AUTHORITY, "")
			return
		}
	}

	// Add this note.

	id, err := sv.ns.AddNote(classID, meetingID, t, key, subKey, body, session)
	if err != nil {
		sv.Send(w, r, -11, err.Error(), "")
		return
	}

	// Say everything is fine.
	result := `"` + common.FIELD_ID + `":` + strconv.Itoa(id)
	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpGetNote(w http.ResponseWriter, r *http.Request) {
	// Check authority.
	session, err := sv.ss.CheckHttpSessionForUser(w, r)
	if err != nil {
		return
	}

	// Check inputs.

	s := common.Prune(r.FormValue(common.FIELD_CLASS))
	if len(s) > 0 {
		classID, err := strconv.Atoi(s)
		if err != nil {
			sv.Send(w, r, -1, err.Error(), "")
			return
		}

		result, err := sv.ns.GetMyNote(classID, session)
		if err != nil {
			sv.Send(w, r, -2, err.Error(), "")
			return
		}

		sv.Send(w, r, 0, "", result.ToJSON())

	} else {
		t, err := strconv.Atoi(r.FormValue(common.FIELD_TYPE))
		if err != nil {
			sv.Send(w, r, -3, "Invalid note type.", "")
			return
		}

		key := strings.TrimSpace(r.FormValue(common.FIELD_KEY))
		if len(key) == 0 {
			sv.Send(w, r, -4, "Invalid note key.", "")
			return
		}

		result, err := sv.ns.GetTypedNote(t, key, session)
		if err != nil {
			sv.Send(w, r, -5, err.Error(), "")
			return
		}

		sv.Send(w, r, 0, "", result.ToJSON())
	}
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpDeleteNote(w http.ResponseWriter, r *http.Request) {
	// Check authority.
	session, err := sv.ss.CheckHttpSessionForUser(w, r)
	if err != nil {
		return
	}

	// Check inputs.

	id, err := strconv.Atoi(r.FormValue(common.FIELD_ID))
	if err != nil {
		sv.Send(w, r, -1, "Invalid note ID.", "")
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -2, "Invalid class ID.", "")
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -3, "Invalid meeting ID.", "")
		return
	}

	t, err := strconv.Atoi(r.FormValue(common.FIELD_TYPE))
	if err != nil {
		sv.Send(w, r, -4, "Invalid note type.", "")
		return
	}

	key := strings.TrimSpace(r.FormValue(common.FIELD_KEY))
	if len(key) == 0 {
		sv.Send(w, r, -5, "Invalid note key.", "")
		return
	}

	// Delete this note.

	err = sv.ns.DeleteNote(id, classID, meetingID, t, key, session)
	if err != nil {
		sv.Send(w, r, -10, err.Error(), "")
		return
	}

	// Say everything is fine.

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------
