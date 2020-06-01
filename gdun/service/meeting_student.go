package service

import (
	"container/list"
	"fmt"
	"github.com/wangmhgo/go-project/gdun/common"
	"strconv"
	"time"
)

//----------------------------------------------------------------------------
// Database: Required.
// Cache   : Compatible.

func (ms *MeetingService) SetCoursewareProgress(meetingID int, coursewareID string, session *Session) error {
	if ms.db == nil {
		return common.ERR_NO_SERVICE
	}

	if !session.IsStudent() {
		return common.ERR_NO_AUTHORITY
	}

	mi, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return err
	}
	sCoursewareID := common.Escape(coursewareID)
	_, existing := common.InStringArrayByKey(sCoursewareID, mi.Coursewares)
	if !existing {
		return common.ERR_INVALID_COURSEWARE
	}

	umpi, err := ms.GetUserProgress(meetingID, session.UserID)
	if err != nil {
		return err
	}
	existing = common.InMap(coursewareID, umpi.CoursewareProgress)
	if existing {
		return nil
	}

	sUpdateTime := common.GetTimeString()
	sUpdater := strconv.Itoa(session.UserID)
	sUpdateIP := session.IP

	sMeetingID := strconv.Itoa(meetingID)

	//----------------------------------------------------

	key := common.KEY_PREFIX_MEETING + sMeetingID + ":" + sUpdater

	if session.IsExperienceStudent() {
		if ms.cache == nil {
			return common.ERR_NO_SERVICE
		}

		s, changed := common.AddResourceToMap(sCoursewareID, sUpdateTime, umpi.CoursewareProgress)
		if changed {
			return ms.cache.SetField(key, common.FIELD_COURSEWARE_A, s)
		}

	} else {
		err = (func() error {
			tx, err := ms.db.Transaction()
			if err != nil {
				return err
			}

			s := "SELECT " +
				common.FIELD_COURSEWARE_A +
				" FROM " +
				common.TABLE_USER_MEETING +
				" WHERE " +
				common.FIELD_USER_ID + "=" + sUpdater +
				" AND " + common.FIELD_MEETING_ID + "=" + sMeetingID +
				" FOR UPDATE;"

			row := tx.QueryRow(s)
			cws := ""
			if err = row.Scan(&cws); err != nil {
				tx.Rollback()
				return err
			}

			changed := false
			cws, changed = common.AddResourceToMap(sCoursewareID, sUpdateTime, cws)
			if !changed {
				if ms.cache != nil {
					m := make(map[string]string)

					m[common.FIELD_COURSEWARE_A] = cws
					m[common.FIELD_UPDATE_IP] = sUpdateIP
					m[common.FIELD_UPDATE_TIME] = sUpdateTime
					m[common.FIELD_UPDATER] = sUpdater

					if err = ms.cache.SetFields(key, m); err != nil {
						//
					}
				}

				tx.Rollback()
				return nil
			}

			s = "UPDATE " +
				common.TABLE_USER_MEETING +
				" SET " +
				common.FIELD_COURSEWARE_A + "='" + cws + "'," +
				common.FIELD_UPDATER + "=" + sUpdater + "," +
				common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
				common.FIELD_UPDATE_TIME + "=" + sUpdateTime +
				" WHERE " +
				common.FIELD_USER_ID + "=" + sUpdater +
				" AND " + common.FIELD_MEETING_ID + "=" + sMeetingID + ";"

			if _, err = tx.Exec(s); err != nil {
				tx.Rollback()
				return err
			}

			if ms.cache != nil {
				m := make(map[string]string)

				m[common.FIELD_COURSEWARE_A] = cws
				m[common.FIELD_UPDATE_IP] = sUpdateIP
				m[common.FIELD_UPDATE_TIME] = sUpdateTime
				m[common.FIELD_UPDATER] = sUpdater

				if err = ms.cache.SetFields(key, m); err != nil {
					tx.Rollback()
					return err
				}
			}

			if err = tx.Commit(); err != nil {
				tx.Rollback()
				return err
			}

			return nil
		})()
		if err != nil {
			return err
		}
	}

	return nil
}

//----------------------------------------------------------------------------
// Database: Required.
// Cache   : Compatible.

