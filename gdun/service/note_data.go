package service

import (
	"github.com/wangmhgo/go-project/gdun/common"
	"strconv"
	"strings"
)

//----------------------------------------------------------------------------

type Note struct {
	ID         int
	ClassID    int
	MeetingID  int
	UserID     int
	Type       int
	Key        string
	SubKey     int
	Body       string
	UpdateIP   string
	UpdateTime int
}

func NewNoteFromString(s string) *Note {
	if len(s) == 0 {
		return nil
	}
	arr := strings.Split(s, "_")
	if len(arr) != 10 {
		return nil
	}

	var err error
	n := new(Note)
	if n.ID, err = strconv.Atoi(arr[0]); err != nil {
		return nil
	}
	if n.ClassID, err = strconv.Atoi(arr[1]); err != nil {
		return nil
	}
	if n.MeetingID, err = strconv.Atoi(arr[2]); err != nil {
		return nil
	}
	if n.UserID, err = strconv.Atoi(arr[3]); err != nil {
		return nil
	}
	if n.Type, err = strconv.Atoi(arr[4]); err != nil {
		return nil
	}
	n.Key = arr[5]
	if n.SubKey, err = strconv.Atoi(arr[6]); err != nil {
		return nil
	}
	n.UpdateIP = arr[7]
	if n.UpdateTime, err = strconv.Atoi(arr[8]); err != nil {
		return nil
	}
	n.Body = arr[9]
	return n
}

func (n *Note) GetStringPrefix() string {
	return strconv.Itoa(n.ID) + "_" +
		strconv.Itoa(n.ClassID) + "_" +
		strconv.Itoa(n.MeetingID) + "_" +
		strconv.Itoa(n.UserID) + "_" +
		strconv.Itoa(n.Type) + "_"
}

func (n *Note) ToString() string {
	return n.GetStringPrefix() +
		n.Key + "_" +
		strconv.Itoa(n.SubKey) + "_" +
		n.UpdateIP + "_" +
		strconv.Itoa(n.UpdateTime) + "_" +
		n.Body
}

func (n *Note) ToJSON() string {
	return `"` + common.FIELD_ID + `":` + strconv.Itoa(n.ID) + `,` +
		`"` + common.FIELD_CLASS_ID + `":` + strconv.Itoa(n.ClassID) + `,` +
		`"` + common.FIELD_MEETING_ID + `":` + strconv.Itoa(n.MeetingID) + `,` +
		`"` + common.FIELD_USER_ID + `":` + strconv.Itoa(n.UserID) + `,` +
		`"` + common.FIELD_TYPE + `":` + strconv.Itoa(n.Type) + `,` +
		`"` + common.FIELD_KEY + `":"` + common.UnescapeForJSON(n.Key) + `",` +
		`"` + common.FIELD_SUB_KEY + `":` + strconv.Itoa(n.SubKey) + `,` +
		`"` + common.FIELD_UPDATE_IP + `":"` + common.UnescapeForJSON(n.UpdateIP) + `",` +
		`"` + common.FIELD_UPDATE_TIME + `":` + strconv.Itoa(n.UpdateTime*1000) + `,` +
		`"` + common.FIELD_BODY + `":"` + common.UnescapeForJSON(n.Body) + `"`
}

//----------------------------------------------------------------------------

type NoteMap map[string]string

func (nm *NoteMap) ToJSON() string {
	r := `"` + common.FIELD_NOTE + `":[`
	first := true
	for _, s := range *nm {
		if first {
			first = false
		} else {
			r += `,`
		}

		ns := strings.Replace(s, "\r", "\\r", -1)
		ns = strings.Replace(ns, "\n", "\\n", -1)
		r += `{` + ns + `}`
	}
	r += `]`

	return r
}

//----------------------------------------------------------------------------

type NoteTask struct {
	isAdd bool
	n     *Note
}

//----------------------------------------------------------------------------
