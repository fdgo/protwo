package service

import (
	"container/list"
	"errors"
	"github.com/wangmhgo/go-project/gdun/common"
	"math/rand"
	"strconv"
	"time"
)

//----------------------------------------------------------------------------

const (
	MAX_MEETING_DURATION = 60 * 60 * 24 * 5 // In seconds.
)

//----------------------------------------------------------------------------

type MeetingService struct {
	db            *common.Database
	cache         *common.Cache
	liveServerUrl string
	ts            *TranscodingService
	es            *ExamService
}

func NewMeetingService(db *common.Database, cache *common.Cache, liveServerUrl string, ts *TranscodingService, es *ExamService) (*MeetingService, error) {
	ms := new(MeetingService)
	ms.db = db
	ms.cache = cache
	ms.liveServerUrl = liveServerUrl
	ms.ts = ts
	ms.es = es

	rand.Seed(time.Now().UnixNano())

	err := ms.Init()
	if err != nil {
		return nil, err
	}

	return ms, nil
}

//----------------------------------------------------------------------------

func (ms *MeetingService) Init() error {
	if ms.db == nil {
		return common.ERR_NO_DATABASE
	}

	sql := "CREATE TABLE IF NOT EXISTS `" + common.TABLE_MEETING + "` ("
	sql += " `" + common.FIELD_ID + "` INT NOT NULL AUTO_INCREMENT,"
	sql += " `" + common.FIELD_NAME + "` VARCHAR(512) NOT NULL,"
	sql += " `" + common.FIELD_SUBJECT_LIST + "` TEXT NOT NULL,"
	sql += " `" + common.FIELD_SECTION + "` INT NOT NULL DEFAULT 0,"
	sql += " `" + common.FIELD_TYPE + "` INT NOT NULL DEFAULT 0," // New field.
	sql += " `" + common.FIELD_DATA + "` TEXT NOT NULL,"          // New field.
	sql += " `" + common.FIELD_ALLY + "` INT NOT NULL DEFAULT 0,"
	sql += " `" + common.FIELD_SCORE + "` TEXT NOT NULL,"                // New field.
	sql += " `" + common.FIELD_SCORE_COUNT + "` INT NOT NULL DEFAULT 0," // New field.
	// sql += " `" + common.FIELD_TEACHER_LIST + "` TEXT NOT NULL,"
	// sql += " `" + common.FIELD_STUDENT_LIST + "` TEXT NOT NULL,"
	sql += " `" + common.FIELD_NUMBER_OF_ATTENDEE + "` INT NOT NULL DEFAULT 0,"
	sql += " `" + common.FIELD_COURSEWARE_LIST + "` TEXT NOT NULL,"
	sql += " `" + common.FIELD_VIDEO_LIST + "` TEXT NOT NULL,"
	sql += " `" + common.FIELD_EXAM_LIST + "` TEXT NOT NULL," // New field.
	sql += " `" + common.FIELD_REPLAY_LIST + "` TEXT NOT NULL,"
	sql += " `" + common.FIELD_CLASS_ID + "` INT NOT NULL DEFAULT 0,"
	sql += " `" + common.FIELD_GROUP_ID + "` INT NOT NULL,"
	sql += " `" + common.FIELD_START_TIME + "` BIGINT NOT NULL,"
	sql += " `" + common.FIELD_END_TIME + "` BIGINT NOT NULL DEFAULT 0,"
	sql += " `" + common.FIELD_DURATION + "` INT NOT NULL,"
	sql += " `" + common.FIELD_IS_FAKE + "` TINYINT(1) NOT NULL DEFAULT 0,"
	sql += " `" + common.FIELD_UPDATE_IP + "` VARCHAR(32) NOT NULL,"
	sql += " `" + common.FIELD_UPDATE_TIME + "` BIGINT NOT NULL,"
	sql += " `" + common.FIELD_UPDATER + "` INT NOT NULL,"
	sql += " PRIMARY KEY (`" + common.FIELD_ID + "`),"
	sql += " KEY (`" + common.FIELD_GROUP_ID + "`),"
	sql += " KEY (`" + common.FIELD_START_TIME + "`)"
	sql += ") ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;"

	_, err := ms.db.Exec(sql)
	if err != nil {
		return err
	}

	sql = "CREATE TABLE IF NOT EXISTS `" + common.TABLE_USER_MEETING + "` ("
	sql += " `" + common.FIELD_USER_ID + "` INT NOT NULL,"
	sql += " `" + common.FIELD_MEETING_ID + "` INT NOT NULL,"
	// sql += " `" + common.FIELD_COURSEWARE + "` INT NOT NULL DEFAULT 0,"
	sql += " `" + common.FIELD_COURSEWARE_A + "` TEXT NOT NULL,"
	// sql += " `" + common.FIELD_VIDEO + "` INT NOT NULL DEFAULT 0,"
	sql += " `" + common.FIELD_VIDEO_A + "` TEXT NOT NULL,"
	sql += " `" + common.FIELD_MEETING + "` INT NOT NULL DEFAULT 0,"
	sql += " `" + common.FIELD_LOG + "` TEXT NOT NULL,"
	sql += " `" + common.FIELD_EXAM + "` TEXT NOT NULL,"                  // New field.
	sql += " `" + common.FIELD_EXAM_CORRECT + "` INT NOT NULL DEFAULT 0," // New field.
	sql += " `" + common.FIELD_EXAM_TOTAL + "` INT NOT NULL DEFAULT 0,"   // New field.
	sql += " `" + common.FIELD_REPLAY + "` INT NOT NULL DEFAULT 0,"       // New field.
	sql += " `" + common.FIELD_SCORE + "` TEXT NOT NULL,"                 // New field.
	sql += " `" + common.FIELD_UPDATE_IP + "` VARCHAR(32) NOT NULL,"
	sql += " `" + common.FIELD_UPDATE_TIME + "` BIGINT NOT NULL,"
	sql += " `" + common.FIELD_UPDATER + "` INT NOT NULL,"
	sql += "KEY (`" + common.FIELD_USER_ID + "`),"
	sql += "KEY (`" + common.FIELD_MEETING_ID + "`)"
	sql += ") ENGINE=InnoDB DEFAULT CHARSET=utf8;"

	_, err = ms.db.Exec(sql)
	if err != nil {
		return err
	}

	sql = "CREATE TABLE IF NOT EXISTS `" + common.TABLE_MEETING_FEEDBACK + "` ("
	sql += " `" + common.FIELD_MEETING_ID + "` INT NOT NULL,"
	sql += " `" + common.FIELD_FEEDBACK + "` TEXT NOT NULL,"
	sql += "PRIMARY KEY (`" + common.FIELD_MEETING_ID + "`)"
	sql += ") ENGINE=InnoDB DEFAULT CHARSET=utf8;"

	_, err = ms.db.Exec(sql)
	return err
}

//----------------------------------------------------------------------------
// This routine yields:
//
// 1. A map from meeting ID to meeting name.
// 2. Meeting Hash tables.
// 3. User-meeting progress Hash tables.

func (ms *MeetingService) Preload(meetingOnly bool) (int, int, error) {
	// Check requirements.
	if ms.db == nil {
		return 0, 0, common.ERR_NO_DATABASE
	}
	if ms.cache == nil {
		return 0, 0, common.ERR_NO_CACHE
	}

	//----------------------------------------------------
	// Step 1.

	n1, err := (func() (int, error) {
		rest, err := ms.db.Count(common.TABLE_MEETING)
		if err != nil {
			return 0, err
		}

		i := 0
		for i < rest {
			n, err := ms.preloadMeetings(i, common.DATABASE_PRELOAD_SIZE)
			if err != nil {
				return i, err
			}
			i += n
		}
		return i, nil
	})()
	if err != nil {
		return n1, 0, err
	}

	if meetingOnly {
		return n1, 0, nil
	}

	//----------------------------------------------------
	// Step 2.

	n2, err := (func() (int, error) {
		rest, err := ms.db.Count(common.TABLE_USER_MEETING)
		if err != nil {
			return 0, err
		}

		i := 0
		for i < rest {
			n, err := ms.preloadUserProgress(i, common.DATABASE_PRELOAD_SIZE)
			if err != nil {
				return i, err
			}
			i += n
		}

		return i, nil
	})()
	if err != nil {
		return n1, n2, err
	}

	return n1, n2, nil
}

