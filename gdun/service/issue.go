package service

import (
	"container/list"
	"errors"
	"github.com/wangmhgo/go-project/gdun/common"
	"strconv"
	"strings"
	"time"
)

//----------------------------------------------------------------------------

var issueTableNames = []string{common.TABLE_COURSEWARE_ISSUE, common.TABLE_EXAM_ISSUE, common.TABLE_MEETING_ISSUE, common.TABLE_VIDEO_ISSUE}
var issueTableFieldTypes = []string{"VARCHAR(128)", "INT", "INT", "VARCHAR(64)"}
var issueTypeNum = 4

//----------------------------------------------------------------------------

type IssueService struct {
	db    *common.Database
	cache *common.Cache
	cs    *ClassService
	ms    *MeetingService
}

func NewIssueService(db *common.Database, cache *common.Cache, cs *ClassService, ms *MeetingService) (*IssueService, error) {
	is := new(IssueService)
	is.db = db
	is.cache = cache
	is.cs = cs
	is.ms = ms

	err := is.Init()
	if err != nil {
		return is, err
	}

	return is, nil
}

//----------------------------------------------------------------------------

func (is *IssueService) Init() error {
	if is.db == nil {
		return common.ERR_NO_DATABASE
	}

	// Get group IDs.
	ls, err := (func() (*list.List, error) {
		s := "SELECT " +
			common.FIELD_ID +
			" FROM " +
			common.TABLE_GROUP +
			" WHERE " +
			common.FIELD_ID + ">" + strconv.Itoa(common.GROUP_ID_FOR_KEEPER) + ";"

		rows, err := is.db.Select(s)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		ls := list.New()
		for rows.Next() {
			id := 0
			if err = rows.Scan(&id); err != nil {
				return ls, err
			}
			ls.PushBack(id)
		}

		return ls, nil
	})()
	if err != nil {
		return err
	}

	for e := ls.Front(); e != nil; e = e.Next() {
		groupID, okay := e.Value.(int)
		if !okay {
			continue
		}

		s := "CREATE TABLE IF NOT EXISTS `" + common.TABLE_ISSUE + strconv.Itoa(groupID) + "` (" +
			"`" + common.FIELD_ID + "` INT NOT NULL AUTO_INCREMENT," +
			"`" + common.FIELD_CLASS_ID + "` INT NOT NULL DEFAULT 0," +
			"`" + common.FIELD_MEETING_ID + "` INT NOT NULL DEFAULT 0," +
			"`" + common.FIELD_TYPE + "` INT NOT NULL," +
			"`" + common.FIELD_KEY + "` VARCHAR(128) NOT NULL," +
			"`" + common.FIELD_SUB_KEY + "` INT NOT NULL DEFAULT 0," +
			"`" + common.FIELD_GAODUN_QUESTION_ID + "` INT NOT NULL DEFAULT 0," +
			"`" + common.FIELD_QUESTION_BODY + "` TEXT NOT NULL," +
			"`" + common.FIELD_QUESTION_UPDATE_IP + "` VARCHAR(32) NOT NULL," +
			"`" + common.FIELD_QUESTION_UPDATE_TIME + "` BIGINT NOT NULL," +
			"`" + common.FIELD_QUESTION_UPDATER + "` INT NOT NULL," +
			"`" + common.FIELD_ANSWER_BODY + "` TEXT NOT NULL," +
			"`" + common.FIELD_ANSWER_UPDATE_IP + "` VARCHAR(32) NOT NULL," +
			"`" + common.FIELD_ANSWER_UPDATE_TIME + "` BIGINT NOT NULL," +
			"`" + common.FIELD_ANSWER_UPDATER + "` INT NOT NULL," +
			"PRIMARY KEY (`" + common.FIELD_ID + "`)," +
			"KEY (`" + common.FIELD_CLASS_ID + "`)," +
			"KEY (`" + common.FIELD_MEETING_ID + "`)," +
			"KEY (`" + common.FIELD_TYPE + "`)," +
			"KEY (`" + common.FIELD_KEY + "`)" +
			") ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;"

		if _, err := is.db.Exec(s); err != nil {
			return err
		}
	}

	return nil
}

//----------------------------------------------------------------------------

func (is *IssueService) Preload() (int, error) {
	if is.db == nil {
		return 0, common.ERR_NO_DATABASE
	}
	if is.cache == nil {
		return 0, common.ERR_NO_CACHE
	}

	return 0, nil
}

//----------------------------------------------------------------------------

