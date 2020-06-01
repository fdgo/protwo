package service

import (
	"github.com/wangmhgo/go-project/gdun/common"
	"net/http"
	"strconv"
	"strings"
)

//----------------------------------------------------------------------------

func (ss *SessionService) GetHttpSession(r *http.Request) (*Session, error) {
	// Check the HTTP method.
	if strings.Compare(r.Method, "POST") != 0 {
		return nil, common.ERR_NO_SERVICE
	}

	// Get them via HTTP headers.
	userID, token, app := (func() (int, string, int) {
		err := r.ParseForm()
		if err != nil {
			return 0, "", 0
		}

		s := r.FormValue(common.FIELD_SESSION)
		if len(s) == 0 {
			return 0, "", 0
		}
		userID, err := strconv.Atoi(s)
		if err != nil {
			return 0, "", 0
		}

		token := common.Prune(r.FormValue(common.FIELD_TOKEN))

		app, err := strconv.Atoi(r.FormValue(common.FIELD_APP))
		if err != nil {
			app = 0
		}

		return userID, token, app
	})()
	if userID <= 0 {
		return nil, common.ERR_INVALID_SESSION
	}

	// Get the session.
	session, err := ss.GetSession(userID)
	if err != nil {
		return nil, err
	}

	// Check the token.
	if app == 0 {
		if !session.CheckToken(token) {
			return nil, common.ERR_INVALID_TOKEN
		}
	} else if app == 1 {
		if !session.CheckAppToken(token) {
			return nil, common.ERR_INVALID_TOKEN
		}
	} else {
		if !session.CheckWeixinToken(token) {
			return nil, common.ERR_INVALID_TOKEN
		}
	}

	return session, nil
}

//----------------------------------------------------------------------------

func (ss *SessionService) SetHttpSession(id int, group int, nickname string, w http.ResponseWriter, r *http.Request) string {

	app, err := strconv.Atoi(r.FormValue(common.FIELD_APP))
	if err != nil {
		app = 0
	}

	device := common.Prune(r.FormValue(common.FIELD_DEVICE))

	sessionID, token := ss.SetSession(id, group, nickname, r.RemoteAddr, app, device)

	// Try to retrieve Weixin form ID.
	if app == 2 {
		formID := common.Prune(r.FormValue(common.FIELD_FORM_ID))
		if len(formID) > 0 {
			// Save this form ID.
			if err := ss.cache.Push(common.KEY_PREFIX_SESSION+strconv.Itoa(id)+":"+common.FIELD_FORM_ID, formID); err != nil {
				// TODO:
			}
		}
	}

	// Save private information on the client side for this user.

	//----------------------------------------------------
	// HTTP headers.
	w.Header().Set("Access-Control-Expose-Headers", common.FIELD_TOKEN+","+common.FIELD_DATE+","+common.FIELD_TIMESTAMP)
	w.Header().Set(common.FIELD_TOKEN, token)
	w.Header().Set(common.FIELD_TIMESTAMP, common.GetMillisecondString())

	// Return the session ID.
	return sessionID
}

//----------------------------------------------------------------------------

func (ss *SessionService) UpdateHttpSession(session *Session, w http.ResponseWriter, r *http.Request) error {

	app, err := strconv.Atoi(r.FormValue(common.FIELD_APP))
	if err != nil {
		app = 0
	}

	// Try to retrieve Weixin form ID.
	if app == 2 {
		formID := common.Prune(r.FormValue(common.FIELD_FORM_ID))
		if len(formID) > 0 {
			// Save this form ID.
			if err := ss.cache.Push(common.KEY_PREFIX_SESSION+strconv.Itoa(session.UserID)+":"+common.FIELD_FORM_ID, formID); err != nil {
				// TODO:
			}
		}
	}

	// Update cache.
	err = ss.UpdateSessionToken(session, app)
	if err != nil {
		return err
	}

	// Update HTTP header.
	w.Header().Set("Access-Control-Expose-Headers", common.FIELD_TOKEN+","+common.FIELD_DATE+","+common.FIELD_TIMESTAMP)
	switch app {
	case 1:
		w.Header().Set(common.FIELD_TOKEN, session.AppToken)
	case 2:
		w.Header().Set(common.FIELD_TOKEN, session.WeixinToken)
	default:
		w.Header().Set(common.FIELD_TOKEN, session.Token)
	}

	w.Header().Set(common.FIELD_TIMESTAMP, common.GetMillisecondString())

	return nil
}