func (ms *MeetingService) preloadMeetings(start int, length int) (int, error) {
	sql := "SELECT " +
		common.FIELD_ID + "," +
		common.FIELD_NAME + "," +
		common.FIELD_SUBJECT_LIST + "," +
		common.FIELD_SECTION + "," +
		"`" + common.FIELD_TYPE + "`," +
		common.FIELD_DATA + "," +
		common.FIELD_ALLY + "," +
		common.FIELD_SCORE + "," +
		common.FIELD_SCORE_COUNT + "," +
		// common.FIELD_TEACHER_LIST + "," +
		// common.FIELD_STUDENT_LIST + "," +
		common.FIELD_NUMBER_OF_ATTENDEE + "," +
		common.FIELD_COURSEWARE_LIST + "," +
		common.FIELD_VIDEO_LIST + "," +
		common.FIELD_EXAM_LIST + "," +
		common.FIELD_REPLAY_LIST + "," +
		common.FIELD_CLASS_ID + "," +
		common.FIELD_GROUP_ID + "," +
		common.FIELD_START_TIME + "," +
		common.FIELD_END_TIME + "," +
		common.FIELD_DURATION + "," +
		common.FIELD_IS_FAKE + "," +
		common.FIELD_UPDATE_IP + "," +
		common.FIELD_UPDATE_TIME + "," +
		common.FIELD_UPDATER +
		" FROM " +
		common.TABLE_MEETING +
		" LIMIT " + strconv.Itoa(start) + "," + strconv.Itoa(length) + ";"

	rows, err := ms.db.Select(sql)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	m := make(map[string]string)
	key := common.KEY_PREFIX_MEETING + common.FIELD_ID

	cnt := 0

	id := 0
	name := ""
	sbl := ""
	section := 0
	t := 0
	data := ""
	ally := 0
	score := ""
	scoreCnt := 0
	// tl := ""
	// sl := ""
	noa := 0
	cl := ""
	vl := ""
	el := ""
	rl := ""
	cID := 0
	gID := 0
	st := 0
	et := 0
	d := 0
	fake := 0
	uip := ""
	ut := 0
	ur := 0
	for rows.Next() {
		err = rows.Scan(&id, &name, &sbl, &section, &t, &data, &ally, &score, &scoreCnt /*&tl, &sl,*/, &noa, &cl, &vl, &el, &rl, &cID, &gID, &st, &et, &d, &fake, &uip, &ut, &ur)
		if err != nil {
			return cnt, err
		}

		//------------------------------------------------
		// A map from meeting ID to meeting name.

		sID := strconv.Itoa(id)
		err = ms.cache.SetField(key, sID, name)
		if err != nil {
			return cnt, err
		}

		//------------------------------------------------
		// The meeting itself.

		m[common.FIELD_NAME] = name
		m[common.FIELD_SUBJECT_LIST] = sbl
		m[common.FIELD_SECTION] = strconv.Itoa(section)
		m[common.FIELD_TYPE] = strconv.Itoa(t)
		m[common.FIELD_DATA] = data
		m[common.FIELD_ALLY] = strconv.Itoa(ally)
		m[common.FIELD_SCORE] = score
		m[common.FIELD_SCORE_COUNT] = strconv.Itoa(scoreCnt)
		// m[common.FIELD_TEACHER_LIST] = tl
		m[common.FIELD_NUMBER_OF_ATTENDEE] = strconv.Itoa(noa)
		// m[common.FIELD_STUDENT_LIST] = sl
		m[common.FIELD_COURSEWARE_LIST] = cl
		m[common.FIELD_VIDEO_LIST] = vl
		m[common.FIELD_EXAM_LIST] = el
		m[common.FIELD_REPLAY_LIST] = rl
		m[common.FIELD_CLASS_ID] = strconv.Itoa(cID)
		m[common.FIELD_GROUP_ID] = strconv.Itoa(gID)
		m[common.FIELD_START_TIME] = strconv.Itoa(st)
		m[common.FIELD_END_TIME] = strconv.Itoa(et)
		m[common.FIELD_DURATION] = strconv.Itoa(d)
		m[common.FIELD_IS_FAKE] = strconv.Itoa(fake)
		m[common.FIELD_UPDATE_IP] = uip
		m[common.FIELD_UPDATE_TIME] = strconv.Itoa(ut)
		m[common.FIELD_UPDATER] = strconv.Itoa(ur)

		err = ms.cache.SetFields(common.KEY_PREFIX_MEETING+sID, m)
		if err != nil {
			return cnt, err
		}

		cnt++
	}

	return cnt, nil
}

func (ms *MeetingService) preloadUserProgress(start int, length int) (int, error) {
	sql := "SELECT " +
		common.FIELD_USER_ID + "," +
		common.FIELD_MEETING_ID + "," +
		// common.FIELD_COURSEWARE + "," +
		common.FIELD_COURSEWARE_A + "," +
		// common.FIELD_VIDEO + "," +
		common.FIELD_VIDEO_A + "," +
		common.FIELD_MEETING + "," +
		common.FIELD_LOG + "," +
		common.FIELD_EXAM + "," +
		common.FIELD_EXAM_CORRECT + "," +
		common.FIELD_EXAM_TOTAL + "," +
		common.FIELD_REPLAY + "," +
		common.FIELD_SCORE + "," +
		common.FIELD_UPDATE_IP + "," +
		common.FIELD_UPDATE_TIME + "," +
		common.FIELD_UPDATER +
		" FROM " +
		common.TABLE_USER_MEETING +
		" LIMIT " + strconv.Itoa(start) + "," + strconv.Itoa(length) + ";"

	rows, err := ms.db.Select(sql)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	m := make(map[string]string)
	cnt := 0

	uID := 0
	mID := 0
	// cw := 0
	cwa := ""
	// vd := 0
	va := ""
	mt := 0
	lg := ""
	e := ""
	ec := 0
	et := 0
	rp := 0
	score := ""
	uip := ""
	ut := 0
	ur := 0

	for rows.Next() {
		err = rows.Scan(&uID, &mID, &cwa, &va, &mt, &lg, &e, &ec, &et, &rp, &score, &uip, &ut, &ur)
		if err != nil {
			return cnt, err
		}

		// m[common.FIELD_COURSEWARE] = strconv.Itoa(cw)
		m[common.FIELD_COURSEWARE_A] = cwa
		// m[common.FIELD_VIDEO] = strconv.Itoa(vd)
		m[common.FIELD_VIDEO_A] = va
		m[common.FIELD_MEETING] = strconv.Itoa(mt)
		m[common.FIELD_LOG] = lg
		m[common.FIELD_EXAM] = e
		m[common.FIELD_EXAM_CORRECT] = strconv.Itoa(ec)
		m[common.FIELD_EXAM_TOTAL] = strconv.Itoa(et)
		m[common.FIELD_REPLAY] = strconv.Itoa(rp)
		m[common.FIELD_SCORE] = score
		m[common.FIELD_UPDATE_IP] = uip
		m[common.FIELD_UPDATE_TIME] = strconv.Itoa(ut)
		m[common.FIELD_UPDATER] = strconv.Itoa(ur)

		err = ms.cache.SetFields(common.KEY_PREFIX_MEETING+strconv.Itoa(mID)+":"+strconv.Itoa(uID), m)
		if err != nil {
			return cnt, err
		}
		cnt++
	}

	return cnt, nil
}

