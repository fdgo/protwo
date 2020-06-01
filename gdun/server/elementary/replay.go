package elementary

import (
	"github.com/wangmhgo/go-project/gdun/common"
	"net/http"
	"strconv"
	"strings"
)

//----------------------------------------------------------------------------

func (sv *Server) onHttpAddReplay(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForTeacher(w, r)
	if err != nil {
		return
	}

	videoID := r.FormValue(common.FIELD_VIDEO)

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	err = sv.ms.AddReplay(videoID, meetingID, session)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpDeleteReplay(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForTeacher(w, r)
	if err != nil {
		return
	}

	videoID := r.FormValue(common.FIELD_VIDEO)

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	err = sv.ms.DeleteReplay(videoID, meetingID, session)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpFinishReplay(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForStudent(w, r)
	if err != nil {
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	err = sv.ms.SetReplayProgress(meetingID, session)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpAuthorizeReplays(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForStudent(w, r)
	if err != nil {
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, "Invalid meeting ID.", "")
		return
	}

	mi, err := sv.ms.GetMeeting(meetingID, session, true)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	result := `"` + common.FIELD_REPLAY + `":[`
	first := true
	for i := 0; i < len(mi.Replays); i++ {
		if len(mi.Replays[i]) == 0 {
			continue
		}
		if strings.Index(mi.Replays[i], "http") == 0 {
			continue
		}

		vi, err := sv.vs.GetVideoAuthorizeInfo(mi.Replays[i])
		if err != nil {
			sv.Send(w, r, -3, err.Error(), "")
			return
		}

		if first {
			first = false
		} else {
			result += `,`
		}
		result += `{` + vi.ToJSON() + `}`
	}
	result += `]`

	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------