//----------------------------------------------------------------------------

func (ss *SessionService) Go404(w http.ResponseWriter, r *http.Request) {
	if ss.accessLog != nil {
		go ss.accessLog.Error(r.RemoteAddr + " " + r.Host + r.RequestURI + " (404) (" + r.UserAgent() + ") (" + r.Referer() + ")")
	}

	w.WriteHeader(404)
}

//----------------------------------------------------------------------------

func (ss *SessionService) CheckHttpSessionForSystem(w http.ResponseWriter, r *http.Request) (*Session, error) {
	session, err := ss.GetHttpSession(r)
	if err != nil {
		ss.Go404(w, r)
		return nil, err
	}

	if !session.IsSystem() {
		ss.Go404(w, r)
		return nil, common.ERR_INVALID_SESSION
	}

	err = ss.UpdateHttpSession(session, w, r)
	if err != nil {
		return nil, err
	}
	return session, nil
}

//----------------------------------------------------------------------------

func (ss *SessionService) CheckHttpSessionForAssitant(w http.ResponseWriter, r *http.Request) (*Session, error) {
	session, err := ss.GetHttpSession(r)
	if err != nil {
		ss.Go404(w, r)
		return nil, err
	}

	if !session.IsAssistantOrAbove() {
		ss.Go404(w, r)
		return nil, common.ERR_INVALID_SESSION
	}

	err = ss.UpdateHttpSession(session, w, r)
	if err != nil {
		return nil, err
	}
	return session, nil
}

//----------------------------------------------------------------------------

func (ss *SessionService) CheckHttpSessionForTeacher(w http.ResponseWriter, r *http.Request) (*Session, error) {
	session, err := ss.GetHttpSession(r)
	if err != nil {
		ss.Go404(w, r)
		return nil, err
	}

	if !session.IsTeacherOrAbove() {
		ss.Go404(w, r)
		return nil, common.ERR_INVALID_SESSION
	}

	err = ss.UpdateHttpSession(session, w, r)
	if err != nil {
		return nil, err
	}
	return session, nil
}

//----------------------------------------------------------------------------

func (ss *SessionService) CheckHttpSessionForKeeper(w http.ResponseWriter, r *http.Request) (*Session, error) {
	session, err := ss.GetHttpSession(r)
	if err != nil {
		ss.Go404(w, r)
		return nil, err
	}

	if !session.IsKeeperOrAbove() {
		ss.Go404(w, r)
		return nil, common.ERR_INVALID_SESSION
	}

	err = ss.UpdateHttpSession(session, w, r)
	if err != nil {
		return nil, err
	}
	return session, err
}

//----------------------------------------------------------------------------

func (ss *SessionService) CheckHttpSessionForStudent(w http.ResponseWriter, r *http.Request) (*Session, error) {
	session, err := ss.GetHttpSession(r)
	if err != nil {
		ss.Go404(w, r)
		return nil, err
	}

	if (!session.IsStudent()) && (!session.IsSystem()) {
		ss.Go404(w, r)
		return nil, common.ERR_INVALID_SESSION
	}

	err = ss.UpdateHttpSession(session, w, r)
	if err != nil {
		return nil, err
	}
	return session, err
}

//----------------------------------------------------------------------------

func (ss *SessionService) CheckHttpSessionForUser(w http.ResponseWriter, r *http.Request) (*Session, error) {
	session, err := ss.GetHttpSession(r)
	if err != nil {
		ss.Go404(w, r)
		return nil, err
	}

	err = ss.UpdateHttpSession(session, w, r)
	if err != nil {
		return nil, err
	}
	return session, nil
}

//----------------------------------------------------------------------------
