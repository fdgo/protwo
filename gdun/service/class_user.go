package service

import (
	"container/list"
	"database/sql"
	"github.com/wangmhgo/go-project/gdun/common"
	"strconv"
	"time"
)

//----------------------------------------------------------------------------
// Database: Required.
// Cache   : Compatible.

func (cs *ClassService) ChangeUser(userID int, classID int, isTeacher bool, add bool, session *Session) error {
	// Check requirements.
	if cs.db == nil {
		return common.ERR_NO_SERVICE
	}

	// Check authority.
	ci, err := cs.GetClass(classID, session)
	if err != nil {
		return err
	}

	fieldName := ""

	// Check whether this user exists or not.
	if isTeacher {
		if add == common.InIntArray(userID, ci.Teachers) {
			return nil
		}
		fieldName = common.FIELD_TEACHER_LIST
	} else {
		if add == common.InIntArray(userID, ci.Students) {
			return nil
		}
		fieldName = common.FIELD_STUDENT_LIST
	}

	sUserID := strconv.Itoa(userID)
	sClassID := strconv.Itoa(classID)

	sUpdateTime := common.GetTimeString()
	sUpdateIP := session.IP
	sUpdater := strconv.Itoa(session.UserID)

	if err = (func() error {

		// Create a transaction.
		tx, err := cs.db.Transaction()
		if err != nil {
			return err
		}

		//------------------------------------------------
		// 1. Update class info.

		// Get existing value.
		s := "SELECT " +
			fieldName + "," +
			common.FIELD_MEETING_LIST +
			" FROM " +
			common.TABLE_CLASS +
			" WHERE " +
			common.FIELD_ID + "=" + sClassID +
			" FOR UPDATE;"

		ul := ""
		ml := ""
		if err = tx.QueryRow(s).Scan(&ul, &ml); err != nil {
			tx.Rollback()
			return err
		}

		// Compute new value.
		changed := false
		if add {
			ul, changed = common.AddToList(sUserID, ul)
		} else {
			ul, changed = common.DeleteFromList(sUserID, ul)
		}
		if changed {
			// Update database.
			s = "UPDATE " +
				common.TABLE_CLASS +
				" SET " +
				fieldName + "='" + ul + "'," +
				common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
				common.FIELD_UPDATE_TIME + "=" + sUpdateTime + "," +
				common.FIELD_UPDATER + "=" + sUpdater +
				" WHERE " +
				common.FIELD_ID + "=" + sClassID + ";"

			if _, err = tx.Exec(s); err != nil {
				tx.Rollback()
				return err
			}

			// Update cache.
			if cs.cache != nil {
				m := make(map[string]string)

				m[fieldName] = ul
				m[common.FIELD_UPDATE_IP] = sUpdateIP
				m[common.FIELD_UPDATE_TIME] = sUpdateTime
				m[common.FIELD_UPDATER] = sUpdater

				if err = cs.cache.SetFields(common.KEY_PREFIX_CLASS+sClassID, m); err != nil {
					tx.Rollback()
					return err
				}
			}
		}

		//------------------------------------------------
		// 2. Update meeting info.

		if len(ml) > 0 {
			s = "UPDATE " +
				common.TABLE_MEETING +
				" SET " +
				fieldName + "='" + ul + "'," +
				common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
				common.FIELD_UPDATE_TIME + "=" + sUpdateTime + "," +
				common.FIELD_UPDATER + "=" + sUpdater +
				" WHERE " +
				common.FIELD_ID + " IN (" + ml + ");"

			if _, err = tx.Exec(s); err != nil {
				tx.Rollback()
				return err
			}

			if cs.cache != nil {
				m := make(map[string]string)

				m[fieldName] = ul
				m[common.FIELD_UPDATE_IP] = sUpdateIP
				m[common.FIELD_UPDATE_TIME] = sUpdateTime
				m[common.FIELD_UPDATER] = sUpdater

				arr := common.StringToStringArray(ml)
				for i := 0; i < len(arr); i++ {
					if len(arr[i]) == 0 {
						continue
					}
					if err = cs.cache.SetFields(common.KEY_PREFIX_MEETING+arr[i], m); err != nil {
						tx.Rollback()
						return err
					}
				}
			}
		}

		//------------------------------------------------
		// 3. Update user-class info.

		s = "SELECT " +
			common.FIELD_CLASS_LIST +
			" FROM " +
			common.TABLE_USER_CLASS +
			" WHERE " +
			common.FIELD_USER_ID + "=" + sUserID +
			" FOR UPDATE;"

		cl := ""
		err = tx.QueryRow(s).Scan(&cl)
		if err == nil {
			// Update this record.
			if add {
				cl, changed = common.AddToSortedIntList(classID, cl)
			} else {
				cl, changed = common.DeleteFromList(sClassID, cl)
			}
			if changed {
				// Update database.
				s = "UPDATE " +
					common.TABLE_USER_CLASS +
					" SET " +
					common.FIELD_CLASS_LIST + "='" + cl + "'," +
					common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
					common.FIELD_UPDATE_TIME + "=" + sUpdateTime + "," +
					common.FIELD_UPDATER + "=" + sUpdater +
					" WHERE " +
					common.FIELD_USER_ID + "=" + sUserID + ";"

				if _, err = tx.Exec(s); err != nil {
					tx.Rollback()
					return err
				}

				// Update cache.
				if cs.cache != nil {
					m := make(map[string]string)

					m[common.FIELD_CLASS_LIST] = cl
					m[common.FIELD_UPDATE_IP] = sUpdateIP
					m[common.FIELD_UPDATE_TIME] = sUpdateTime
					m[common.FIELD_UPDATER] = sUpdater

					if err = cs.cache.SetFields(common.KEY_PREFIX_USER+sUserID, m); err != nil {
						tx.Rollback()
						return err
					}
				}
			}
		} else if err == sql.ErrNoRows {
			// Insert a new record.
			s = "INSERT INTO " +
				common.TABLE_USER_CLASS +
				" (" +
				common.FIELD_USER_ID + "," +
				common.FIELD_CLASS_LIST + "," +
				common.FIELD_UPDATE_IP + "," +
				common.FIELD_UPDATE_TIME + "," +
				common.FIELD_UPDATER +
				") VALUES (" +
				sUserID + "," +
				"'" + sClassID + "'," +
				"'" + sUpdateIP + "'," +
				sUpdateTime + "," +
				sUpdater +
				");"

			if _, err = tx.Exec(s); err != nil {
				tx.Rollback()
				return err
			}

			if cs.cache != nil {
				m := make(map[string]string)

				m[common.FIELD_USER_ID] = sUserID
				m[common.FIELD_CLASS_LIST] = sClassID
				m[common.FIELD_UPDATE_IP] = sUpdateIP
				m[common.FIELD_UPDATE_TIME] = sUpdateTime
				m[common.FIELD_UPDATER] = sUpdater

				if err = cs.cache.SetFields(common.KEY_PREFIX_USER+sUserID, m); err != nil {
					tx.Rollback()
					return err
				}
			}
		} else {
			tx.Rollback()
			return err
		}

		//------------------------------------------------
		// 4. Update user-meeting info.

		if add && (!isTeacher) {
			s = "SELECT " +
				common.FIELD_MEETING_ID +
				" FROM " +
				common.TABLE_USER_MEETING +
				" WHERE " +
				common.FIELD_USER_ID + "=" + sUserID + ";"

			rows, err := tx.Query(s)
			if err != nil {
				tx.Rollback()
				return err
			}
			defer rows.Close()

			arr := common.StringToIntArray(ml)
			existing := make(map[int]bool)
			for i := 0; i < len(arr); i++ {
				existing[arr[i]] = true
			}

			for rows.Next() {
				id := 0
				if err = rows.Scan(&id); err != nil {
					tx.Rollback()
					return err
				}
				delete(existing, id)
			}

			if len(existing) > 0 {
				s = "INSERT INTO " +
					common.TABLE_USER_MEETING +
					" (" +
					common.FIELD_USER_ID + "," +
					common.FIELD_MEETING_ID + "," +
					common.FIELD_COURSEWARE_A + "," +
					common.FIELD_VIDEO_A + "," +
					common.FIELD_MEETING + "," +
					common.FIELD_EXAM + "," +
					common.FIELD_EXAM_CORRECT + "," +
					common.FIELD_EXAM_TOTAL + "," +
					common.FIELD_UPDATE_IP + "," +
					common.FIELD_UPDATE_TIME + "," +
					common.FIELD_UPDATER +
					")" +
					" VALUES "

				first := true
				for id, _ := range existing {
					if first {
						first = false
					} else {
						s += ","
					}

					s += "(" +
						sUserID + "," +
						strconv.Itoa(id) + "," +
						"'','',0,'',0,0," +
						"'" + sUpdateIP + "'," +
						sUpdateTime + "," +
						sUpdater +
						")"
				}
				s += ";"

				if _, err = tx.Exec(s); err != nil {
					tx.Rollback()
					return err
				}

				if cs.cache != nil {
					m := make(map[string]string)

					m[common.FIELD_COURSEWARE_A] = ""
					m[common.FIELD_VIDEO_A] = ""
					m[common.FIELD_MEETING] = "0"
					m[common.FIELD_EXAM] = ""
					m[common.FIELD_EXAM_CORRECT] = "0"
					m[common.FIELD_EXAM_TOTAL] = "0"
					m[common.FIELD_UPDATE_IP] = sUpdateIP
					m[common.FIELD_UPDATE_TIME] = sUpdateTime
					m[common.FIELD_UPDATER] = sUpdater

					for id, _ := range existing {
						err = cs.cache.SetFields(common.KEY_PREFIX_MEETING+strconv.Itoa(id)+":"+sUserID, m)
						if err != nil {
							tx.Rollback()
							return err
						}
					}
				}
			}
		}

		// Commit the transaction.
		if err = tx.Commit(); err != nil {
			tx.Rollback()
			return err
		}

		return nil
	})(); err != nil {
		return err
	}

	return nil
}