//----------------------------------------------------------------------------

func (ms *MeetingService) GetUUID() string {
	return strconv.FormatInt(rand.Int63(), 32)
}

//----------------------------------------------------------------------------
// Database: Required
// Cache   : Compatible

func (ms *MeetingService) AddMeeting(name string, subjects []int, section int /*teacherList string,*/, studentList string, startTime int, duration int, t int, data string, classID int, groupID int, session *Session, unset bool) (int, error) {
	// Check requirements.
	if ms.db == nil {
		return 0, common.ERR_NO_SERVICE
	}

	// Check authority.
	isAllowed := false
	if session.IsSystem() {
		isAllowed = true
	} else if session.IsAssistant() {
		if session.GroupID == groupID {
			isAllowed = true
		}
	}
	if !isAllowed {
		return 0, common.ERR_NO_AUTHORITY
	}

	// Check inputs.
	if len(name) == 0 {
		return 0, errors.New("Empty meeting name.")
	}
	sName := common.Escape(name)
	if len(sName) == 0 {
		return 0, errors.New("Invalid meeting name.")
	}

	sSubjectList := ""
	if (subjects != nil) && (len(subjects) > 0) {
		sSubjectList = common.IntArrayToString(subjects)
	}

	students := common.StringToStringArray(studentList)
	nStudents := len(students)

	if !unset {
		if t != common.MEETING_TYPE_FOR_SELF_STUDYING {
			if int(time.Now().Unix()) > startTime {
				return 0, errors.New("Invalid start time.")
			}
		}

		// Check duration.
		if duration < 1 || duration > MAX_MEETING_DURATION {
			return 0, errors.New("Invalid meeting duration.")
		}
	}

	sSection := strconv.Itoa(section)
	sType := strconv.Itoa(t)
	sData := common.Escape(data)
	sClassID := strconv.Itoa(classID)
	sGroupID := strconv.Itoa(groupID)
	sStartTime := strconv.Itoa(startTime)
	sDuration := strconv.Itoa(duration)
	timestamp := common.GetTimeString()
	sUserID := strconv.Itoa(session.UserID)

	sql := "INSERT INTO " + common.TABLE_MEETING +
		" (" +
		common.FIELD_NAME + "," +
		common.FIELD_SUBJECT_LIST + "," +
		common.FIELD_SECTION + "," +
		"`" + common.FIELD_TYPE + "`," +
		common.FIELD_DATA + "," +
		common.FIELD_ALLY + "," +
		common.FIELD_SCORE + "," +
		common.FIELD_SCORE_COUNT + "," +
		common.FIELD_TEACHER_LIST + "," +
		common.FIELD_STUDENT_LIST + "," +
		common.FIELD_NUMBER_OF_ATTENDEE + "," +
		common.FIELD_COURSEWARE_LIST + "," +
		common.FIELD_VIDEO_LIST + "," +
		common.FIELD_EXAM_LIST + "," +
		common.FIELD_REPLAY_LIST + "," +
		common.FIELD_CLASS_ID + "," +
		common.FIELD_GROUP_ID + "," +
		common.FIELD_START_TIME + "," +
		common.FIELD_DURATION + "," +
		common.FIELD_UPDATE_IP + "," +
		common.FIELD_UPDATE_TIME + "," +
		common.FIELD_UPDATER +
		") VALUES (" +
		"'" + sName + "'," +
		"'" + sSubjectList + "'," +
		sSection + "," +
		sType + "," + // type
		"'" + sData + "'," + // data
		"0," + // ally
		"'',0," + // scores and their count
		"''," +
		"''," +
		"0," + // number of attendee
		"''," + // courseware list
		"''," + // video list
		"''," + // exam list
		"''," + // replay list
		sClassID + "," +
		sGroupID + "," +
		sStartTime + "," +
		sDuration + ",'" +
		session.IP + "'," +
		timestamp + "," +
		sUserID + ");"

	meetingID, err := ms.db.Insert(sql, 1)
	if err != nil {
		return 0, err
	}

	sMeetingID := strconv.FormatInt(meetingID, 10)

	if nStudents > 0 {
		sql = "INSERT INTO " + common.TABLE_USER_MEETING +
			" (" +
			common.FIELD_USER_ID + "," +
			common.FIELD_MEETING_ID + "," +
			common.FIELD_COURSEWARE_A + "," +
			common.FIELD_VIDEO_A + "," +
			common.FIELD_LOG + "," +
			common.FIELD_EXAM + "," +
			common.FIELD_EXAM_CORRECT + "," +
			common.FIELD_EXAM_TOTAL + "," +
			common.FIELD_REPLAY + "," +
			common.FIELD_SCORE + "," +
			common.FIELD_UPDATE_TIME + "," +
			common.FIELD_UPDATE_IP + "," +
			common.FIELD_UPDATER +
			") VALUES "

		first := true
		for i := 0; i < nStudents; i++ {
			if len(students[i]) == 0 {
				continue
			}

			if first {
				first = false
			} else {
				sql += ","
			}

			sql += "(" +
				students[i] + "," +
				sMeetingID + ",'','','','',0,0,0,''," +
				timestamp + "," +
				"'" + session.IP + "'," +
				sUserID +
				")"
		}
		sql += ";"

		_, err = ms.db.Insert(sql, int64(nStudents))
		if err != nil {
			return 0, err
		}
	}

	if ms.cache != nil {
		key := common.KEY_PREFIX_MEETING + sMeetingID

		m := make(map[string]string)
		m[common.FIELD_NAME] = sName
		m[common.FIELD_SUBJECT_LIST] = sSubjectList
		m[common.FIELD_SECTION] = sSection
		m[common.FIELD_TYPE] = sType
		m[common.FIELD_DATA] = sData
		m[common.FIELD_ALLY] = "0"
		m[common.FIELD_SCORE] = ""
		m[common.FIELD_SCORE_COUNT] = "0"
		// m[common.FIELD_TEACHER_LIST] = teacherList
		// m[common.FIELD_STUDENT_LIST] = studentList
		m[common.FIELD_NUMBER_OF_ATTENDEE] = "0"
		m[common.FIELD_COURSEWARE_LIST] = ""
		m[common.FIELD_VIDEO_LIST] = ""
		m[common.FIELD_EXAM_LIST] = ""
		m[common.FIELD_START_TIME] = sStartTime
		m[common.FIELD_DURATION] = sDuration
		m[common.FIELD_END_TIME] = "0"
		m[common.FIELD_REPLAY_LIST] = ""
		m[common.FIELD_CLASS_ID] = sClassID
		m[common.FIELD_GROUP_ID] = sGroupID
		m[common.FIELD_UPDATE_IP] = session.IP
		m[common.FIELD_UPDATE_TIME] = timestamp
		m[common.FIELD_UPDATER] = sUserID

		err = ms.cache.SetFields(key, m)
		if err != nil {
			return 0, err
		}

		//------------------------------------------------

		for k, _ := range m {
			delete(m, k)
		}

		// m[common.FIELD_COURSEWARE] = "0"
		m[common.FIELD_COURSEWARE_A] = ""
		// m[common.FIELD_VIDEO] = "0"
		m[common.FIELD_VIDEO_A] = ""
		m[common.FIELD_MEETING] = "0"
		m[common.FIELD_LOG] = ""
		m[common.FIELD_EXAM] = ""
		m[common.FIELD_EXAM_CORRECT] = "0"
		m[common.FIELD_EXAM_TOTAL] = "0"
		m[common.FIELD_REPLAY] = "0"
		m[common.FIELD_SCORE] = ""
		m[common.FIELD_UPDATE_IP] = session.IP
		m[common.FIELD_UPDATE_TIME] = timestamp
		m[common.FIELD_UPDATER] = sUserID

		for i := 0; i < nStudents; i++ {
			err = ms.cache.SetFields(key+":"+students[i], m)
			if err != nil {
				return 0, err
			}
		}
	}

	return int(meetingID), nil
}