func (is *IssueService) preloadExamQuestions(t int, start int, length int) (int, error) {
	return 0, nil
}

//----------------------------------------------------------------------------

func (is *IssueService) Ask(meetingID int, t int, key string, subKey int, body string, session *Session) (int, error) {
	// Check authority.
	mi, err := is.ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return 0, err
	}
	ci, err := is.cs.GetClass(mi.ClassID, session)
	if err != nil {
		return 0, err
	}

	now := time.Now().Unix()

	sGroupID := strconv.Itoa(mi.GroupID)
	sClassID := strconv.Itoa(mi.ClassID)
	sMeetingID := strconv.Itoa(meetingID)
	sType := strconv.Itoa(t)
	sKey := common.Escape(key)
	sSubKey := strconv.Itoa(subKey)
	sBody := common.Escape(body)
	sUpdateTime := strconv.FormatInt(now, 10)
	sUpdater := strconv.Itoa(session.UserID)
	sUpdateIP := session.IP

	issueID, err := (func() (int, error) {
		s := "INSERT INTO " +
			common.TABLE_ISSUE + sGroupID +
			" (" +
			common.FIELD_CLASS_ID + "," +
			common.FIELD_MEETING_ID + "," +
			"`" + common.FIELD_TYPE + "`," +
			"`" + common.FIELD_KEY + "`," +
			common.FIELD_SUB_KEY + "," +
			common.FIELD_QUESTION_BODY + "," +
			common.FIELD_QUESTION_UPDATER + "," +
			common.FIELD_QUESTION_UPDATE_TIME + "," +
			common.FIELD_QUESTION_UPDATE_IP + "," +
			common.FIELD_ANSWER_BODY + "," +
			common.FIELD_ANSWER_UPDATER + "," +
			common.FIELD_ANSWER_UPDATE_TIME + "," +
			common.FIELD_ANSWER_UPDATE_IP +
			") VALUES (" +
			sClassID + "," +
			sMeetingID + "," +
			sType + "," +
			"'" + sKey + "'," +
			sSubKey + "," +
			"'" + sBody + "'," +
			sUpdater + "," +
			sUpdateTime + "," +
			"'" + sUpdateIP + "'," +
			"'',0,0,''" +
			");"

		id, err := is.db.Insert(s, 1)
		if err != nil {
			return 0, err
		}

		return int(id), nil
	})()
	if err != nil {
		return 0, err
	}

	err = (func() error {
		sIssueID := strconv.Itoa(issueID)

		// Save issue itself.

		m := make(map[string]string)
		m[common.FIELD_CLASS_ID] = sClassID
		m[common.FIELD_MEETING_ID] = sMeetingID
		m[common.FIELD_TYPE] = sType
		m[common.FIELD_KEY] = key
		m[common.FIELD_SUB_KEY] = sSubKey
		m[common.FIELD_GAODUN_QUESTION_ID] = "0"
		m[common.FIELD_QUESTION_BODY] = body
		m[common.FIELD_QUESTION_UPDATER] = sUpdater
		m[common.FIELD_QUESTION_UPDATE_TIME] = sUpdateTime
		m[common.FIELD_QUESTION_UPDATE_IP] = session.IP
		m[common.FIELD_ANSWER_BODY] = ""
		m[common.FIELD_ANSWER_UPDATER] = "0"
		m[common.FIELD_ANSWER_UPDATE_TIME] = "0"
		m[common.FIELD_ANSWER_UPDATE_IP] = ""

		if err := is.cache.SetFields(is.getIssueKey(mi.GroupID, issueID), m); err != nil {
			return err
		}

		// Set an access to this issue.

		ii := new(IssueInfo)
		ii.ID = issueID
		ii.ClassID = mi.ClassID
		ii.MeetingID = meetingID
		ii.Type = t
		ii.Key = key
		ii.SubKey = subKey
		ii.QuestionBody = body
		ii.QuestionUpdateIP = session.IP
		ii.QuestionUpdateTime = int(now)
		ii.QuestionUpdater = session.UserID
		ii.AnswerBody = ""
		ii.AnswerUpdateIP = ""
		ii.AnswerUpdateTime = 0
		ii.AnswerUpdater = 0

		if err := is.cache.SetField(is.getUnfinishedIssueKey(mi.ClassID), sIssueID, ii.ToJSON()); err != nil {
			// TODO:
		}

		return nil
	})()
	if err != nil {
		return issueID, err
	}

	if (len(ci.Subjects) > 0 || len(mi.Subjects) > 0) && (mi.GroupID > 0) {
		gdStudentID := 0
		if s, err := is.cache.GetField(common.KEY_PREFIX_USER+strconv.Itoa(session.UserID), common.FIELD_GAODUN_STUDENT_ID); err == nil {
			if gdStudentID, err = strconv.Atoi(s); err != nil {
				gdStudentID = session.UserID
			}
		}
		if gdStudentID == 0 {
			return issueID, nil
		}

		//------------------------------------------------
		// Send to Care system.

		//subjectID := 0
		//if len(mi.Subjects) > 0 {
		//	subjectID = mi.Subjects[0]
		//} else {
		//	subjectID = ci.Subjects[0]
		//}

		//	gdQuestionID, err := common.AddCareQuestion(ci.GdCourseID, gdStudentID, body, mi.GroupID, subjectID, issueID)
		//	if err != nil {
		//		return issueID, err
		//	}
		//
		//	sGdQuestionID := strconv.Itoa(gdQuestionID)
		//
		//	s := "UPDATE " +
		//		common.TABLE_ISSUE + sGroupID +
		//		" SET " +
		//		common.FIELD_GAODUN_QUESTION_ID + "=" + sGdQuestionID +
		//		" WHERE " +
		//		common.FIELD_ID + "=" + strconv.Itoa(issueID) + ";"
		//
		//	if _, err = is.db.Exec(s); err != nil {
		//		return issueID, err
		//	}
		//
		//	// Record question ID in Care system.
		//	if err = is.cache.SetField(is.getIssueKey(ci.GroupID, issueID), common.FIELD_GAODUN_QUESTION_ID, strconv.Itoa(gdQuestionID)); err != nil {
		//		return issueID, err
		//	}
		//
		//	//------------------------------------------------
		//	// Cache issue resources.
		//
		//	s, err = (func() (string, error) {
		//		switch t {
		//		case common.TYPE_FOR_COURSEWARE:
		//			return is.computeCoursewareResource(key, subKey)
		//
		//		case common.TYPE_FOR_EXAM:
		//			n, err := strconv.Atoi(key)
		//			if err != nil {
		//				return "", nil
		//			}
		//			return is.computeExamResource(n, subKey)
		//
		//		case common.TYPE_FOR_MEETING:
		//			n, err := strconv.Atoi(key)
		//			if err != nil {
		//				return "", nil
		//			}
		//			return is.computeMeetingResource(n, subKey)
		//
		//		case common.TYPE_FOR_VIDEO:
		//			return is.computeVideoResource(key, subKey)
		//
		//		default:
		//			return "", nil
		//		}
		//	})()
		//	if (err == nil) && (len(s) > 0) {
		//		if err = is.cache.SetKey(is.GetIssueResourceKey(mi.GroupID, issueID), s); err != nil {
		//			return issueID, err
		//		}
		//	}
		//}
		//
		//return issueID, nil
	}
	return 12, nil
}