//----------------------------------------------------------------------------
// Database: Compatible.
// Cache   : Compatible.

func (cs *ClassService) QueryUsers(classID int, isTeacher bool, session *Session) (*UserInfoArray, error) {
	ci, err := cs.GetClass(classID, session)
	if err != nil {
		return nil, err
	}

	if cs.cache != nil {
		arr := []int{}
		if isTeacher {
			arr = ci.Teachers
		} else {
			if session.IsKeeper() {
				s, err := cs.GetAssociatedStudentsForKeeper(classID, session.UserID)
				if err != nil {
					return nil, err
				}

				arr = common.StringToIntArray(s)
			} else {
				arr = ci.Students
			}
		}

		if len(arr) == 0 {
			return new(UserInfoArray), nil
		}

		if uia, err := (func() (*UserInfoArray, error) {

			r := new(UserInfoArray)
			r.Users = list.New()

			for i := 0; i < len(arr); i++ {
				m, err := cs.cache.GetAllFields(common.KEY_PREFIX_USER + strconv.Itoa(arr[i]))
				if err != nil {
					return nil, err
				}

				ui := NewUserInfoFromMap(m, arr[i])
				if ui == nil {
					return nil, common.ERR_INVALID_USER
				}
				r.Users.PushBack(ui)
			}

			return r, nil
		})(); err == nil {
			return uia, nil
		}
	}

	if cs.db != nil {
		ul := ""
		if isTeacher {
			ul = common.IntArrayToString(ci.Teachers)
		} else {
			if session.IsKeeper() {
				s, err := cs.GetAssociatedStudentsForKeeper(classID, session.UserID)
				if err != nil {
					return nil, err
				}

				ul = s
			} else {
				ul = common.IntArrayToString(ci.Students)
			}
		}

		if len(ul) == 0 {
			return new(UserInfoArray), nil
		}

		s := "SELECT " +
			common.FIELD_ID + "," +
			common.FIELD_NICKNAME + "," +
			common.FIELD_REMARK + "," +
			common.FIELD_GAODUN_STUDENT_ID + "," +
			common.FIELD_GROUP_ID +
			" FROM " +
			common.TABLE_USER +
			" WHERE " +
			common.FIELD_ID + " IN (" + ul + ");"

		rows, err := cs.db.Select(s)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		uia := new(UserInfoArray)
		uia.Users = list.New()

		for rows.Next() {
			ui := new(UserInfo)
			err = rows.Scan(&ui.ID, &ui.Nickname, &ui.Remark, &ui.GdStudentID, &ui.GroupID)
			if err != nil {
				return nil, err
			}
			uia.Users.PushBack(ui)
		}

		return uia, nil
	}

	return nil, common.ERR_NO_SERVICE
}

