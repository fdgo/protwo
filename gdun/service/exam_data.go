package service

import (
	"github.com/wangmhgo/go-project/gdun/common"
	"strconv"
	"strings"
)

//----------------------------------------------------------------------------

type ExamInfo struct {
	ID           int
	Name         string
	Key          string
	IV           string
	Count        int
	GdExamID int
	Answer       string
	GroupID      int
	UpdateIP     string
	UpdateTime   int
	Updater      int
}

func NewExamInfoFromMap(m map[string]string, id int) *ExamInfo {
	name, okay := m[common.FIELD_NAME]
	if !okay {
		return nil
	}

	key, okay := m[common.FIELD_KEY]
	if !okay {
		return nil
	}
	iv, okay := m[common.FIELD_IV]
	if !okay {
		return nil
	}

	s, okay := m[common.FIELD_COUNT]
	if !okay {
		return nil
	}
	cnt, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}

	s, okay = m[common.FIELD_GAODUN_EXAM_ID]
	if !okay {
		return nil
	}
	gdExamID, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}

	answer, okay := m[common.FIELD_ANSWER]
	if !okay {
		return nil
	}

	s, okay = m[common.FIELD_GROUP_ID]
	if !okay {
		return nil
	}
	groupID, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}

	updateIP, okay := m[common.FIELD_UPDATE_IP]
	if !okay {
		return nil
	}

	s, okay = m[common.FIELD_UPDATE_TIME]
	if !okay {
		return nil
	}
	updateTime, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}

	s, okay = m[common.FIELD_UPDATER]
	if !okay {
		return nil
	}
	updater, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}

	ei := new(ExamInfo)
	ei.ID = id
	ei.Name = name
	ei.Key = key
	ei.IV = iv
	ei.Count = cnt
	ei.GdExamID = gdExamID
	ei.Answer = answer
	ei.GroupID = groupID
	ei.UpdateIP = updateIP
	ei.UpdateTime = updateTime
	ei.Updater = updater

	return ei
}

func (ei *ExamInfo) ToJSON() string {
	r := `"` + common.FIELD_ID + `":` + strconv.Itoa(ei.ID) + `,` +
		`"` + common.FIELD_NAME + `":"` + common.UnescapeForJSON(ei.Name) + `",` +
		`"` + common.FIELD_KEY + `":"` + common.UnescapeForJSON(ei.Key) + `",` +
		`"` + common.FIELD_IV + `":"` + common.UnescapeForJSON(ei.IV) + `",` +
		`"` + common.FIELD_COUNT + `":` + strconv.Itoa(ei.Count) + `,` +
		`"` + common.FIELD_GAODUN_EXAM_ID + `":` + strconv.Itoa(ei.GdExamID) + `,` +
		`"` + common.FIELD_ANSWER + `":"` + common.UnescapeForJSON(ei.Answer) + `",` +
		`"` + common.FIELD_GROUP + `":` + strconv.Itoa(ei.GroupID)

	return r
}

//----------------------------------------------------------------------------

type ExamResult struct {
	UserID     int
	Answer     string
	Rank       int
	Count      int
	UpdateTime int
}

func NewExamResultFromString(s string) *ExamResult {
	if len(s) == 0 {
		return nil
	}

	arr := strings.Split(s, ":")
	if len(arr) != 4 {
		return nil
	}

	t, err := strconv.Atoi(arr[0])
	if err != nil {
		return nil
	}

	r, err := strconv.Atoi(arr[2])
	if err != nil {
		return nil
	}

	c, err := strconv.Atoi(arr[3])
	if err != nil {
		return nil
	}

	er := new(ExamResult)
	er.Answer = arr[1]
	er.Rank = r
	er.Count = c
	er.UpdateTime = t

	return er
}

func (er *ExamResult) ToJSON() string {
	r := `"` + common.FIELD_UPDATE_TIME + `":` + strconv.Itoa(er.UpdateTime*1000) + `,` +
		`"` + common.FIELD_RANK + `":` + strconv.Itoa(er.Rank) + `,` +
		`"` + common.FIELD_COUNT + `":` + strconv.Itoa(er.Count) + `,` +
		`"` + common.FIELD_ANSWER + `":"` + er.Answer + `"`
	return r
}

//----------------------------------------------------------------------------

type ExamResultSlice []*ExamResult

func (s ExamResultSlice) Len() int {
	return len(s)
}
func (s ExamResultSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ExamResultSlice) Less(i, j int) bool {
	if s[i].Rank < s[j].Rank {
		return true
	} else {
		return false
	}
}

//----------------------------------------------------------------------------

type ExamTask struct {
	ExamID int
	UserID int
	Answer string
}

//----------------------------------------------------------------------------