//----------------------------------------------------------------------------
// Database: Required.
// Cache   : Compatible.

func (ms *MeetingService) ChangeMeeting(meetingID int, name string, subjects []int, section int, startTime int, duration int, t int, data string, session *Session) (int, error) {
	// Check requirements.
	if ms.db == nil {
		return -1, common.ERR_NO_SERVICE
	}

	// Check authority.
	mi, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return -2, err
	}

	// Check inputs.
	sName := common.Escape(name)
	if len(sName) == 0 {
		return -3, common.ERR_INVALID_NAME
	}
	if (t != common.MEETING_TYPE_FOR_SELF_STUDYING) && (duration > MAX_MEETING_DURATION) {
		return -4, common.ERR_OUT_OF_TIME
	}

	sMeetingID := strconv.Itoa(meetingID)
	sSubjectList := common.IntArrayToString(subjects)
	sSection := strconv.Itoa(section)
	sStartTime := strconv.Itoa(startTime)
	sDuration := strconv.Itoa(duration)
	sType := strconv.Itoa(t)
	sData := common.Escape(data)

	sUpdateTime := common.GetTimeString()
	sUpdateIP := session.IP
	sUpdater := strconv.Itoa(session.UserID)

	s := "UPDATE " +
		common.TABLE_MEETING +
		" SET " +
		common.FIELD_NAME + "='" + sName + "'," +
		common.FIELD_SUBJECT_LIST + "='" + sSubjectList + "'," +
		common.FIELD_SECTION + "=" + sSection + "," +
		common.FIELD_START_TIME + "=" + sStartTime + "," +
		common.FIELD_DURATION + "=" + sDuration + "," +
		"`" + common.FIELD_TYPE + "`=" + sType + "," +
		common.FIELD_DATA + "='" + sData + "'," +
		common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
		common.FIELD_UPDATE_TIME + "=" + sUpdateTime + "," +
		common.FIELD_UPDATER + "=" + sUpdater +
		" WHERE " +
		common.FIELD_ID + "=" + sMeetingID + ";"

	_, err = ms.db.Exec(s)
	if err != nil {
		return -5, err
	}

	if ms.cache != nil {
		m := make(map[string]string)
		m[common.FIELD_NAME] = sName
		m[common.FIELD_SUBJECT_LIST] = sSubjectList
		m[common.FIELD_SECTION] = sSection
		m[common.FIELD_START_TIME] = sStartTime
		m[common.FIELD_DURATION] = sDuration
		m[common.FIELD_TYPE] = sType
		m[common.FIELD_DATA] = sData
		m[common.FIELD_UPDATE_IP] = sUpdateIP
		m[common.FIELD_UPDATE_TIME] = sUpdateTime
		m[common.FIELD_UPDATER] = sUpdater

		err = ms.cache.SetFields(common.KEY_PREFIX_MEETING+sMeetingID, m)
		if err != nil {
			return -6, err
		}
	}

	if mi.StartTime != startTime || mi.Duration != duration || mi.Type != t || mi.Name != sName {
		return 1, nil
	} else {
		return 0, nil
	}
}

//----------------------------------------------------------------------------
// Database: Required.
// Cache   : Compatible.

func (ms *MeetingService) CloneMeeting(meetingID int, destClassID int, session *Session) (int, error) {
	// Check requirements.
	if ms.db == nil {
		return 0, common.ERR_NO_SERVICE
	}

	// Check authority.
	mi, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return 0, err
	}

	sbl := common.IntArrayToString(mi.Subjects)
	cwl := common.StringArrayToString(mi.Coursewares)
	vl := common.StringArrayToString(mi.Videos)
	el := common.StringArrayToString(mi.Exams)

	sSection := strconv.Itoa(mi.Section)
	sStartTime := strconv.Itoa(mi.StartTime)
	sDuration := strconv.Itoa(mi.Duration)
	sType := strconv.Itoa(mi.Type)
	sClassID := strconv.Itoa(destClassID)
	sGroupID := strconv.Itoa(mi.GroupID)
	sUpdater := strconv.Itoa(session.UserID)
	sUpdateTime := common.GetTimeString()
	sUpdateIP := session.IP

	sql := "INSERT INTO " + common.TABLE_MEETING +
		" (" +
		common.FIELD_NAME + "," +
		common.FIELD_SUBJECT_LIST + "," +
		common.FIELD_SECTION + "," +
		"`" + common.FIELD_TYPE + "`," +
		common.FIELD_DATA + "," +
		common.FIELD_ALLY + "," +
		common.FIELD_SCORE + "," +
		common.FIELD_SCORE_COUNT + "," +
		common.FIELD_TEACHER_LIST + "," +
		common.FIELD_STUDENT_LIST + "," +
		common.FIELD_NUMBER_OF_ATTENDEE + "," +
		common.FIELD_COURSEWARE_LIST + "," +
		common.FIELD_VIDEO_LIST + "," +
		common.FIELD_EXAM_LIST + "," +
		common.FIELD_REPLAY_LIST + "," +
		common.FIELD_CLASS_ID + "," +
		common.FIELD_GROUP_ID + "," +
		common.FIELD_START_TIME + "," +
		common.FIELD_DURATION + "," +
		common.FIELD_UPDATE_IP + "," +
		common.FIELD_UPDATE_TIME + "," +
		common.FIELD_UPDATER +
		") VALUES (" +
		"'" + mi.Name + "'," +
		"'" + sbl + "'," +
		sSection + "," +
		sType + "," + // type
		"'" + mi.Data + "'," + // data
		"0," + // ally
		"'',0," + // scores and their count
		"''," + // teacher list
		"''," + // student list
		"0," + // number of attendee
		"'" + cwl + "'," + // courseware list
		"'" + vl + "'," + // video list
		"'" + el + "'," + // exam list
		"''," + // replay list
		// "0," +
		sClassID + "," +
		sGroupID + "," +
		sStartTime + "," + // start time
		sDuration + "," + // duration
		"'" + sUpdateIP + "'," +
		sUpdateTime + "," +
		sUpdater + ");"

	id, err := ms.db.Insert(sql, 1)
	if err != nil {
		return int(id), err
	}

	if ms.cache != nil {
		key := common.KEY_PREFIX_MEETING + strconv.FormatInt(id, 10)

		m := make(map[string]string)
		m[common.FIELD_NAME] = mi.Name
		m[common.FIELD_SUBJECT_LIST] = sbl
		m[common.FIELD_SECTION] = sSection
		m[common.FIELD_TYPE] = sType
		m[common.FIELD_DATA] = mi.Data
		m[common.FIELD_ALLY] = "0"
		m[common.FIELD_SCORE] = ""
		m[common.FIELD_SCORE_COUNT] = "0"
		// m[common.FIELD_TEACHER_LIST] = ""
		// m[common.FIELD_STUDENT_LIST] = ""
		m[common.FIELD_NUMBER_OF_ATTENDEE] = "0"
		m[common.FIELD_COURSEWARE_LIST] = cwl
		m[common.FIELD_VIDEO_LIST] = vl
		m[common.FIELD_EXAM_LIST] = el
		m[common.FIELD_START_TIME] = sStartTime
		m[common.FIELD_DURATION] = sDuration
		m[common.FIELD_END_TIME] = "0"
		m[common.FIELD_REPLAY_LIST] = ""
		m[common.FIELD_CLASS_ID] = sClassID
		m[common.FIELD_GROUP_ID] = sGroupID
		m[common.FIELD_UPDATE_IP] = sUpdateIP
		m[common.FIELD_UPDATE_TIME] = sUpdateTime
		m[common.FIELD_UPDATER] = sUpdater

		if err = ms.cache.SetFields(key, m); err != nil {
			return int(id), err
		}
	}

	return int(id), nil
}