func (is *IssueService) computeCoursewareResource(coursewareID string, page int) (string, error) {
	return `"slides_id":"//glive.gitlab.hfjy.com/gdun/cw/` + coursewareID + `/` + strconv.Itoa(page) + `.png"`, nil
}

func (is *IssueService) computeExamResource(examID int, questionID int) (string, error) {
	id, err := is.ms.es.GetGdQuestionID(examID, questionID)
	if err != nil {
		return "", err
	}

	return `"item_id":` + strconv.Itoa(id), nil
}

func (is *IssueService) computeMeetingResource(meetingID int, position int) (string, error) {
	s, err := is.cache.GetField(common.KEY_PREFIX_MEETING+strconv.Itoa(meetingID), common.FIELD_REPLAY_LIST)
	if err != nil {
		return "", err
	}

	arr := strings.Split(s, ",")
	// TODO:
	videoID := arr[0]

	sDuration, err := is.cache.GetField(common.KEY_PREFIX_VIDEO+videoID, common.FIELD_DURATION)
	if err != nil {
		sDuration = "0"
	}

	return `"video_id":"` + videoID + `","video_title":"","video_duration":` + sDuration + `,"video_position":` + strconv.Itoa(position), nil
}

func (is *IssueService) computeVideoResource(videoID string, position int) (string, error) {
	sDuration, err := is.cache.GetField(common.KEY_PREFIX_VIDEO+videoID, common.FIELD_DURATION)
	if err != nil {
		sDuration = "0"
	}

	return `"video_id":"` + videoID + `","video_title":"","video_duration":` + sDuration + `,"video_position":` + strconv.Itoa(position), nil
}

