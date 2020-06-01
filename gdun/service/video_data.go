package service

import (
	"github.com/wangmhgo/go-project/gdun/common"
	"strconv"
)

type VideoAuthorizeInfo struct {
	ID       string
	AESKey   string
	AESIV    string
	Duration int
}

func NewVideoAuthorizeInfoFromMap(m map[string]string, id string) *VideoAuthorizeInfo {
	aesKey, okay := m[common.FIELD_KEY]
	if !okay {
		return nil
	}

	aesIV, okay := m[common.FIELD_IV]
	if !okay {
		return nil
	}

	var err error
	duration := 0
	s, okay := m[common.FIELD_DURATION]
	if okay {
		duration, err = strconv.Atoi(s)
		if err != nil {
			return nil
		}
	}

	vai := new(VideoAuthorizeInfo)
	vai.ID = id
	vai.AESKey = aesKey
	vai.AESIV = aesIV
	vai.Duration = duration

	return vai
}

//----------------------------------------------------------------------------

func (vai *VideoAuthorizeInfo) ToJSON() string {
	r := `"` + common.FIELD_ID + `":"` + vai.ID + `",` +
		`"` + common.FIELD_KEY + `":"` + vai.AESKey + `",` +
		`"` + common.FIELD_IV + `":"` + vai.AESIV + `",` +
		`"` + common.FIELD_DURATION + `":` + strconv.Itoa(vai.Duration)

	return r
}

//----------------------------------------------------------------------------

type VideoInfo struct {
	ID         string
	Title      string
	Duration   int
	Width      int
	Height     int
	Encryption int
	UpdateTime int
}

//----------------------------------------------------------------------------

func (vi *VideoInfo) ToJSON() string {
	r := `"` + common.FIELD_ID + `":"` + vi.ID + `",` +
		`"` + common.FIELD_TITLE + `":"` + common.ReplaceForJSON(vi.Title) + `",` +
		`"` + common.FIELD_DURATION + `":` + strconv.Itoa(vi.Duration) + `,` +
		`"` + common.FIELD_WIDTH + `":` + strconv.Itoa(vi.Width) + `,` +
		`"` + common.FIELD_HEIGHT + `":` + strconv.Itoa(vi.Height) + `,` +
		`"` + common.FIELD_ENCRYPTION + `":` + strconv.Itoa(vi.Encryption) + `,` +
		`"` + common.FIELD_UPDATE_TIME + `":` + strconv.Itoa(vi.UpdateTime)
	return r
}

//----------------------------------------------------------------------------