//----------------------------------------------------------------------------
// Database: Required.
// Cache   : Compatible.

func (ms *MeetingService) SyncMeeting(from int, to []int, data string, session *Session) error {
	if ms.db == nil {
		return common.ERR_NO_SERVICE
	}

	okay := false
	for i := 0; i < len(data); i++ {
		if data[i] == '1' {
			okay = true
			break
		}
	}
	if !okay {
		return nil
	}

	fromMi, err := ms.GetMeeting(from, session, true)
	if err != nil {
		return err
	}

	sTo := ""
	first := true
	n := len(to)
	for i := 0; i < n; i++ {
		_, err = ms.GetMeeting(to[i], session, true)
		if err != nil {
			return err
		}

		if first {
			first = false
		} else {
			sTo += ","
		}
		sTo += strconv.Itoa(to[i])
	}

	sUpdater := strconv.Itoa(session.UserID)
	sUpdateTime := common.GetTimeString()
	sUpdateIP := session.IP

	/*
		0: name
		1: type
		2: data
		3: section
		4: startTime
		5: duration
		6: endTime
		7: replay
		8: courseware
		9: exam
		10: video
		11: subject
	*/

	if err = (func() error {
		m := make(map[string]string)

		s := "UPDATE " +
			common.TABLE_MEETING +
			" SET "

		if data[0] == '1' {
			s += common.FIELD_NAME + "='" + fromMi.Name + "',"
			m[common.FIELD_NAME] = fromMi.Name
		}
		if data[1] == '1' {
			v := strconv.Itoa(fromMi.Type)
			s += "`" + common.FIELD_TYPE + "`=" + v + ","
			m[common.FIELD_TYPE] = v
		}
		if data[2] == '1' {
			s += common.FIELD_DATA + "='" + fromMi.Data + "',"
			m[common.FIELD_DATA] = fromMi.Data
		}
		if data[3] == '1' {
			v := strconv.Itoa(fromMi.Section)
			s += common.FIELD_SECTION + "=" + v + ","
			m[common.FIELD_SECTION] = v
		}
		if data[4] == '1' {
			v := strconv.Itoa(fromMi.StartTime)
			s += common.FIELD_START_TIME + "=" + v + ","
			m[common.FIELD_START_TIME] = v
		}
		if data[5] == '1' {
			v := strconv.Itoa(fromMi.Duration)
			s += common.FIELD_DURATION + "=" + v + ","
			m[common.FIELD_DURATION] = v
		}
		if data[6] == '1' {
			v := strconv.Itoa(fromMi.EndTime)
			s += common.FIELD_END_TIME + "=" + v + ","
			m[common.FIELD_END_TIME] = v
		}
		if data[7] == '1' {
			v := common.StringArrayToString(fromMi.Replays)
			s += common.FIELD_REPLAY_LIST + "='" + v + "',"
			m[common.FIELD_REPLAY_LIST] = v
		}
		if data[8] == '1' {
			v := common.StringArrayToString(fromMi.Coursewares)
			s += common.FIELD_COURSEWARE_LIST + "='" + v + "',"
			m[common.FIELD_COURSEWARE_LIST] = v
		}
		if data[9] == '1' {
			v := common.StringArrayToString(fromMi.Exams)
			s += common.FIELD_EXAM_LIST + "='" + v + "',"
			m[common.FIELD_EXAM_LIST] = v
		}
		if data[10] == '1' {
			v := common.StringArrayToString(fromMi.Videos)
			s += common.FIELD_VIDEO_LIST + "='" + v + "',"
			m[common.FIELD_VIDEO_LIST] = v
		}
		if data[11] == '1' {
			v := common.IntArrayToString(fromMi.Subjects)
			s += common.FIELD_SUBJECT_LIST + "='" + v + "',"
			m[common.FIELD_SUBJECT_LIST] = v
		}

		s += common.FIELD_UPDATER + "=" + sUpdater + "," +
			common.FIELD_UPDATE_TIME + "=" + sUpdateTime + "," +
			common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'" +
			" WHERE " +
			common.FIELD_ID + " IN (" + sTo + ");"

		_, err := ms.db.Exec(s)
		if err != nil {
			return err
		}

		if len(m) > 0 {
			m[common.FIELD_UPDATER] = sUpdater
			m[common.FIELD_UPDATE_IP] = sUpdateIP
			m[common.FIELD_UPDATE_TIME] = sUpdateTime

			if ms.cache != nil {
				for i := 0; i < n; i++ {
					err = ms.cache.SetFields(common.KEY_PREFIX_MEETING+strconv.Itoa(to[i]), m)
					if err != nil {
						return err
					}
				}
			}
		}

		return nil
	})(); err != nil {
		return err
	}

	return nil
}

//----------------------------------------------------------------------------
// Database: Required
// Cache   : Compatible

func (ms *MeetingService) ChangeMeetingSubjects(meetingID int, subjects []int, session *Session) error {
	// Check requirements.
	if ms.db == nil {
		return common.ERR_NO_SERVICE
	}

	// Check authority.
	_, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return err
	}

	if subjects == nil {
		return nil
	}

	sSubjectList := common.IntArrayToString(subjects)
	sMeetingID := strconv.Itoa(meetingID)

	sUpdater := strconv.Itoa(session.UserID)
	sUpdateIP := session.IP
	sUpdateTime := common.GetTimeString()

	// Update database.
	sql := "UPDATE " +
		common.TABLE_MEETING +
		" SET " +
		common.FIELD_SUBJECT_LIST + "='" + sSubjectList + "'," +
		common.FIELD_UPDATER + "=" + sUpdater + "," +
		common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
		common.FIELD_UPDATE_TIME + "=" + sUpdateTime +
		" WHERE " +
		common.FIELD_ID + "=" + sMeetingID + ";"

	_, err = ms.db.Exec(sql)
	if err != nil {
		return err
	}

	// Update cache.
	if ms.cache != nil {
		m := make(map[string]string)

		m[common.FIELD_SUBJECT_LIST] = sSubjectList
		m[common.FIELD_UPDATER] = sUpdater
		m[common.FIELD_UPDATE_IP] = sUpdateIP
		m[common.FIELD_UPDATE_TIME] = sUpdateTime

		err = ms.cache.SetFields(common.KEY_PREFIX_MEETING+sMeetingID, m)
		if err != nil {
			return err
		}
	}

	return nil
}

//----------------------------------------------------------------------------
// Database: Required
// Cache   : Compatible

func (ms *MeetingService) ChangeMeetingSection(meetingID int, section int, session *Session) error {
	// Check requirements.
	if ms.db == nil {
		return common.ERR_NO_SERVICE
	}

	// Check authority.
	mi, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return err
	}

	// Check via cache.
	if mi.Section == section {
		return nil
	}

	sSection := strconv.Itoa(section)
	sMeetingID := strconv.Itoa(meetingID)

	sUpdater := strconv.Itoa(session.UserID)
	sUpdateIP := session.IP
	sUpdateTime := common.GetTimeString()

	// Update database.
	sql := "UPDATE " +
		common.TABLE_MEETING +
		" SET " +
		common.FIELD_SECTION + "=" + sSection + "," +
		common.FIELD_UPDATER + "=" + sUpdater + "," +
		common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
		common.FIELD_UPDATE_TIME + "=" + sUpdateTime +
		" WHERE " +
		common.FIELD_ID + "=" + sMeetingID + ";"

	_, err = ms.db.Exec(sql)
	if err != nil {
		return err
	}

	// Update cache.
	if ms.cache != nil {
		m := make(map[string]string)

		m[common.FIELD_SECTION] = sSection
		m[common.FIELD_UPDATER] = sUpdater
		m[common.FIELD_UPDATE_IP] = sUpdateIP
		m[common.FIELD_UPDATE_TIME] = sUpdateTime

		err = ms.cache.SetFields(common.KEY_PREFIX_MEETING+sMeetingID, m)
		if err != nil {
			return err
		}
	}

	return nil
}