//----------------------------------------------------------------------------

func (is *IssueService) GetIssueResource(groupID int, issueID int) (string, error) {
	s, err := is.cache.GetKey(is.GetIssueResourceKey(groupID, issueID))
	if err == nil {
		return s, nil
	}

	// Reload this resource.
	k := is.getIssueKey(groupID, issueID)

	s, err = is.cache.GetField(k, common.FIELD_TYPE)
	if err != nil {
		return "", err
	}
	t, err := strconv.Atoi(s)
	if err != nil {
		return "", err
	}

	key, err := is.cache.GetField(k, common.FIELD_KEY)
	if err != nil {
		return "", err
	}

	s, err = is.cache.GetField(k, common.FIELD_SUB_KEY)
	if err != nil {
		return "", err
	}
	subKey, err := strconv.Atoi(s)
	if err != nil {
		return "", err
	}

	return (func() (string, error) {
		switch t {
		case common.TYPE_FOR_COURSEWARE:
			return is.computeCoursewareResource(key, subKey)

		case common.TYPE_FOR_EXAM:
			n, err := strconv.Atoi(key)
			if err != nil {
				return "", nil
			}
			return is.computeExamResource(n, subKey)

		case common.TYPE_FOR_MEETING:
			n, err := strconv.Atoi(key)
			if err != nil {
				return "", nil
			}
			return is.computeMeetingResource(n, subKey)

		case common.TYPE_FOR_VIDEO:
			return is.computeVideoResource(key, subKey)

		default:
			return "", nil
		}
	})()
}

func (is *IssueService) GetIssueQuestion(groupID int, issueID int) (string, error) {
	key := is.getIssueKey(groupID, issueID)

	if is.cache != nil {
		s, err := (func() (string, error) {
			s, err := is.cache.GetField(key, common.FIELD_QUESTION_BODY)
			if err != nil {
				return "", err
			}

			arr := strings.Split(s, "\\n")
			s = ""
			for i := 0; i < len(arr); i++ {
				if len(arr[i]) == 0 {
					continue
				}

				if strings.HasPrefix(arr[i], "data:image/") {
					s += `<img src="` + arr[i] + `">`
				} else {
					s += arr[i]
				}
				s += `<br>`
			}

			return s, err
		})()
		if err == nil {
			return s, nil
		}
	}

	return "", common.ERR_NO_SERVICE
}

//----------------------------------------------------------------------------

func (is *IssueService) finishIssue(groupID int, classID int, issueID int, body string, userID int, updateTime int, updateIP string) error {

	ufk := is.getUnfinishedIssueKey(classID)
	fk := is.getFinishedIssueKey(classID)
	sIssueID := strconv.Itoa(issueID)

	// Check whether such an unfinished issue exists or not.
	okay, err := is.cache.FieldExist(ufk, sIssueID)
	if err != nil {
		return err
	}
	if !okay {
		return common.ERR_NO_RECORD
	}

	//----------------------------------------------------
	// Update cache.

	// Get the issue itself.
	k := is.getIssueKey(groupID, issueID)
	m, err := is.cache.GetAllFields(k)
	if err != nil {
		return err
	}

	ii, n := NewIssueInfoFromMap(m, groupID, issueID)
	if n != 0 {
		return errors.New(strconv.Itoa(n))
	}
	ii.AnswerBody = body
	ii.AnswerUpdateIP = updateIP
	ii.AnswerUpdateTime = updateTime
	ii.AnswerUpdater = userID

	// Fill blank fields.
	if err = is.cache.SetField(k, common.FIELD_ANSWER_BODY, body); err != nil {
		return err
	}
	if err = is.cache.SetField(k, common.FIELD_UPDATE_IP, updateIP); err != nil {
		return err
	}
	if err = is.cache.SetField(k, common.FIELD_UPDATE_TIME, strconv.Itoa(updateTime)); err != nil {
		return err
	}
	if err = is.cache.SetField(k, common.FIELD_UPDATER, strconv.Itoa(userID)); err != nil {
		return err
	}

	// Append to finished issues.
	if err = is.cache.Append(fk, ","+ii.ToJSON()); err != nil {
		return err
	}

	// Delete the unfinished issue.
	if err = is.cache.DelField(ufk, sIssueID); err != nil {
		return err
	}

	//----------------------------------------------------
	// Update database.

	err = (func() error {
		s := "UPDATE " +
			common.TABLE_ISSUE + strconv.Itoa(groupID) +
			" SET " +
			common.FIELD_ANSWER_BODY + "='" + common.Escape(body) + "'," +
			common.FIELD_ANSWER_UPDATER + "=" + strconv.Itoa(userID) + "," +
			common.FIELD_ANSWER_UPDATE_TIME + "=" + strconv.Itoa(updateTime) + "," +
			common.FIELD_ANSWER_UPDATE_IP + "='" + updateIP + "'" +
			" WHERE " +
			common.FIELD_ID + "=" + strconv.Itoa(issueID) + ";"

		_, err := is.db.Exec(s)
		if err != nil {
			return err
		}

		return nil
	})()
	if err != nil {
		return err
	}

	return nil
}

