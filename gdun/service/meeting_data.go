package service

import (
	"container/list"
	"github.com/wangmhgo/go-project/gdun/common"
	"strconv"
	"strings"
)

//----------------------------------------------------------------------------

type MeetingInfo struct {
	ID               int
	Name             string
	Subjects         []int
	Section          int
	Type             int
	Data             string
	Ally             int
	Scores           []int
	ScoreCount       int
	NumberOfAttendee int
	Coursewares      []string
	Videos           []string
	Exams            []string
	Replays          []string
	ClassID          int
	GroupID          int
	StartTime        int
	Duration         int
	EndTime          int
	UpdateTime       int
	UpdateIP         string
	Updater          int

	// Teachers         []int
	// Students         []int
}

func NewMeetingInfoFromMap(m map[string]string, id int) *MeetingInfo {
	name, okay := m[common.FIELD_NAME]
	if !okay {
		return nil
	}

	sbl, okay := m[common.FIELD_SUBJECT_LIST]
	if !okay {
		sbl = ""
	}

	s, okay := m[common.FIELD_SECTION]
	if !okay {
		return nil
	}
	section, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}

	s, okay = m[common.FIELD_TYPE]
	if !okay {
		return nil
	}
	t, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}

	data, okay := m[common.FIELD_DATA]
	if !okay {
		return nil
	}

	ally := 0
	if s, okay = m[common.FIELD_ALLY]; okay {
		if ally, err = strconv.Atoi(s); err != nil {
			ally = 0
		}
	}

	scores, okay := m[common.FIELD_SCORE]
	if !okay {
		return nil
	}

	s, okay = m[common.FIELD_SCORE_COUNT]
	if !okay {
		return nil
	}
	scoreCount, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}

	// tl, okay := m[common.FIELD_TEACHER_LIST]
	// if !okay {
	// 	return nil
	// }

	// sl, okay := m[common.FIELD_STUDENT_LIST]
	// if !okay {
	// 	return nil
	// }

	noa := 0
	if s, okay = m[common.FIELD_NUMBER_OF_ATTENDEE]; okay {
		if noa, err = strconv.Atoi(s); err != nil {
			noa = 0
		}
	}

	cwl, okay := m[common.FIELD_COURSEWARE_LIST]
	if !okay {
		return nil
	}

	vl, okay := m[common.FIELD_VIDEO_LIST]
	if !okay {
		return nil
	}

	el, okay := m[common.FIELD_EXAM_LIST]
	if !okay {
		return nil
	}

	rl, okay := m[common.FIELD_REPLAY_LIST]
	if !okay {
		return nil
	}

	cID := 0
	if s, okay = m[common.FIELD_CLASS_ID]; okay {
		if cID, err = strconv.Atoi(s); err != nil {
			return nil
		}
	}

	s, okay = m[common.FIELD_GROUP_ID]
	if !okay {
		return nil
	}
	gID, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}

	s, okay = m[common.FIELD_START_TIME]
	if !okay {
		return nil
	}
	st, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}

	s, okay = m[common.FIELD_DURATION]
	if !okay {
		return nil
	}
	d, err := strconv.Atoi(s)
	if err != nil {
		return nil
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

	uip, okay := m[common.FIELD_UPDATE_IP]
	if !okay {
		return nil
	}

	s, okay = m[common.FIELD_UPDATER]
	if !okay {
		return nil
	}
	ur, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}

	mi := new(MeetingInfo)
	mi.ID = id
	mi.Name = name
	mi.Subjects = common.StringToIntArray(sbl)
	mi.Section = section
	mi.Type = t
	mi.Data = data
	mi.Ally = ally
	mi.Scores = common.StringToIntArray(scores)
	mi.ScoreCount = scoreCount
	// mi.Teachers = common.StringToIntArray(tl)
	// mi.Students = common.StringToIntArray(sl)
	mi.NumberOfAttendee = noa
	mi.Coursewares = common.StringToStringArray(cwl)
	mi.Videos = common.StringToStringArray(vl)
	mi.Exams = common.StringToStringArray(el)
	mi.Replays = common.StringToStringArray(rl)
	mi.ClassID = cID
	mi.GroupID = gID
	mi.StartTime = st
	mi.Duration = d
	mi.EndTime = et
	mi.UpdateTime = ut
	mi.Updater = ur
	mi.UpdateIP = uip

	return mi
}