//----------------------------------------------------------------------------
// Database: Required
// Cache   : Compatible

func (ms *MeetingService) ChangeMeetingType(meetingID int, t int, data string, session *Session) error {
	// Check requirements.
	if ms.db == nil {
		return common.ERR_NO_SERVICE
	}

	// Check authority.
	mri, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return err
	}

	// Check via cache.
	sData := common.Escape(data)
	if (mri.Type == t) && (mri.Data == sData) {
		return nil
	}

	sType := strconv.Itoa(t)
	sMeetingID := strconv.Itoa(meetingID)

	sUpdater := strconv.Itoa(session.UserID)
	sUpdateIP := session.IP
	sUpdateTime := common.GetTimeString()

	// Update database.
	sql := "UPDATE " +
		common.TABLE_MEETING +
		" SET " +
		"`" + common.FIELD_TYPE + "`=" + sType + "," +
		common.FIELD_DATA + "='" + sData + "'," +
		common.FIELD_UPDATER + "=" + sUpdater + "," +
		common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
		common.FIELD_UPDATE_TIME + "=" + sUpdateTime +
		" WHERE " +
		common.FIELD_ID + "=" + sMeetingID + ";"

	_, err = ms.db.Exec(sql)
	if err != nil {
		return err
	}

	// Update cache.
	if ms.cache != nil {
		m := make(map[string]string)

		m[common.FIELD_TYPE] = sType
		m[common.FIELD_DATA] = sData
		m[common.FIELD_UPDATER] = sUpdater
		m[common.FIELD_UPDATE_IP] = sUpdateIP
		m[common.FIELD_UPDATE_TIME] = sUpdateTime

		err = ms.cache.SetFields(common.KEY_PREFIX_MEETING+sMeetingID, m)
		if err != nil {
			return err
		}
	}

	return nil
}

//----------------------------------------------------------------------------
// Database: Required
// Cache   : Compatible

func (ms *MeetingService) ChangeMeetingAlly(meetingID int, ally int, session *Session) error {
	// Check requirements.
	if ms.db == nil {
		return common.ERR_NO_SERVICE
	}

	if meetingID <= 0 || ally <= 0 {
		return common.ERR_INVALID_MEETING
	}

	// Check authority.
	_, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return err
	}
	_, err = ms.GetMeeting(ally, session, true)
	if err != nil {
		return err
	}

	sMeetingID := strconv.Itoa(meetingID)
	sAlly := strconv.Itoa(ally)
	sUpdater := strconv.Itoa(session.UserID)
	sUpdateIP := session.IP
	sUpdateTime := common.GetTimeString()

	err = (func() error {
		s := "UPDATE " +
			common.TABLE_MEETING +
			" SET " +
			common.FIELD_ALLY + "=" + sAlly + "," +
			common.FIELD_UPDATER + "=" + sUpdater + "," +
			common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
			common.FIELD_UPDATE_TIME + "=" + sUpdateTime +
			" WHERE " +
			common.FIELD_ID + "=" + sMeetingID + ";"

		_, err := ms.db.Exec(s)
		if err != nil {
			return err
		}

		if ms.cache != nil {
			m := make(map[string]string)

			m[common.FIELD_ALLY] = sAlly
			m[common.FIELD_UPDATER] = sUpdater
			m[common.FIELD_UPDATE_IP] = sUpdateIP
			m[common.FIELD_UPDATE_TIME] = sUpdateTime

			err = ms.cache.SetFields(common.KEY_PREFIX_MEETING+sMeetingID, m)
			if err != nil {
				return err
			}
		}

		return nil
	})()
	if err != nil {
		return err
	}

	return nil
}

//----------------------------------------------------------------------------
// Database: Required
// Cache   : Compatible

func (ms *MeetingService) ChangeMeetingName(meetingID int, name string, session *Session) error {
	// Check requirements.
	if ms.db == nil {
		return common.ERR_NO_SERVICE
	}

	// Check authority.
	_, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return err
	}

	// Check meeting name.
	if len(name) == 0 {
		return errors.New("Empty meeting name.")
	}
	sName := common.Escape(name)
	if len(sName) == 0 {
		return errors.New("Invalid meeting name.")
	}

	timestamp := common.GetTimeString()
	sUserID := strconv.Itoa(session.UserID)
	sMeetingID := strconv.Itoa(meetingID)

	sql := "UPDATE " + common.TABLE_MEETING +
		" SET " +
		common.FIELD_NAME + "='" + sName + "'," +
		common.FIELD_UPDATE_IP + "='" + session.IP + "'," +
		common.FIELD_UPDATE_TIME + "=" + timestamp + "," +
		common.FIELD_UPDATER + "=" + sUserID +
		" WHERE " +
		common.FIELD_ID + "=" + sMeetingID + ";"

	_, err = ms.db.Exec(sql)
	if err != nil {
		return err
	}

	if ms.cache != nil {
		m := make(map[string]string)

		m[common.FIELD_NAME] = sName
		m[common.FIELD_UPDATE_IP] = session.IP
		m[common.FIELD_UPDATE_TIME] = timestamp
		m[common.FIELD_UPDATER] = sUserID

		err = ms.cache.SetFields(common.KEY_PREFIX_MEETING+sMeetingID, m)
		if err != nil {
			return err
		}
	}

	return nil
}

//----------------------------------------------------------------------------
// Database: Required
// Cache   : Compatible

func (ms *MeetingService) ChangeMeetingTime(meetingID int, startTime int, duration int, session *Session) error {
	// Check requirements.
	if ms.db == nil {
		return common.ERR_NO_SERVICE
	}

	// Check authority.
	_, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return err
	}

	// Check start time.
	if !session.IsAssistantOrAbove() {
		now := time.Now().Unix()
		if now > int64(startTime) {
			return errors.New("Invalid start time.")
		}
	}

	// Check duration.
	if duration < 1 || duration > MAX_MEETING_DURATION {
		return errors.New("Invalid meeting duration.")
	}

	sStartTime := strconv.Itoa(startTime)
	sDuration := strconv.Itoa(duration)
	timestamp := common.GetTimeString()
	sUserID := strconv.Itoa(session.UserID)
	sMeetingID := strconv.Itoa(meetingID)

	sql := "UPDATE " + common.TABLE_MEETING +
		" SET " +
		common.FIELD_START_TIME + "=" + sStartTime + "," +
		common.FIELD_DURATION + "=" + sDuration + "," +
		common.FIELD_UPDATE_IP + "='" + session.IP + "'," +
		common.FIELD_UPDATE_TIME + "=" + timestamp + "," +
		common.FIELD_UPDATER + "=" + sUserID +
		" WHERE " +
		common.FIELD_ID + "=" + sMeetingID + ";"

	_, err = ms.db.Exec(sql)
	if err != nil {
		return err
	}

	if ms.cache != nil {
		m := make(map[string]string)

		m[common.FIELD_START_TIME] = sStartTime
		m[common.FIELD_DURATION] = sDuration
		m[common.FIELD_UPDATE_IP] = session.IP
		m[common.FIELD_UPDATE_TIME] = timestamp
		m[common.FIELD_UPDATER] = sUserID

		err = ms.cache.SetFields(common.KEY_PREFIX_MEETING+sMeetingID, m)
		if err != nil {
			return err
		}
	}

	return nil
}

//----------------------------------------------------------------------------
// Database: Required
// Cache   : Compatible

