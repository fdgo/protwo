package service

import (
	"container/list"
	"errors"
	"fmt"
	"github.com/wangmhgo/go-project/gdun/common"
	"strconv"
	"strings"
)

//----------------------------------------------------------------------------

type ExamQuestionService struct {
	db    *common.Database
	cache *common.Cache
}

func NewExamQuestionService(db *common.Database, cache *common.Cache) (*ExamQuestionService, error) {
	qs := new(ExamQuestionService)
	qs.db = db
	qs.cache = cache

	err := qs.Init()
	if err != nil {
		return nil, err
	}

	return qs, nil
}

//----------------------------------------------------------------------------

func (qs *ExamQuestionService) Init() error {
	if qs.db == nil {
		return common.ERR_NO_DATABASE
	}

	sql := "CREATE TABLE IF NOT EXISTS `" + common.TABLE_QUESTION_SELECTION + "` ("
	sql += " `" + common.FIELD_ID + "` INT NOT NULL AUTO_INCREMENT,"
	sql += " `" + common.FIELD_BODY + "` TEXT NOT NULL,"
	sql += " `" + common.FIELD_CHOICE + "` TEXT NOT NULL,"
	sql += " `" + common.FIELD_ANSWER + "` VARCHAR(16) NOT NULL,"
	sql += " `" + common.FIELD_ANALYSIS + "` TEXT NOT NULL,"
	sql += " `" + common.FIELD_KNOWLEDGE_ID + "` INT DEFAULT 0,"
	sql += " `" + common.FIELD_UPDATE_IP + "` VARCHAR(32) NOT NULL,"
	sql += " `" + common.FIELD_UPDATE_TIME + "` BIGINT NOT NULL,"
	sql += " `" + common.FIELD_UPDATER + "` INT NOT NULL,"
	sql += " PRIMARY KEY (`" + common.FIELD_ID + "`)"
	sql += ") ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;"

	_, err := qs.db.Exec(sql)
	return err
}

//----------------------------------------------------------------------------

func (qs *ExamQuestionService) Preload() (int, error) {
	if qs.db == nil {
		return 0, common.ERR_NO_DATABASE
	}
	if qs.cache == nil {
		return 0, common.ERR_NO_CACHE
	}

	rest, err := qs.db.Count(common.TABLE_QUESTION_SELECTION)
	if err != nil {
		return 0, err
	}

	i := 0
	for i < rest {
		n, err := qs.preload(i, common.DATABASE_PRELOAD_SIZE)
		if err != nil {
			return i + n, err
		}
		i += n
	}

	return i, nil
}

func (qs *ExamQuestionService) preload(start int, length int) (int, error) {
	sql := "SELECT " +
		common.FIELD_ID + "," +
		common.FIELD_BODY + "," +
		common.FIELD_CHOICE + "," +
		common.FIELD_ANSWER + "," +
		common.FIELD_ANALYSIS + "," +
		common.FIELD_KNOWLEDGE_ID +
		" FROM " +
		common.TABLE_QUESTION_SELECTION +
		" LIMIT " + strconv.Itoa(start) + "," + strconv.Itoa(length) + ";"

	rows, err := qs.db.Select(sql)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	m := make(map[string]string)

	cnt := 0

	id := 0
	body := ""
	choice := ""
	answer := ""
	analysis := ""
	knowledgeID := 0
	for rows.Next() {
		err = rows.Scan(&id, &body, &choice, &answer, &analysis, &knowledgeID)
		if err != nil {
			return cnt, err
		}

		m[common.FIELD_BODY] = body
		m[common.FIELD_CHOICE] = choice
		m[common.FIELD_ANSWER] = answer
		m[common.FIELD_ANALYSIS] = analysis
		m[common.FIELD_KNOWLEDGE_ID] = strconv.Itoa(knowledgeID)

		err = qs.cache.SetFields(common.KEY_PREFIX_QUESTION+strconv.Itoa(id), m)
		if err != nil {
			return cnt, err
		}

		cnt++
	}

	return cnt, nil
}

//----------------------------------------------------------------------------