//----------------------------------------------------------------------------

func (cs *ClassService) ChangeKeeperStudentRelation(classID int, studentID int, keeperID int, add bool, session *Session) error {
	// Check requirements.
	if cs.db == nil {
		return common.ERR_NO_SERVICE
	}

	// Check authority.
	_, err := cs.GetClass(classID, session)
	if err != nil {
		return err
	}

	sClassID := strconv.Itoa(classID)
	sStudentID := strconv.Itoa(studentID)
	sKeeperID := strconv.Itoa(keeperID)

	key := common.KEY_PREFIX_CLASS + sClassID + ":" + common.FIELD_KEEPER

	// Check via cache.
	if cs.cache != nil {
		if err := (func() error {
			sl, err := cs.cache.GetField(key, sKeeperID)
			if err != nil {
				return err
			}

			if add == common.InList(sStudentID, sl) {
				return nil
			}

			return common.ERR_NO_USER
		})(); err == nil {
			return nil
		}
	}

	if err := (func() error {
		tx, err := cs.db.Transaction()
		if err != nil {
			return nil
		}

		s := "SELECT " +
			common.FIELD_STUDENT_LIST +
			" FROM " +
			common.TABLE_KEEPER_CLASS +
			" WHERE " +
			common.FIELD_USER_ID + "=" + sKeeperID + " AND " +
			common.FIELD_CLASS_ID + "=" + sClassID + " FOR UPDATE;"

		row := tx.QueryRow(s)

		sl := ""
		err = row.Scan(&sl)
		if err != nil {
			if err != sql.ErrNoRows {
				tx.Rollback()
				return err
			}

			s = "INSERT INTO " +
				common.TABLE_KEEPER_CLASS +
				" VALUES (" +
				sKeeperID + "," +
				sClassID + "," +
				sStudentID + ");"

			if _, err = tx.Exec(s); err != nil {
				tx.Rollback()
				return err
			}
		} else {
			// Compute new value.
			changed := false
			if add {
				sl, changed = common.AddToList(sStudentID, sl)
			} else {
				sl, changed = common.DeleteFromList(sStudentID, sl)
			}

			if !changed {
				tx.Rollback()

				if cs.cache != nil {
					if err = cs.cache.SetField(key, sKeeperID, sl); err != nil {
						return err
					}
				}
				return nil
			}

			s = "UPDATE " +
				common.TABLE_KEEPER_CLASS +
				" SET " +
				common.FIELD_STUDENT_LIST + "='" + sl + "'" +
				" WHERE " +
				common.FIELD_USER_ID + "=" + sKeeperID + " AND " +
				common.FIELD_CLASS_ID + "=" + sClassID + ";"

			if _, err = tx.Exec(s); err != nil {
				tx.Rollback()
				return err
			}
		}

		if cs.cache != nil {
			if err := cs.cache.SetField(key, sKeeperID, sl); err != nil {
				tx.Rollback()
				return err
			}
		}

		if err = tx.Commit(); err != nil {
			tx.Rollback()
			return err
		}

		return nil
	})(); err != nil {
		return err
	}

	return nil
}

