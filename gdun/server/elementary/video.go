package elementary

import (
	"github.com/wangmhgo/go-project/gdun/common"
	"net/http"
	"strconv"
	"strings"
)

//----------------------------------------------------------------------------

func (sv *Server) onHttpAddVideo(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForTeacher(w, r)
	if err != nil {
		return
	}

	videoID := r.FormValue(common.FIELD_VIDEO)
	videoName := r.FormValue(common.FIELD_NAME)

	preparation, err := strconv.Atoi(r.FormValue(common.FIELD_PREPARATION))
	if err != nil {
		preparation = 1
	}
	// if preparation != 0 {
	// 	preparation = 1
	// }

	necessary, err := strconv.Atoi(r.FormValue(common.FIELD_NECESSARY))
	if err != nil {
		necessary = 0
	}
	if necessary != 0 {
		necessary = 1
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	err = sv.ms.AddVideo(videoID, videoName, meetingID, preparation, necessary, session)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpDeleteVideo(w http.ResponseWriter, r *http.Request) {
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

	err = sv.ms.DeleteVideo(videoID, meetingID, session)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpFinishVideo(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForStudent(w, r)
	if err != nil {
		return
	}

	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, common.S_INVALID_MEETING, "")
		return
	}

	videoID := common.Prune(r.FormValue(common.FIELD_VIDEO))
	if len(videoID) == 0 {
		sv.Send(w, r, -2, common.S_INVALID_VIDEO, "")
		return
	}

	err = sv.ms.SetVideoProgress(meetingID, videoID, session)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpAuthorizeVideos(w http.ResponseWriter, r *http.Request) {
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

	result := `"` + common.FIELD_VIDEO + `":[`
	first := true
	for i := 0; i < len(mi.Videos); i++ {
		if len(mi.Videos[i]) == 0 {
			continue
		}

		id := (strings.Split(mi.Videos[i], ":"))[0]
		if len(id) == 0 {
			continue
		}

		vi, err := sv.vs.GetVideoAuthorizeInfo(id)
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

func (sv *Server) OnHttpGetM3U8(w http.ResponseWriter, r *http.Request) {
	arr := strings.Split(r.RequestURI, "/")
	n := len(arr)
	if n != 6 {
		sv.Go404(w, r, common.S_INVALID_VIDEO)
		return
	}

	//----------------------------------------------------

	userID, err := strconv.Atoi(arr[1])
	if err != nil {
		sv.Go404(w, r, common.S_INVALID_USER)
		return
	}
	token := arr[2]

	// Check session.
	session, err := sv.ss.GetSession(userID)
	if err != nil {
		sv.Go404(w, r, err.Error())
		return
	}
	if !session.CheckToken(token) {
		if !session.CheckAppToken(token) {
			if !session.CheckWeixinToken(token) {
				sv.Go404(w, r, common.S_INVALID_TOKEN)
				return
			}
		}
	}

	//----------------------------------------------------

	videoID := common.Prune(arr[3])
	if len(videoID) == 0 {
		sv.Go404(w, r, "Empty video ID.")
		return
	}
	if !sv.vs.CheckInternalIP(videoID, r.RemoteAddr) {
		sv.Go404(w, r, "Internal video.")
		return
	}

	//----------------------------------------------------

	resolution := common.Prune(arr[4])
	if len(resolution) == 0 {
		sv.Go404(w, r, "Empty video resolution.")
		return
	}

	//----------------------------------------------------

	arr = strings.Split(arr[5], ".")
	if len(arr) != 2 {
		sv.Go404(w, r, common.S_INVALID_VIDEO)
		return
	}
	lineID, err := strconv.Atoi(arr[0])
	if err != nil {
		sv.Go404(w, r, "Invalid video line ID.")
		return
	}

	//----------------------------------------------------

	// Get M3U8 content.
	result, err := sv.vs.GetDowngradedM3U8(videoID, resolution, lineID, (r.TLS != nil), userID, token)
	if err != nil {
		sv.Go404(w, r, err.Error())
		return
	}

	//----------------------------------------------------

	if sv.accessLog != nil {
		sv.accessLog.Info(sv.createLogLine(r))
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/x-mpegurl")
	w.Write(([]byte)(result))
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpAuthorizeM3U8(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		sv.Go404(w, r, err.Error())
		return
	}

	//----------------------------------------------------

	videoID := common.Prune(r.FormValue(common.FIELD_ID))
	if len(videoID) == 0 {
		sv.Go404(w, r, "Empty video ID.")
		return
	}
	if !sv.vs.CheckInternalIP(videoID, r.RemoteAddr) {
		sv.Go404(w, r, "Internal video.")
		return
	}

	//----------------------------------------------------

	userID, err := strconv.Atoi(r.FormValue(common.FIELD_SESSION))
	if err != nil {
		sv.Go404(w, r, "Invalid user ID.")
		return
	}
	token := common.Prune(r.FormValue(common.FIELD_TOKEN))
	if len(token) == 0 {
		sv.Go404(w, r, "Empty user token.")
		return
	}

	session, err := sv.ss.GetSession(userID)
	if err != nil {
		sv.Go404(w, r, err.Error())
		return
	}
	if !session.CheckToken(token) {
		if !session.CheckAppToken(token) {
			if !session.CheckWeixinToken(token) {
				sv.Go404(w, r, "Invalid user token.")
				return
			}
		}
	}

	//----------------------------------------------------

	aesKey, err := sv.vs.GetVideoKey(videoID)
	if err != nil {
		sv.Go404(w, r, err.Error())
		return
	}

	if sv.accessLog != nil {
		sv.accessLog.Info(sv.createLogLine(r))
	}

	if sv.cache != nil {
		sv.cache.Publish(common.CHANNEL_REALTIME_VOD, r.RemoteAddr+","+videoID)
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(aesKey)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpQueryVideo(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	start, err := strconv.Atoi(r.FormValue(common.FIELD_START))
	if err != nil {
		start = 0
	}

	groupID, err := strconv.Atoi(r.FormValue(common.FIELD_GROUP))
	if err != nil {
		if session.IsAssistant() {
			groupID = session.GroupID
		} else {
			sv.Send(w, r, -1, common.S_NO_GROUP, "")
			return
		}
	}

	var keywords []string = nil
	s := common.Prune(r.FormValue(common.FIELD_KEYWORD))
	if len(s) > 0 {
		keywords = strings.Split(s, ",")
	}

	var IDs []string = nil
	s = common.Prune(r.FormValue(common.FIELD_ID))
	if len(s) > 0 {
		IDs = strings.Split(s, ",")
	}

	result, err := sv.vs.QueryVideo(start, groupID, keywords, IDs, session)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpUpdateVideoStatistics(w http.ResponseWriter, r *http.Request) {
	// session, err := sv.ss.CheckHttpSessionForStudent(w, r)
	// if err != nil {
	// 	return
	// }

	// id := common.Prune(r.FormValue(common.FIELD_ID))
	// if len(id) == 0 {
	// 	sv.Send(w, r, -1, common.S_INVALID_ID, "")
	// 	return
	// }

	// duration, err := strconv.Atoi(r.FormValue(common.field_sub))
	// if err != nil {
	// 	sv.Send(w, r, -2, err.Error(), "")
	// 	return
	// }

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------
