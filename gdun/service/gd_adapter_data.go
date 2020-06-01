package service

import (
	"container/list"
	"fmt"
	"github.com/wangmhgo/go-project/gdun/common"
	"sort"
	"strconv"
	"strings"
)

//----------------------------------------------------------------------------

type GdVideoResource struct {
	ID   string
	Name string
}

func (gvr *GdVideoResource) ToJSON() string {
	r := `"` + common.FIELD_ID + `":"` + common.UnescapeForJSON(gvr.ID) + `",` +
		`"` + common.FIELD_NAME + `":"` + common.UnescapeForJSON(gvr.Name) + `"`
	return r
}

//----------------------------------------------------------------------------

type GdCourseware struct {
	ID        int
	Name      string
	Resources *list.List
}

func NewGdCourseware(id int, name string) *GdCourseware {
	gcw := new(GdCourseware)
	gcw.ID = id
	gcw.Name = name
	gcw.Resources = list.New()

	return gcw
}

//----------------------------------------------------------------------------

func (gcw *GdCourseware) ToJSON() string {
	r := `"` + common.FIELD_ID + `":` + strconv.Itoa(gcw.ID) + `,` +
		`"` + common.FIELD_NAME + `":"` + common.UnescapeForJSON(gcw.Name) + `"` + `,` +
		`"` + common.FIELD_RESOURCE_LIST + `":[`
	if gcw.Resources != nil {
		first := true
		for e := gcw.Resources.Front(); e != nil; e = e.Next() {
			if video, okay := e.Value.(*GdVideoResource); okay {
				if first {
					first = false
				} else {
					r += `,`
				}

				r += `{` + video.ToJSON() + `}`

			} else if exam, okay := e.Value.(*GdTikuExam); okay {
				if first {
					first = false
				} else {
					r += `,`
				}

				r += `{` + exam.ToJSON() + `}`
			}
		}
	}
	r += `]`

	return r
}

//----------------------------------------------------------------------------

type GdCourse struct {
	ID          int
	Coursewares *list.List
}

func NewGdCourse(id int) *GdCourse {
	gc := new(GdCourse)
	gc.ID = id
	gc.Coursewares = list.New()

	return gc
}

//----------------------------------------------------------------------------

func (gc *GdCourse) ToJSON() string {
	r := `"` + common.FIELD_ID + `":` + strconv.Itoa(gc.ID) + `,` +
		`"` + common.FIELD_COURSEWARE_LIST + `":[`
	if gc.Coursewares != nil {
		first := true
		for e := gc.Coursewares.Front(); e != nil; e = e.Next() {
			cw, okay := e.Value.(*GdCourseware)
			if !okay {
				continue
			}

			if first {
				first = false
			} else {
				r += `,`
			}

			r += `{` + cw.ToJSON() + `}`
		}
	}
	r += `]`

	return r
}

//----------------------------------------------------------------------------

type GdTikuItem struct {
	GdQuestionID int
	Type             int
	Body             string
	Choice           string
	Answer           string
	Analysis         string
}

func (gti *GdTikuItem) ToCSV() string {
	s := strconv.Itoa(gti.GdQuestionID) + "," +
		common.UnescapeForCSV(gti.Body) + "," +
		common.UnescapeForCSV(gti.Choice) + "," +
		common.UnescapeForCSV(gti.Answer) + "," +
		common.UnescapeForCSV(gti.Analysis) + "\n"

	return s
}

func (gti *GdTikuItem) ToJSON(id *int) string {
	r := `"` + common.FIELD_ID + `":` + strconv.Itoa(*id) + `,` +
		`"` + common.FIELD_GAODUN_QUESTION_ID + `":` + strconv.Itoa(gti.GdQuestionID) + `,` +
		`"` + common.FIELD_TYPE + `":` + strconv.Itoa(gti.Type) + `,` +
		`"` + common.FIELD_BODY + `":"` + common.UnescapeForJSON(gti.Body) + `",` +
		`"` + common.FIELD_CHOICE + `":"` + common.UnescapeForJSON(gti.Choice) + `",` +
		`"` + common.FIELD_ANSWER + `":"` + common.UnescapeForJSON(gti.Answer) + `",` +
		`"` + common.FIELD_ANALYSIS + `":"` + common.UnescapeForJSON(gti.Analysis) + `"`

	(*id)++

	return r
}

//----------------------------------------------------------------------------

type GdTikuComposedItem struct {
	GdQuestionID int
	Body             string
	Items            *list.List
}

func NewGdTikuComposedItem() *GdTikuComposedItem {
	gtci := new(GdTikuComposedItem)
	gtci.Items = list.New()

	return gtci
}

func (gtci *GdTikuComposedItem) ToJSON(id *int) string {
	r := `"` + common.FIELD_GAODUN_QUESTION_ID + `":` + strconv.Itoa(gtci.GdQuestionID) + `,` +
		`"` + common.FIELD_BODY + `":"` + common.UnescapeForJSON(gtci.Body) + `",` +
		`"` + common.FIELD_QUESTION + `":[`

	first := true
	for e := gtci.Items.Front(); e != nil; e = e.Next() {
		item, okay := e.Value.(*GdTikuItem)
		if !okay {
			continue
		}

		if first {
			first = false
		} else {
			r += `,`
		}
		r += `{` + item.ToJSON(id) + `}`
	}

	r += `]`

	return r
}