func (ms *MeetingService) SetVideoProgress(meetingID int, videoID string, session *Session) error {
	// Check requirements.
	if ms.db == nil {
		return common.ERR_NO_SERVICE
	}

	// Check authority.
	if !session.IsStudent() {
		return common.ERR_NO_AUTHORITY
	}
	mi, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return err
	}
	// if mi.EndTime > 0 {
	// 	return common.ERR_MEETING_CLOSED
	// }

	// Check whether this video belongs to the meeting.
	sVideoID := common.EscapeForStr64(videoID)
	_, okay := common.InStringArrayByKey(sVideoID, mi.Videos)
	if !okay {
		return common.ERR_INVALID_VIDEO
	}

	// Check via cache.
	umpi, err := ms.GetUserProgress(meetingID, session.UserID)
	if err != nil {
		return err
	}
	if common.InMap(sVideoID, umpi.VideoProgress) {
		return nil
	}

	sMeetingID := strconv.Itoa(meetingID)

	sUpdateTime := common.GetTimeString()
	sUpdateIP := session.IP
	sUpdater := strconv.Itoa(session.UserID)

	key := common.KEY_PREFIX_MEETING + sMeetingID + ":" + sUpdater

	if session.IsExperienceStudent() {
		if ms.cache != nil {
			s, changed := common.AddResourceToMap(sVideoID, "1_"+sUpdateTime, umpi.VideoProgress)
			if changed {
				return ms.cache.SetField(key, common.FIELD_VIDEO_A, s)
			}
		} else {
			return common.ERR_NO_SERVICE
		}
	} else {
		if err = (func() error {
			// Create a new transaction.
			tx, err := ms.db.Transaction()
			if err != nil {
				return err
			}

			// Get existing value.
			s := "SELECT " +
				common.FIELD_VIDEO_A +
				" FROM " +
				common.TABLE_USER_MEETING +
				" WHERE " +
				common.FIELD_USER_ID + "=" + sUpdater +
				" AND " + common.FIELD_MEETING_ID + "=" + sMeetingID +
				" FOR UPDATE;"

			row := tx.QueryRow(s)

			vm := ""
			err = row.Scan(&vm)
			if err != nil {
				tx.Rollback()
				return err
			}

			// Compute new value.
			changed := false
			vm, changed = common.AddResourceToMap(sVideoID, "1_"+sUpdateTime, vm)
			if !changed {
				tx.Rollback()

				// Update cache.
				if ms.cache != nil {
					m := make(map[string]string)

					m[common.FIELD_VIDEO_A] = vm
					m[common.FIELD_UPDATER] = sUpdater
					m[common.FIELD_UPDATE_IP] = sUpdateIP
					m[common.FIELD_UPDATE_TIME] = sUpdateTime

					if err := ms.cache.SetFields(key, m); err != nil {
						return err
					}
				}
				return nil
			}

			// Update database.
			s = "UPDATE " +
				common.TABLE_USER_MEETING +
				" SET " +
				common.FIELD_VIDEO_A + "='" + vm + "'," +
				common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
				common.FIELD_UPDATE_TIME + "=" + sUpdateTime + "," +
				common.FIELD_UPDATER + "=" + sUpdater +
				" WHERE " +
				common.FIELD_USER_ID + "=" + sUpdater +
				" AND " + common.FIELD_MEETING_ID + "=" + sMeetingID + ";"

			if _, err := tx.Exec(s); err != nil {
				tx.Rollback()
				return err
			}

			if ms.cache != nil {
				m := make(map[string]string)

				m[common.FIELD_VIDEO_A] = vm
				m[common.FIELD_UPDATER] = sUpdater
				m[common.FIELD_UPDATE_IP] = sUpdateIP
				m[common.FIELD_UPDATE_TIME] = sUpdateTime

				if err := ms.cache.SetFields(key, m); err != nil {
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
	}

	return nil
}

//----------------------------------------------------------------------------
// Database: Required.
// Cache   : Compatible.

func (ms *MeetingService) SetReplayProgress(meetingID int, session *Session) error {
	// Check requirements.
	if ms.db == nil {
		return common.ERR_NO_SERVICE
	}

	// Check authority.
	if !session.IsStudent() {
		return common.ERR_NO_AUTHORITY
	}
	umpi, err := ms.GetUserProgress(meetingID, session.UserID)
	if err != nil {
		return err
	}
	if umpi.ReplayProgress > 0 {
		return nil
	}
	// if mi.EndTime > 0 {
	// 	return common.ERR_MEETING_CLOSED
	// }

	timestamp := common.GetTimeString()
	sUserID := strconv.Itoa(session.UserID)
	sMeetingID := strconv.Itoa(meetingID)

	//----------------------------------------------------
	// Step 1. Update database.

	if !session.IsExperienceStudent() {
		sql := "UPDATE " +
			common.TABLE_USER_MEETING +
			" SET " +
			common.FIELD_REPLAY + "=1," +
			common.FIELD_UPDATE_IP + "='" + session.IP + "'," +
			common.FIELD_UPDATE_TIME + "=" + timestamp + "," +
			common.FIELD_UPDATER + "=" + sUserID +
			" WHERE " +
			common.FIELD_USER_ID + "=" + sUserID +
			" AND " + common.FIELD_MEETING_ID + "=" + sMeetingID + ";"

		_, err = ms.db.Exec(sql)
		if err != nil {
			return err
		}
	}

	//----------------------------------------------------
	// Step 2. Update cache.

	if ms.cache != nil {
		m := make(map[string]string)

		m[common.FIELD_REPLAY] = "1"
		m[common.FIELD_UPDATE_IP] = session.IP
		m[common.FIELD_UPDATE_TIME] = timestamp
		m[common.FIELD_UPDATER] = sUserID

		err = ms.cache.SetFields(common.KEY_PREFIX_MEETING+sMeetingID+":"+sUserID, m)
		if err != nil {
			return err
		}
	}

	return nil
}

//----------------------------------------------------------------------------
// Database: Required.
// Cache   : Compatible.

func (ms *MeetingService) SetMeetingProgress(meetingID int, seconds int, session *Session) error {
	// Check requirements.
	if ms.db == nil {
		return common.ERR_NO_SERVICE
	}

	// Check authority.
	if !session.IsStudent() {
		return common.ERR_NO_AUTHORITY
	}
	mi, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return err
	}
	if mi.EndTime > 0 {
		return common.ERR_MEETING_CLOSED
	}
	umpi, err := ms.GetUserProgress(meetingID, session.UserID)
	if err != nil {
		return common.ERR_NO_AUTHORITY
	}

	// Check existing value.
	if mi.Type == common.MEETING_TYPE_FOR_TEACHING {
		if (seconds > 0) && (umpi.MeetingProgress >= seconds) {
			return nil
		}
	}

	if err = ms.setMeetingProgress(meetingID, session.UserID, seconds, "", session); err != nil {
		return err
	}

	// Adjust the attendee list.
	if mi.Type == common.MEETING_TYPE_FOR_TEACHING {
		if (seconds > 0) && (umpi.MeetingProgress <= 0) {
			if err = ms.increaseMeetingAttendees(meetingID); err != nil {
				return err
			}
		}
	} else if mi.Type == common.MEETING_TYPE_FOR_OFFLINE {
		if (seconds < -1 || seconds > 0) && (umpi.MeetingProgress <= 0) {
			if err = ms.increaseMeetingAttendees(meetingID); err != nil {
				return err
			}
		}
	}

	return nil
}

func (ms *MeetingService) increaseMeetingAttendees(meetingID int) error {
	sMeetingID := strconv.Itoa(meetingID)

	if ms.db != nil {
		s := "UPDATE " +
			common.TABLE_MEETING +
			" SET " +
			common.FIELD_NUMBER_OF_ATTENDEE + "=" + common.FIELD_NUMBER_OF_ATTENDEE + "+1" +
			" WHERE " +
			common.FIELD_ID + "=" + sMeetingID + ";"

		if _, err := ms.db.Exec(s); err != nil {
			return err
		}
	}

	if ms.cache != nil {
		if _, err := ms.cache.IncrField(common.KEY_PREFIX_MEETING+sMeetingID, common.FIELD_NUMBER_OF_ATTENDEE); err != nil {
			return err
		}
	}
	return nil
}

func (ms *MeetingService) setMeetingProgress(meetingID int, studentID int, seconds int, accessLog string, session *Session) error {
	sMeetingID := strconv.Itoa(meetingID)
	sSeconds := strconv.Itoa(seconds)
	sStudentID := strconv.Itoa(studentID)

	sUpdateTime := common.GetTimeString()
	sUpdater := strconv.Itoa(session.UserID)
	sUpdateIP := session.IP

	if ms.db != nil {
		if studentID < common.VALUE_MINIMAL_TEMPERARY_USER_ID {
			s := "UPDATE " +
				common.TABLE_USER_MEETING +
				" SET " +
				common.FIELD_MEETING + "=" + sSeconds + "," +
				common.FIELD_LOG + "='" + accessLog + "'," +
				common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
				common.FIELD_UPDATER + "=" + sUpdater + "," +
				common.FIELD_UPDATE_TIME + "=" + sUpdateTime +
				" WHERE " +
				common.FIELD_USER_ID + "=" + sStudentID +
				" AND " + common.FIELD_MEETING_ID + "=" + sMeetingID + ";"

			if _, err := ms.db.Exec(s); err != nil {
				return err
			}
		}
	}

	if ms.cache != nil {
		m := make(map[string]string)

		m[common.FIELD_MEETING] = sSeconds
		m[common.FIELD_LOG] = accessLog
		m[common.FIELD_UPDATE_IP] = sUpdateIP
		m[common.FIELD_UPDATER] = sUpdater
		m[common.FIELD_UPDATE_TIME] = sUpdateTime

		if err := ms.cache.SetFields(common.KEY_PREFIX_MEETING+sMeetingID+":"+sStudentID, m); err != nil {
			return err
		}
	}

	return nil
}

//----------------------------------------------------------------------------
// Database: Required.
// Cache   : Compatible.

func (ms *MeetingService) LeaveMeeting(meetingID int, cancel int, session *Session) error {
	// Check requirements.
	if ms.db == nil {
		return common.ERR_NO_SERVICE
	}

	// Check authority.
	if !session.IsStudent() {
		return common.ERR_NO_AUTHORITY
	}
	mi, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return err
	}
	if mi.EndTime > 0 {
		return common.ERR_MEETING_CLOSED
	}

	timestamp := common.GetTimeString()
	sUserID := strconv.Itoa(session.UserID)
	sMeetingID := strconv.Itoa(meetingID)

	//----------------------------------------------------
	// Step 1. Update database.

	if !session.IsExperienceStudent() {
		sql := "UPDATE " +
			common.TABLE_USER_MEETING +
			" SET "
		if cancel == 0 {
			sql += common.FIELD_MEETING + "=-1,"
		} else {
			sql += common.FIELD_MEETING + "=0,"
		}
		sql += common.FIELD_UPDATE_IP + "='" + session.IP + "'," +
			common.FIELD_UPDATE_TIME + "=" + timestamp + "," +
			common.FIELD_UPDATER + "=" + sUserID +
			" WHERE " +
			common.FIELD_USER_ID + "=" + sUserID +
			" AND " + common.FIELD_MEETING_ID + "=" + sMeetingID + ";"

		_, err = ms.db.Exec(sql)
		if err != nil {
			return err
		}
	}

	//----------------------------------------------------
	// Step 2. Update cache.

	if ms.cache != nil {
		m := make(map[string]string)

		if cancel == 0 {
			m[common.FIELD_MEETING] = "-1"
		} else {
			m[common.FIELD_MEETING] = "0"
		}
		m[common.FIELD_UPDATE_IP] = session.IP
		m[common.FIELD_UPDATE_TIME] = timestamp
		m[common.FIELD_UPDATER] = sUserID

		err = ms.cache.SetFields(common.KEY_PREFIX_MEETING+sMeetingID+":"+sUserID, m)
		if err != nil {
			return err
		}
	}

	return nil
}

//----------------------------------------------------------------------------
// Database: Required.
// Cache   : Compatible.

func (ms *MeetingService) ScoreMeeting(meetingID int, scores []int, feedback string, session *Session) error {
	// Check requirements.
	if ms.db == nil {
		return common.ERR_NO_SERVICE
	}

	// Check inputs.
	if len(scores) == 0 {
		return common.ERR_INVALID_SCORE
	} else {
		for i := 0; i < len(scores); i++ {
			if scores[i] < 0 || scores[i] > 5 {
				return common.ERR_INVALID_SCORE
			}
		}
	}

	// Check authority.
	_, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return err
	}

	// Check whether he scored this meeting or not.
	umpi, err := ms.GetUserProgress(meetingID, session.UserID)
	if err != nil {
		return err
	}
	if len(umpi.Scores) > 0 {
		return common.ERR_DUPLICATED_SCORE
	}

	sMeetingID := strconv.Itoa(meetingID)
	sScores := common.IntArrayToString(scores)

	sUpdateIP := session.IP
	sUpdateTime := common.GetTimeString()
	sUpdater := strconv.Itoa(session.UserID)

	if session.IsExperienceStudent() {
		//
	} else {
		if err = (func() error {

			// Create a new transaction.
			tx, err := ms.db.Transaction()
			if err != nil {
				return err
			}

			//------------------------------------------------
			// 1. Update user progresses.

			sql := "UPDATE " +
				common.TABLE_USER_MEETING +
				" SET " +
				common.FIELD_SCORE + "='" + sScores + "'," +
				common.FIELD_UPDATE_TIME + "=" + sUpdateTime + "," +
				common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
				common.FIELD_UPDATER + "=" + sUpdater +
				" WHERE " +
				common.FIELD_MEETING_ID + "=" + sMeetingID + " AND " +
				common.FIELD_USER_ID + "=" + sUpdater + ";"

			if _, err = tx.Exec(sql); err != nil {
				tx.Rollback()
				return err
			}

			// Update cache.
			if ms.cache != nil {
				m := make(map[string]string)

				m[common.FIELD_SCORE] = sScores

				m[common.FIELD_UPDATE_TIME] = sUpdateTime
				m[common.FIELD_UPDATE_IP] = sUpdateIP
				m[common.FIELD_UPDATER] = sUpdater

				if err = ms.cache.SetFields(common.KEY_PREFIX_MEETING+sMeetingID+":"+sUpdater, m); err != nil {
					tx.Rollback()
					return err
				}
			}

			//------------------------------------------------
			// 2. Update meeting info.

			// Get existing values.
			sql = "SELECT " +
				common.FIELD_SCORE + "," +
				common.FIELD_SCORE_COUNT +
				" FROM " +
				common.TABLE_MEETING +
				" WHERE " +
				common.FIELD_ID + "=" + sMeetingID +
				" FOR UPDATE;"

			row := tx.QueryRow(sql)

			sl := ""
			sc := 0
			if err = row.Scan(&sl, &sc); err != nil {
				tx.Rollback()
				return err
			}

			// Compute new value.
			if len(sl) == 0 {
				// This is the first score.
				sl = sScores
				sc = 1
			} else {
				sl = common.IntArrayToString(common.CombineIntArrayNumerically(common.StringToIntArray(sl), scores))
				sc++
			}

			// Update database.
			sql = "UPDATE " +
				common.TABLE_MEETING +
				" SET " +
				common.FIELD_SCORE + "='" + sl + "'," +
				common.FIELD_SCORE_COUNT + "=" + strconv.Itoa(sc) + "," +
				common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
				common.FIELD_UPDATE_TIME + "=" + sUpdateTime + "," +
				common.FIELD_UPDATER + "=" + sUpdater +
				" WHERE " +
				common.FIELD_ID + "=" + sMeetingID + ";"

			if _, err = tx.Exec(sql); err != nil {
				tx.Rollback()
				return err
			}

			// Update cache.
			if ms.cache != nil {
				m := make(map[string]string)

				m[common.FIELD_SCORE] = sl
				m[common.FIELD_SCORE_COUNT] = strconv.Itoa(sc)

				m[common.FIELD_UPDATE_TIME] = sUpdateTime
				m[common.FIELD_UPDATE_IP] = sUpdateIP
				m[common.FIELD_UPDATER] = sUpdater

				if err = ms.cache.SetFields(common.KEY_PREFIX_MEETING+sMeetingID, m); err != nil {
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
	}

	if len(feedback) > 0 {
		if err = (func() error {
			sFeedback := `"` + sUpdater + `":{"` + common.FIELD_TIMESTAMP + `":` + sUpdateTime + `000,"` + common.FIELD_FEEDBACK + `":"` + common.Escape(feedback) + `"}`

			s := "INSERT INTO " +
				common.TABLE_MEETING_FEEDBACK +
				" VALUES (" +
				sMeetingID + "," +
				"'" + sFeedback + "'" +
				");"
			if _, err := ms.db.Exec(s); err == nil {
				return nil
			}

			s = "UPDATE " +
				common.TABLE_MEETING_FEEDBACK +
				" SET " +
				common.FIELD_FEEDBACK + "=CONCAT(" + common.FIELD_FEEDBACK + ",'," + sFeedback + "')" +
				" WHERE " +
				common.FIELD_MEETING_ID + "=" + sMeetingID + ";"
			if _, err := ms.db.Exec(s); err != nil {
				return nil
			}

			return nil
		})(); err != nil {
			return err
		}
	}

	return nil
}

func (ms *MeetingService) GetMeetingFeedback(meetingID int, session *Session) (string, error) {
	_, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return "", common.ERR_NO_AUTHORITY
	}

	s := "SELECT " +
		common.FIELD_FEEDBACK +
		" FROM " +
		common.TABLE_MEETING_FEEDBACK +
		" WHERE " +
		common.FIELD_MEETING_ID + "=" + strconv.Itoa(meetingID) + ";"

	rows, err := ms.db.Select(s)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	if !rows.Next() {
		return "", nil
	}

	fb := ""
	err = rows.Scan(&fb)
	if err != nil {
		return "", nil
	}

	return `"` + common.FIELD_FEEDBACK + `":{` + fb + `}`, nil
}

//----------------------------------------------------------------------------
// Database: Compatible.
// Cache   : Compatible.

func (ms *MeetingService) getCachedProgressKey(meetingID int, userID int) string {
	return common.KEY_PREFIX_MEETING + strconv.Itoa(meetingID) + ":" + strconv.Itoa(userID) + ":" + common.FIELD_TEXT
}

func (ms *MeetingService) DeleteCachedUserProgress(meetingID int, userID int) error {
	if ms.cache == nil {
		return common.ERR_NO_CACHE
	}

	return ms.cache.Del(ms.getCachedProgressKey(meetingID, userID))
}

func (ms *MeetingService) GetCachedUserProgress(meetingID int, userID int, withHead bool) (string, error) {
	key := ms.getCachedProgressKey(meetingID, userID)

	// Retrieve it from cache.
	if ms.cache != nil {
		if s, err := ms.cache.GetKey(key); err == nil {
			if withHead {
				s = `"` + common.FIELD_USER_ID + `":` + strconv.Itoa(userID) + `,` +
					`"` + common.FIELD_MEETING_ID + `":` + strconv.Itoa(meetingID) + `,` +
					s
			}

			return s, nil
		}
	}

	upi, err := ms.GetUserProgress(meetingID, userID)
	if err != nil {
		return "", err
	}

	s := upi.ToJSON(false)
	if ms.cache != nil {
		if err = ms.cache.SetKey(key, s); err != nil {
			// TODO:
			fmt.Println(err.Error())
		}
	}

	if withHead {
		s = `"` + common.FIELD_USER_ID + `":` + strconv.Itoa(userID) + `,` +
			`"` + common.FIELD_MEETING_ID + `":` + strconv.Itoa(meetingID) + `,` +
			s
	}

	fmt.Printf("GetCachedUserProgress() %d %d\n", meetingID, userID)
	return s, nil
}

//----------------------------------------------------------------------------
// Database: Compatible.
// Cache   : Compatible.

func (ms *MeetingService) GetUserProgress(meetingID int, userID int) (*UserMeetingProgressInfo, error) {

	sMeetingID := strconv.Itoa(meetingID)
	sUserID := strconv.Itoa(userID)

	if ms.cache != nil {
		key := common.KEY_PREFIX_MEETING + sMeetingID + ":" + sUserID

		if m, err := ms.cache.GetAllFields(key); err == nil {
			if umpi := NewUserMeetingProgressInfoFromMap(m, userID, meetingID); umpi != nil {
				return umpi, nil
			}
		}

		// fmt.Println(strconv.Itoa(meetingID) + "," + strconv.Itoa(userID))

		if userID > common.VALUE_MINIMAL_TEMPERARY_USER_ID {
			umpi := new(UserMeetingProgressInfo)
			umpi.UserID = userID
			umpi.MeetingID = meetingID
			umpi.CoursewareProgress = ""
			umpi.VideoProgress = ""
			umpi.MeetingProgress = 0
			umpi.MeetingLog = ""
			umpi.ExamAnswers = ""
			umpi.ExamCorrect = 0
			umpi.ExamTotal = 0
			umpi.ReplayProgress = 0
			umpi.Scores = []int{}
			umpi.UpdateTime = 0
			umpi.UpdateIP = ""
			umpi.Updater = userID

			m := make(map[string]string)
			m[common.FIELD_COURSEWARE_A] = ""
			m[common.FIELD_VIDEO_A] = ""
			m[common.FIELD_MEETING] = "0"
			m[common.FIELD_LOG] = ""
			m[common.FIELD_EXAM] = ""
			m[common.FIELD_EXAM_CORRECT] = "0"
			m[common.FIELD_EXAM_TOTAL] = "0"
			m[common.FIELD_REPLAY] = "0"
			m[common.FIELD_SCORE] = ""
			m[common.FIELD_UPDATE_IP] = "127.0.0.1"
			m[common.FIELD_UPDATE_TIME] = common.GetTimeString()
			m[common.FIELD_UPDATER] = sUserID

			if err := ms.cache.SetFields(key, m); err != nil {
				return umpi, err
			}
			if err := ms.cache.Expire(key, 12*time.Hour); err != nil {
				return umpi, err
			}

			return umpi, nil
		}
	}

	if ms.db != nil {
		sql := "SELECT " +
			common.FIELD_COURSEWARE_A + "," +
			common.FIELD_VIDEO_A + "," +
			common.FIELD_MEETING + "," +
			common.FIELD_LOG + "," +
			common.FIELD_EXAM + "," +
			common.FIELD_EXAM_CORRECT + "," +
			common.FIELD_EXAM_TOTAL + "," +
			common.FIELD_REPLAY + "," +
			common.FIELD_SCORE + "," +
			common.FIELD_UPDATE_TIME + "," +
			common.FIELD_UPDATE_IP + "," +
			common.FIELD_UPDATER +
			" FROM " +
			common.TABLE_USER_MEETING +
			" WHERE " +
			common.FIELD_USER_ID + "=" + sUserID +
			" AND " + common.FIELD_MEETING_ID + "=" + sMeetingID + ";"
		rows, err := ms.db.Select(sql)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		if !rows.Next() {
			return nil, common.ERR_NO_USER
		}

		scores := ""
		umpi := new(UserMeetingProgressInfo)
		err = rows.Scan(
			&umpi.CoursewareProgress,
			&umpi.VideoProgress,
			&umpi.MeetingProgress, &umpi.MeetingLog,
			&umpi.ExamAnswers, &umpi.ExamCorrect, &umpi.ExamTotal,
			&umpi.ReplayProgress,
			&scores,
			&umpi.UpdateTime, &umpi.UpdateIP, &umpi.Updater)
		if err != nil {
			return nil, err
		}
		umpi.Scores = common.StringToIntArray(scores)
		umpi.UserID = userID
		umpi.MeetingID = meetingID

		if ms.cache != nil {
			m := make(map[string]string)

			m[common.FIELD_COURSEWARE_A] = umpi.CoursewareProgress
			m[common.FIELD_VIDEO_A] = umpi.VideoProgress
			m[common.FIELD_MEETING] = strconv.Itoa(umpi.MeetingProgress)
			m[common.FIELD_LOG] = umpi.MeetingLog
			m[common.FIELD_EXAM] = umpi.ExamAnswers
			m[common.FIELD_EXAM_CORRECT] = strconv.Itoa(umpi.ExamCorrect)
			m[common.FIELD_EXAM_TOTAL] = strconv.Itoa(umpi.ExamTotal)
			m[common.FIELD_REPLAY] = strconv.Itoa(umpi.ReplayProgress)
			m[common.FIELD_SCORE] = common.IntArrayToString(umpi.Scores)
			m[common.FIELD_UPDATE_IP] = umpi.UpdateIP
			m[common.FIELD_UPDATE_TIME] = strconv.Itoa(umpi.UpdateTime)
			m[common.FIELD_UPDATER] = strconv.Itoa(umpi.Updater)

			if err = ms.cache.SetFields(common.KEY_PREFIX_MEETING+sMeetingID+":"+sUserID, m); m != nil {
				return umpi, err
			}
		}

		return umpi, nil
	}

	return nil, common.ERR_NO_SERVICE
}

//----------------------------------------------------------------------------
// Database: Compatible.
// Cache   : Compatible.

func (ms *MeetingService) GetCachedUserProgressesByMeeting(meetingID int, studentIDs []int, withHead bool) (string, error) {
	r := `"` + common.FIELD_PROGRESS + `":[`
	first := true
	for i := 0; i < len(studentIDs); i++ {
		s, err := ms.GetCachedUserProgress(meetingID, studentIDs[i], withHead)
		if err != nil {
			return "", err
		}

		if first {
			first = false
		} else {
			r += `,`
		}

		if withHead {
			s = `"` + common.FIELD_USER_ID + `":` + strconv.Itoa(studentIDs[i]) + `,` +
				`"` + common.FIELD_MEETING_ID + `":` + strconv.Itoa(meetingID) + `,` +
				s
		}
		r += `{` + s + `}`
	}
	r += `]`

	return r, nil
}

func (ms *MeetingService) GetMeetingProgresses(meetingID int, session *Session) (*UserMeetingProgressInfoArray, error) {
	// Check authority.
	mi, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return nil, err
	}

	s, err := ms.cache.GetField(common.KEY_PREFIX_CLASS+strconv.Itoa(mi.ClassID), common.FIELD_STUDENT_LIST)
	if err != nil {
		return nil, err
	}
	studentIDs := common.StringToIntArray(s)

	// Result.
	umpia := new(UserMeetingProgressInfoArray)
	umpia.Status = list.New()

	for i := 0; i < len(studentIDs); i++ {
		umpi, err := ms.GetUserProgress(meetingID, studentIDs[i])
		if err != nil {
			return umpia, err
		}

		umpia.Status.PushBack(umpi)
	}

	return umpia, nil
}

//----------------------------------------------------------------------------
// Database: Required.
// Cache   : Required.

func (ms *MeetingService) JoinMeeting(meetingID int, session *Session) (string, error) {
	// Check requirements.
	if ms.db == nil || ms.cache == nil {
		return "", common.ERR_NO_SERVICE
	}

	// Check authority.
	mi, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return "", err
	}
	if mi.EndTime > 0 {
		return "", common.ERR_MEETING_CLOSED
	}

	// Check whether such a meeting is running right now.
	liveKey := common.KEY_PREFIX_LIVE + strconv.Itoa(meetingID)
	okay, err := ms.cache.Exists(liveKey)
	if err != nil {
		return "", err
	}
	if !okay {
		if session.IsStudent() {
			return "", common.ERR_NO_MEETING
		} else {
			liveFields := make(map[string]string)
			liveFields[common.FIELD_GROUP_ID] = strconv.Itoa(mi.GroupID)
			liveFields[common.FIELD_NAME] = mi.Name

			err = ms.cache.SetFields(liveKey, liveFields)
			if err != nil {
				return "", err
			}
		}
	}

	// Set an common token for this user.
	sMeetingID := strconv.Itoa(meetingID)
	sUserID := strconv.Itoa(session.UserID)
	password := ms.GetUUID()
	tokenKey := common.KEY_PREFIX_LIVE_TOKEN + sMeetingID
	tokenField := sUserID
	tokenValue := strconv.Itoa(session.GroupID) + "-" + password + "-" + session.Nickname

	err = ms.cache.SetField(tokenKey, tokenField, tokenValue)
	if err != nil {
		return "", err
	}

	return `"` + common.FIELD_URL + `":"` + ms.liveServerUrl + `?` + sMeetingID + "-" + sUserID + "-" + password + `"`, nil
}

//----------------------------------------------------------------------------
