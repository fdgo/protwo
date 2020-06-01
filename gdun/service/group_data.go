package service

import (
	"github.com/wangmhgo/go-project/gdun/common"
	"strconv"
)

//----------------------------------------------------------------------------

type GroupInfo struct {
	ID          int
	Name        string
	InstituteID int
}

func (gi *GroupInfo) ToJSON(isSystem bool) string {
	r := `"` + common.FIELD_ID + `":` + strconv.Itoa(gi.ID) + `,` +
		`"` + common.FIELD_NAME + `":"` + common.UnescapeForJSON(gi.Name) + `"`

	if isSystem {
		r += `,"` + common.FIELD_INSTITUTE_ID + `":` + strconv.Itoa(gi.InstituteID)
	}

	return r
}

//----------------------------------------------------------------------------
