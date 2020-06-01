package elementary

import (
	"fmt"
	"github.com/wangmhgo/go-project/gdun/common"
	// "gitlab.hfjy.com/gdun/server/live2"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//----------------------------------------------------------------------------

/*
{
    "result": "ok",
    "message": "登录成功",
    "user":{
        "id": "E6A232B2DEDF69469C33DC5901307461",
        "name': "学员A"
    }
}
roomid		字符串	直播间ID
viewername	字符串	userID_meetingID
viewertoken	字符串	userToken
*/

func (sv *Server) responseBokeccLogin(msg string, userID string, userName string, classID int, r *http.Request) []byte {
	if len(msg) == 0 {
		buf := `{` +
			`"result":"ok",` +
			`"message":"登录成功",` +
			`"user":{` +
			`"id":"` + strconv.Itoa(classID) + "_" + userID + `",` +
			`"name":"` + userName + `",` +
			`"viewercustommark":"` + strconv.Itoa(classID) + `"` +
			`}}`

		if sv.authorizeLog != nil {
			s := r.RemoteAddr + " (" + r.Host + ")" + r.RequestURI + " (" + userID + ")(" + userName + ")(" + strconv.Itoa(classID) + ")"
			go sv.authorizeLog.Info(s)
		}
		return []byte(buf)
	}

	if sv.authorizeLog != nil {
		s := r.RemoteAddr + " (" + r.Host + ")" + r.RequestURI + " (" + r.FormValue("viewername") + "," + r.FormValue("roomid") + "," + r.FormValue("viewertoken") + ")"
		go sv.authorizeLog.Error(s + " (" + msg + ")")
	}

	return []byte(`{"result":"error","message":"` + msg + `"}`)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpBokeccLogin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.Write(sv.responseBokeccLogin("Invalid HTTP form.", "", "", 0, r))
		return
	}

	viewername := common.Prune(r.FormValue("viewername"))

	params := strings.Split(viewername, "_")
	if len(params) != 2 {
		w.Write(sv.responseBokeccLogin("Invalid viewer name.", "", "", 0, r))
		return
	}

	userID, err := strconv.Atoi(params[0])
	if err != nil {
		w.Write(sv.responseBokeccLogin("Invalid user ID.", "", "", 0, r))
		return
	}
	meetingID, err := strconv.Atoi(params[1])
	if err != nil {
		w.Write(sv.responseBokeccLogin("Invalid meeting ID.", "", "", 0, r))
		return
	}

	roomID := common.Prune(r.FormValue("roomid"))
	if len(roomID) == 0 {
		w.Write(sv.responseBokeccLogin("Empty room ID.", "", "", 0, r))
		return
	}

	token := common.Prune(r.FormValue("viewertoken"))
	if len(token) == 0 {
		w.Write(sv.responseBokeccLogin("Empty user token.", "", "", 0, r))
		return
	}

	// Check user session.
	session, err := sv.ss.GetSession(userID)
	if err != nil {
		w.Write(sv.responseBokeccLogin("Invalid user session.", "", "", 0, r))
		return
	}
	// Check user token.
	if (!session.CheckToken(token)) && (!session.CheckAppToken(token)) && (!session.CheckWeixinToken(token)) {
		w.Write(sv.responseBokeccLogin("Invalid user token.", "", "", 0, r))
		return
	}
	if session.IsSystem() {
		userID = int(time.Now().UnixNano() / 1000000)
	} else if session.IsStudent() {
		if err = sv.ms.SetMeetingProgress(meetingID, 1, session); err != nil {
			// TODO:
		}

		if sv.cache != nil {
			go (func() {
				result := `"` + common.FIELD_IS_TEACHER + `":false,"` + common.FIELD_NAME + `":"` + common.UnescapeForJSON(session.Nickname) + `"`
				msg := common.PlainCommandJSONMessage(0, "", common.COMMAND_USER, common.COMMAND_USER_JOINED, session.UserID, result)
				gossipMsg := fmt.Sprintf("%08x%08x%08x0%s", 0, meetingID, common.ATTENDEE_GROUP_TEACHER, string(msg))

				if err := sv.cache.Publish(common.CHANNEL_GOSSIP, gossipMsg); err != nil {
					// TODO:
					fmt.Println("onHttpBokeccLogin() " + err.Error())
				}
			})()
		}

		// go (func() {
		// 	if lSv := live2.GetCurrentServer(); lSv != nil {
		// 		result := `"` + common.FIELD_IS_TEACHER + `":false,"` + common.FIELD_NAME + `":"` + common.UnescapeForJSON(session.Nickname) + `"`
		// 		msg := common.PlainCommandJSONMessage(0, "", common.COMMAND_USER, common.COMMAND_USER_JOINED, session.UserID, result)
		// 		lSv.Broadcast(meetingID, common.ATTENDEE_GROUP_TEACHER, string(msg))
		// 	}
		// })()
	}

	classID, err := sv.ms.GetMeetingClassID(meetingID)
	if err != nil {
		classID = 0
	}

	w.Write(sv.responseBokeccLogin("", strconv.Itoa(userID), common.Unescape(session.Nickname), classID, r))
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpBokeccInternalRegister(w http.ResponseWriter, r *http.Request) {
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
	nickname := common.Prune(r.FormValue(common.FIELD_NICKNAME))
	if len(nickname) == 0 {
		sv.Send(w, r, -4, "Invalid nickname.", "")
		return
	}

	// Save id.
	id := sv.ss.GetUUID()
	err = sv.cache.SetField("bokecc:internal:id", id, common.Escape(nickname))
	if err != nil {
		sv.Send(w, r, -5, err.Error(), "")
		return
	}

	result := `"` + common.FIELD_ID + `":"` + id + `"`
	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpBokeccInternalPass(w http.ResponseWriter, r *http.Request) {
	// Check requirements.
	if sv.cache == nil {
		sv.Send(w, r, -1, common.S_NO_SERVICE, "")
		return
	}

	// Get parameters.
	err := r.ParseForm()
	if err != nil {
		w.Write(sv.responseBokeccLogin("Invalid HTTP form.", "", "", 0, r))
		return
	}
	viewername := common.Prune(r.FormValue("viewername"))
	if len(viewername) == 0 {
		w.Write(sv.responseBokeccLogin("Invalid viewer name.", "", "", 0, r))
		return
	}
	roomID := common.Prune(r.FormValue("roomid"))
	if len(roomID) == 0 {
		w.Write(sv.responseBokeccLogin("Empty room ID.", "", "", 0, r))
		return
	}
	id := common.Prune(r.FormValue("viewertoken"))
	if len(id) == 0 {
		w.Write(sv.responseBokeccLogin("Empty user token.", "", "", 0, r))
		return
	}

	// Check authority.
	nickname, err := sv.cache.GetField("bokecc:internal:id", id)
	if err != nil {
		w.Write(sv.responseBokeccLogin("Empty user token.", "", "", 0, r))
		return
	}
	err = sv.cache.DelField("bokecc:internal:id", id)
	if err != nil {
		// TODO:
	}

	w.Write(sv.responseBokeccLogin("", id, common.Unescape(nickname), 0, r))
}

//----------------------------------------------------------------------------
