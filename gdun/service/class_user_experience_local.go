package service

import (
	"github.com/wangmhgo/go-project/gdun/common"
	"strconv"
	"strings"
	"time"
	// "fmt"
)

//----------------------------------------------------------------------------

func (cs *ClassService) checkVerificationCode(phone int, code int) error {
	// TODO:
	return nil

	s, err := cs.cache.GetKey(common.KEY_PREFIX_VERIFICATION + strconv.Itoa(phone))
	if err != nil {
		return err
	}

	n, err := strconv.Atoi(s)
	if err != nil {
		return err
	}

	if n != code {
		return common.ERR_INVALID_TOKEN
	}

	return nil
}

func (cs *ClassService) getClassInvitationInfo(classID int, teacherID int, token string, endTime int) (*ClassInvitationInfo, error) {
	var err error
	var s string

	sTeacherID := strconv.Itoa(teacherID)
	key := common.KEY_PREFIX_CLASS + strconv.Itoa(classID) + ":"
	if teacherID > 0 {
		s, err = cs.cache.GetField(key+common.FIELD_CHANNEL, sTeacherID)
	} else {
		s, err = cs.cache.GetField(key+common.FIELD_TOKEN, token)
	}
	if err != nil {
		return nil, err
	}

	// Get each value, respectively.
	arr := strings.Split(s, ":")
	if len(arr) < 5 {
		return nil, common.ERR_INVALID_KEY
	}

	cii := new(ClassInvitationInfo)
	if cii.UpdateTime, err = strconv.Atoi(arr[0]); err != nil {
		return nil, err
	}
	if cii.EndTime, err = strconv.Atoi(arr[1]); err != nil {
		return nil, err
	}
	if endTime != cii.EndTime {
		return nil, common.ERR_INVALID_KEY
	}
	if cii.Duration, err = strconv.Atoi(arr[2]); err != nil {
		return nil, err
	}
	if teacherID == 0 {
		if len(arr) < 6 {
			return nil, common.ERR_INVALID_KEY
		}
		cii.Channel = arr[3]
		if cii.Updater, err = strconv.Atoi(arr[4]); err != nil {
			return nil, err
		}
		cii.UpdateIP = arr[5]
	} else {
		cii.Teacher = teacherID
		if cii.Updater, err = strconv.Atoi(arr[3]); err != nil {
			return nil, err
		}
		cii.UpdateIP = arr[4]
	}

	// Check whether this invitation is expired or not.
	now := int(time.Now().Unix())
	if now >= cii.EndTime {
		// Delete this invitation.
		if teacherID == 0 {
			if err = cs.cache.DelField(key+common.FIELD_CHANNEL, sTeacherID); err != nil {
				// TODO:
			}
		} else {
			if err = cs.cache.DelField(key+common.FIELD_TOKEN, token); err != nil {
				// TODO:
			}
		}
		return nil, common.ERR_OUT_OF_TIME
	}

	if cii.Duration == 0 {
		cii.ExpiredTime = cii.EndTime
	} else {
		cii.ExpiredTime = now + cii.Duration
	}

	// Remove this token.
	if teacherID == 0 {
		if err = cs.cache.DelField(key+common.FIELD_TOKEN, token); err != nil {
			// TODO:
		}
	}

	return cii, nil
}

//----------------------------------------------------------------------------

func (cs *ClassService) buildInfoForExperienceUser(userID int, phone int, ip string) *UserInfo {
	ui := new(UserInfo)
	ui.ID = userID
	ui.Nickname = strconv.Itoa(phone/100000000) + "****" + strconv.Itoa(phone%10000)
	ui.GroupID = common.GROUP_ID_FOR_STUDENT

	(func() {
		sUserID := strconv.Itoa(ui.ID)

		// Assign a session.
		key := common.KEY_PREFIX_SESSION + sUserID

		m := make(map[string]string)
		m[common.FIELD_ID] = sUserID
		m[common.FIELD_GROUP] = strconv.Itoa(ui.GroupID)
		m[common.FIELD_NICKNAME] = ui.Nickname
		m[common.FIELD_TOKEN] = ""
		m[common.FIELD_IP] = ip

		if err := cs.cache.SetFields(key, m); err != nil {
			// TODO:
		}
	})()

	return ui
}

//----------------------------------------------------------------------------

func (cs *ClassService) getOrCreateIDForExperienceUser(phone int, ip string) (*UserInfo, error) {
	key := common.KEY_PREFIX_EXPERIENCE + strconv.Itoa(phone)

	// Check whether such a user exists or not.
	if s, err := cs.cache.GetKey(key); err == nil {
		if id, err := strconv.Atoi(s); err == nil {
			return cs.buildInfoForExperienceUser(id, phone, ip), nil
		}
	}

	var id int
	n, err := cs.cache.Incr(common.COUNTER_TEMPERARY_USER)
	if err != nil {
		// TODO: Assign a random number or return error straightly.
		id = common.VALUE_MINIMAL_TEMPERARY_USER_ID
	} else {
		id = int(n)
		if id < common.VALUE_MINIMAL_TEMPERARY_USER_ID {
			id += common.VALUE_MINIMAL_TEMPERARY_USER_ID
		}
	}

	// Set up a map from phone to user ID.
	if err = cs.cache.SetKey(key, strconv.Itoa(id)); err != nil {
		// TODO: DO WHAT?
	}

	return cs.buildInfoForExperienceUser(id, phone, ip), nil
}

