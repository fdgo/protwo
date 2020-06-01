package service

import (
	"github.com/wangmhgo/go-project/gdun/common"
	"strconv"
	"strings"
	"time"
)

//----------------------------------------------------------------------------

func (cs *ClassService) GenerateInvitationToken(classID int, endTime int, duration int, teacherID int, channel string, size int, session *Session) error {
	if cs.cache == nil {
		return common.ERR_NO_SERVICE
	}

	if teacherID == 0 {
		key := common.KEY_PREFIX_CLASS + strconv.Itoa(classID) + ":" + common.FIELD_TOKEN
		value := common.GetTimeString() + ":" + strconv.Itoa(endTime) + ":" + strconv.Itoa(duration) +
			":" + common.Escape(channel) +
			":" + strconv.Itoa(session.UserID) + ":" + common.Escape(session.IP)

		for i := 0; i < size; i++ {
			token := cs.ms.GetUUID()
			cs.cache.SetField(key, token, value)
		}
	} else {
		key := common.KEY_PREFIX_CLASS + strconv.Itoa(classID) + ":" + common.FIELD_CHANNEL
		value := common.GetTimeString() + ":" + strconv.Itoa(endTime) + ":" + strconv.Itoa(duration) +
			":" + strconv.Itoa(session.UserID) + ":" + common.Escape(session.IP)

		cs.cache.SetField(key, strconv.Itoa(teacherID), value)
	}

	return nil
}

//----------------------------------------------------------------------------

func (cs *ClassService) QueryInvitationToken(classID int, isToken bool, session *Session) (string, error) {
	if cs.cache == nil {
		return "", common.ERR_NO_SERVICE
	}

	now := int(time.Now().Unix())

	if isToken {
		key := common.KEY_PREFIX_CLASS + strconv.Itoa(classID) + ":" + common.FIELD_TOKEN
		m, err := cs.cache.GetAllFields(key)
		if err != nil {
			return "", err
		}

		r := ``
		first := true
		for token, value := range m {
			arr := strings.Split(value, ":")

			endTime, err := strconv.Atoi(arr[1])
			if err != nil || endTime <= now {
				cs.cache.DelField(key, token)
				continue
			}

			if first {
				first = false
			} else {
				r += `,`
			}

			duration, err := strconv.Atoi(arr[2])
			if err != nil {
				duration = 0
			}
			updateTime, err := strconv.Atoi(arr[0])
			if err != nil {
				updateTime = 0
			}

			r += `{` +
				`"` + common.FIELD_TOKEN + `":"` + token + `",` +
				`"` + common.FIELD_END_TIME + `":` + strconv.Itoa(endTime*1000) + `,` +
				`"` + common.FIELD_DURATION + `":` + strconv.Itoa(duration*1000) + `,` +
				`"` + common.FIELD_CHANNEL + `":"` + common.UnescapeForJSON(arr[3]) + `",` +
				`"` + common.FIELD_UPDATER + `":` + arr[4] + `,` +
				`"` + common.FIELD_UPDATE_TIME + `":` + strconv.Itoa(updateTime*1000) + `,` +
				`"` + common.FIELD_UPDATE_IP + `":"` + common.UnescapeForJSON(arr[5]) + `"` +
				`}`
		}

		return `"` + common.FIELD_INVITATION + `":[` + r + `]`, nil
	} else {
		key := common.KEY_PREFIX_CLASS + strconv.Itoa(classID) + ":" + common.FIELD_CHANNEL
		m, err := cs.cache.GetAllFields(key)
		if err != nil {
			return "", err
		}

		r := ``
		first := true
		for id, value := range m {
			arr := strings.Split(value, ":")

			endTime, err := strconv.Atoi(arr[1])
			if err != nil || endTime <= now {
				cs.cache.DelField(key, id)
				continue
			}

			if first {
				first = false
			} else {
				r += `,`
			}

			duration, err := strconv.Atoi(arr[2])
			if err != nil {
				duration = 0
			}
			updateTime, err := strconv.Atoi(arr[0])
			if err != nil {
				updateTime = 0
			}

			r += `{` +
				`"` + common.FIELD_TEACHER + `":"` + id + `",` +
				`"` + common.FIELD_END_TIME + `":` + strconv.Itoa(endTime*1000) + `,` +
				`"` + common.FIELD_DURATION + `":` + strconv.Itoa(duration*1000) + `,` +
				`"` + common.FIELD_UPDATER + `":` + arr[3] + `,` +
				`"` + common.FIELD_UPDATE_TIME + `":` + strconv.Itoa(updateTime*1000) + `,` +
				`"` + common.FIELD_UPDATE_IP + `":"` + common.UnescapeForJSON(arr[4]) + `"` +
				`}`
		}

		return `"` + common.FIELD_INVITATION + `":[` + r + `]`, nil
	}
}

