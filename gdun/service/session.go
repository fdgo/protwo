package service

import (
	"github.com/wangmhgo/go-project/gdun/common"
	"github.com/wangmhgo/go-project/gdun/log"
	"math/rand"
	"strconv"
	"time"
)

//----------------------------------------------------------------------------

type SessionService struct {
	cache     *common.Cache
	accessLog *log.Logger
}

func NewSessionService(cache *common.Cache, accessLog *log.Logger) *SessionService {
	ss := new(SessionService)
	ss.cache = cache
	ss.accessLog = accessLog
	rand.Seed(time.Now().UnixNano())

	return ss
}

//----------------------------------------------------------------------------

func (ss *SessionService) GetUUID() string {
	return strconv.FormatInt(rand.Int63(), 32)
}

//----------------------------------------------------------------------------

func (ss *SessionService) SetSession(id int, group int, nickname string, ip string, app int, device string) (string, string) {
	// Generate a session ID.
	sUserID := strconv.Itoa(id)
	key := common.KEY_PREFIX_SESSION + sUserID

	// Generate a token.
	token := ss.GetUUID()

	// Save private information on the server side for this user.
	m := make(map[string]string)
	m[common.FIELD_ID] = sUserID
	m[common.FIELD_GROUP] = strconv.Itoa(group)
	m[common.FIELD_NICKNAME] = nickname
	if app == 0 {
		m[common.FIELD_TOKEN] = token
	} else if app == 1 {
		m[common.FIELD_APP_TOKEN] = token
	} else {
		m[common.FIELD_WEIXIN_TOKEN] = token
	}
	m[common.FIELD_IP] = ip
	m[common.FIELD_DEVICE] = device

	if err := ss.cache.SetFields(key, m); err != nil {
		// TODO:
	}

	return sUserID, token
}

//----------------------------------------------------------------------------

func (ss *SessionService) GetSession(userID int) (*Session, error) {
	m, err := ss.cache.GetAllFields(common.KEY_PREFIX_SESSION + strconv.Itoa(userID))
	if err != nil {
		return nil, err
	}

	r := NewSessionFromMap(m, userID)
	if r == nil {
		return nil, common.ERR_INVALID_SESSION
	}

	return r, nil
}

//----------------------------------------------------------------------------

func (ss *SessionService) ExpireSession(userID int, duration time.Duration) error {
	return ss.cache.Expire(common.KEY_PREFIX_SESSION+strconv.Itoa(userID), duration)
}

//----------------------------------------------------------------------------

func (ss *SessionService) UpdateSessionNickname(session *Session, nickname string) error {
	return ss.cache.SetField(common.KEY_PREFIX_SESSION+strconv.Itoa(session.UserID), common.FIELD_NICKNAME, nickname)
}

//----------------------------------------------------------------------------

func (ss *SessionService) UpdateSessionToken(session *Session, app int) error {
	key := common.KEY_PREFIX_SESSION + strconv.Itoa(session.UserID)
	token := ss.GetUUID()

	switch app {
	case 1:
		if err := ss.cache.SetField(key, common.FIELD_APP_TOKEN, token); err != nil {
			return err
		}
		session.AppToken = token

	case 2:
		if err := ss.cache.SetField(key, common.FIELD_WEIXIN_TOKEN, token); err != nil {
			return err
		}
		session.WeixinToken = token

	case 0:
		if err := ss.cache.SetField(key, common.FIELD_TOKEN, token); err != nil {
			return err
		}
		session.Token = token
	}

	return nil
}

//----------------------------------------------------------------------------

func (ss *SessionService) SetWeixinOpenID(session *Session, openID string) error {
	if err := ss.cache.SetField(common.KEY_PREFIX_SESSION+strconv.Itoa(session.UserID), common.FIELD_WEIXIN_OPEN_ID, openID); err != nil {
		return err
	}

	session.WeixinOpenID = openID
	return nil
}

//----------------------------------------------------------------------------
