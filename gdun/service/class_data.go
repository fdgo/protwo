package service

import (
	// "container/list"
	"github.com/wangmhgo/go-project/gdun/common"
	"sort"
	"strconv"
)

//----------------------------------------------------------------------------

type ClassInfo struct {
	ID                      int
	Name                    string
	Subjects                []int
	Teachers                []int
	Keepers                 string
	Students                []int
	Meetings                []int
	Deleted                 []int
	NumberOfFinishedMeeting int
	Parent                  []int
	GroupID                 int
	GdCourseID          int
	Template                int
	PlatformID              int
	PlatformData            string
	Ally                    int
	StartTime               int
	EndTime                 int
	Schedule                string
	NextTime                int
	UpdateTime              int
	UpdateIP                string
	Updater                 int
}

func NewClassInfoFromMap(m map[string]string, id int) *ClassInfo {
	name, okay := m[common.FIELD_NAME]
	if !okay {
		return nil
	}

	sbl, okay := m[common.FIELD_SUBJECT_LIST]
	if !okay {
		sbl = ""
	}

	tl, okay := m[common.FIELD_TEACHER_LIST]
	if !okay {
		return nil
	}

	kl, okay := m[common.FIELD_KEEPER_LIST]
	if !okay {
		kl = ""
	}

	sl, okay := m[common.FIELD_STUDENT_LIST]
	if !okay {
		return nil
	}

	ml, okay := m[common.FIELD_MEETING_LIST]
	if !okay {
		return nil
	}

	dl, okay := m[common.FIELD_DELETED]
	if !okay {
		return nil
	}

	s, okay := m[common.FIELD_NUMBER_OF_FINISHED_MEETING]
	if !okay {
		return nil
	}
	nfm, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}

	s, okay = m[common.FIELD_GAODUN_COURSE_ID]
	if !okay {
		return nil
	}
	gdCourseID, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}

	template := 0
	if s, okay = m[common.FIELD_TEMPLATE]; okay {
		if template, err = strconv.Atoi(s); err != nil {
			template = 0
		}
	}

	var parent []int = nil
	s, okay = m[common.FIELD_PARENT]
	if okay {
		parent = common.StringToIntArray(s)
	} else {
		parent = []int{}
	}

	s, okay = m[common.FIELD_GROUP_ID]
	if !okay {
		return nil
	}
	gID, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}

	s, okay = m[common.FIELD_PLATFORM_ID]
	if !okay {
		return nil
	}
	platformID, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}

	platformData, okay := m[common.FIELD_PLATFORM_DATA]
	if !okay {
		return nil
	}

	ally := 0
	if s, okay = m[common.FIELD_ALLY]; okay {
		if ally, err = strconv.Atoi(s); err != nil {
			ally = 0
		}
	}

	s, okay = m[common.FIELD_START_TIME]
	if !okay {
		return nil
	}
	st, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}

	nt := 0
	s, okay = m[common.FIELD_NEXT_TIME]
	if okay {
		nt, err = strconv.Atoi(s)
		if err != nil {
			nt = 0
		}
	}

	s, okay = m[common.FIELD_END_TIME]
	if !okay {
		return nil
	}
	et, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}

	s, okay = m[common.FIELD_UPDATE_TIME]
	if !okay {
		return nil
	}
	ut, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}

	sch, okay := m[common.FIELD_SCHEDULE]
	if !okay {
		// TODO: Update it according to the meeting list.
		sch = ""
	}

	s, okay = m[common.FIELD_UPDATER]
	if !okay {
		return nil
	}
	ur, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}

	uip, okay := m[common.FIELD_UPDATE_IP]
	if !okay {
		return nil
	}

	ci := new(ClassInfo)
	ci.ID = id

	ci.Name = name
	ci.Subjects = common.StringToIntArray(sbl)
	ci.Teachers = common.StringToIntArray(tl)
	ci.Keepers = kl
	ci.Students = common.StringToIntArray(sl)
	ci.Meetings = common.StringToIntArray(ml)
	ci.Deleted = common.StringToIntArray(dl)
	ci.NumberOfFinishedMeeting = nfm
	ci.Parent = parent
	ci.GroupID = gID
	ci.GdCourseID = gdCourseID
	ci.Template = template
	ci.PlatformID = platformID
	ci.PlatformData = platformData
	ci.Ally = ally
	ci.StartTime = st
	ci.EndTime = et
	ci.NextTime = nt
	ci.Schedule = sch
	ci.UpdateTime = ut
	ci.Updater = ur
	ci.UpdateIP = uip

	return ci
}