//----------------------------------------------------------------------------

func (cs *ClassService) RefreshKeeperList(classID int) error {
	sClassID := strconv.Itoa(classID)

	// Get keeper list.
	m, err := (func() (map[string]string, error) {
		key := common.KEY_PREFIX_CLASS + sClassID + ":" + common.FIELD_KEEPER

		if cs.cache != nil {
			m, err := cs.cache.GetAllFields(key)
			if err == nil {
				return m, nil
			}
		}

		if cs.db != nil {
			s := "SELECT " +
				common.FIELD_USER_ID + "," +
				common.FIELD_STUDENT_LIST +
				" FROM " +
				common.TABLE_KEEPER_CLASS +
				" WHERE " +
				common.FIELD_CLASS_ID + "=" + sClassID + ";"

			rows, err := cs.db.Select(s)
			if err != nil {
				return nil, err
			}
			defer rows.Close()

			id := 0
			sl := ""
			m := make(map[string]string)
			for rows.Next() {
				err = rows.Scan(&id, &sl)
				if err != nil {
					return m, err
				}

				m[strconv.Itoa(id)] = sl
			}

			if cs.cache != nil {
				if err = cs.cache.SetFields(key, m); err != nil {
					// TODO:
				}
			}

			return m, nil
		}

		return nil, common.ERR_NO_SERVICE
	})()
	if err != nil {
		return err
	}

	// Set keeper list.
	if err := (func() error {
		r := ``
		first := true
		for id, sl := range m {
			if first {
				first = false
			} else {
				r += `,`
			}

			r += `"` + id + `":[` + sl + `]`
		}

		s := "UPDATE " +
			common.TABLE_CLASS +
			" SET " +
			common.FIELD_KEEPER_LIST + "='" + r + "'" +
			" WHERE " +
			common.FIELD_ID + "=" + sClassID + ";"

		if _, err := cs.db.Exec(s); err != nil {
			return err
		}

		if cs.cache != nil {
			if err := cs.cache.SetField(common.KEY_PREFIX_CLASS+sClassID, common.FIELD_KEEPER_LIST, r); err != nil {
				return err
			}
		}

		return nil
	})(); err != nil {
		return err
	}

	return nil
}

