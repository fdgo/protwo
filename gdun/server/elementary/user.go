package elementary

import (
	"crypto/md5"
	"fmt"
	"github.com/wangmhgo/go-project/gdun/common"
	"github.com/wangmhgo/go-project/gdun/server/live2"
	"github.com/wangmhgo/go-project/gdun/service"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

//----------------------------------------------------------------------------

func (sv *Server) onHttpRegister(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	err = sv.us.RegisterViaInvitationToken(
		r.FormValue(common.FIELD_NAME),
		r.FormValue(common.FIELD_PASSWORD),
		r.FormValue(common.FIELD_TOKEN),
		r.RemoteAddr)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpLogin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	result, err := sv.us.Login(r.FormValue(common.FIELD_NAME), r.FormValue(common.FIELD_PASSWORD), r.RemoteAddr)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	sv.ss.SetHttpSession(result.ID, result.GroupID, result.Nickname, w, r)
	sv.Send(w, r, 0, "", result.ToJSON())
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpScan(w http.ResponseWriter, r *http.Request) {
	// Get teacher session.
	session, err := sv.ss.CheckHttpSessionForTeacher(w, r)
	if err != nil {
		return
	}

	// Get inputs.
	meetingID, err := strconv.Atoi(r.FormValue(common.FIELD_MEETING))
	if err != nil {
		sv.Send(w, r, -1, common.S_INVALID_MEETING, "")
		return
	}
	userID, err := strconv.Atoi(r.FormValue(common.FIELD_USER))
	if err != nil {
		sv.Send(w, r, -2, common.S_INVALID_USER, "")
		return
	}

	// Check authority.
	mi, err := sv.ms.GetMeeting(meetingID, session, true)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}
	if (mi.Type != common.MEETING_TYPE_FOR_OFFLINE) && (mi.Type != common.MEETING_TYPE_FOR_TEACHING) {
		sv.Send(w, r, -4, common.S_INVALID_MEETING, "")
		return
	}

	ci, err := sv.cs.GetClass(mi.ClassID, session)
	if err != nil {
		sv.Send(w, r, -5, err.Error(), "")
		return
	}
	if !common.InIntArray(userID, ci.Students) {
		sv.Send(w, r, -6, common.S_INVALID_USER, "")
		return
	}

	now := (int)(time.Now().Unix())
	data := now

	password := common.Prune(r.FormValue(common.FIELD_PASSWORD))
	if len(password) == 0 {
		// if !session.IsAssistantOrAbove() {
		// 	sv.Send(w, r, -5, "Empty password.", "")
		// 	return
		// }

		// Get student session.
		session, err = sv.ss.GetSession(userID)
		if err != nil {
			sv.Send(w, r, -7, err.Error(), "")
			return
		}

		data, err = strconv.Atoi(r.FormValue(common.FIELD_DATA))
		if err != nil {
			// sv.Send(w, r, -7, "Empty password.", "")
			// return
			data = -1 * now
		}
	} else {
		// Get student session.
		session, err = sv.ss.GetSession(userID)
		if err != nil {
			sv.Send(w, r, -8, err.Error(), "")
			return
		}

		// Check authority.
		if (!session.CheckToken(password)) && (!session.CheckAppToken(password) && (!session.CheckWeixinToken(password))) {
			sv.Send(w, r, -9, "Invalid password.", "")
			return
		}
	}

	// Update student's progress.
	if err = sv.ms.SetMeetingProgress(meetingID, data, session); err != nil {
		sv.Send(w, r, -10, err.Error(), "")
		return
	}

	go (func() {
		if lSv := live2.GetCurrentServer(); lSv != nil {
			result := `"` + common.FIELD_IS_TEACHER + `":false,"` + common.FIELD_NAME + `":"` + common.UnescapeForJSON(session.Nickname) + `"`
			msg := common.PlainCommandJSONMessage(0, "", common.COMMAND_USER, common.COMMAND_USER_JOINED, session.UserID, result)
			lSv.Broadcast(meetingID, common.ATTENDEE_GROUP_TEACHER, string(msg))
		}
	})()

	result := `"` + common.FIELD_ID + `":` + strconv.Itoa(session.UserID) + `,` +
		`"` + common.FIELD_NAME + `":"` + common.UnescapeForJSON(session.Nickname) + `",` +
		`"` + common.FIELD_IP + `":"` + session.IP + `"`
	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpLoginAsGdUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	// Get parameters.
	name := common.Prune(r.FormValue(common.FIELD_NAME))
	if len(name) == 0 {
		sv.Send(w, r, -2, common.S_INVALID_NAME, "")
		return
	}
	password := common.Prune(r.FormValue(common.FIELD_PASSWORD))
	if len(password) == 0 {
		sv.Send(w, r, -3, common.S_INVALID_PASSWORD, "")
		return
	}

	passwordMD5 := fmt.Sprintf("%x", md5.Sum(([]byte)(password)))

	ui, err := sv.us.Login(name, passwordMD5, r.RemoteAddr)
	if err != nil {
		// Get Gd student ID.
		gdStudentID, err := sv.gdp.Login(name, password)
		if err != nil {
			sv.Send(w, r, -4, err.Error(), "")
			return
		}

		// Get user ID and nickname.
		userID, nickname, _, err := sv.us.GetOrAddGdStudent(gdStudentID, r.RemoteAddr, false)
		if err != nil {
			sv.Send(w, r, -5, err.Error(), "")
			return
		}

		ui = new(service.UserInfo)
		ui.ID = userID
		ui.Nickname = nickname
	}

	// go sv.accessLog

	sv.ss.SetHttpSession(ui.ID, common.GROUP_ID_FOR_STUDENT, ui.Nickname, w, r)

	result := `"` + common.FIELD_ID + `":` + strconv.Itoa(ui.ID) + `,"` + common.FIELD_NICKNAME + `":"` + common.UnescapeForJSON(ui.Nickname) + `"`
	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpLoginAs3rdUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	// Get parameters.
	id := common.Prune(r.FormValue(common.FIELD_ID))
	if len(id) == 0 {
		sv.Send(w, r, -2, common.S_INVALID_USER, "")
		return
	}
	t, err := strconv.Atoi(r.FormValue(common.FIELD_TYPE))
	if err != nil {
		sv.Send(w, r, -3, common.S_INVALID_TYPE, "")
		return
	}

	// Get Gd student ID.
	gdStudentID, err := sv.gdp.LoginAs3rd(id, t)
	if err != nil {
		sv.Send(w, r, -4, err.Error(), "")
		return
	}

	// Get user ID and nickname.
	userID, nickname, _, err := sv.us.GetOrAddGdStudent(gdStudentID, r.RemoteAddr, false)
	if err != nil {
		sv.Send(w, r, -5, err.Error(), "")
		return
	}

	sv.ss.SetHttpSession(userID, common.GROUP_ID_FOR_STUDENT, nickname, w, r)

	result := `"` + common.FIELD_ID + `":` + strconv.Itoa(userID) + `,"` + common.FIELD_NICKNAME + `":"` + common.UnescapeForJSON(nickname) + `"`
	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpLoginAsGdStudent(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	ui, classID, showDate, status, err := sv.gdp.Pass(r.FormValue(common.FIELD_PASSWORD), r.RemoteAddr)
	if err != nil {
		sv.Send(w, r, status-1, err.Error(), "")
		return
	}

	sv.ss.SetHttpSession(ui.ID, ui.GroupID, ui.Nickname, w, r)
	sv.Send(w, r, 0, "", ui.ToJSON()+`,"`+common.FIELD_CLASS+`":`+strconv.Itoa(classID)+`,"`+common.FIELD_DATE+`":`+strconv.Itoa(showDate))
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpExperienceLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		sv.Send(w, r, -1, "Invalid form.", "")
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil {
		sv.Send(w, r, -2, "Invalid class ID.", "")
		return
	}

	phone, err := strconv.Atoi(r.FormValue(common.FIELD_PHONE))
	if err != nil {
		sv.Send(w, r, -3, "Invalid phone number.", "")
		return
	}

	result, id, data, err := sv.cs.LoginAsExperienceUser(classID, phone, r.RemoteAddr)
	if err != nil {
		sv.Send(w, r, -4, err.Error(), "")
		return
	}

	sv.ss.SetHttpSession(result.ID, result.GroupID, result.Nickname, w, r)
	sv.Send(w, r, 0, "", result.ToJSON()+`,"`+common.FIELD_PLATFORM_ID+`":`+strconv.Itoa(id)+`,"`+common.FIELD_PLATFORM_DATA+`":"`+data+`"`)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpExperienceRegister(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		sv.Send(w, r, -1, "Invalid form.", "")
		return
	}

	classID, err := strconv.Atoi(r.FormValue(common.FIELD_CLASS))
	if err != nil || classID < 0 {
		sv.Send(w, r, -2, "Invalid class ID.", "")
		return
	}

	teacherID, err := strconv.Atoi(r.FormValue(common.FIELD_TEACHER))
	if err != nil || teacherID < 0 {
		teacherID = 0
	}

	endTime, err := strconv.Atoi(r.FormValue(common.FIELD_END_TIME))
	if err != nil || endTime < 0 {
		endTime = 0
	}

	token := common.Prune(r.FormValue(common.FIELD_PASSWORD))
	if (teacherID == 0) && (token == "") {
		sv.Send(w, r, -3, "Invalid teacher ID or invitation token.", "")
		return
	}

	phone, err := strconv.Atoi(r.FormValue(common.FIELD_PHONE))
	if err != nil {
		sv.Send(w, r, -4, "Invalid phone number.", "")
		return
	}

	verificationCode, err := strconv.Atoi(r.FormValue(common.FIELD_VERIFICATION))
	if err != nil {
		sv.Send(w, r, -5, "Invalid verification code.", "")
		return
	}

	result, id, data, err := sv.cs.RegisterExperienceUser(classID, teacherID, endTime, token, phone, verificationCode, r.RemoteAddr)
	if err != nil {
		sv.Send(w, r, -6, err.Error(), "")
		return
	}

	sv.ss.SetHttpSession(result.ID, result.GroupID, result.Nickname, w, r)
	sv.Send(w, r, 0, "", result.ToJSON()+`,"`+common.FIELD_PLATFORM_ID+`":`+strconv.Itoa(id)+`,"`+common.FIELD_PLATFORM_DATA+`":"`+data+`"`)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpSendVerificationCode(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		sv.Send(w, r, -1, "Invalid form.", "")
		return
	}

	phone, err := strconv.Atoi(r.FormValue(common.FIELD_PHONE))
	if err != nil {
		sv.Send(w, r, -2, "Invalid phone number.", "")
		return
	}

	// Generate a verification code.
	verificationCode := rand.Int() % 10000
	if verificationCode < 1000 {
		verificationCode += 1000
	}

	// Save it.
	key := common.KEY_PREFIX_VERIFICATION + strconv.Itoa(phone)
	if err := sv.cache.SetKey(key, strconv.Itoa(verificationCode)); err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}
	if err := sv.cache.Expire(key, 10*time.Minute); err != nil {
		sv.Send(w, r, -4, err.Error(), "")
		return
	}

	// if err := sv.cs.SendVerificationCodeToExperienceUser(phone, verificationCode); err != nil {
	// 	sv.Send(w, r, -5, err.Error(), "")
	// 	return
	// }
	ip := r.RemoteAddr
	arr := strings.Split(r.RemoteAddr, ":")
	if len(arr) == 2 {
		ip = arr[0]
	}
	if err := common.SendVerificationCode(phone, verificationCode, ip); err != nil {
		sv.Send(w, r, -5, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpChangePassword(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForUser(w, r)
	if err != nil {
		return
	}

	err = sv.us.ChangePassword(session.UserID, r.FormValue(common.FIELD_OLD_PASSWORD), r.FormValue(common.FIELD_PASSWORD))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpChangeProfile(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForUser(w, r)
	if err != nil {
		return
	}

	err = sv.us.ChangeProfile(session.UserID,
		r.FormValue(common.FIELD_NICKNAME),
		r.FormValue(common.FIELD_MAIL),
		r.FormValue(common.FIELD_PHONE),
		r.FormValue(common.FIELD_QQ),
		r.FormValue(common.FIELD_WEIXIN),
		r.FormValue(common.FIELD_WEIBO),
		session)
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpQueryUser(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	name := r.FormValue(common.FIELD_NAME)

	groupID, err := strconv.Atoi(r.FormValue(common.FIELD_GROUP))
	if err != nil {
		sv.Send(w, r, -1, "Invalid group ID.", "")
		return
	}

	result, err := sv.us.QueryUsersByNickname(name, groupID, session)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", result.ToJSON())
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpGenerateInvitationToken(w http.ResponseWriter, r *http.Request) {
	_, err := sv.ss.CheckHttpSessionForSystem(w, r)
	if err != nil {
		return
	}

	s := r.FormValue(common.FIELD_SIZE)
	size, err := strconv.Atoi(s)
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	s = r.FormValue(common.FIELD_GROUP)
	groupID, err := strconv.Atoi(s)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	tokens, err := sv.us.GenerateInvitationToken(size, groupID)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	first := true
	s = `"` + common.FIELD_TOKEN + `":[`
	for i := 0; i < len(tokens); i++ {
		if first {
			first = false
		} else {
			s += `,`
		}
		s += `"` + tokens[i] + `"`
	}
	s += `]`

	sv.Send(w, r, 0, "", s)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpQueryInvitationToken(w http.ResponseWriter, r *http.Request) {
	_, err := sv.ss.CheckHttpSessionForSystem(w, r)
	if err != nil {
		return
	}

	tokens, err := sv.us.QueryInvitationToken()
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	keys := make([]string, len(tokens))
	i := 0
	for k, _ := range tokens {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	first := true
	s := `"` + common.FIELD_TOKEN + `":[`
	for _, k := range keys {
		if first {
			first = false
		} else {
			s += `,`
		}

		s += `{"` + common.FIELD_NAME + `":"` + k + `","` + common.FIELD_GROUP + `":` + tokens[k] + `}`
	}
	s += `]`

	sv.Send(w, r, 0, "", s)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpDeleteInvitationToken(w http.ResponseWriter, r *http.Request) {
	_, err := sv.ss.CheckHttpSessionForSystem(w, r)
	if err != nil {
		return
	}

	err = sv.us.DeleteInvitationToken(r.FormValue(common.FIELD_ID))
	if err != nil {
		sv.Send(w, r, -1, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpQueryGdStudent(w http.ResponseWriter, r *http.Request) {
	_, err := sv.ss.CheckHttpSessionForTeacher(w, r)
	if err != nil {
		return
	}

	id, err := strconv.Atoi(r.FormValue(common.FIELD_ID))
	if err != nil {
		sv.Send(w, r, -1, common.S_INVALID_STUDENT, "")
		return
	}

	name, phone, email := sv.gdp.GetSutdentInfo(id)
	result := `"` + common.FIELD_NAME + `":"` + name + `","` +
		common.FIELD_PHONE + `":"` + phone + `","` +
		common.FIELD_EMAIL + `":"` + email + `"`

	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpResyncGdStudentName(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForAssitant(w, r)
	if err != nil {
		return
	}

	userID, err := strconv.Atoi(r.FormValue(common.FIELD_ID))
	if err != nil {
		sv.Send(w, r, -1, common.S_INVALID_STUDENT, "")
		return
	}

	gdStudentID, err := sv.us.QueryGdStudentID(userID)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	name, err := sv.gdp.GetStudentName(gdStudentID)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	err = sv.us.Remark(userID, name, "", session)
	if err != nil {
		sv.Send(w, r, -4, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------
