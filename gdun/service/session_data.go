package service

import (
	"github.com/wangmhgo/go-project/gdun/common"
	"strconv"
)

//----------------------------------------------------------------------------

type Session struct {
	UserID        int
	Nickname      string
	GroupID       int
	IP            string
	Token         string
	AppToken      string
	WeixinToken   string
	WeixinOpenID  string
	UMengDeviceID string
}

func NewSessionFromMap(m map[string]string, userID int) *Session {
	nickname, okay := m[common.FIELD_NICKNAME]
	if !okay {
		return nil
	}

	s, okay := m[common.FIELD_GROUP]
	if !okay {
		return nil
	}
	groupID, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}

	ip, okay := m[common.FIELD_IP]
	if !okay {
		return nil
	}

	token, okay := m[common.FIELD_TOKEN]
	if !okay {
		token = ""
	}
	appToken, okay := m[common.FIELD_APP_TOKEN]
	if !okay {
		appToken = ""
	}
	weixinToken, okay := m[common.FIELD_WEIXIN_TOKEN]
	if !okay {
		weixinToken = ""
	}
	umengDeviceID, okay := m[common.FIELD_DEVICE]
	if !okay {
		umengDeviceID = ""
	}
	weixinOpenID, okay := m[common.FIELD_WEIXIN_OPEN_ID]
	if !okay {
		weixinOpenID = ""
	}

	r := new(Session)
	r.UserID = userID
	r.Nickname = nickname
	r.GroupID = groupID
	r.IP = ip
	r.Token = token
	r.AppToken = appToken
	r.WeixinToken = weixinToken
	r.UMengDeviceID = umengDeviceID
	r.WeixinOpenID = weixinOpenID

	return r
}

//----------------------------------------------------------------------------

func (s *Session) IsExperienceStudent() bool {
	return s.UserID > common.VALUE_MINIMAL_TEMPERARY_USER_ID
}

//----------------------------------------------------------------------------

func (s *Session) IsStudent() bool {
	return s.GroupID == common.GROUP_ID_FOR_STUDENT
}

//----------------------------------------------------------------------------

func (s *Session) IsKeeper() bool {
	return s.GroupID == common.GROUP_ID_FOR_KEEPER
}

func (s *Session) IsKeeperOrAbove() bool {
	if s.GroupID == 0 {
		return false
	}

	switch s.GroupID {
	case common.GROUP_ID_FOR_STUDENT, common.GROUP_ID_FOR_TEACHER:
		return false

	default:
		return true
	}
}

//----------------------------------------------------------------------------

func (s *Session) IsTeacher() bool {
	return s.GroupID == common.GROUP_ID_FOR_TEACHER
}

func (s *Session) IsTeacherOrAbove() bool {
	if s.GroupID == 0 {
		return false
	}

	switch s.GroupID {
	case common.GROUP_ID_FOR_STUDENT:
		return false

	default:
		return true
	}
}

//----------------------------------------------------------------------------

func (s *Session) IsAssistant() bool {
	if s.GroupID == 0 {
		return false
	}

	switch s.GroupID {
	case common.GROUP_ID_FOR_STUDENT, common.GROUP_ID_FOR_KEEPER, common.GROUP_ID_FOR_TEACHER, common.GROUP_ID_FOR_SYSTEM:
		return false

	default:
		return true
	}
}

func (s *Session) IsAssistantOrAbove() bool {
	if s.GroupID == 0 {
		return false
	}

	switch s.GroupID {
	case common.GROUP_ID_FOR_STUDENT, common.GROUP_ID_FOR_KEEPER, common.GROUP_ID_FOR_TEACHER:
		return false

	default:
		return true
	}
}

//----------------------------------------------------------------------------

func (s *Session) IsAdmin() bool {
	return s.GroupID < 0 || s.GroupID == common.GROUP_ID_FOR_SYSTEM
}

//----------------------------------------------------------------------------

func (s *Session) IsSystem() bool {
	return s.GroupID == common.GROUP_ID_FOR_SYSTEM
}

//----------------------------------------------------------------------------

func (s *Session) CheckToken(token string) bool {
	if len(token) == 0 {
		return false
	}
	if token != s.Token {
		return false
	}
	return true
}

//----------------------------------------------------------------------------

func (s *Session) CheckAppToken(token string) bool {
	if len(token) == 0 {
		return false
	}
	if token != s.AppToken {
		return false
	}
	return true
}

//----------------------------------------------------------------------------

func (s *Session) CheckWeixinToken(token string) bool {
	if len(token) == 0 {
		return false
	}
	if token != s.WeixinToken {
		return false
	}
	return true
}

//----------------------------------------------------------------------------