//----------------------------------------------------------------------------

// func (cs *ClassService) GetCachedUserProgress(classID int, userID int) (string, error) {
// 	return "", nil
// }

func (cs *ClassService) QueryUserProgress(classID int, userID int, session *Session) (*UserMeetingProgressInfoArray, error) {
	// Check authority.
	if session.IsStudent() {
		if userID != session.UserID {
			return nil, common.ERR_NO_AUTHORITY
		}
	}

	ci, err := cs.GetClass(classID, session)
	if err != nil {
		return nil, err
	}

	umpia := new(UserMeetingProgressInfoArray)
	umpia.Status = list.New()

	for i := 0; i < len(ci.Meetings); i++ {
		umpi, err := cs.ms.GetUserProgress(ci.Meetings[i], userID)
		if err != nil {
			return umpia, err
		}

		umpia.Status.PushBack(umpi)
	}

	return umpia, nil
}

//----------------------------------------------------------------------------

func (cs *ClassService) QueryUserProgresses(classID int, session *Session) (string, error) {
	ci, err := cs.GetClass(classID, session)
	if err != nil {
		return "", err
	}

	mia, err := cs.ms.GetMeetings(ci.Meetings, session, false)
	if err != nil {
		return "", err
	}

	r := `"` + common.FIELD_PROGRESS + `":{`
	firstMeeting := true
	now := int(time.Now().Unix())
	for e := mia.Meetinigs.Front(); e != nil; e = e.Next() {
		mi, okay := e.Value.(*MeetingInfo)
		if !okay {
			continue
		}

		if mi.StartTime-now > 10800 {
			continue
		}

		if firstMeeting {
			firstMeeting = false
		} else {
			r += `,`
		}

		r += `"` + strconv.Itoa(mi.ID) + `":{`
		firstStudent := true
		for i := 0; i < len(ci.Students); i++ {
			umpi, err := cs.ms.GetUserProgress(mi.ID, ci.Students[i])
			if err != nil {
				continue
			}

			if firstStudent {
				firstStudent = false
			} else {
				r += `,`
			}

			r += `"` + strconv.Itoa(umpi.UserID) + `":{`
			r += umpi.ToJSON(false)
			r += `}`
		}
		r += `}`
	}
	r += `}`

	return r, nil
}

func (cs *ClassService) GetCachedUserProgresses(meetingIDs []int, studentIDs []int) (string, error) {
	r := `"` + common.FIELD_PROGRESS + `":{`
	firstMeeting := true
	for i := 0; i < len(meetingIDs); i++ {
		if firstMeeting {
			firstMeeting = false
		} else {
			r += `,`
		}

		firstStudent := true
		r += `"` + strconv.Itoa(meetingIDs[i]) + `":{`
		for j := 0; j < len(studentIDs); j++ {
			s, err := cs.ms.GetCachedUserProgress(meetingIDs[i], studentIDs[j], false)
			if err != nil {
				return "", err
			}

			if firstStudent {
				firstStudent = false
			} else {
				r += `,`
			}
			r += `"` + strconv.Itoa(studentIDs[j]) + `":{` + s + `}`
		}
		r += `}`
	}
	r += `}`

	return r, nil
}