//----------------------------------------------------------------------------

func (cs *ClassService) authorizeClassToExperienceUser(userID int, classID int, expiredTime int) error {
	now := int(time.Now().Unix())
	if now >= expiredTime {
		return common.ERR_OUT_OF_TIME
	}

	sUserID := strconv.Itoa(userID)
	sClassID := strconv.Itoa(classID)

	// TODO: Check whether this class had been registered before hand.

	sExpiredTime := strconv.Itoa(expiredTime)
	err := cs.cache.SetField(common.KEY_PREFIX_USER+sUserID+":"+common.KEY_PREFIX_EXPERIENCE+common.FIELD_CLASS, sClassID, sExpiredTime)
	if err != nil {
		return err
	}

	return nil
}

func (cs *ClassService) checkAuthorityOfExperienceUser(userID int, classID int) error {
	sUserID := strconv.Itoa(userID)
	sClassID := strconv.Itoa(classID)

	// Check via class list.
	// cl, err := cs.cache.GetField(common.KEY_PREFIX_USER+sUserID, common.FIELD_CLASS_LIST)
	// if err != nil {
	// 	return err
	// }
	// if common.InList(sClassID, cl) {
	// 	return nil
	// }

	// Check via expired time.
	s, err := cs.cache.GetField(common.KEY_PREFIX_USER+sUserID+":"+common.KEY_PREFIX_EXPERIENCE+common.FIELD_CLASS, sClassID)
	if err != nil {
		return err
	}
	expiredTime, err := strconv.Atoi(s)
	if err != nil {
		return err
	}

	now := int(time.Now().Unix())
	if now >= expiredTime {
		return common.ERR_OUT_OF_TIME
	}

	return nil
}

func (cs *ClassService) queryClassListForExperienceUser(userID int) []int {
	m, err := cs.cache.GetAllFields(common.KEY_PREFIX_USER + strconv.Itoa(userID) + ":" + common.KEY_PREFIX_EXPERIENCE + common.FIELD_CLASS)
	if err != nil {
		return []int{}
	}

	arr := make([]int, len(m))
	i := 0
	for id, _ := range m {
		arr[i], err = strconv.Atoi(id)
		if err == nil {
			i++
		}
	}

	return arr
}

//----------------------------------------------------------------------------

func (cs *ClassService) recordRegisterOfExperienceUser(classID int, token string, userID int, phone int, cii *ClassInvitationInfo, ip string) error {
	key := common.KEY_PREFIX_CLASS + strconv.Itoa(classID) + ":"
	if cii.Teacher == 0 {
		key += common.FIELD_TOKEN
	} else {
		key += common.FIELD_CHANNEL
	}
	key += ":" + common.FIELD_LOG

	okay, err := cs.cache.Exists(key)
	if err != nil {
		return err
	}

	s := `{` +
		`"` + common.FIELD_END_TIME + `":` + strconv.Itoa(cii.EndTime*1000) + `,` +
		`"` + common.FIELD_DURATION + `":` + strconv.Itoa(cii.Duration*1000) + `,` +
		`"` + common.FIELD_UPDATER + `":` + strconv.Itoa(cii.Updater) + `,` +
		`"` + common.FIELD_UPDATE_IP + `":"` + common.UnescapeForJSON(cii.UpdateIP) + `",`

	if cii.Teacher == 0 {
		s += `"` + common.FIELD_TOKEN + `":"` + token + `",` +
			`"` + common.FIELD_CHANNEL + `":"` + common.UnescapeForJSON(cii.Channel) + `",`
	} else {
		s += `"` + common.FIELD_TEACHER + `":` + strconv.Itoa(cii.Teacher) + `,`
	}
	s += `"` + common.FIELD_UPDATE_TIME + `":` + strconv.Itoa(cii.UpdateTime*1000) + `,` +
		`"` + common.FIELD_PHONE + `":` + strconv.Itoa(phone) + `,` +
		`"` + common.FIELD_IP + `":"` + ip + `",` +
		`"` + common.FIELD_TIMESTAMP + `":` + common.GetTimeString() + `,` +
		`"` + common.FIELD_USER + `":` + strconv.Itoa(userID) +
		`}`

	if okay {
		if err = cs.cache.Append(key, `,`+s); err != nil {
			return err
		}
	} else {
		if err = cs.cache.SetKey(key, s); err != nil {
			return err
		}
	}

	return nil
}

//----------------------------------------------------------------------------

func (cs *ClassService) recordLoginOfExperienceUser(classID int, phone int) error {
	sClassID := strconv.Itoa(classID)
	sPhone := strconv.Itoa(phone)

	err1 := cs.cache.SetField(common.KEY_PREFIX_CLASS+sClassID+":"+common.KEY_PREFIX_EXPERIENCE+common.FIELD_TIMESTAMP, sPhone, common.GetTimeString())
	_, err2 := cs.cache.IncrField(common.KEY_PREFIX_CLASS+sClassID+":"+common.KEY_PREFIX_EXPERIENCE+common.FIELD_COUNT, sPhone)

	if err1 != nil {
		return err1
	} else if err2 != nil {
		return err2
	} else {
		return nil
	}
}

//----------------------------------------------------------------------------
