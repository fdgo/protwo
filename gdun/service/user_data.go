package service

import (
	"container/list"
	"github.com/wangmhgo/go-project/gdun/common"
	"strconv"
)

//----------------------------------------------------------------------------

type UserInfo struct {
	ID              int
	Nickname        string
	Remark          string
	GdStudentID int
	GroupID         int
	Weixin          string
	// Privilege       int
}

func NewUserInfoFromMap(m map[string]string, id int) *UserInfo {
	nickname, okay := m[common.FIELD_NICKNAME]
	if !okay {
		return nil
	}

	remark, okay := m[common.FIELD_REMARK]
	if !okay {
		return nil
	}

	s, okay := m[common.FIELD_GROUP_ID]
	if !okay {
		return nil
	}
	groupID, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}

	gdStudentID := 0
	if s, okay = m[common.FIELD_GAODUN_STUDENT_ID]; okay {
		if gdStudentID, err = strconv.Atoi(s); err != nil {
			gdStudentID = 0
		}
	}

	weixin, okay := m[common.FIELD_WEIXIN]
	if !okay {
		weixin = ""
	}

	// privilege := 0
	// if s, okay = m[common.FIELD_PRIVILEGE]; okay {
	// 	if privilege, err = strconv.Atoi(s); err != nil {
	// 		privilege = 0
	// 	}
	// }

	ui := new(UserInfo)
	ui.ID = id
	ui.Nickname = nickname
	ui.Remark = remark
	ui.GdStudentID = gdStudentID
	ui.GroupID = groupID
	ui.Weixin = weixin
	// ui.Privilege = privilege

	return ui
}

func (ui *UserInfo) ToJSON() string {
	s := `"` + common.FIELD_ID + `":` + strconv.Itoa(ui.ID) + `,` +
		`"` + common.FIELD_NICKNAME + `":"` + common.UnescapeForJSON(ui.Nickname) + `",` +
		`"` + common.FIELD_GROUP + `":` + strconv.Itoa(ui.GroupID) + `,` +
		`"` + common.FIELD_ROLE + `":`
	switch ui.GroupID {
	case common.GROUP_ID_FOR_SYSTEM:
		s += `"SYSTEM"`
	case common.GROUP_ID_FOR_TEACHER:
		s += `"TEACHER"`
	case common.GROUP_ID_FOR_KEEPER:
		s += `"KEEPER"`
	case common.GROUP_ID_FOR_STUDENT:
		s += `"STUDENT"`
	default:
		s += `"ASSISTANT","` + common.FIELD_GROUP + `":` + strconv.Itoa(ui.GroupID)
	}
	if ui.GdStudentID > 0 {
		s += `,"` + common.FIELD_GAODUN_STUDENT_ID + `":` + strconv.Itoa(ui.GdStudentID)
	}

	return s
}

//----------------------------------------------------------------------------

type UserInfoArray struct {
	Users *list.List
}

func (uia *UserInfoArray) ToJSON() string {
	s := `"` + common.FIELD_USER + `":{`

	if uia.Users != nil {
		first := true
		for p := uia.Users.Front(); p != nil; p = p.Next() {
			ui, okay := p.Value.(*UserInfo)
			if !okay {
				continue
			}

			if first {
				first = false
			} else {
				s += `,`
			}

			s += `"` + strconv.Itoa(ui.ID) + `":{` +
				`"` + common.FIELD_NICKNAME + `":"` + common.UnescapeForJSON(ui.Nickname) + `",` +
				`"` + common.FIELD_REMARK + `":"` + common.UnescapeForJSON(ui.Remark) + `",` +
				`"` + common.FIELD_ROLE + `":`
			switch ui.GroupID {
			case common.GROUP_ID_FOR_SYSTEM:
				s += `"SYSTEM"`
			case common.GROUP_ID_FOR_KEEPER:
				s += `"KEEPER"`
			case common.GROUP_ID_FOR_TEACHER:
				s += `"TEACHER"`
			case common.GROUP_ID_FOR_STUDENT:
				s += `"STUDENT"`
			default:
				s += `"ASSISTANT","` + common.FIELD_GROUP + `":` + strconv.Itoa(ui.GroupID)
			}
			if ui.GdStudentID > 0 {
				s += `,"` + common.FIELD_GAODUN_STUDENT_ID + `":` + strconv.Itoa(ui.GdStudentID)
			}
			s += `}`
		}
	}

	s += `},"` + common.FIELD_COUNT + `":`

	if uia.Users != nil {
		s += strconv.Itoa(uia.Users.Len())
	} else {
		s += `0`
	}

	return s
}

//----------------------------------------------------------------------------