func (cs *ClassService) QueryUserProgressesA(classID int, session *Session) (string, error) {
	// Check authority.
	ci, err := cs.GetClass(classID, session)
	if err != nil {
		return "", err
	}

	// Get student list.
	var students []int = nil
	if session.IsKeeper() {
		sl, err := cs.GetAssociatedStudentsForKeeper(classID, session.UserID)
		if err != nil {
			return "", err
		}

		students = common.StringToIntArray(sl)
	} else {
		students = ci.Students
	}

	now := int(time.Now().Unix())

	r := `"` + common.FIELD_PROGRESS + `":{`
	firstMeeting := true

	// Visit each meeting, respectively.
	for i := 0; i < len(ci.Meetings); i++ {
		if cs.cache != nil {
			if s, err := cs.cache.GetField(common.KEY_PREFIX_MEETING+strconv.Itoa(ci.Meetings[i]), common.FIELD_START_TIME); err == nil {
				if startTime, err := strconv.Atoi(s); err == nil {
					if startTime-now > 10800 {
						continue
					}
				}
			}
		}

		if firstMeeting {
			firstMeeting = false
		} else {
			r += `,`
		}

		firstStudent := true
		r += `"` + strconv.Itoa(ci.Meetings[i]) + `":{`
		for j := 0; j < len(students); j++ {
			umpi, err := cs.ms.GetUserProgress(ci.Meetings[i], students[j])
			if err != nil {
				continue
			}

			if firstStudent {
				firstStudent = false
			} else {
				r += `,`
			}
			r += `"` + strconv.Itoa(students[j]) + `":{` + umpi.ToJSON(false) + `}`
		}
		r += `}`
	}
	r += `}`

	return r, nil
}

//----------------------------------------------------------------------------

func (cs *ClassService) GetAssociatedStudentsForKeeper(classID int, userID int) (string, error) {
	sClassID := strconv.Itoa(classID)
	sKeeperID := strconv.Itoa(userID)

	if cs.cache != nil {
		sl, err := cs.cache.GetField(common.KEY_PREFIX_CLASS+sClassID+":"+common.FIELD_KEEPER, sKeeperID)
		if err == nil {
			return sl, nil
		}
	}

	if cs.db != nil {
		s := "SELECT " +
			common.FIELD_STUDENT_LIST +
			" FROM " +
			common.TABLE_KEEPER_CLASS +
			" WHERE " +
			common.FIELD_USER_ID + "=" + sKeeperID + " AND " +
			common.FIELD_CLASS_ID + "=" + sClassID + ";"

		rows, err := cs.db.Select(s)
		if err != nil {
			return "", err
		}
		defer rows.Close()

		if !rows.Next() {
			return "", common.ERR_NO_AUTHORITY
		}

		sl := ""
		if err = rows.Scan(&sl); err != nil {
			return "", err
		}

		if cs.cache != nil {
			if err = cs.cache.SetField(common.KEY_PREFIX_CLASS+sClassID+":"+common.FIELD_KEEPER, sKeeperID, sl); err != nil {
				// TODO:
			}
			if err = cs.RefreshKeeperList(classID); err != nil {
				// TODO:
			}
		}

		return sl, nil
	}

	return "", common.ERR_NO_SERVICE
}

//----------------------------------------------------------------------------

func (cs *ClassService) QueryClassBriefs(session *Session) (string, error) {
	cl, err := cs.GetUserClassRelation(session.UserID)
	if err != nil {
		return "", nil
	}

	r := `"` + common.FIELD_CLASS_LIST + `":{`
	first := true

	classIDs := common.StringToIntArray(cl)
	for i := 0; i < len(classIDs); i++ {
		ci, err := cs.GetClass(classIDs[i], session)
		if err != nil {
			continue
		}
		if ci.GdCourseID == 0 {
			continue
		}

		if first {
			first = false
		} else {
			r += `,`
		}

		r += `"` + strconv.Itoa(ci.GdCourseID) + `":{` +
			`"` + common.FIELD_NAME + `":"` + common.UnescapeForJSON(ci.Name) + `",` +
			`"` + common.FIELD_START_TIME + `":` + strconv.Itoa(ci.StartTime*1000) + `,` +
			`"` + common.FIELD_END_TIME + `":` + strconv.Itoa(ci.EndTime*1000) + `,` +
			`"` + common.FIELD_NEXT_TIME + `":` + strconv.Itoa(ci.NextTime*1000) + `,` +
			`"` + common.FIELD_IS_RUNNING + `":0,` +
			`"` + common.FIELD_IS_TODAY + `":0,` +
			`"` + common.FIELD_NUMBER_OF_FINISHED_MEETING + `":` + strconv.Itoa(ci.NumberOfFinishedMeeting) + `,` +
			`"` + common.FIELD_NUMBER_OF_MEETING + `":` + strconv.Itoa(len(ci.Meetings)) +
			`}`
	}
	r += `}`

	return r, nil
}

//----------------------------------------------------------------------------