func (ms *MeetingService) EndMeeting(meetingID int, session *Session) error {
	// Check requirements.
	if ms.db == nil {
		return common.ERR_NO_SERVICE
	}

	// Check authority.
	mi, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return err
	}
	if mi.EndTime > 0 {
		return common.ERR_MEETING_CLOSED
	}

	now := (int)(time.Now().Unix())
	if now < mi.StartTime {
		return common.ERR_OUT_OF_TIME
	}

	sUserID := strconv.Itoa(session.UserID)
	sMeetingID := strconv.Itoa(meetingID)
	timestamp := common.GetTimeString()

	sql := "UPDATE " + common.TABLE_MEETING +
		" SET " +
		common.FIELD_END_TIME + "=" + timestamp + "," +
		common.FIELD_UPDATE_IP + "='" + session.IP + "'," +
		common.FIELD_UPDATE_TIME + "=" + timestamp + "," +
		common.FIELD_UPDATER + "=" + sUserID +
		" WHERE " +
		common.FIELD_ID + "=" + sMeetingID + ";"

	_, err = ms.db.Exec(sql)
	if err != nil {
		return err
	}

	//----------------------------------------------------

	if ms.cache != nil {
		key := common.KEY_PREFIX_MEETING + sMeetingID

		m := make(map[string]string)
		m[common.FIELD_END_TIME] = timestamp
		m[common.FIELD_UPDATE_IP] = session.IP
		m[common.FIELD_UPDATE_TIME] = timestamp
		m[common.FIELD_UPDATER] = sUserID

		err = ms.cache.SetFields(key, m)
		if err != nil {
			return err
		}
	}

	return nil
}

//----------------------------------------------------------------------------
// Database: Compatible.
// Cache   : Compatible.

func (ms *MeetingService) GetMeeting(meetingID int, session *Session, checkAuthority bool) (*MeetingInfo, error) {
	// Check inputs.
	if meetingID <= 0 {
		return nil, common.ERR_INVALID_MEETING
	}
	sMeetingID := strconv.Itoa(meetingID)

	// Get it via cache.
	if ms.cache != nil {
		if m, err := ms.cache.GetAllFields(common.KEY_PREFIX_MEETING + sMeetingID); err == nil {
			if mi := NewMeetingInfoFromMap(m, meetingID); mi != nil {
				if !checkAuthority {
					return mi, nil
				}

				// Check authority.
				allowed := false
				switch session.GroupID {
				case common.GROUP_ID_FOR_STUDENT, common.GROUP_ID_FOR_TEACHER, common.GROUP_ID_FOR_KEEPER:
					if session.UserID >= common.VALUE_MINIMAL_TEMPERARY_USER_ID {
						allowed = true
					} else {
						// TODO: The following code fragment should be moved to class service.

						// Get the class list of this user via cache.
						s, err := ms.cache.GetField(common.KEY_PREFIX_USER+strconv.Itoa(session.UserID), common.FIELD_CLASS_LIST)
						if err == nil {
							if common.InList(strconv.Itoa(mi.ClassID), s) {
								allowed = true
							}
						}
					}
					// allowed = common.InIntArray(session.UserID, mi.Teachers)
				case common.GROUP_ID_FOR_SYSTEM:
					allowed = true
				default:
					allowed = (mi.GroupID == session.GroupID)
				}
				if !allowed {
					return nil, common.ERR_NO_AUTHORITY
				}

				return mi, nil
			}
		}
	}

	// Get it via database.
	if ms.db != nil {
		sql := "SELECT " +
			common.FIELD_ID + "," +
			common.FIELD_NAME + "," +
			common.FIELD_SECTION + "," +
			"`" + common.FIELD_TYPE + "`," +
			common.FIELD_DATA + "," +
			common.FIELD_SCORE + "," +
			common.FIELD_SCORE_COUNT + "," +
			// common.FIELD_TEACHER_LIST + "," +
			// common.FIELD_STUDENT_LIST + "," +
			common.FIELD_NUMBER_OF_ATTENDEE + "," +
			common.FIELD_COURSEWARE_LIST + "," +
			common.FIELD_VIDEO_LIST + "," +
			common.FIELD_EXAM_LIST + "," +
			common.FIELD_REPLAY_LIST + "," +
			common.FIELD_CLASS_ID + "," +
			common.FIELD_GROUP_ID + "," +
			common.FIELD_START_TIME + "," +
			common.FIELD_DURATION + "," +
			common.FIELD_END_TIME + "," +
			common.FIELD_UPDATE_TIME + "," +
			common.FIELD_UPDATE_IP + "," +
			common.FIELD_UPDATER +
			" FROM " +
			common.TABLE_MEETING +
			" WHERE " +
			common.FIELD_ID + "=" + sMeetingID + ";"

		rows, err := ms.db.Select(sql)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		if !rows.Next() {
			return nil, common.ERR_NO_MEETING
		}

		scores := ""
		// tl := ""
		// sl := ""
		cwl := ""
		el := ""
		vl := ""
		rl := ""

		mi := new(MeetingInfo)
		err = rows.Scan(&mi.ID, &mi.Name,
			&mi.Section,
			&mi.Type, &mi.Data,
			&scores, &mi.ScoreCount,
			// &tl, &sl,
			&mi.NumberOfAttendee,
			&cwl,
			&vl, &el, &rl,
			&mi.ClassID,
			&mi.GroupID,
			&mi.StartTime, &mi.Duration, &mi.EndTime,
			&mi.UpdateTime, &mi.UpdateIP, &mi.Updater)
		if err != nil {
			return nil, err
		}

		mi.Scores = common.StringToIntArray(scores)
		// mi.Teachers = common.StringToIntArray(tl)
		// mi.Students = common.StringToIntArray(sl)
		mi.Coursewares = common.StringToStringArray(cwl)
		mi.Exams = common.StringToStringArray(el)
		mi.Videos = common.StringToStringArray(vl)
		mi.Replays = common.StringToStringArray(rl)

		// Save it to cache.
		if ms.cache != nil {
			m := make(map[string]string)

			m[common.FIELD_NAME] = mi.Name
			m[common.FIELD_SECTION] = strconv.Itoa(mi.Section)
			m[common.FIELD_TYPE] = strconv.Itoa(mi.Type)
			m[common.FIELD_DATA] = mi.Data
			m[common.FIELD_SCORE] = scores
			m[common.FIELD_SCORE_COUNT] = strconv.Itoa(mi.ScoreCount)
			// m[common.FIELD_TEACHER_LIST] = tl
			// m[common.FIELD_STUDENT_LIST] = sl
			m[common.FIELD_NUMBER_OF_ATTENDEE] = strconv.Itoa(mi.NumberOfAttendee)
			m[common.FIELD_COURSEWARE_LIST] = cwl
			m[common.FIELD_VIDEO_LIST] = vl
			m[common.FIELD_EXAM_LIST] = el
			m[common.FIELD_REPLAY_LIST] = rl
			m[common.FIELD_START_TIME] = strconv.Itoa(mi.StartTime)
			m[common.FIELD_DURATION] = strconv.Itoa(mi.Duration)
			m[common.FIELD_END_TIME] = strconv.Itoa(mi.EndTime)
			m[common.FIELD_CLASS_ID] = strconv.Itoa(mi.ClassID)
			m[common.FIELD_GROUP_ID] = strconv.Itoa(mi.GroupID)
			m[common.FIELD_UPDATE_IP] = mi.UpdateIP
			m[common.FIELD_UPDATE_TIME] = strconv.Itoa(mi.UpdateTime)
			m[common.FIELD_UPDATER] = strconv.Itoa(mi.Updater)

			if err = ms.cache.SetFields(common.KEY_PREFIX_MEETING+sMeetingID, m); err != nil {
				return nil, err
			}
		}

		if !checkAuthority {
			return mi, nil
		}

		// Check authority.
		allowed := false
		switch session.GroupID {
		case common.GROUP_ID_FOR_STUDENT, common.GROUP_ID_FOR_TEACHER, common.GROUP_ID_FOR_KEEPER:
			if session.UserID >= common.VALUE_MINIMAL_TEMPERARY_USER_ID {
				allowed = true
			} else {
				// TODO: The following code fragment should be moved to class service.
				allowed = (func() bool {
					sUserID := strconv.Itoa(session.UserID)
					sClassID := strconv.Itoa(mi.ClassID)

					if ms.cache != nil {
						// Get the class list of this user via cache.
						s, err := ms.cache.GetField(common.KEY_PREFIX_USER+sUserID, common.FIELD_CLASS_LIST)
						if err == nil {
							if common.InList(sClassID, s) {
								return true
							} else {
								return false
							}
						}
					}

					if ms.db != nil {
						// Get the class list of this user via database.
						s := "SELECT " +
							common.FIELD_CLASS_LIST +
							" FROM " +
							common.TABLE_USER_CLASS +
							" WHERE " +
							common.FIELD_USER_ID + "=" + sUserID + ";"

						rows, err := ms.db.Select(s)
						if err != nil {
							return false
						}
						defer rows.Close()

						if !rows.Next() {
							return false
						}

						if err = rows.Scan(&s); err != nil {
							return false
						}

						if ms.cache != nil {
							if err = ms.cache.SetField(common.KEY_PREFIX_USER+sUserID, common.FIELD_CLASS_LIST, s); err != nil {
								// TODO:
							}
						}

						if common.InList(sClassID, s) {
							return true
						}

						return false
					}

					return false
				})()
			}

		case common.GROUP_ID_FOR_SYSTEM:
			allowed = true
		default:
			allowed = (mi.GroupID == session.GroupID)
		}
		if !allowed {
			return nil, common.ERR_NO_AUTHORITY
		}

		return mi, nil
	}

	return nil, common.ERR_NO_SERVICE
}