func (mi *MeetingInfo) ToJSON(isTeacher bool) string {
	r := `"` + common.FIELD_ID + `":` + strconv.Itoa(mi.ID) + `,` +
		`"` + common.FIELD_NAME + `":"` + common.UnescapeForJSON(mi.Name) + `",` +
		`"` + common.FIELD_SUBJECT + `":` + common.IntArrayToJSON(mi.Subjects) + `,` +
		`"` + common.FIELD_SECTION + `":` + strconv.Itoa(mi.Section) + `,` +
		`"` + common.FIELD_TYPE + `":` + strconv.Itoa(mi.Type) + `,` +
		`"` + common.FIELD_DATA + `":"` + common.UnescapeForJSON(mi.Data) + `",` +
		`"` + common.FIELD_SCORE + `":` + common.IntArrayToJSON(mi.Scores) + `,` +
		`"` + common.FIELD_SCORE_COUNT + `":` + strconv.Itoa(mi.ScoreCount) + `,`
	if isTeacher {
		r += `"` + common.FIELD_NUMBER_OF_ATTENDEE + `":` + strconv.Itoa(mi.NumberOfAttendee) + `,` +
			// `"` + common.FIELD_TEACHER_LIST + `":` + common.IntArrayToJSON(mi.Teachers) + `,` +
			`"` + common.FIELD_ALLY + `":` + strconv.Itoa(mi.Ally) + `,`
	}
	r += `"` + common.FIELD_COURSEWARE + `":` + common.ResourceArrayToJSON(mi.Coursewares, false) + `,` +
		`"` + common.FIELD_VIDEO + `":` + common.ResourceArrayToJSON(mi.Videos, false) + `,` +
		`"` + common.FIELD_EXAM + `":` + common.ResourceArrayToJSON(mi.Exams, true) + `,` +
		`"` + common.FIELD_REPLAY + `":` + common.StringArrayToJSON(mi.Replays) + `,` +
		`"` + common.FIELD_START_TIME + `":` + strconv.Itoa(mi.StartTime*1000) + `,` +
		`"` + common.FIELD_DURATION + `":` + strconv.Itoa(mi.Duration) + `,` +
		`"` + common.FIELD_CLASS_ID + `":` + strconv.Itoa(mi.ClassID) + `,` +
		`"` + common.FIELD_GROUP_ID + `":` + strconv.Itoa(mi.GroupID) + `,` +
		`"` + common.FIELD_UPDATE_TIME + `":` + strconv.Itoa(mi.UpdateTime*1000) + `,` +
		`"` + common.FIELD_END_TIME + `":` + strconv.Itoa(mi.EndTime*1000)

	return r
}

//----------------------------------------------------------------------------

type MeetingInfoArray struct {
	Meetinigs *list.List
}

func (mia *MeetingInfoArray) ToJSON(isTeacher bool) string {
	r := ""
	if mia.Meetinigs != nil {
		first := true
		for e := mia.Meetinigs.Front(); e != nil; e = e.Next() {
			mi, okay := e.Value.(*MeetingInfo)
			if !okay {
				continue
			}

			if first {
				first = false
			} else {
				r += ","
			}
			r += "{" + mi.ToJSON(isTeacher) + "}"
		}
	}

	return `"` + common.FIELD_MEETING + `":[` + r + `]`
}

//----------------------------------------------------------------------------

type UserMeetingProgressInfo struct {
	UserID             int
	MeetingID          int
	CoursewareProgress string
	VideoProgress      string
	MeetingProgress    int
	MeetingLog         string
	ExamAnswers        string
	ExamCorrect        int
	ExamTotal          int
	ReplayProgress     int
	Scores             []int
	UpdateTime         int
	UpdateIP           string
	Updater            int
}

func NewUserMeetingProgressInfoFromMap(m map[string]string, userID int, meetingID int) *UserMeetingProgressInfo {
	cwap, okay := m[common.FIELD_COURSEWARE_A]
	if !okay {
		cwap = ""
	}

	vap, okay := m[common.FIELD_VIDEO_A]
	if !okay {
		vap = ""
	}

	s, okay := m[common.FIELD_MEETING]
	if !okay {
		return nil
	}
	mp, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}

	ml, okay := m[common.FIELD_LOG]
	if !okay {
		return nil
	}

	ea, okay := m[common.FIELD_EXAM]
	if !okay {
		return nil
	}

	s, okay = m[common.FIELD_EXAM_CORRECT]
	if !okay {
		return nil
	}
	ec, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}

	s, okay = m[common.FIELD_EXAM_TOTAL]
	if !okay {
		return nil
	}
	et, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}

	rp := 0
	if s, okay = m[common.FIELD_REPLAY]; okay {
		rp, err = strconv.Atoi(s)
		if err != nil {
			rp = 0
		}
	}

	scores, okay := m[common.FIELD_SCORE]
	if !okay {
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

	uip, okay := m[common.FIELD_UPDATE_IP]
	if !okay {
		return nil
	}

	s, okay = m[common.FIELD_UPDATER]
	if !okay {
		return nil
	}
	ur, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}

	umpi := new(UserMeetingProgressInfo)

	umpi.UserID = userID
	umpi.MeetingID = meetingID
	umpi.CoursewareProgress = cwap
	umpi.VideoProgress = vap
	umpi.MeetingProgress = mp
	umpi.MeetingLog = ml
	umpi.ExamAnswers = ea
	umpi.ExamCorrect = ec
	umpi.ExamTotal = et
	umpi.ReplayProgress = rp
	umpi.Scores = common.StringToIntArray(scores)
	umpi.UpdateTime = ut
	umpi.UpdateIP = uip
	umpi.Updater = ur

	return umpi
}

