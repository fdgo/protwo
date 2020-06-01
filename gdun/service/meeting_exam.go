package service

import (
	"github.com/wangmhgo/go-project/gdun/common"
	"strconv"
	"strings"
)

//----------------------------------------------------------------------------
// Database: Required.

func (ms *MeetingService) AddExam(meetingID int, examID int, gdExamID int, name string, startTime int, duration int, preparation int, isNecessary bool, groupID int, session *Session) error {
	// Check requirements.
	if ms.db == nil || ms.es == nil {
		return common.ERR_NO_SERVICE
	}

	//----------------------------------------------------
	// Check authority.

	mi, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return err
	}

	gID := 0
	if session.IsSystem() {
		gID = groupID
	} else if session.IsAssistant() {
		gID = session.GroupID
	} else {
		return common.ERR_NO_AUTHORITY
	}

	//----------------------------------------------------

	key := ""

	value := common.Escape(name) + "_" + strconv.Itoa(preparation)
	if isNecessary {
		value += "_1"
	} else {
		value += "_0"
	}
	value += "_" + strconv.Itoa(startTime) + "_" + strconv.Itoa(duration)

	if examID > 0 {
		if gdExamID > 0 {
			// Re-import this exam for Gd.
			_, cnt, err := ms.es.Import(name, gID, examID, gdExamID, session)
			if err != nil {
				return err
			}

			value += "_" + strconv.Itoa(cnt)
		} else {
			// Get the number of questions within this exam.
			ei, err := ms.es.GetExam(examID, session)
			if err != nil {
				return err
			}
			value += "_" + strconv.Itoa(ei.Count)
		}

		key = strconv.Itoa(examID)

		// Check whether this video resides in the meeting via cache.
		target := key + ":" + value
		existing := false
		for i := 0; i < len(mi.Exams); i++ {
			if mi.Exams[i] == target {
				existing = true
				break
			}
		}
		if existing {
			return nil
		}
	} else {
		id, cnt, err := ms.es.Import(name, gID, 0, gdExamID, session)
		if err != nil {
			return err
		}

		value += "_" + strconv.Itoa(cnt)

		// It must be a new exam ID.
		key = strconv.Itoa(id)
	}

	//----------------------------------------------------
	// Update the meeting table.

	sMeetingID := strconv.Itoa(meetingID)

	sUpdateTime := common.GetTimeString()
	sUpdater := strconv.Itoa(session.UserID)
	sUpdateIP := session.IP

	if err = (func() error {
		// Create a transaction.
		tx, err := ms.db.Transaction()
		if err != nil {
			return err
		}

		// Get existing exams.
		sql := "SELECT " +
			common.FIELD_EXAM_LIST +
			" FROM " +
			common.TABLE_MEETING +
			" WHERE " +
			common.FIELD_ID + "=" + strconv.Itoa(meetingID) +
			" FOR UPDATE;"

		row := tx.QueryRow(sql)

		el := ""
		err = row.Scan(&el)
		if err != nil {
			tx.Rollback()
			return err
		}

		// Compute new value.
		el, _ = common.AddResourceToMap(key, value, el)
		// TODO:

		// Update database.
		sql = "UPDATE " +
			common.TABLE_MEETING +
			" SET " +
			common.FIELD_EXAM_LIST + "='" + el + "'," +
			common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
			common.FIELD_UPDATE_TIME + "=" + sUpdateTime + "," +
			common.FIELD_UPDATER + "=" + sUpdater +
			" WHERE " +
			common.FIELD_ID + "=" + sMeetingID + ";"

		_, err = tx.Exec(sql)
		if err != nil {
			tx.Rollback()
			return err
		}

		// Update cache.
		if ms.cache != nil {
			m := make(map[string]string)

			m[common.FIELD_EXAM_LIST] = el
			m[common.FIELD_UPDATE_IP] = sUpdateIP
			m[common.FIELD_UPDATE_TIME] = sUpdateTime
			m[common.FIELD_UPDATER] = sUpdater

			err = ms.cache.SetFields(common.KEY_PREFIX_MEETING+sMeetingID, m)
			if err != nil {
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

	return nil
}

//----------------------------------------------------------------------------

func (ms *MeetingService) ResyncExam(meetingID int, examID int, session *Session) error {
	// Check authority.
	mi, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return err
	}

	// Check whether this exam resides in the meeting.
	if _, okay := common.InStringArrayByKey(strconv.Itoa(examID), mi.Exams); !okay {
		return common.ERR_NO_EXAM
	}

	// Reload the exam.
	if _, err = ms.es.Resync(examID, session); err != nil {
		return err
	}

	// TODO: Change existing exam answers and other data.

	return nil
}

//----------------------------------------------------------------------------

func (ms *MeetingService) DeleteExam(meetingID int, examID int, session *Session) error {
	// Check requirements.
	if ms.db == nil || ms.cache == nil {
		return common.ERR_NO_SERVICE
	}

	// Check authority.
	mi, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return err
	}

	sExamID := strconv.Itoa(examID)

	// Check via cache.
	existing := false
	for i := 0; i < len(mi.Exams); i++ {
		if strings.HasPrefix(mi.Exams[i], sExamID+":") {
			existing = true
			break
		}
	}
	if !existing {
		return nil
	}

	sMeetingID := strconv.Itoa(meetingID)

	sUpdater := strconv.Itoa(session.UserID)
	sUpdateTime := common.GetTimeString()
	sUpdateIP := session.IP

	if err := (func() error {
		// Create a new transaction.
		tx, err := ms.db.Transaction()
		if err != nil {
			return err
		}

		sql := "SELECT " +
			common.FIELD_EXAM_LIST +
			" FROM " +
			common.TABLE_MEETING +
			" WHERE " +
			common.FIELD_ID + "=" + sMeetingID +
			" FOR UPDATE;"

		row := tx.QueryRow(sql)

		el := ""
		err = row.Scan(&el)
		if err != nil {
			tx.Rollback()
			return err
		}

		// Check again.
		okay := false
		el, okay = common.DeleteFromMap(sExamID, el)
		if !okay {
			tx.Rollback()

			if ms.cache != nil {
				err = ms.cache.SetField(common.KEY_PREFIX_MEETING+sMeetingID, common.FIELD_EXAM_LIST, el)
				if err != nil {
					return err
				}
			}
			return nil
		}

		// Update database.
		sql = "UPDATE " +
			common.TABLE_MEETING +
			" SET " +
			common.FIELD_EXAM_LIST + "='" + el + "'," +
			common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
			common.FIELD_UPDATE_TIME + "=" + sUpdateTime + "," +
			common.FIELD_UPDATER + "=" + sUpdater +
			" WHERE " +
			common.FIELD_ID + "=" + sMeetingID + ";"

		_, err = tx.Exec(sql)
		if err != nil {
			tx.Rollback()
			return err
		}

		// Update cache.
		if ms.cache != nil {
			m := make(map[string]string)

			m[common.FIELD_EXAM_LIST] = el
			m[common.FIELD_UPDATE_IP] = sUpdateIP
			m[common.FIELD_UPDATE_TIME] = sUpdateTime
			m[common.FIELD_UPDATER] = sUpdater

			err = ms.cache.SetFields(common.KEY_PREFIX_MEETING+sMeetingID, m)
			if err != nil {
				return err
			}
		}

		// Commit this transaction.
		err = tx.Commit()
		if err != nil {
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

func (ms *MeetingService) AuthorizeExam(meetingID int, examID int, session *Session) (*ExamInfo, error) {
	if ms.es == nil {
		return nil, common.ERR_NO_SERVICE
	}

	// Check authority to visit this meeting.
	mi, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return nil, err
	}

	// Check authority to visit this exam.
	_, okay := common.InStringArrayByKey(strconv.Itoa(examID), mi.Exams)
	if !okay {
		return nil, common.ERR_NO_EXAM
	}

	ei, err := ms.es.GetExam(examID, session)
	if err != nil {
		return nil, err
	}

	return ei, err
}

//----------------------------------------------------------------------------

func (ms *MeetingService) AnswerExam(meetingID int, examID int, answer string, session *Session) error {
	// Check requirements.
	if ms.db == nil || ms.es == nil {
		return common.ERR_NO_SERVICE
	}

	// Check authority.
	mi, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return err
	}

	// Check authority to visit this exam.
	_, okay := common.InStringArrayByKey(strconv.Itoa(examID), mi.Exams)
	if !okay {
		return common.ERR_NO_EXAM
	}

	//----------------------------------------------------
	// Check whether this exam had been answered before.

	umpi, err := ms.GetUserProgress(meetingID, session.UserID)
	if err != nil {
		return err
	}

	sExamID := strconv.Itoa(examID)

	if common.InMap(sExamID, umpi.ExamAnswers) {
		return common.ERR_DUPLICATED_ANSWER
	}

	// Get the correct answers.
	correctAnswer, err := ms.cache.GetField(common.KEY_PREFIX_EXAM+sExamID, common.FIELD_ANSWER)
	if err != nil {
		return err
	}

	//----------------------------------------------------
	// Check the user's answer.

	n := len(answer)
	totalCount := len(correctAnswer)
	if (n != totalCount) || (n%2 != 0) {
		return common.ERR_INVALID_ANSWER
	}

	correctCount := 0
	for i := 0; i < totalCount; i += 2 {
		if answer[i] != correctAnswer[i] || answer[i+1] != correctAnswer[i+1] {
			continue
		} else {
			correctCount++
		}
	}

	totalCount /= 2

	//----------------------------------------------------
	// Update database.

	sMeetingID := strconv.Itoa(meetingID)

	sUpdater := strconv.Itoa(session.UserID)
	sUpdateTime := common.GetTimeString()
	sUpdateIP := session.IP

	// value := sUpdateTime + "_" + common.Escape(answer)
	value := sUpdateTime + "_" + strconv.Itoa(correctCount) + "_" + strconv.Itoa(totalCount)

	if err = (func() error {
		// Create a transaction.
		tx, err := ms.db.Transaction()
		if err != nil {
			return err
		}

		// Get existing exam answers.
		sql := "SELECT " +
			common.FIELD_EXAM + "," +
			common.FIELD_EXAM_CORRECT + "," +
			common.FIELD_EXAM_TOTAL +
			" FROM " +
			common.TABLE_USER_MEETING +
			" WHERE " +
			common.FIELD_USER_ID + "=" + sUpdater +
			" AND " + common.FIELD_MEETING_ID + "=" + sMeetingID + " FOR UPDATE;"
		row := tx.QueryRow(sql)

		el := ""
		ec := 0
		et := 0
		if err = row.Scan(&el, &ec, &et); err != nil {
			tx.Rollback()
			return err
		}

		// Check whether this exam had been answered before.
		el, changed := common.AddResourceToMap(sExamID, value, el)
		if !changed {
			tx.Rollback()
			return nil
		}

		// Compute new value.

		ec += correctCount
		et += totalCount

		sCorrectCount := strconv.Itoa(ec)
		sTotalCount := strconv.Itoa(et)

		// Update database.
		sql = "UPDATE " +
			common.TABLE_USER_MEETING +
			" SET " +
			common.FIELD_EXAM + "='" + el + "'," +
			common.FIELD_EXAM_CORRECT + "=" + sCorrectCount + "," +
			common.FIELD_EXAM_TOTAL + "=" + sTotalCount + "," +
			common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
			common.FIELD_UPDATE_TIME + "=" + sUpdateTime + "," +
			common.FIELD_UPDATER + "=" + sUpdater +
			" WHERE " +
			common.FIELD_USER_ID + "=" + sUpdater +
			" AND " + common.FIELD_MEETING_ID + "=" + sMeetingID + ";"

		if _, err = tx.Exec(sql); err != nil {
			tx.Rollback()
			return err
		}

		if ms.cache != nil {
			m := make(map[string]string)
			m[common.FIELD_EXAM] = el
			m[common.FIELD_EXAM_CORRECT] = sCorrectCount
			m[common.FIELD_EXAM_TOTAL] = sTotalCount
			m[common.FIELD_UPDATE_IP] = sUpdateIP
			m[common.FIELD_UPDATE_TIME] = sUpdateTime
			m[common.FIELD_UPDATER] = sUpdater

			err = ms.cache.SetFields(common.KEY_PREFIX_MEETING+sMeetingID+":"+sUpdater, m)
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

	if err = ms.saveExamAnswers(meetingID, examID, answer, session); err != nil {
		return err
	}

	return nil
}

//----------------------------------------------------------------------------
// Cache: Required.

func (ms *MeetingService) getTempExamAnswerKey(meetingID int, userID int) string {
	return common.KEY_PREFIX_MEETING + strconv.Itoa(meetingID) + ":" + common.KEY_PREFIX_EXAM + strconv.Itoa(userID)
}

func (ms *MeetingService) saveExamAnswers(meetingID int, examID int, objectiveAnswers string, session *Session) error {
	key := ms.getTempExamAnswerKey(meetingID, session.UserID)
	m, err := ms.cache.GetAllFields(key)
	if err != nil {
		return err
	}

	// Check inputs.
	if (len(m) == 0) && (len(objectiveAnswers) == 0) {
		// Nothing have to be saved.
		return nil
	}

	// Get answers for subjective questions.
	subjectiveAnswers := `{`
	first := true
	for id, answer := range m {
		if first {
			first = false
		} else {
			subjectiveAnswers += `,`
		}
		subjectiveAnswers += `"` + id + `":"` + answer + `"`
	}
	subjectiveAnswers += `}`

	if err = ms.es.Answer(meetingID, examID, objectiveAnswers, subjectiveAnswers, session); err != nil {
		return err
	}

	if err = ms.cache.Del(key); err != nil {
		// TODO:
	}

	return nil
}

func (ms *MeetingService) AnswerExamQuestion(meetingID int, examID int, questionID int, answer string, session *Session) error {
	// Check requirements.
	if ms.cache == nil {
		return common.ERR_NO_SERVICE
	}

	// Check inputs.
	if questionID < 0 {
		return common.ERR_NO_QUESTION
	}
	sAnswer := common.Escape(strings.Replace(answer, "\"", "\\\"", -1))
	if len(sAnswer) == 0 {
		return common.ERR_INVALID_ANSWER
	}

	// Save the answer of this question.
	err := ms.cache.SetField(ms.getTempExamAnswerKey(meetingID, session.UserID), strconv.Itoa(questionID), sAnswer)
	if err != nil {
		return err
	}

	return nil
}

//----------------------------------------------------------------------------

func (ms *MeetingService) GetMyExamResult(meetingID int, examID int, session *Session) (string, error) {
	// Check authority.
	mi, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return "", err
	}

	// Check authority to visit this exam.
	_, okay := common.InStringArrayByKey(strconv.Itoa(examID), mi.Exams)
	if !okay {
		return "", common.ERR_NO_EXAM
	}

	return ms.es.GetMyResult(meetingID, examID, session)
}

//----------------------------------------------------------------------------

func (ms *MeetingService) GetExamResults(meetingID int, examID int, session *Session) (string, error) {
	// Check authority.
	mi, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return "", err
	}

	// Check authority to visit this exam.
	_, okay := common.InStringArrayByKey(strconv.Itoa(examID), mi.Exams)
	if !okay {
		return "", common.ERR_NO_EXAM
	}

	return ms.es.GetResults(meetingID, examID, session)
}

//----------------------------------------------------------------------------
