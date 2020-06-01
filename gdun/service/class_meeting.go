package service

import (
	"github.com/wangmhgo/go-project/gdun/common"
	//"gitlab.hfjy.com/gdun/vender/bokecc"
	"sort"
	"strconv"
	"time"
)

//----------------------------------------------------------------------------
// Database: Required.
// Cache   : Compatible.

func (cs *ClassService) AddMeeting(name string, subjects []int, section int, startTime int, duration int, classID int, t int, data string, session *Session, unset bool) (int, error) {
	// Check requirements.
	if cs.db == nil || cs.ms == nil {
		return 0, common.ERR_NO_SERVICE
	}

	//----------------------------------------------------
	// Check authority.

	ci, err := cs.GetClass(classID, session)
	if err != nil {
		return 0, err
	}

	//----------------------------------------------------
	// Create a meeting.

	meetingID, err := cs.ms.AddMeeting(name, subjects, section /*common.IntArrayToString(ci.Teachers),*/, common.IntArrayToString(ci.Students), startTime, duration, t, data, ci.ID, ci.GroupID, session, unset)
	if err != nil {
		return 0, err
	}

	//----------------------------------------------------
	// Update class information.

	sMeetingID := strconv.Itoa(meetingID)
	sClassID := strconv.Itoa(classID)

	sUpdater := strconv.Itoa(session.UserID)
	sUpdateIP := session.IP
	sUpdateTime := common.GetTimeString()

	if err = (func() error {
		// Create a transaction.
		tx, err := cs.db.Transaction()
		if err != nil {
			return err
		}

		// Get exiting meeting list.
		sql := "SELECT " +
			common.FIELD_MEETING_LIST +
			" FROM " +
			common.TABLE_CLASS +
			" WHERE " +
			common.FIELD_ID + "=" + sClassID +
			" FOR UPDATE;"

		row := tx.QueryRow(sql)

		ml := ""
		if err = row.Scan(&ml); err != nil {
			tx.Rollback()
			return err
		}

		// Compute the new meeting list.
		ml, _ = common.AddToList(sMeetingID, ml)

		// Update database.
		sql = "UPDATE " +
			common.TABLE_CLASS +
			" SET " +
			common.FIELD_MEETING_LIST + "='" + ml + "'," +
			common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
			common.FIELD_UPDATE_TIME + "=" + sUpdateTime + "," +
			common.FIELD_UPDATER + "=" + sUpdater +
			" WHERE " +
			common.FIELD_ID + "=" + sClassID + ";"

		if _, err = tx.Exec(sql); err != nil {
			tx.Rollback()
			return err
		}

		// Update cache.
		if cs.cache != nil {
			m := make(map[string]string)

			m[common.FIELD_MEETING_LIST] = ml
			m[common.FIELD_UPDATER] = sUpdater
			m[common.FIELD_UPDATE_IP] = sUpdateIP
			m[common.FIELD_UPDATE_TIME] = sUpdateTime

			if err = cs.cache.SetFields(common.KEY_PREFIX_CLASS+sClassID, m); err != nil {
				tx.Rollback()
				return err
			}
		}

		// Commit it.
		if err = tx.Commit(); err != nil {
			tx.Rollback()
			return err
		}

		return nil
	})(); err != nil {
		return meetingID, err
	}

	if err = cs.UpdateMeetingTime(classID, session); err != nil {
		// TODO:
	}
	// if err = cs.UpdateTimeOfNextMeeting(classID, session); err != nil {
	// 	// TODO:
	// }

	return meetingID, nil
}

func (cs *ClassService) UpdateMeetingTime(classID int, session *Session) error {
	if cs.cache == nil {
		return common.ERR_NO_CACHE
	}

	// Check authority.
	ci, err := cs.GetClass(classID, session)
	if err != nil {
		return err
	}

	key := common.KEY_PREFIX_CLASS + strconv.Itoa(classID)

	// If no meeting left.
	if len(ci.Meetings) == 0 {
		if err = cs.cache.SetField(key, common.FIELD_NEXT_TIME, ""); err != nil {
			// TODO:
		}
		if err = cs.cache.SetField(key, common.FIELD_SCHEDULE, ""); err != nil {
			// TODO:
		}

		return nil
	}

	// Sort meeting start times.
	arr := make([]*MeetingInfo, len(ci.Meetings))
	for i := 0; i < len(ci.Meetings); i++ {
		// arr[i], err = cs.ms.GetMeetingTime(ci.Meetings[i])
		// if err != nil {
		// 	return err
		// }
		arr[i], err = cs.ms.GetMeeting(ci.Meetings[i], session, false)
		if err != nil {
			return err
		}
	}
	// sort.Ints(arr)
	sort.Sort(MeetingInfoSlice(arr))

	// Find next time.
	now := int(time.Now().Unix())
	nt := 0
	for i := 0; i < len(arr); i++ {
		if arr[i].StartTime >= now {
			nt = arr[i].StartTime
			break
		}
	}

	// Update cache.
	if err = cs.cache.SetField(key, common.FIELD_NEXT_TIME, strconv.Itoa(nt)); err != nil {
		// TODO:
	}
	if err = cs.cache.SetField(key, common.FIELD_SCHEDULE, MeetingInfoSlice(arr).ToJSON()); err != nil {
		// TODO:
	}

	return nil
}