func (umpi *UserMeetingProgressInfo) ToJSON(withHead bool) string {
	r := ``

	if withHead {
		r += `"` + common.FIELD_USER_ID + `":` + strconv.Itoa(umpi.UserID) + `,` +
			`"` + common.FIELD_MEETING_ID + `":` + strconv.Itoa(umpi.MeetingID) + `,`
	}
	r += `"` + common.FIELD_COURSEWARE + `":{`

	first := true
	arr := common.StringToStringArray(umpi.CoursewareProgress)
	for i := 0; i < len(arr); i++ {
		kvs := strings.Split(arr[i], ":")
		if len(kvs) != 2 {
			continue
		}

		if first {
			first = false
		} else {
			r += `,`
		}
		r += `"` + kvs[0] + `":` + kvs[1] + `000`
	}

	r += `},` +
		`"` + common.FIELD_VIDEO + `":{`

	first = true
	arr = common.StringToStringArray(umpi.VideoProgress)
	for i := 0; i < len(arr); i++ {
		kvs := strings.Split(arr[i], ":")
		if len(kvs) != 2 {
			continue
		}

		vs := strings.Split(kvs[1], "_")
		if len(vs) < 1 {
			continue
		}

		if first {
			first = false
		} else {
			r += `,`
		}
		r += `"` + kvs[0] + `":{` +
			`"` + common.FIELD_DURATION + `":` + vs[0] + `,` +
			`"` + common.FIELD_TIMESTAMP + `":`
		if len(vs) == 2 {
			r += vs[1] + `000`
		} else {
			r += `0`
		}
		r += `}`
	}

	r += `},` +
		`"` + common.FIELD_MEETING + `":` + strconv.Itoa(umpi.MeetingProgress) + `,` +
		`"` + common.FIELD_LOG + `":[` + umpi.MeetingLog + `],` +
		`"` + common.FIELD_EXAM + `":{`

	first = true
	arr = common.StringToStringArray(umpi.ExamAnswers)
	for i := 0; i < len(arr); i++ {
		kvs := strings.Split(arr[i], ":")
		if len(kvs) != 2 {
			continue
		}

		vs := strings.Split(kvs[1], "_")

		if first {
			first = false
		} else {
			r += `,`
		}

		switch len(vs) {
		case 3:
			r += `"` + kvs[0] + `":{"` + common.FIELD_TIMESTAMP + `":` + vs[0] + `000,"` + common.FIELD_CORRECT + `":` + vs[1] + `,"` + common.FIELD_TOTAL + `":` + vs[2] + `}`

		default:
			r += `"` + kvs[0] + `":` + vs[0] + `000`
		}
	}

	r += `},` +
		`"` + common.FIELD_EXAM_CORRECT + `":` + strconv.Itoa(umpi.ExamCorrect) + `,` +
		`"` + common.FIELD_EXAM_TOTAL + `":` + strconv.Itoa(umpi.ExamTotal) + `,` +
		`"` + common.FIELD_REPLAY + `":` + strconv.Itoa(umpi.ReplayProgress) + `,` +
		`"` + common.FIELD_SCORE + `":` + common.IntArrayToJSON(umpi.Scores)
	return r
}

//----------------------------------------------------------------------------

type UserMeetingProgressInfoArray struct {
	Status *list.List
}

func (umpia *UserMeetingProgressInfoArray) ToJSON(withHead bool) string {
	r := `"` + common.FIELD_PROGRESS + `":[`

	if umpia.Status != nil {
		first := true
		for p := umpia.Status.Front(); p != nil; p = p.Next() {
			umpi, okay := p.Value.(*UserMeetingProgressInfo)
			if !okay {
				continue
			}

			if first {
				first = false
			} else {
				r += `,`
			}

			r += `{` + umpi.ToJSON(withHead) + `}`
		}
	}
	r += `]`
	return r
}

//----------------------------------------------------------------------------

type MeetingInfoSlice []*MeetingInfo

func (s MeetingInfoSlice) Len() int {
	return len(s)
}
func (s MeetingInfoSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s MeetingInfoSlice) Less(i, j int) bool {
	if s[i].StartTime < s[j].StartTime {
		return true
	} else {
		return false
	}
}
func (s MeetingInfoSlice) ToJSON() string {
	first := true
	r := ``
	for i := 0; i < len(s); i++ {
		if first {
			first = false
		} else {
			r += `,`
		}
		r += `{` +
			`"` + common.FIELD_ID + `":` + strconv.Itoa(s[i].ID) + `,` +
			`"` + common.FIELD_NAME + `":"` + common.UnescapeForJSON(s[i].Name) + `",` +
			`"` + common.FIELD_TYPE + `":` + strconv.Itoa(s[i].Type) + `,` +
			`"` + common.FIELD_START_TIME + `":` + strconv.Itoa(s[i].StartTime*1000) + `,` +
			`"` + common.FIELD_DURATION + `":` + strconv.Itoa(s[i].Duration) +
			`}`
	}

	return r
}

//----------------------------------------------------------------------------