func (qs *ExamQuestionService) AddSelection(body string, choices string, answer string, analysis string, knowledgeID int, session *Session) (int, error) {

	//----------------------------------------------------
	// Check the inputs.

	if len(body) == 0 {
		return 0, errors.New("Empty body.")
	}

	arr := strings.Split(choices, "\n")
	nc := len(arr)
	if nc < 2 {
		return 0, errors.New("The number of choices is less than 2.")
	}
	for i := 0; i < nc; i++ {
		if len(arr[i]) == 0 {
			return 0, errors.New("One choice is empty at least.")
		}
	}

	na := len(answer)
	if nc != na {
		return 0, errors.New("Invalid number of answers.")
	}

	okay := false
	for i := 0; i < na; i++ {
		if answer[i] == '1' {
			okay = true
			break
		}
	}
	if !okay {
		return 0, errors.New("No correct answer.")
	}

	//----------------------------------------------------
	// Encoding them.

	sBody := common.Escape(body)
	if len(sBody) == 0 {
		return 0, errors.New("Empty body.")
	}

	sChoices := common.Escape(choices)
	if len(sChoices) == 0 {
		return 0, errors.New("Empty choice.")
	}

	sAnswer := common.Escape(answer)
	if len(sAnswer) == 0 {
		return 0, errors.New("Empty answer.")
	}

	sAnalysis := common.Escape(analysis)

	timestamp := common.GetTimeString()

	if qs.db == nil {
		return 0, common.ERR_NO_SERVICE
	}

	// Step 1. Save it to database.
	sql := "INSERT INTO " + common.TABLE_QUESTION_SELECTION + " (" +
		common.FIELD_BODY + "," +
		common.FIELD_CHOICE + "," +
		common.FIELD_ANSWER + "," +
		common.FIELD_ANALYSIS + "," +
		common.FIELD_KNOWLEDGE_ID + "," +
		common.FIELD_UPDATE_TIME + "," +
		common.FIELD_UPDATE_IP + "," +
		common.FIELD_UPDATER +
		") VALUES (" +
		"'" + sBody + "'," +
		"'" + sChoices + "'," +
		"'" + sAnswer + "'," +
		"'" + sAnalysis + "'," +
		strconv.Itoa(knowledgeID) + "," +
		timestamp + "," +
		"'" + session.IP + "'," +
		strconv.Itoa(session.UserID) +
		");"
	id, err := qs.db.Insert(sql, 1)
	if err != nil {
		return 0, err
	}

	// Step 2. Save it to cache.
	if qs.cache != nil {
		m := make(map[string]string)
		m[common.FIELD_BODY] = sBody
		m[common.FIELD_CHOICE] = sChoices
		m[common.FIELD_ANSWER] = sAnswer
		m[common.FIELD_ANALYSIS] = sAnalysis
		m[common.FIELD_KNOWLEDGE_ID] = strconv.Itoa(knowledgeID)

		err = qs.cache.SetFields(common.KEY_PREFIX_QUESTION+strconv.FormatInt(id, 10), m)
		if err != nil {
			return 0, err
		}
	}

	return int(id), nil
}

//----------------------------------------------------------------------------

func (qs *ExamQuestionService) DeleteSelection(id int, session *Session) error {
	okay := false

	if qs.db != nil {
		go (func() error {
			sql := "DELETE FROM " + common.TABLE_QUESTION_SELECTION +
				" WHERE " +
				common.FIELD_ID + "=" + strconv.Itoa(id) + ";"

			_, err := qs.db.Exec(sql)
			return err
		})()

		okay = true
	}

	if qs.cache != nil {
		go (func() error {
			return qs.cache.Del(common.KEY_PREFIX_QUESTION + strconv.Itoa(id))
		})()

		okay = true
	}

	if okay {
		return nil
	} else {
		return common.ERR_NO_SERVICE
	}
}

//----------------------------------------------------------------------------
// Database: Compatible.
// Cache   : Compatible.

func (qs *ExamQuestionService) GetSelections(selectionIDs []int, session *Session) (*QuestionInfoArray, error) {
	if selectionIDs == nil || len(selectionIDs) == 0 {
		fmt.Println("selectionIDs == nil || len(selectionIDs) == 0")
		return new(QuestionInfoArray), nil
	}

	if qs.cache != nil {
		qia, err := (func() (*QuestionInfoArray, error) {
			qia := new(QuestionInfoArray)
			qia.Questions = list.New()

			for i := 0; i < len(selectionIDs); i++ {
				m, err := qs.cache.GetAllFields(common.KEY_PREFIX_QUESTION + strconv.Itoa(selectionIDs[i]))
				if err != nil {
					return nil, err
				}

				qi := NewQuestionInfoFromMap(m, selectionIDs[i])
				if qi != nil {
					qia.Questions.PushBack(qi)
				} else {
					fmt.Printf("%d qi == nil\n", selectionIDs[i])
				}
			}

			return qia, nil
		})()

		if err == nil {
			return qia, nil
		} else {
			fmt.Println(err.Error())
		}
	}

	if qs.db != nil {
		sql := "SELECT " +
			common.FIELD_ID + "," +
			common.FIELD_BODY + "," +
			common.FIELD_CHOICE + "," +
			common.FIELD_ANSWER + "," +
			common.FIELD_ANALYSIS +
			" FROM " +
			common.TABLE_QUESTION_SELECTION +
			" WHERE " +
			common.FIELD_ID + " IN (" + common.IntArrayToString(selectionIDs) + ");"

		rows, err := qs.db.Select(sql)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		qia := new(QuestionInfoArray)
		qia.Questions = list.New()

		b := ""
		cl := ""
		al := ""
		for rows.Next() {
			qi := new(QuestionInfo)
			err = rows.Scan(&qi.ID, &b, &cl, &al, &qi.Analysis)
			if err != nil {
				return nil, err
			}

			qi.Body = strings.Split(b, "%0A")

			qi.Choices = strings.Split(cl, "%0A")
			n := len(qi.Choices)
			if len(al) != n {
				return nil, errors.New("Invalid choices and answers.")
			}

			qi.Answer = make([]int, n)
			for i := 0; i < n; i++ {
				if al[i] == '0' {
					qi.Answer[i] = 0
				} else {
					qi.Answer[i] = 1
				}
			}

			qia.Questions.PushBack(qi)
		}

		return qia, nil
	}

	return nil, common.ERR_NO_SERVICE
}