//----------------------------------------------------------------------------

type GdTikuSubExam struct {
	Name  string
	Items *list.List
}

func NewGdTikuSubExam() *GdTikuSubExam {
	gtes := new(GdTikuSubExam)
	gtes.Items = list.New()

	return gtes
}

func (gtes *GdTikuSubExam) ToJSON(id *int) string {
	r := `"` + common.FIELD_NAME + `":"` + common.UnescapeForJSON(gtes.Name) + `",` +
		`"` + common.FIELD_QUESTION + `":[`

	first := true
	for e := gtes.Items.Front(); e != nil; e = e.Next() {
		if gti, okay := e.Value.(*GdTikuItem); okay {
			if first {
				first = false
			} else {
				r += `,`
			}
			r += `{` + gti.ToJSON(id) + `}`

		} else if gtci, okay := e.Value.(*GdTikuComposedItem); okay {
			if first {
				first = false
			} else {
				r += `,`
			}
			r += `{` + gtci.ToJSON(id) + `}`
		}
	}

	r += `]`

	return r
}

type GdTikuSubExamSlice []*GdTikuSubExam

func (s GdTikuSubExamSlice) Len() int {
	return len(s)
}
func (s GdTikuSubExamSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s GdTikuSubExamSlice) Less(i, j int) bool {
	if strings.Compare(s[i].Name, s[j].Name) < 0 {
		return true
	}

	return false
}

//----------------------------------------------------------------------------

type GdTikuExam struct {
	ID       int
	Name     string
	Count    int
	Answer   string
	Subjects *list.List
}

func NewGdTikuExam(gdExamID int) *GdTikuExam {
	gte := new(GdTikuExam)
	gte.ID = gdExamID
	gte.Name = ""
	gte.Count = 0
	gte.Answer = ""
	gte.Subjects = list.New()

	return gte
}

func (gte *GdTikuExam) solve(q *GdTikuItem) {
	if (q.Type != 4) && (q.Type != 6) {
		n := 0
		for i := 0; i < len(q.Answer); i++ {
			if q.Answer[i] == '1' {
				n |= (0x01 << uint(i))
			}
		}
		gte.Answer += fmt.Sprintf("%02x", n)
	}
	gte.Count++
}

func (gte *GdTikuExam) Solve() {
	gte.Count = 0
	gte.Answer = ""

	for e := gte.Subjects.Front(); e != nil; e = e.Next() {
		se, okay := e.Value.(*GdTikuSubExam)
		if !okay {
			continue
		}

		for p := se.Items.Front(); p != nil; p = p.Next() {
			if si, okay := p.Value.(*GdTikuItem); okay {
				gte.solve(si)
			} else if ci, okay := p.Value.(*GdTikuComposedItem); okay {
				for q := ci.Items.Front(); q != nil; q = q.Next() {
					if si, okay := q.Value.(*GdTikuItem); okay {
						gte.solve(si)
					}
				}
			}
		}
	}
}

func (gte *GdTikuExam) ToJSON() string {
	r := `"` + common.FIELD_NAME + `":"` + common.UnescapeForJSON(gte.Name) + `",` +
		`"` + common.FIELD_GAODUN_EXAM_ID + `":` + strconv.Itoa(gte.ID) + `,` +
		`"` + common.FIELD_SUBJECT + `":[`

	id := 1
	first := true
	for e := gte.Subjects.Front(); e != nil; e = e.Next() {
		subject, okay := e.Value.(*GdTikuSubExam)
		if !okay {
			continue
		}

		if first {
			first = false
		} else {
			r += `,`
		}
		r += `{` + subject.ToJSON(&id) + `}`
	}

	r += `],"` + common.FIELD_COUNT + `":` + strconv.Itoa(id-1)

	return strings.Replace(r, "%2B", "+", -1)
}

func (gte *GdTikuExam) Sort() {
	n := gte.Subjects.Len()
	if n == 0 {
		return
	}

	arr := make([]*GdTikuSubExam, n)
	i := 0
	for e := gte.Subjects.Front(); e != nil; e = e.Next() {
		subject, okay := e.Value.(*GdTikuSubExam)
		if (!okay) || (subject == nil) {
			continue
		}

		arr[i] = subject
		i++
	}

	sort.Sort(GdTikuSubExamSlice(arr))

	gte.Subjects.Init()
	for i = 0; i < len(arr); i++ {
		gte.Subjects.PushBack(arr[i])
	}
}

//----------------------------------------------------------------------------

type IDPositionPair struct {
	ID       int
	Position int
}

type IDPositionSlice []*IDPositionPair

func (s IDPositionSlice) Len() int           { return len(s) }
func (s IDPositionSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s IDPositionSlice) Less(i, j int) bool { return s[i].Position < s[j].Position }

//----------------------------------------------------------------------------
