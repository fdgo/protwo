package service

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/wangmhgo/go-project/gdun/common"
	"math/rand"
	"strconv"
	"strings"
)

//----------------------------------------------------------------------------

type ExamService struct {
	db        *common.Database
	cache     *common.Cache
	oss       *common.ObjectStorage
	ossPrefix string
	gda       *GdAdapter
}

func NewExamService(db *common.Database, cache *common.Cache, oss *common.ObjectStorage, prefix string, gda *GdAdapter) (*ExamService, error) {
	es := new(ExamService)
	es.db = db
	es.cache = cache
	es.oss = oss
	es.ossPrefix = prefix
	es.gda = gda

	err := es.Init()
	if err != nil {
		return nil, err
	}

	return es, nil
}

//----------------------------------------------------------------------------

func (es *ExamService) Init() error {
	if es.db == nil {
		return common.ERR_NO_DATABASE
	}

	sql := "CREATE TABLE IF NOT EXISTS `" + common.TABLE_EXAM + "` ("
	sql += " `" + common.FIELD_ID + "` INT NOT NULL AUTO_INCREMENT,"
	sql += " `" + common.FIELD_NAME + "` VARCHAR(512) NOT NULL,"
	sql += " `" + common.FIELD_KEY + "` VARCHAR(512) NOT NULL DEFAULT '',"
	sql += " `" + common.FIELD_IV + "` VARCHAR(512) NOT NULL DEFAULT '',"
	sql += " `" + common.FIELD_COUNT + "` INT NOT NULL DEFAULT 0,"
	sql += " `" + common.FIELD_GAODUN_EXAM_ID + "` INT NOT NULL DEFAULT 0,"
	sql += " `" + common.FIELD_ANSWER + "` TEXT NOT NULL," // New field.
	sql += " `" + common.FIELD_GROUP_ID + "` INT NOT NULL,"
	sql += " `" + common.FIELD_UPDATE_IP + "` VARCHAR(32) NOT NULL,"
	sql += " `" + common.FIELD_UPDATE_TIME + "` BIGINT NOT NULL,"
	sql += " `" + common.FIELD_UPDATER + "` INT NOT NULL,"
	sql += " PRIMARY KEY (`" + common.FIELD_ID + "`),"
	sql += " KEY (`" + common.FIELD_GAODUN_EXAM_ID + "`),"
	sql += " KEY (`" + common.FIELD_GROUP_ID + "`)"
	sql += ") ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;"

	if _, err := es.db.Exec(sql); err != nil {
		return err
	}

	sql = "CREATE TABLE IF NOT EXISTS `" + common.TABLE_USER_EXAM + "` ("
	sql += " `" + common.FIELD_USER_ID + "` INT NOT NULL,"
	sql += " `" + common.FIELD_MEETING_ID + "` INT NOT NULL,"
	sql += " `" + common.FIELD_EXAM_ID + "` INT NOT NULL,"
	sql += " `" + common.FIELD_OBJECTIVE + "` TEXT NOT NULL,"
	sql += " `" + common.FIELD_SUBJECTIVE + "` TEXT NOT NULL,"
	sql += " `" + common.FIELD_UPDATE_IP + "` VARCHAR(32) NOT NULL,"
	sql += " `" + common.FIELD_UPDATE_TIME + "` BIGINT NOT NULL,"
	sql += " PRIMARY KEY (`" + common.FIELD_USER_ID + "`,`" + common.FIELD_EXAM_ID + "`,`" + common.FIELD_MEETING_ID + "`)"
	sql += ") ENGINE=InnoDB DEFAULT CHARSET=utf8;"

	if _, err := es.db.Exec(sql); err != nil {
		return err
	}

	return nil
}

//----------------------------------------------------------------------------