//----------------------------------------------------------------------------
// Database: Compatible.
// Cache   : Compatible.

func (ms *MeetingService) GetMeetings(meetingIDs []int, session *Session, checkAuthority bool) (*MeetingInfoArray, error) {
	mia := new(MeetingInfoArray)
	mia.Meetinigs = list.New()

	for i := 0; i < len(meetingIDs); i++ {
		mi, err := ms.GetMeeting(meetingIDs[i], session, checkAuthority)
		if err != nil {
			mia.Meetinigs.Init()
			return nil, err
		}
		mia.Meetinigs.PushBack(mi)
	}

	return mia, nil
}

//----------------------------------------------------------------------------
// Database: Required.
// Cache   : Compatible.

func (ms *MeetingService) DeleteMeeting(meetingID int, session *Session) error {
	// Check requirements.
	if ms.db == nil {
		return common.ERR_NO_SERVICE
	}

	_, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return err
	}

	sMeetingID := strconv.Itoa(meetingID)

	if err := (func() error {
		tx, err := ms.db.Transaction()
		if err != nil {
			return err
		}

		s := "DELETE FROM " +
			common.TABLE_MEETING +
			" WHERE " +
			common.FIELD_ID + "=" + sMeetingID + ";"
		if _, err = tx.Exec(s); err != nil {
			tx.Rollback()
			return err
		}

		s = "DELETE FROM " +
			common.TABLE_USER_MEETING +
			" WHERE " +
			common.FIELD_MEETING_ID + "=" + sMeetingID + ";"
		if _, err = tx.Exec(s); err != nil {
			tx.Rollback()
			return err
		}

		// TODO: We do not remove them from cache yet.
		if ms.cache != nil {
			key := common.KEY_PREFIX_MEETING + sMeetingID
			// for i := 0; i < len(mi.Students); i++ {
			// 	if err = ms.cache.Del(key + ":" + strconv.Itoa(mi.Students[i])); err != nil {
			// 		// TODO:
			// 	}
			// }

			if err = ms.cache.Del(key); err != nil {
				// TODO:
			}

			if err = ms.cache.DelField(common.KEY_PREFIX_MEETING+common.FIELD_ID, sMeetingID); err != nil {
				// TODO:
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
// Database: Required.

func (ms *MeetingService) Count() (int, error) {
	if ms.db == nil {
		return 0, common.ERR_NO_DATABASE
	}

	return ms.db.Count(common.TABLE_MEETING)
}

//----------------------------------------------------------------------------
// Database: Required.
// Cache   : Compatible.

func (ms *MeetingService) ArrangeResource(meetingID int, resourceType int, expectedOrder []string, session *Session) error {
	// Check authority.
	mi, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return err
	}

	// Get original value.
	var arr []string = nil
	f := ""
	switch resourceType {
	case common.TYPE_FOR_COURSEWARE:
		arr = mi.Coursewares
		f = common.FIELD_COURSEWARE_LIST

	case common.TYPE_FOR_EXAM:
		arr = mi.Exams
		f = common.FIELD_EXAM_LIST

	case common.TYPE_FOR_VIDEO:
		arr = mi.Videos
		f = common.FIELD_VIDEO_LIST

	default:
		return common.ERR_INVALID_SOURCE
	}

	// Compute new value.
	r := ``
	first := true
	for i := 0; i < len(expectedOrder); i++ {
		v, okay := common.InStringArrayByKey(expectedOrder[i], arr)
		if okay {
			if first {
				first = false
			} else {
				r += `,`
			}

			r += v
		}
	}

	sMeetingID := strconv.Itoa(meetingID)
	sUpdater := strconv.Itoa(session.UserID)
	sUpdateIP := session.IP
	sUpdateTime := common.GetTimeString()

	// Update database.
	if ms.db != nil {
		s := "UPDATE " +
			common.TABLE_MEETING +
			" SET " +
			f + "='" + r + "'," +
			common.FIELD_UPDATER + "=" + sUpdater + "," +
			common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
			common.FIELD_UPDATE_TIME + "=" + sUpdateTime +
			" WHERE " +
			common.FIELD_ID + "=" + sMeetingID + ";"

		if _, err = ms.db.Exec(s); err != nil {
			return err
		}
	}

	// Update cache.
	if ms.cache != nil {
		m := make(map[string]string)
		m[f] = r
		m[common.FIELD_UPDATER] = sUpdater
		m[common.FIELD_UPDATE_IP] = sUpdateIP
		m[common.FIELD_UPDATE_TIME] = sUpdateTime

		if err = ms.cache.SetFields(common.KEY_PREFIX_MEETING+sMeetingID, m); err != nil {
			return err
		}
	}

	return nil
}

//----------------------------------------------------------------------------

func (ms *MeetingService) GetMeetingClassID(meetingID int) (int, error) {
	if ms.cache != nil {
		if s, err := ms.cache.GetField(common.KEY_PREFIX_MEETING+strconv.Itoa(meetingID), common.FIELD_CLASS_ID); err == nil {
			if classID, err := strconv.Atoi(s); err == nil {
				return classID, nil
			}
		}
	}

	if ms.db != nil {
		s := "SELECT " +
			common.FIELD_CLASS_ID +
			" FROM " +
			common.TABLE_MEETING +
			" WHERE " +
			common.FIELD_ID + "=" + strconv.Itoa(meetingID) + ";"

		rows, err := ms.db.Select(s)
		if err != nil {
			return 0, err
		}
		defer rows.Close()

		if !rows.Next() {
			return 0, common.ERR_NO_MEETING
		}

		classID := 0
		if err = rows.Scan(&classID); err != nil {
			return 0, err
		}

		if ms.cache != nil {
			if err = ms.cache.SetField(common.KEY_PREFIX_MEETING+strconv.Itoa(meetingID), common.FIELD_CLASS_ID, strconv.Itoa(classID)); err != nil {
				// TODO:
			}
		}

		return classID, nil
	}

	return 0, common.ERR_NO_SERVICE
}

//----------------------------------------------------------------------------