//----------------------------------------------------------------------------

func (cs *ClassService) RegisterExperienceUser(classID int, teacherID int, endTime int, token string, phone int, verificationCode int, ip string) (*UserInfo, int, string, error) {
	if cs.cache == nil {
		return nil, 0, "", common.ERR_NO_SERVICE
	}

	// Check verification code.
	if err := cs.checkVerificationCode(phone, verificationCode); err != nil {
		return nil, 0, "", common.ERR_INVALID_TOKEN
	}

	// Get info on this invitation.
	cii, err := cs.getClassInvitationInfo(classID, teacherID, token, endTime)
	if err != nil {
		return nil, 0, "", common.ERR_INVALID_KEY
	}

	// Assign a user ID.
	ui, err := cs.getOrCreateIDForExperienceUser(phone, ip)
	if err != nil {
		return nil, 0, "", err
	}

	// Authorize this user to visit the class.
	if err = cs.authorizeClassToExperienceUser(ui.ID, classID, cii.ExpiredTime); err != nil {
		return nil, 0, "", err
	}

	// Save a registration record.
	if err := cs.recordRegisterOfExperienceUser(classID, token, ui.ID, phone, cii, ip); err != nil {
		return nil, 0, "", err
	}

	// Get platform ID and data.
	platformID, platformData, err := cs.getPlatformInfo(classID)
	if err != nil {
		return nil, 0, "", common.ERR_INVALID_CLASS
	}

	return ui, platformID, platformData, nil
}

//----------------------------------------------------------------------------

func (cs *ClassService) LoginAsExperienceUser(classID int, phone int, ip string) (*UserInfo, int, string, error) {
	if cs.cache == nil {
		return nil, 0, "", common.ERR_NO_SERVICE
	}

	// Get user info.
	ui, err := cs.getOrCreateIDForExperienceUser(phone, ip)
	if err != nil {
		return nil, 0, "", common.ERR_NO_USER
	}

	// Check authority.
	if err = cs.checkAuthorityOfExperienceUser(ui.ID, classID); err != nil {
		return nil, 0, "", common.ERR_NO_AUTHORITY
	}

	// Update login log.
	if err = cs.recordLoginOfExperienceUser(classID, phone); err != nil {
		// WE DO NOTHING YET.
	}

	// Get platform ID and data.
	platformID, platformData, err := cs.getPlatformInfo(classID)
	if err != nil {
		return nil, 0, "", common.ERR_INVALID_CLASS
	}

	return ui, platformID, platformData, nil
}

//----------------------------------------------------------------------------

func (cs *ClassService) QueryExperienceUserLog(classID int, isToken bool, session *Session) (string, error) {
	if cs.cache == nil {
		return "", common.ERR_NO_SERVICE
	}

	sClassID := strconv.Itoa(classID)

	key := common.KEY_PREFIX_CLASS + sClassID + ":"
	if isToken {
		key += common.FIELD_TOKEN
	} else {
		key += common.FIELD_CHANNEL
	}
	key += ":" + common.FIELD_LOG

	s, err := cs.cache.GetKey(key)
	if err != nil {
		s = ""
	}

	mTime, err := cs.cache.GetAllFields(common.KEY_PREFIX_CLASS + sClassID + ":" + common.KEY_PREFIX_EXPERIENCE + common.FIELD_TIMESTAMP)
	if err != nil {
		mTime = make(map[string]string)
	}
	mCount, err := cs.cache.GetAllFields(common.KEY_PREFIX_CLASS + sClassID + ":" + common.KEY_PREFIX_EXPERIENCE + common.FIELD_COUNT)
	if err != nil {
		mCount = make(map[string]string)
	}

	r := `"` + common.FIELD_LOG + `":[` + s + `],` + common.FIELD_USER + `:{`
	first := true
	for sPhone, sCount := range mCount {
		sTime, okay := mTime[sPhone]
		if !okay {
			sTime = "0"
		}

		if first {
			first = false
		} else {
			r += `,`
		}

		r += `"` + sPhone + `":{"` + common.FIELD_TIMESTAMP + `":` + sTime + `,"` + common.FIELD_COUNT + `":` + sCount + `}`
	}
	r += `}`

	return r, nil
}

//----------------------------------------------------------------------------