//----------------------------------------------------------------------------
// Database: Required.
// Cache   : Compatible.

func (cs *ClassService) CopyMeetingTo(meetingID int, destClassID int, session *Session) error {
	// Check requirements.
	if cs.db == nil {
		return common.ERR_NO_SERVICE
	}

	// Get the source meeting.
	mi, err := cs.ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return err
	}

	// Get the destination class.
	dest, err := cs.GetClass(destClassID, session)
	if err != nil {
		return err
	}

	// Create a new meeting.
	newMeetingID, err := cs.AddMeeting(common.Unescape(mi.Name), mi.Subjects, mi.Section, mi.StartTime, mi.Duration, dest.ID, mi.Type, common.Unescape(mi.Data), session, true)
	if err != nil {
		return err
	}

	// Add coursewares, exams, replays, and videos to the new meeting.
	if err = (func() error {
		cwl := common.StringArrayToString(mi.Coursewares)
		el := common.StringArrayToString(mi.Exams)
		vl := common.StringArrayToString(mi.Videos)

		sMeetingID := strconv.Itoa(newMeetingID)
		sUpdater := strconv.Itoa(session.UserID)
		sUpdateIP := session.IP
		sUpdateTime := common.GetTimeString()

		s := "UPDATE " +
			common.TABLE_MEETING +
			" SET " +
			common.FIELD_COURSEWARE_LIST + "='" + cwl + "'," +
			common.FIELD_EXAM_LIST + "='" + el + "'," +
			common.FIELD_VIDEO_LIST + "='" + vl + "'," +
			common.FIELD_UPDATER + "=" + sUpdater + "," +
			common.FIELD_UPDATE_TIME + "=" + sUpdateTime + "," +
			common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'" +
			" WHERE " +
			common.FIELD_ID + "=" + sMeetingID + ";"

		if _, err = cs.db.Exec(s); err != nil {
			return err
		}

		if cs.cache != nil {
			m := make(map[string]string)
			m[common.FIELD_COURSEWARE_LIST] = cwl
			m[common.FIELD_EXAM_LIST] = el
			m[common.FIELD_VIDEO_LIST] = vl
			m[common.FIELD_UPDATER] = sUpdater
			m[common.FIELD_UPDATE_IP] = sUpdateIP
			m[common.FIELD_UPDATE_TIME] = sUpdateTime

			if err = cs.cache.SetFields(common.KEY_PREFIX_MEETING+sMeetingID, m); err != nil {
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
// Database: Required.
// Cache   : Compatible.

func (cs *ClassService) EndMeeting(classID int, meetingID int, session *Session) error {
	// Check requirements.
	if cs.db == nil || cs.ms == nil {
		return common.ERR_NO_SERVICE
	}

	// Check authority.
	ci, err := cs.GetClass(classID, session)
	if err != nil {
		return err
	}
	if !common.InIntArray(meetingID, ci.Meetings) {
		return common.ERR_NO_MEETING
	}
	mi, err := cs.ms.GetMeeting(meetingID, session, false)
	if err != nil {
		return err
	}

	// End the meeting.
	err = cs.ms.EndMeeting(meetingID, session)
	if err != nil {
		return err
	}

	sClassID := strconv.Itoa(classID)

	sUpdateIP := session.IP
	sUpdateTime := common.GetTimeString()
	sUpdater := strconv.Itoa(session.UserID)

	s := "UPDATE " +
		common.TABLE_CLASS +
		" SET " +
		common.FIELD_NUMBER_OF_FINISHED_MEETING + "=" + common.FIELD_NUMBER_OF_FINISHED_MEETING + "+1," +
		common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
		common.FIELD_UPDATE_TIME + "=" + sUpdateTime + "," +
		common.FIELD_UPDATER + "=" + sUpdater +
		" WHERE " +
		common.FIELD_ID + "=" + sClassID + ";"

	if _, err = cs.db.Exec(s); err != nil {
		return err
	}

	if cs.cache != nil {
		m := make(map[string]string)

		m[common.FIELD_NUMBER_OF_FINISHED_MEETING] = strconv.Itoa(ci.NumberOfFinishedMeeting + 1)
		m[common.FIELD_UPDATER] = sUpdater
		m[common.FIELD_UPDATE_IP] = sUpdateIP
		m[common.FIELD_UPDATE_TIME] = sUpdateTime

		if err = cs.cache.SetFields(common.KEY_PREFIX_CLASS+sClassID, m); err != nil {
			return err
		}
	}

	if ci.PlatformID == common.PLATFORM_ID_FOR_BOKECC {
		if err = cs.updateMeetingProgressViaBokecc(ci, mi, session); err != nil {
			// TODO:
		}
	}

	if err = cs.UpdateMeetingTime(classID, session); err != nil {
		// TODO:
	}
	// if err = cs.UpdateTimeOfNextMeeting(classID, session); err != nil {
	// 	// TODO:
	// }

	return nil
}

func (cs *ClassService) updateMeetingProgressViaBokecc(ci *ClassInfo, mi *MeetingInfo, session *Session) error {
	//layout := "2006-01-02 15:04:05"
	//
	//start := time.Unix(int64(mi.StartTime), 0)
	//// end := time.Now()
	//end := time.Unix(int64(mi.StartTime+mi.Duration), 0)

	//n, resp := bokecc.GetRoomStatistics(ci.PlatformData, start.Format(layout), end.Format(layout))
	//if n != 0 {
	//	return common.ERR_NO_RECORD
	//}
	//
	//log := make(map[int]string)
	//m := make(map[int]int)
	//for i := 0; i < len(resp.UserActions); i++ {
	//	ua := resp.UserActions[i]
	//
	//	// Retrieve class ID and user ID.
	//	sClassID := ""
	//	sUserID := ""
	//	arr := strings.Split(ua.UserID, "_")
	//	if len(arr) == 2 {
	//		sClassID = arr[0]
	//		sUserID = arr[1]
	//	} else {
	//		sUserID = ua.UserID
	//	}
	//
	//	// Check class ID.
	//	if len(sClassID) > 0 {
	//		classID, err := strconv.Atoi(sClassID)
	//		if err != nil || classID != ci.ID {
	//			continue
	//		}
	//	}
	//
	//	// Check user ID.
	//	userID, err := strconv.Atoi(sUserID)
	//	if err != nil || userID <= 0 {
	//		continue
	//	}
	//	if !common.InIntArray(userID, ci.Students) {
	//		continue
	//	}
	//
	//	start, err = time.Parse(layout, ua.EnterTime)
	//	if err != nil {
	//		continue
	//	}
	//	end, err = time.Parse(layout, ua.LeaveTime)
	//	if err != nil {
	//		continue
	//	}
	//
	//	s := `{"` + common.FIELD_START_TIME + `":` + strconv.FormatInt(start.Unix()*1000, 10) + `,` +
	//		`"` + common.FIELD_END_TIME + `":` + strconv.FormatInt(end.Unix()*1000, 10) + `,` +
	//		`"` + common.FIELD_IP + `":"` + ua.UserIP + `"}`
	//
	//	d := int(end.Unix() - start.Unix())
	//	t, okay := m[userID]
	//	if okay {
	//		m[userID] = t + d
	//	} else {
	//		m[userID] = d
	//	}
	//
	//	l, okay := log[userID]
	//	if okay {
	//		log[userID] = l + "," + s
	//	} else {
	//		log[userID] = s
	//	}
	//}
	//for studentID, seconds := range m {
	//	s, okay := log[studentID]
	//	if !okay {
	//		s = ""
	//	}
	//	if err := cs.ms.setMeetingProgress(mi.ID, studentID, seconds, s, session); err != nil {
	//		// TODO:
	//	}
	//}

	return nil
}

//----------------------------------------------------------------------------
// Database: Required.
// Cache   : Compatible.

func (cs *ClassService) DeleteMeeting(meetingID int, classID int, session *Session) error {
	// Check requirements.
	if cs.db == nil {
		return common.ERR_NO_SERVICE
	}

	ci, err := cs.GetClass(classID, session)
	if err != nil {
		return err
	}
	if !session.IsSystem() {
		if ci.EndTime > 0 {
			return common.ERR_CLASS_CLOSED
		}
	}
	if !common.InIntArray(meetingID, ci.Meetings) {
		return common.ERR_NO_MEETING
	}

	mi, err := cs.ms.GetMeeting(meetingID, session, false)
	if err != nil {
		return err
	}

	sClassID := strconv.Itoa(classID)
	sMeetingID := strconv.Itoa(meetingID)

	sUpdater := strconv.Itoa(session.UserID)
	sUpdateIP := session.IP
	sUpdateTime := common.GetTimeString()

	if err = (func() error {
		// Create a transaction.
		tx, err := cs.db.Transaction()
		if err != nil {
			return err
		}

		sql := "SELECT " +
			common.FIELD_MEETING_LIST + "," +
			common.FIELD_DELETED +
			" FROM " +
			common.TABLE_CLASS +
			" WHERE " +
			common.FIELD_ID + "=" + sClassID +
			" FOR UPDATE;"

		row := tx.QueryRow(sql)

		ml := ""
		dl := ""
		if err = row.Scan(&ml, &dl); err != nil {
			tx.Rollback()
			return err
		}

		// Compute new meeting list.
		meetingChanged := false
		ml, meetingChanged = common.DeleteFromList(sMeetingID, ml)
		deletedChanged := false
		dl, deletedChanged = common.AddToList(sMeetingID, dl)
		if (!meetingChanged) && (!deletedChanged) {
			tx.Rollback()

			if cs.cache != nil {
				m := make(map[string]string)
				m[common.FIELD_MEETING_LIST] = ml
				m[common.FIELD_DELETED] = dl
				m[common.FIELD_UPDATER] = sUpdater
				m[common.FIELD_UPDATE_TIME] = sUpdateTime
				m[common.FIELD_UPDATE_IP] = sUpdateIP

				err = cs.cache.SetFields(common.KEY_PREFIX_CLASS+sClassID, m)
				if err != nil {
					return err
				}
			}
			return nil
		}

		// Update the meeting list.
		sql = "UPDATE " +
			common.TABLE_CLASS +
			" SET " +
			common.FIELD_MEETING_LIST + "='" + ml + "'," +
			common.FIELD_DELETED + "='" + dl + "',"

		if mi.EndTime > 0 {
			sql += common.FIELD_NUMBER_OF_FINISHED_MEETING + "=" + common.FIELD_NUMBER_OF_FINISHED_MEETING + "-1,"
		}

		sql += common.FIELD_UPDATER + "=" + sUpdater + "," +
			common.FIELD_UPDATE_TIME + "=" + sUpdateTime + "," +
			common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'" +
			" WHERE " +
			common.FIELD_ID + "=" + sClassID + ";"

		_, err = tx.Exec(sql)
		if err != nil {
			tx.Rollback()
			return err
		}

		if cs.cache != nil {
			m := make(map[string]string)
			m[common.FIELD_MEETING_LIST] = ml
			m[common.FIELD_DELETED] = dl

			err = cs.cache.SetFields(common.KEY_PREFIX_CLASS+sClassID, m)
			if err != nil {
				tx.Rollback()
				return err
			}

			if mi.EndTime > 0 {
				n, err := cs.cache.DecrField(common.KEY_PREFIX_CLASS+sClassID, common.FIELD_NUMBER_OF_FINISHED_MEETING)
				if err != nil {
					// TODO:
				}
				if n < 0 {
					if err = cs.cache.SetField(common.KEY_PREFIX_CLASS+sClassID, common.FIELD_NUMBER_OF_FINISHED_MEETING, "0"); err != nil {
						// TODO:
					}
				}
			}
		}

		// Commit it.
		if err = tx.Commit(); err != nil {
			tx.Rollback()
			return err
		}

		return nil

	})(); err != nil {
		return err
	}

	if err = cs.UpdateMeetingTime(classID, session); err != nil {
		// TODO:
	}
	// if err = cs.UpdateTimeOfNextMeeting(classID, session); err != nil {
	// 	// TODO:
	// }

	return nil
}

//----------------------------------------------------------------------------
// Database: Required.
// Cache   : Compatible.

func (cs *ClassService) DeleteMeetingPermanently(meetingID int, classID int, session *Session) error {
	// Check requirements.
	if cs.db == nil {
		return common.ERR_NO_SERVICE
	}

	ci, err := cs.GetClass(classID, session)
	if err != nil {
		return err
	}
	if !session.IsSystem() {
		if ci.EndTime > 0 {
			return common.ERR_CLASS_CLOSED
		}
	}

	if !common.InIntArray(meetingID, ci.Meetings) {
		return common.ERR_NO_MEETING
	}

	sClassID := strconv.Itoa(classID)
	sMeetingID := strconv.Itoa(meetingID)

	sUpdater := strconv.Itoa(session.UserID)
	sUpdateIP := session.IP
	sUpdateTime := common.GetTimeString()

	if err = (func() error {
		tx, err := cs.db.Transaction()
		if err != nil {
			return err
		}

		// Get existing value.
		s := "SELECT " +
			common.FIELD_DELETED +
			" FROM " +
			common.TABLE_CLASS +
			" WHERE " +
			common.FIELD_ID + "=" + sClassID +
			" FOR UPDATE;"
		row := tx.QueryRow(s)

		dl := ""
		if err = row.Scan(&dl); err != nil {
			tx.Rollback()
			return err
		}

		// Compute new value.
		changed := false
		dl, changed = common.DeleteFromList(sMeetingID, dl)
		if !changed {
			tx.Rollback()

			if cs.cache != nil {
				m := make(map[string]string)
				m[common.FIELD_DELETED] = dl

				err = cs.cache.SetFields(common.KEY_PREFIX_CLASS+sClassID, m)
				if err != nil {
					return err
				}
			}
			return nil
		}

		// Update database.
		s = "UPDATE " +
			common.TABLE_CLASS +
			" SET " +
			common.FIELD_DELETED + "='" + dl + "'," +
			common.FIELD_UPDATER + "=" + sUpdater + "," +
			common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
			common.FIELD_UPDATE_TIME + "=" + sUpdateTime +
			" WHERE " +
			common.FIELD_ID + "=" + sClassID + ";"

		if _, err = tx.Exec(s); err != nil {
			tx.Rollback()
			return err
		}

		if cs.cache != nil {
			m := make(map[string]string)
			m[common.FIELD_DELETED] = dl

			err = cs.cache.SetFields(common.KEY_PREFIX_CLASS+sClassID, m)
			if err != nil {
				tx.Rollback()
				return err
			}
		}

		// Commit it.
		if err = tx.Commit(); err != nil {
			tx.Rollback()
			return err
		}

		return nil
	})(); err != nil {
		return err
	}

	if err = cs.ms.DeleteMeeting(meetingID, session); err != nil {
		// TODO: Record this error.
	}

	return nil
}

//----------------------------------------------------------------------------
// Database: Required.
// Cache   : Compatible.

func (cs *ClassService) GetMeetingList(classID int) (string, error) {
	sClassID := strconv.Itoa(classID)

	if cs.cache != nil {
		ml, err := cs.cache.GetField(common.KEY_PREFIX_CLASS+sClassID, common.FIELD_MEETING_LIST)
		if err == nil {
			return ml, nil
		}
	}

	if cs.db != nil {
		s := "SELECT " +
			common.FIELD_MEETING_LIST +
			" FROM " +
			common.TABLE_CLASS +
			" WHERE " +
			common.FIELD_ID + "=" + sClassID + ";"

		rows, err := cs.db.Select(s)
		if err != nil {
			return "", err
		}
		defer rows.Close()

		if !rows.Next() {
			return "", common.ERR_NO_CLASS
		}

		ml := ""
		if err = rows.Scan(&ml); err != nil {
			return "", err
		}

		if cs.cache != nil {
			if err = cs.cache.SetField(common.KEY_PREFIX_CLASS+sClassID, common.FIELD_MEETING_LIST, ml); err != nil {
				// TODO:
			}
		}

		return ml, nil
	}

	return "", common.ERR_NO_SERVICE
}

func (cs *ClassService) GetMeetings(classID int, session *Session) (*MeetingInfoArray, int, error) {
	// Check requirements.
	if cs.db == nil || cs.ms == nil {
		return nil, -1, common.ERR_NO_SERVICE
	}

	// Check authority.
	ci, err := cs.GetClass(classID, session)
	if err != nil {
		return nil, -2, err
	}

	var arr []int = nil
	if (ci.PlatformID != common.PLATFORM_ID_FOR_PACKAGE) && (ci.Ally > 0) {
		ml, err := cs.GetMeetingList(classID)
		if err != nil {
			return nil, -3, err
		}

		arr = common.StringToIntArray(ml)
	} else {
		arr = ci.Meetings
	}

	mia, err := cs.ms.GetMeetings(arr, session, false)
	if err != nil {
		return nil, -4, err
	}

	return mia, 0, nil
}

//----------------------------------------------------------------------------
// Database: Compatible.
// Cache   : Compatible.

func (cs *ClassService) GetMeeting(meetingID int, session *Session) (*MeetingInfo, int, error) {
	if cs.cache != nil {
		// cs.ms.GetMeeting(meetingID, session)
	}

	if cs.db != nil {
	}

	return nil, 0, common.ERR_NO_SERVICE
}

//----------------------------------------------------------------------------