func (es *ExamService) Preload() (int, int, error) {
	if es.db == nil {
		return 0, 0, common.ERR_NO_DATABASE
	}
	if es.cache == nil {
		return 0, 0, common.ERR_NO_CACHE
	}

	n1, err := (func() (int, error) {
		rest, err := es.db.Count(common.TABLE_EXAM)
		if err != nil {
			return 0, err
		}

		i := 0
		for i < rest {
			n, err := es.preloadExams(i, common.DATABASE_PRELOAD_SIZE)
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

	n2, err := (func() (int, error) {
		rest, err := es.db.Count(common.TABLE_USER_EXAM)
		if err != nil {
			return 0, err
		}

		i := 0
		for i < rest {
			n, err := es.preloadUserExams(i, common.DATABASE_PRELOAD_SIZE)
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

func (es *ExamService) preloadExams(start int, length int) (int, error) {
	sql := "SELECT " +
		common.FIELD_ID + "," +
		common.FIELD_NAME + "," +
		"`" + common.FIELD_KEY + "`," + common.FIELD_IV + "," +
		"`" + common.FIELD_COUNT + "`," +
		common.FIELD_GAODUN_EXAM_ID + "," +
		common.FIELD_ANSWER + "," +
		common.FIELD_GROUP_ID + "," +
		common.FIELD_UPDATE_IP + "," + common.FIELD_UPDATE_TIME + "," + common.FIELD_UPDATER +
		" FROM " +
		common.TABLE_EXAM +
		" LIMIT " + strconv.Itoa(start) + "," + strconv.Itoa(length) + ";"

	rows, err := es.db.Select(sql)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	m := make(map[string]string)

	cnt := 0

	id := 0
	name := ""
	key := ""
	iv := ""
	count := 0
	gdExamID := 0
	answer := ""
	groupID := 0
	updateIP := ""
	updateTime := 0
	updater := 0

	for rows.Next() {
		err = rows.Scan(&id, &name, &key, &iv, &count, &gdExamID, &answer, &groupID, &updateIP, &updateTime, &updater)
		if err != nil {
			return cnt, err
		}

		m[common.FIELD_NAME] = name
		m[common.FIELD_KEY] = key
		m[common.FIELD_IV] = iv
		m[common.FIELD_COUNT] = strconv.Itoa(count)
		m[common.FIELD_GAODUN_EXAM_ID] = strconv.Itoa(gdExamID)
		m[common.FIELD_ANSWER] = answer
		m[common.FIELD_GROUP_ID] = strconv.Itoa(groupID)
		m[common.FIELD_UPDATE_IP] = updateIP
		m[common.FIELD_UPDATE_TIME] = strconv.Itoa(updateTime)
		m[common.FIELD_UPDATER] = strconv.Itoa(updater)

		err = es.cache.SetFields(common.KEY_PREFIX_EXAM+strconv.Itoa(id), m)
		if err != nil {
			return cnt, err
		}

		cnt++
	}

	return cnt, nil
}

func (es *ExamService) preloadUserExams(start int, length int) (int, error) {
	sql := "SELECT " +
		common.FIELD_USER_ID + "," +
		common.FIELD_MEETING_ID + "," +
		common.FIELD_EXAM_ID + "," +
		common.FIELD_OBJECTIVE + "," +
		common.FIELD_SUBJECTIVE +
		" FROM " +
		common.TABLE_USER_EXAM +
		" LIMIT " + strconv.Itoa(start) + "," + strconv.Itoa(length) + ";"

	rows, err := es.db.Select(sql)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	cnt := 0

	userID := 0
	meetingID := 0
	examID := 0
	objective := ""
	subjective := ""

	for rows.Next() {
		err = rows.Scan(&userID, &meetingID, &examID, &objective, &subjective)
		if err != nil {
			return cnt, err
		}

		err = es.cache.SetField(common.KEY_PREFIX_EXAM+strconv.Itoa(examID)+":"+common.FIELD_OBJECTIVE, strconv.Itoa(userID)+":"+strconv.Itoa(meetingID), objective)
		if err != nil {
			return cnt, err
		}
		err = es.cache.SetField(common.KEY_PREFIX_EXAM+strconv.Itoa(examID)+":"+common.FIELD_SUBJECTIVE, strconv.Itoa(userID)+":"+strconv.Itoa(meetingID), subjective)
		if err != nil {
			return cnt, err
		}

		cnt++
	}

	return cnt, nil
}

//----------------------------------------------------------------------------

func (es *ExamService) GetExam(id int, session *Session) (*ExamInfo, error) {
	if es.cache != nil {
		m, err := es.cache.GetAllFields(common.KEY_PREFIX_EXAM + strconv.Itoa(id))
		if err == nil {
			if ei := NewExamInfoFromMap(m, id); err != nil {
				return ei, nil
			}
		}
	}

	if es.db != nil {
		sID := strconv.Itoa(id)

		sql := "SELECT " +
			common.FIELD_NAME + "," +
			"`" + common.FIELD_KEY + "`," + common.FIELD_IV + "," +
			"`" + common.FIELD_COUNT + "`," +
			common.FIELD_GAODUN_EXAM_ID + "," +
			common.FIELD_ANSWER + "," +
			common.FIELD_GROUP_ID + "," +
			common.FIELD_UPDATE_IP + "," + common.FIELD_UPDATE_TIME + "," + common.FIELD_UPDATER +
			" FROM " +
			common.TABLE_EXAM +
			" WHERE " +
			common.FIELD_ID + "=" + sID + ";"

		rows, err := es.db.Select(sql)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		if !rows.Next() {
			return nil, common.ERR_NO_EXAM
		}

		//------------------------------------------------

		ei := new(ExamInfo)
		err = rows.Scan(
			&ei.Name,
			&ei.Key, &ei.IV,
			&ei.Count, &ei.GdExamID, &ei.Answer,
			&ei.GroupID,
			&ei.UpdateIP, &ei.UpdateTime, &ei.Updater)
		if err != nil {
			return nil, err
		}
		ei.ID = id

		//------------------------------------------------

		if es.cache != nil {
			m := make(map[string]string)

			m[common.FIELD_ID] = sID
			m[common.FIELD_NAME] = ei.Name
			m[common.FIELD_KEY] = ei.Key
			m[common.FIELD_IV] = ei.IV
			m[common.FIELD_COUNT] = strconv.Itoa(ei.Count)
			m[common.FIELD_GAODUN_EXAM_ID] = strconv.Itoa(ei.GdExamID)
			m[common.FIELD_ANSWER] = ei.Answer
			m[common.FIELD_GROUP_ID] = strconv.Itoa(ei.GroupID)
			m[common.FIELD_UPDATE_IP] = ei.UpdateIP
			m[common.FIELD_UPDATE_TIME] = strconv.Itoa(ei.UpdateTime)
			m[common.FIELD_UPDATER] = strconv.Itoa(ei.Updater)

			err = es.cache.SetFields(common.KEY_PREFIX_EXAM+sID, m)
			if err != nil {
				return nil, err
			}
		}

		return ei, nil
	}

	return nil, common.ERR_NO_SERVICE
}

func (es *ExamService) GetGdQuestionID(examID int, questionID int) (int, error) {
	sExamID := strconv.Itoa(examID)
	sQuestionID := strconv.Itoa(questionID)
	key := common.KEY_PREFIX_EXAM + sExamID + ":" + common.FIELD_GAODUN_QUESTION_ID

	// Get Gd question ID via cache.
	if s, err := es.cache.GetField(key, sQuestionID); err == nil {
		if n, err := strconv.Atoi(s); err == nil {
			return n, nil
		}
	}

	// Get Gd exam ID.
	gdExamID, err := (func() (int, error) {
		if s, err := es.cache.GetField(common.KEY_PREFIX_EXAM+sExamID, common.FIELD_GAODUN_EXAM_ID); err == nil {
			if n, err := strconv.Atoi(s); err == nil {
				return n, nil
			}
		}

		s := "SELECT " +
			common.FIELD_GAODUN_EXAM_ID +
			" FROM " +
			common.TABLE_EXAM +
			" WHERE " +
			common.FIELD_ID + "=" + sExamID + ";"

		rows, err := es.db.Select(s)
		if err != nil {
			return 0, err
		}
		defer rows.Close()

		if !rows.Next() {
			return 0, common.ERR_NO_EXAM
		}

		n := 0
		if err = rows.Scan(&n); err != nil {
			return 0, err
		}

		return n, nil
	})()
	if err != nil {
		return 0, err
	}

	// Get the exam via Gd Tiku.
	gte, err := es.gda.GetExam(gdExamID)
	if err != nil {
		return 0, err
	}

	result := 0
	id := 0
	for e := gte.Subjects.Front(); e != nil; e = e.Next() {
		sub, okay := e.Value.(*GdTikuSubExam)
		if !okay {
			continue
		}

		for p := sub.Items.Front(); p != nil; p = p.Next() {
			if item, okay := p.Value.(*GdTikuItem); okay {
				id++
				if id == questionID {
					result = item.GdQuestionID
				}

				if err = es.cache.SetField(key, strconv.Itoa(id), strconv.Itoa(item.GdQuestionID)); err != nil {
					// TODO:
				}
			} else if composed, okay := p.Value.(*GdTikuComposedItem); okay {
				for q := composed.Items.Front(); q != nil; q = q.Next() {
					item, okay := q.Value.(*GdTikuItem)
					if !okay {
						continue
					}

					id++
					if id == questionID {
						result = item.GdQuestionID
					}

					if err = es.cache.SetField(key, strconv.Itoa(id), strconv.Itoa(item.GdQuestionID)); err != nil {
						// TODO:
					}
				}
			}
		}
	}

	return result, nil
}

//----------------------------------------------------------------------------

func (es *ExamService) Import(name string, groupID int, id int, gdExamID int, session *Session) (int, int, error) {
	if es.gda == nil {
		return 0, 0, common.ERR_NO_SERVICE
	}

	// Get the exam.
	gdExam, err := es.gda.GetExam(gdExamID)
	if err != nil {
		return 0, 0, err
	}
	// fmt.Println(gdExam.ToJSON())

	return es.ImportGdExam(name, groupID, id, gdExam, session)
}

func (es *ExamService) ImportGdExam(name string, groupID int, id int, gdExam *GdTikuExam, session *Session) (int, int, error) {

	gdExam.Solve()

	// Get the key and IV.
	key := ""
	iv := ""
	nID := id
	var err error
	if id > 0 {
		// Get key and IV from the existing exam.
		ei, err := es.GetExam(id, session)
		if err != nil {
			return 0, 0, err
		}

		if ei.Count != gdExam.Count || ei.Answer != gdExam.Answer {
			es.updateExam(id, gdExam.Count, gdExam.Answer, session)
		}

		key = ei.Key
		iv = ei.IV
	} else {
		// Generate a pair of key and IV.
		nID, key, iv, err = es.newExam(name, gdExam.Count, gdExam.ID, gdExam.Answer, groupID, session)
		if err != nil {
			return 0, 0, err
		}
	}

	keyBuf, err := hex.DecodeString(key)
	if err != nil {
		return 0, 0, err
	}
	ivBuf, err := hex.DecodeString(iv)
	if err != nil {
		return 0, 0, err
	}

	// Encrypt the exam.
	buf, err := common.Encrypt([]byte("{"+gdExam.ToJSON()+"}"), keyBuf, ivBuf)
	if err != nil {
		return 0, 0, err
	}
	// fmt.Printf("encrypted: %d\n", len(buf))

	s := base64.StdEncoding.EncodeToString(buf)
	// fmt.Printf("base64: %d\n", len(s))

	// Save the encrypted exam.
	err = es.oss.UploadString(es.ossPrefix+strconv.Itoa(nID), s)
	if err != nil {
		// fmt.Println("UploadString: " + err.Error())
		return 0, 0, err
	}

	return nID, gdExam.Count, nil
}

func (es *ExamService) newExam(name string, count int, gdExamID int, answer string, groupID int, session *Session) (int, string, string, error) {
	// Check requirements.
	if es.db == nil {
		return 0, "", "", common.ERR_NO_SERVICE
	}

	// Check authority.
	gID := 0
	if session.IsAssistant() {
		gID = session.GroupID
	} else if session.IsSystem() {
		gID = groupID
	} else {
		return 0, "", "", common.ERR_NO_AUTHORITY
	}

	// Generate key and IV.
	key := fmt.Sprintf("%016x%016x", rand.Int63(), rand.Int63())
	iv := fmt.Sprintf("%016x%016x", rand.Int63(), rand.Int63())

	sName := common.Escape(name)
	sCount := strconv.Itoa(count)
	sGdExamID := strconv.Itoa(gdExamID)
	sAnswer := common.Escape(answer)
	sGroupID := strconv.Itoa(gID)
	sUpdateTime := common.GetTimeString()
	sUpdater := strconv.Itoa(session.UserID)

	// Save it to database.
	sql := "INSERT INTO " + common.TABLE_EXAM +
		" (" +
		common.FIELD_NAME + "," +
		"`" + common.FIELD_COUNT + "`," +
		"`" + common.FIELD_KEY + "`," + common.FIELD_IV + "," +
		common.FIELD_GAODUN_EXAM_ID + "," +
		common.FIELD_ANSWER + "," +
		common.FIELD_GROUP_ID + "," +
		common.FIELD_UPDATE_TIME + "," + common.FIELD_UPDATE_IP + "," + common.FIELD_UPDATER +
		") VALUES (" +
		"'" + sName + "'," +
		sCount + "," +
		"'" + key + "','" + iv + "'," +
		sGdExamID + "," +
		"'" + sAnswer + "'," +
		sGroupID + "," +
		sUpdateTime + ",'" + session.IP + "'," + sUpdater + ");"

	id, err := es.db.Insert(sql, 1)
	if err != nil {
		return 0, "", "", err
	}
	nID := int(id)

	// Save it to cache.
	if es.cache != nil {
		m := make(map[string]string)
		m[common.FIELD_NAME] = sName
		m[common.FIELD_COUNT] = sCount
		m[common.FIELD_KEY] = key
		m[common.FIELD_IV] = iv
		m[common.FIELD_GAODUN_EXAM_ID] = sGdExamID
		m[common.FIELD_ANSWER] = sAnswer
		m[common.FIELD_GROUP_ID] = sGroupID
		m[common.FIELD_UPDATE_TIME] = sUpdateTime
		m[common.FIELD_UPDATE_IP] = session.IP
		m[common.FIELD_UPDATER] = sUpdater

		err = es.cache.SetFields(common.KEY_PREFIX_EXAM+strconv.Itoa(nID), m)
		if err != nil {
			return 0, "", "", err
		}
	}

	return nID, key, iv, nil
}

func (es *ExamService) updateExam(id int, count int, answer string, session *Session) error {

	sExamID := strconv.Itoa(id)
	sCount := strconv.Itoa(count)
	sAnswer := common.Escape(answer)

	sUpdateIP := session.IP
	sUpdateTime := common.GetTimeString()
	sUpdater := strconv.Itoa(session.UserID)

	if es.db != nil {
		s := "UPDATE " +
			common.TABLE_EXAM +
			" SET " +
			"`" + common.FIELD_COUNT + "`=" + sCount + "," +
			common.FIELD_ANSWER + "='" + sAnswer + "'," +
			common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
			common.FIELD_UPDATE_TIME + "=" + sUpdateTime + "," +
			common.FIELD_UPDATER + "=" + sUpdater +
			" WHERE " +
			common.FIELD_ID + "=" + sExamID + ";"

		if _, err := es.db.Exec(s); err != nil {
			return err
		}
	}

	if es.cache != nil {
		m := make(map[string]string)

		m[common.FIELD_COUNT] = sCount
		m[common.FIELD_ANSWER] = sAnswer
		m[common.FIELD_UPDATE_IP] = sUpdateIP
		m[common.FIELD_UPDATE_TIME] = sUpdateTime
		m[common.FIELD_UPDATER] = sUpdater

		if err := es.cache.SetFields(common.KEY_PREFIX_EXAM+sExamID, m); err != nil {
			return err
		}
	}

	return common.ERR_NO_SERVICE
}

//----------------------------------------------------------------------------

func (es *ExamService) Resync(id int, session *Session) (int, error) {
	ei, err := es.GetExam(id, session)
	if err != nil {
		return 0, err
	}

	if ei.GdExamID == 0 {
		return 0, common.ERR_NO_EXAM
	}

	_, cnt, err := es.Import(common.Unescape(ei.Name), ei.GroupID, ei.ID, ei.GdExamID, session)
	if err != nil {
		return 0, err
	}

	return cnt, nil
}

//----------------------------------------------------------------------------

func (es *ExamService) Answer(meetingID int, examID int, objectiveAnswers string, subjectiveAnswers string, session *Session) error {
	// Check requirements.
	if es.db == nil {
		return common.ERR_NO_SERVICE
	}

	sUserID := strconv.Itoa(session.UserID)
	sMeetingID := strconv.Itoa(meetingID)
	sExamID := strconv.Itoa(examID)

	sObjectiveAnswers := common.Escape(objectiveAnswers)
	sSubjectiveAnswers := subjectiveAnswers

	sUpdateIP := session.IP
	sUpdateTime := common.GetTimeString()

	s := "INSERT INTO " +
		common.TABLE_USER_EXAM +
		"(" +
		common.FIELD_USER_ID + "," +
		common.FIELD_MEETING_ID + "," +
		common.FIELD_EXAM_ID + "," +
		common.FIELD_OBJECTIVE + "," +
		common.FIELD_SUBJECTIVE + "," +
		common.FIELD_UPDATE_IP + "," +
		common.FIELD_UPDATE_TIME +
		") VALUES (" +
		sUserID + "," +
		sMeetingID + "," +
		sExamID + "," +
		"'" + sObjectiveAnswers + "'," +
		"'" + sSubjectiveAnswers + "'," +
		"'" + sUpdateIP + "'," +
		sUpdateTime +
		");"
	if _, err := es.db.Exec(s); err != nil {
		return err
	}

	if es.cache != nil {
		err := es.cache.SetField(common.KEY_PREFIX_EXAM+sExamID+":"+common.FIELD_OBJECTIVE, sUserID+":"+sMeetingID, sObjectiveAnswers)
		if err != nil {
			return err
		}

		err = es.cache.SetField(common.KEY_PREFIX_EXAM+sExamID+":"+common.FIELD_SUBJECTIVE, sUserID+":"+sMeetingID, sSubjectiveAnswers)
		if err != nil {
			return err
		}
	}

	return nil
}

//----------------------------------------------------------------------------

func (es *ExamService) GetMyResult(meetingID int, examID int, session *Session) (string, error) {
	sMeetingID := strconv.Itoa(meetingID)
	sExamID := strconv.Itoa(examID)
	sUserID := strconv.Itoa(session.UserID)

	if es.cache != nil {
		key := common.KEY_PREFIX_EXAM + sExamID + ":"

		if result, err := (func() (string, error) {
			obj, err := es.cache.GetField(key+common.FIELD_OBJECTIVE, sUserID+":"+sMeetingID)
			if err != nil {
				// return "", err
				obj = ""
			}

			sub, err := es.cache.GetField(key+common.FIELD_SUBJECTIVE, sUserID+":"+sMeetingID)
			if err != nil {
				// return "", err
				sub = "{}"
			}

			r := `"` + common.FIELD_OBJECTIVE + `":"` + common.Unescape(obj) + `",` +
				`"` + common.FIELD_SUBJECTIVE + `":` + strings.Replace(sub, "+", "%20", -1) // common.UnescapeForJSON(sub)

			return r, nil
		})(); err == nil {
			return result, nil
		}
	}

	return "", common.ERR_NO_SERVICE
}

//----------------------------------------------------------------------------

func (es *ExamService) GetResults(meetingID int, examID int, session *Session) (string, error) {
	sMeetingID := strconv.Itoa(meetingID)
	sExamID := strconv.Itoa(examID)

	if es.cache != nil {
		key := common.KEY_PREFIX_EXAM + sExamID + ":"

		if result, err := (func() (string, error) {
			// Get all objective answers.
			objectiveMap, err := es.cache.GetAllFields(key + common.FIELD_OBJECTIVE)
			if err != nil {
				return "", err
			}

			// Get all subjective answers.
			subjectiveMap, err := es.cache.GetAllFields(key + common.FIELD_SUBJECTIVE)
			if err != nil {
				return "", err
			}

			r := ``
			first := true
			for s, objectiveAnswer := range objectiveMap {
				// Check meeting ID.
				param := strings.Split(s, ":")
				if len(param) != 2 || param[1] != sMeetingID {
					continue
				}

				if first {
					first = false
				} else {
					r += `,`
				}

				subjectiveAnswer, okay := subjectiveMap[s]
				if !okay {
					subjectiveAnswer = "{}"
				}

				r += param[0] + `:{` +
					`"` + common.FIELD_OBJECTIVE + `":"` + common.Unescape(objectiveAnswer) + `",` +
					`"` + common.FIELD_SUBJECTIVE + `":` + strings.Replace(subjectiveAnswer, "+", "%20", -1) + //common.UnescapeForJSON(subjectiveAnswer) +
					`}`
			}

			return r, nil
		})(); err == nil {
			return result, nil
		}
	}

	return "", common.ERR_NO_SERVICE
}

//----------------------------------------------------------------------------