func (ci *ClassInfo) ToJSON(forStudent bool) string {
	r := `"` + common.FIELD_ID + `":` + strconv.Itoa(ci.ID) + `,` +
		`"` + common.FIELD_NAME + `":"` + common.UnescapeForJSON(ci.Name) + `",` +
		`"` + common.FIELD_SUBJECT + `":` + common.IntArrayToJSON(ci.Subjects) + `,` +
		`"` + common.FIELD_MEETING + `":` + common.IntArrayToJSON(ci.Meetings) + `,` +
		`"` + common.FIELD_NUMBER_OF_FINISHED_MEETING + `":` + strconv.Itoa(ci.NumberOfFinishedMeeting) + `,`
	if !forStudent {
		r += `"` + common.FIELD_DELETED + `":` + common.IntArrayToJSON(ci.Deleted) + `,` +
			`"` + common.FIELD_PARENT + `":` + common.IntArrayToJSON(ci.Parent) + `,` +
			`"` + common.FIELD_TEACHER + `":` + common.IntArrayToJSON(ci.Teachers) + `,` +
			`"` + common.FIELD_KEEPER + `":{` + ci.Keepers + `},` +
			`"` + common.FIELD_STUDENT + `":` + common.IntArrayToJSON(ci.Students) + `,` +
			`"` + common.FIELD_ALLY + `":` + strconv.Itoa(ci.Ally) + `,`
	}
	r += `"` + common.FIELD_GROUP_ID + `":` + strconv.Itoa(ci.GroupID) + `,` +
		`"` + common.FIELD_GAODUN_COURSE_ID + `":` + strconv.Itoa(ci.GdCourseID) + `,` +
		`"` + common.FIELD_TEMPLATE + `":` + strconv.Itoa(ci.Template) + `,` +
		`"` + common.FIELD_PLATFORM_ID + `":` + strconv.Itoa(ci.PlatformID) + `,` +
		`"` + common.FIELD_PLATFORM_DATA + `":"` + common.UnescapeForJSON(ci.PlatformData) + `",` +
		`"` + common.FIELD_UPDATE_TIME + `":` + strconv.Itoa(ci.UpdateTime*1000) + `,` +
		`"` + common.FIELD_START_TIME + `":` + strconv.Itoa(ci.StartTime*1000) + `,` +
		`"` + common.FIELD_END_TIME + `":` + strconv.Itoa(ci.EndTime*1000) + `,` +
		`"` + common.FIELD_NEXT_TIME + `":` + strconv.Itoa(ci.NextTime*1000) + `,` +
		`"` + common.FIELD_SCHEDULE + `":[` + ci.Schedule + `]`

	return r
}

//----------------------------------------------------------------------------

type ClassInfoSlice []*ClassInfo

func (s ClassInfoSlice) Len() int {
	return len(s)
}
func (s ClassInfoSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ClassInfoSlice) Less(i, j int) bool {
	if s[i] == nil {
		return true
	}
	if s[j] == nil {
		return false
	}
	if s[i].ID < s[j].ID {
		return true
	}

	return false
}
func (s ClassInfoSlice) ToJSON(forStudent bool) string {
	sort.Sort(s)

	r := `"` + common.FIELD_CLASS_LIST + `":[`

	first := true
	for i := len(s) - 1; i >= 0; i-- {
		ci := s[i]
		if ci == nil {
			continue
		}

		if first {
			first = false
		} else {
			r += `,`
		}
		r += `{` + ci.ToJSON(forStudent) + `}`
	}
	r += `]`

	return r
}

//----------------------------------------------------------------------------

type ClassInvitationInfo struct {
	EndTime     int
	Duration    int
	ExpiredTime int
	Teacher     int
	Channel     string
	UpdateTime  int
	Updater     int
	UpdateIP    string
}

//----------------------------------------------------------------------------
