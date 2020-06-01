package service

import (
	"container/list"
	"github.com/wangmhgo/go-project/gdun/common"
	"strconv"
	"strings"
)

//----------------------------------------------------------------------------

type QuestionInfo struct {
	ID          int
	Body        []string
	Choices     []string
	Answer      []int
	Analysis    []string
	KnowledgeID int
}

func NewQuestionInfoFromMap(m map[string]string, id int) *QuestionInfo {
	s, okay := m[common.FIELD_BODY]
	if !okay {
		// fmt.Println("common.FIELD_BODY")
		return nil
	}
	body := strings.Split(s, "%0A")

	s, okay = m[common.FIELD_CHOICE]
	if !okay {
		// fmt.Println("common.FIELD_CHOICE")
		return nil
	}
	choice := strings.Split(s, "%0A")

	s, okay = m[common.FIELD_ANSWER]
	if !okay {
		// fmt.Println("common.FIELD_ANSWER")
		return nil
	}

	n := len(s)
	if n != len(choice) {
		// fmt.Println("n != len(choice)")
		return nil
	}
	answer := make([]int, n)
	for i := 0; i < n; i++ {
		if s[i] == '0' {
			answer[i] = 0
		} else {
			answer[i] = 1
		}
	}

	s, okay = m[common.FIELD_ANALYSIS]
	if !okay {
		return nil
	}
	analysis := strings.Split(s, "%0A")

	qi := new(QuestionInfo)
	qi.ID = id
	qi.Body = body
	qi.Choices = choice
	qi.Answer = answer
	qi.Analysis = analysis

	return qi
}

func (qi *QuestionInfo) ToJSON() string {
	s := `"` + common.FIELD_QUESTION + `":{` +
		`"` + common.FIELD_ID + `":` + strconv.Itoa(qi.ID) + `,` +
		`"` + common.FIELD_BODY + `":` + common.StringArrayToJSON(qi.Body) + `,` +
		`"` + common.FIELD_CHOICE + `":` + common.StringArrayToJSON(qi.Choices) + `,` +
		`"` + common.FIELD_ANSWER + `":` + common.IntArrayToJSON(qi.Answer) + `,` +
		`"` + common.FIELD_ANALYSIS + `":` + common.StringArrayToJSON(qi.Analysis) +
		`}`
	return s
}

//----------------------------------------------------------------------------

type QuestionInfoArray struct {
	Questions *list.List
}

func (qia *QuestionInfoArray) ToJSON() string {
	s := `"` + common.FIELD_QUESTION + `":[`

	if qia.Questions != nil {
		first := true
		if qia.Questions != nil {
			for p := qia.Questions.Front(); p != nil; p = p.Next() {
				qi, okay := p.Value.(*QuestionInfo)
				if !okay {
					continue
				}

				if first {
					first = false
				} else {
					s += `,`
				}

				s += `{` +
					`"` + common.FIELD_ID + `":` + strconv.Itoa(qi.ID) + `,` +
					`"` + common.FIELD_BODY + `":` + common.StringArrayToJSON(qi.Body) + `,` +
					`"` + common.FIELD_CHOICE + `":` + common.StringArrayToJSON(qi.Choices) + `,` +
					`"` + common.FIELD_ANSWER + `":` + common.IntArrayToJSON(qi.Answer) + `,` +
					`"` + common.FIELD_ANALYSIS + `":` + common.StringArrayToJSON(qi.Analysis) +
					`}`
			}
		}
	}

	s += `]`
	return s
}

//----------------------------------------------------------------------------