func (is *IssueService) Answer(classID int, issueID int, body string, session *Session) error {
	if is.cache == nil || is.db == nil {
		return common.ERR_NO_SERVICE
	}

	// Check authority.
	ci, err := is.cs.GetClass(classID, session)
	if err != nil {
		return err
	}

	return is.finishIssue(ci.GroupID, classID, issueID, body, session.UserID, int(time.Now().Unix()), session.IP)
}

//----------------------------------------------------------------------------

func (is *IssueService) Change(classID int, issueID int, body string, session *Session) error {
	return nil
}

//----------------------------------------------------------------------------

func (is *IssueService) Get(classID int, session *Session) (string, error) {
	//if is.cache == nil {
	//	return "", common.ERR_NO_SERVICE
	//}
	//
	//// Check authority.
	//ci, err := is.cs.GetClass(classID, session)
	//if err != nil {
	//	return "", err
	//}
	//
	//result := `"` + common.FIELD_UNFINISHED + `":[`
	//
	//m, err := is.cache.GetAllFields(is.getUnfinishedIssueKey(classID))
	//if err == nil {
	//	first := true
	//	for s, item := range m {
	//		issueID, err := strconv.Atoi(s)
	//		if err != nil {
	//			continue
	//		}

			// Check whether this issue has been answered in Care system.
			//answered := (func() bool {
			//	s, err := is.cache.GetField(is.getIssueKey(ci.GroupID, issueID), common.FIELD_GAODUN_QUESTION_ID)
			//	if err != nil {
			//		return false
			//	}

			//gdQuestionID, err := strconv.Atoi(s)
			//if err != nil {
			//	return false
			//}

			//answer, updateTime := common.GetCareQuestion(gdQuestionID)
			//if len(answer) == 0 {
			//	return false
			//}

			//			if err = is.finishIssue(ci.GroupID, classID, issueID, "", 0, 1000, ""); err != nil {
			//				return false
			//			}
			//
			//			return true
			//		})()
			//		if answered {
			//			continue
			//		}
			//
			//		if first {
			//			first = false
			//		} else {
			//			result += `,`
			//		}
			//		result += item
			//	}
			//}
			//
			//result += `],"` + common.FIELD_FINISHED + `":[`
			//
			//s, err := is.cache.GetKey(is.getFinishedIssueKey(classID))
			//if err == nil {
			//	result += s[1:]
			//}
			//
			//result += `]`

	return "", nil
}

//----------------------------------------------------------------------------

func (is *IssueService) getIssueKey(groupID int, issueID int) string {
	return common.KEY_PREFIX_GROUP + strconv.Itoa(groupID) + ":" + common.KEY_PREFIX_ISSUE + strconv.Itoa(issueID)
}

//----------------------------------------------------------------------------

func (is *IssueService) GetIssueResourceKey(groupID int, issueID int) string {
	return common.KEY_PREFIX_GROUP + strconv.Itoa(groupID) + ":" + common.KEY_PREFIX_ISSUE + strconv.Itoa(issueID) + ":" + common.FIELD_TEXT
}

//----------------------------------------------------------------------------

func (is *IssueService) getFinishedIssueKey(classID int) string {
	return common.KEY_PREFIX_CLASS + strconv.Itoa(classID) + ":" + common.FIELD_ISSUE + ":" + common.FIELD_TEXT
}

//----------------------------------------------------------------------------

func (is *IssueService) getUnfinishedIssueKey(classID int) string {
	return common.KEY_PREFIX_CLASS + strconv.Itoa(classID) + ":" + common.FIELD_ISSUE
}

//----------------------------------------------------------------------------
