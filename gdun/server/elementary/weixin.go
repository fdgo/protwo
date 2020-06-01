package elementary

import (
	"github.com/wangmhgo/go-project/gdun/common"
	"net/http"
	"strconv"
)

//----------------------------------------------------------------------------

func (sv *Server) onHttpWeixinLogin(w http.ResponseWriter, r *http.Request) {
	session, err := sv.ss.CheckHttpSessionForUser(w, r)
	if err != nil {
		return
	}

	code := common.Prune(r.FormValue(common.FIELD_CODE))
	if len(code) == 0 {
		sv.Send(w, r, -1, common.S_EMPTY_CODE, "")
		return
	}

	openID, err := sv.studentWeixin.Login(code)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	if err = sv.ss.SetWeixinOpenID(session, openID); err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------

func (sv *Server) onHttpGetWeixinMessageTemplates(w http.ResponseWriter, r *http.Request) {
	_, err := sv.ss.CheckHttpSessionForSystem(w, r)
	if err != nil {
		return
	}

	userID, err := strconv.Atoi(r.FormValue(common.FIELD_USER))
	if err != nil {
		sv.Send(w, r, -1, common.S_INVALID_USER, "")
		return
	}
	sUserID := strconv.Itoa(userID)

	openID, err := sv.cache.GetField(common.KEY_PREFIX_SESSION+sUserID, common.FIELD_WEIXIN_OPEN_ID)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	formID, err := sv.cache.Pop(common.KEY_PREFIX_SESSION + sUserID + ":" + common.FIELD_FORM_ID)
	if err != nil {
		sv.Send(w, r, -3, err.Error(), "")
		return
	}

	templateID := "XrRVYJxCjyk9pU9ncrC3IpFrpsfCNCFlBx5N4_d02Bs"
	page := ""
	data := `{"keyword1":{"value":"测试课程","color":"#173177"}}`

	err = sv.studentWeixin.SendMessage(openID, templateID, page, formID, data)
	if err != nil {
		sv.Send(w, r, -4, err.Error(), "")
		return
	}

	// m, err := sv.studentWeixin.GetTemplates(0, 20)
	// if err != nil {
	// 	sv.Send(w, r, -1, err.Error(), "")
	// 	return
	// }

	// result := ``
	// first := true
	// for id, title := range m {
	// 	if first {
	// 		first = false
	// 	} else {
	// 		result += `,`
	// 	}
	// 	result += `"` + id + `":"` + title + `"`
	// }

	sv.Send(w, r, 0, "", "")
}

//----------------------------------------------------------------------------
