package service

import (
	"io"
	"io/ioutil"
	"strconv"
)

//----------------------------------------------------------------------------

func (cs *ClassService) SetCover(classID int, file io.Reader, session *Session) error {
	_, err := cs.GetClass(classID, session)
	if err != nil {
		return err
	}

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	err = cs.oss.UploadBuffer("class/cover/"+strconv.Itoa(classID), buf)
	if err != nil {
		return err
	}

	return nil
}

//----------------------------------------------------------------------------