//----------------------------------------------------------------------------

func (qs *ExamQuestionService) GetSelection(selectionID int, session *Session) (*QuestionInfo, error) {

	if qs.cache != nil {
		m, err := qs.cache.GetAllFields(common.KEY_PREFIX_QUESTION + strconv.Itoa(selectionID))
		if err != nil {
			return nil, err
		}

		s := ""
		okay := false

		qi := new(QuestionInfo)
		qi.ID = selectionID

		s, okay = m[common.FIELD_BODY]
		if !okay {
			return nil, errors.New("Could not find the question body.")
		}
		qi.Body = strings.Split(s, "%0A")

		s, okay = m[common.FIELD_CHOICE]
		if !okay {
			return nil, errors.New("Could not find the question choice.")
		}
		qi.Choices = strings.Split(s, "%0A")

		s, okay = m[common.FIELD_ANSWER]
		if !okay {
			return nil, errors.New("Could not find the question answer.")
		}
		qi.Answer = common.StringToIntArray(s)

		s, okay = m[common.FIELD_ANALYSIS]
		if !okay {
			return nil, errors.New("Could not find the question analysis.")
		}
		qi.Analysis = common.StringToStringArray(s)

		s, okay = m[common.FIELD_KNOWLEDGE_ID]
		if !okay {
			return nil, errors.New("Could not find the knowledge ID.")
		}
		qi.KnowledgeID, err = strconv.Atoi(s)
		if err != nil {
			return nil, errors.New("Invalid knowledge ID.")
		}

		return qi, nil
	}

	if qs.db != nil {
		sql := "SELECT " +
			common.FIELD_ID + "," +
			common.FIELD_BODY + "," +
			common.FIELD_CHOICE + "," +
			common.FIELD_ANSWER + "," +
			common.FIELD_ANALYSIS +
			" FROM " +
			common.TABLE_QUESTION_SELECTION +
			" WHERE " +
			common.FIELD_ID + "=" + strconv.Itoa(selectionID) + ";"

		rows, err := qs.db.Select(sql)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		if !rows.Next() {
			return nil, errors.New("The question does not exist.")
		}

		b := ""
		cl := ""
		al := ""
		qi := new(QuestionInfo)
		err = rows.Scan(&qi.ID, &b, &cl, &al, &qi.Analysis)
		if err != nil {
			return nil, err
		}

		qi.Body = strings.Split(b, "%0A")

		qi.Choices = strings.Split(cl, "%0A")
		n := len(qi.Choices)
		if len(al) != n {
			return nil, errors.New("Invalid choices and answers.")
		}

		qi.Answer = make([]int, n)
		for i := 0; i < n; i++ {
			if al[i] == '0' {
				qi.Answer[i] = 0
			} else {
				qi.Answer[i] = 1
			}
		}

		return qi, nil
	}

	return nil, common.ERR_NO_SERVICE
}

//----------------------------------------------------------------------------

func (qs *ExamQuestionService) IsCorrect(selectionID int, answer string) (bool, error) {
	// Check the input.
	if len(answer) == 0 {
		return false, nil
	}

	// Check it via cache.
	if qs.cache != nil {
		s, err := qs.cache.GetField(common.KEY_PREFIX_QUESTION+strconv.Itoa(selectionID), common.FIELD_ANSWER)
		if err != nil {
			return false, err
		}

		if s == answer {
			return true, nil
		} else {
			return false, nil
		}
	}

	// Otherwise check it via database.
	if qs.db != nil {
		sql := "SELECT " +
			common.FIELD_ANSWER +
			" FROM " +
			common.TABLE_QUESTION_SELECTION +
			" WHERE " +
			common.FIELD_ID + "=" + strconv.Itoa(selectionID) + ";"
		rows, err := qs.db.Select(sql)
		if err != nil {
			return false, err
		}
		defer rows.Close()

		if !rows.Next() {
			return false, common.ERR_NO_QUESTION
		}

		s := ""
		err = rows.Scan(&s)
		if err != nil {
			return false, err
		}

		if s == answer {
			return true, nil
		} else {
			return false, nil
		}
	}

	return false, common.ERR_NO_SERVICE
}

//----------------------------------------------------------------------------
