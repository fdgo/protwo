package service

import (
	"errors"
	"github.com/wangmhgo/go-project/gdun/common"
	"strconv"
	"strings"
)

//----------------------------------------------------------------------------
// Database: Required.
// Cache   : Compatible.

func (ms *MeetingService) AddPDF(filename string, id string, name string, meetingID int, preparation int, necessary int, session *Session) (string, error) {
	// Check requirements.
	if ms.db == nil || ms.ts == nil {
		return "", common.ERR_NO_SERVICE
	}

	// Check authority.
	mi, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return "", err
	}

	key := common.Escape(id)
	if len(key) == 0 {
		// Convert uploaded PDF file to PNGs.
		key, err = ms.ts.AddFile(filename, PDF, session.IP, session.UserID)
		if err != nil {
			return "", err
		}
	}

	value := common.Escape(name)
	if preparation == 0 {
		value += "_0"
	} else {
		value += "_1"
	}
	if necessary == 0 {
		value += "_0"
	} else {
		value += "_1"
	}
	target := key + ":" + value

	existing := false
	for i := 0; i < len(mi.Coursewares); i++ {
		if mi.Coursewares[i] == target {
			existing = true
			break
		}
	}
	if existing {
		return key, nil
	}

	sMeetingID := strconv.Itoa(meetingID)

	sUpdateIP := session.IP
	sUpdateTime := common.GetTimeString()
	sUpdater := strconv.Itoa(session.UserID)

	if err = (func() error {
		tx, err := ms.db.Transaction()
		if err != nil {
			return err
		}

		sql := "SELECT " +
			common.FIELD_COURSEWARE_LIST +
			" FROM " +
			common.TABLE_MEETING +
			" WHERE " +
			common.FIELD_ID + "=" + sMeetingID +
			" FOR UPDATE;"

		row := tx.QueryRow(sql)

		cwl := ""
		err = row.Scan(&cwl)
		if err != nil {
			tx.Rollback()
			return err
		}

		// Update the courseware list.
		changed := false
		cwl, changed = common.AddResourceToMap(key, value, cwl)
		if !changed {
			tx.Rollback()

			// Update cache.
			if ms.cache != nil {
				err = ms.cache.SetField(common.KEY_PREFIX_MEETING+sMeetingID, common.FIELD_COURSEWARE_LIST, cwl)
				if err != nil {
					return err
				}
			}
			return nil
		}

		sql = "UPDATE " +
			common.TABLE_MEETING +
			" SET " +
			common.FIELD_COURSEWARE_LIST + "='" + cwl + "'," +
			common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
			common.FIELD_UPDATE_TIME + "=" + sUpdateTime + "," +
			common.FIELD_UPDATER + "=" + sUpdater +
			" WHERE " +
			common.FIELD_ID + "=" + sMeetingID + ";"

		_, err = tx.Exec(sql)
		if err != nil {
			tx.Rollback()
			return err
		}

		// Update cache.
		if ms.cache != nil {
			m := make(map[string]string)
			m[common.FIELD_COURSEWARE_LIST] = cwl
			m[common.FIELD_UPDATE_IP] = sUpdateIP
			m[common.FIELD_UPDATE_TIME] = sUpdateTime
			m[common.FIELD_UPDATER] = sUpdater

			err = ms.cache.SetFields(common.KEY_PREFIX_MEETING+sMeetingID, m)
			if err != nil {
				tx.Rollback()
				return err
			}
		}

		// Commit this transaction.
		if err = tx.Commit(); err != nil {
			tx.Rollback()
			return err
		}

		return nil
	})(); err != nil {
		return "", err
	}

	return key, nil
}

//----------------------------------------------------------------------------
// Database: Required.
// Cache   : Compatible.

func (ms *MeetingService) DeleteCourseware(coursewareID string, meetingID int, session *Session) error {
	// Check requirements.
	if ms.db == nil {
		return common.ERR_NO_SERVICE
	}

	// Check inputs.
	if len(coursewareID) == 0 {
		return errors.New("Empty courseware ID.")
	}
	sCoursewareID := common.EscapeForStr64(coursewareID)

	// Check authority.
	mi, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return err
	}

	// Check via cache.
	existing := false
	for i := 0; i < len(mi.Coursewares); i++ {
		if strings.HasPrefix(mi.Coursewares[i], sCoursewareID+":") {
			existing = true
			break
		}
	}
	if !existing {
		return common.ERR_NO_COURSEWARE
	}

	sMeetingID := strconv.Itoa(meetingID)

	sUpdateIP := session.IP
	sUpdateTime := common.GetTimeString()
	sUpdater := strconv.Itoa(session.UserID)

	if err = (func() error {
		// Create a transaction.
		tx, err := ms.db.Transaction()
		if err != nil {
			return err
		}

		// Check via database.
		sql := "SELECT " +
			common.FIELD_COURSEWARE_LIST +
			" FROM " +
			common.TABLE_MEETING +
			" WHERE " +
			common.FIELD_ID + "=" + sMeetingID +
			" FOR UPDATE;"

		row := tx.QueryRow(sql)

		cwl := ""
		err = row.Scan(&cwl)
		if err != nil {
			tx.Rollback()
			return err
		}

		// Check again.
		okay := false
		cwl, okay = common.DeleteFromMap(sCoursewareID, cwl)
		// cwl, okay = common.DeleteFromList(sCoursewareID, cwl)
		if !okay {
			tx.Rollback()
			return common.ERR_NO_COURSEWARE
		}

		sql = "UPDATE " +
			common.TABLE_MEETING +
			" SET " +
			common.FIELD_COURSEWARE_LIST + "='" + cwl + "'," +
			common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
			common.FIELD_UPDATE_TIME + "=" + sUpdateTime + "," +
			common.FIELD_UPDATER + "=" + sUpdater +
			" WHERE " +
			common.FIELD_ID + "=" + sMeetingID + ";"

		_, err = tx.Exec(sql)
		if err != nil {
			tx.Rollback()
			return err
		}

		// Update cache.
		if ms.cache != nil {
			m := make(map[string]string)
			m[common.FIELD_COURSEWARE_LIST] = cwl
			m[common.FIELD_UPDATE_IP] = sUpdateIP
			m[common.FIELD_UPDATE_TIME] = sUpdateTime
			m[common.FIELD_UPDATER] = sUpdater

			err = ms.cache.SetFields(common.KEY_PREFIX_MEETING+sMeetingID, m)
			if err != nil {
				tx.Rollback()
				return err
			}
		}

		// Commit this transaction.
		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			return err
		}

		return nil
	})(); err != nil {
		return err
	}

	return nil
}

//----------------------------------------------------------------------------
// Database: Required.
// Cache   : Compatible.

// func (ms *MeetingService) AddZip(filename string, meetingID int, preparation int, necessary int, session *Session) error {
// 	r, err := zip.OpenReader(filename)
// 	if err != nil {
// 		return err
// 	}
// 	defer r.Close()

// 	for i := 0; i < len(r.File); i++ {
// 		if err = (func(f *zip.File) error {
// 			// Ignore directories and empty files.
// 			info := f.FileInfo()
// 			if info.IsDir() || info.Size() == 0 {
// 				return nil
// 			}

// 			// Open this file.
// 			r, err := f.Open()
// 			if err != nil {
// 				return err
// 			}
// 			defer r.Close()

// 			// Load its content.
// 			buf, err := ioutil.ReadAll(r)
// 			if err != nil {
// 				return err
// 			}

// 			// Save it to another file.
// 			outFilename := ""
// 			ms.GetUUID()

// 			_, err = ms.AddPDF(outFilename, "", f.Name, meetingID, preparation, necessary, session)
// 			if err != nil {
// 				return err
// 			}

// 			return nil
// 		})(r.File[i]); err != nil {
// 			fmt.Println(err.Error())
// 			break
// 		}
// 	}

// 	return nil
// }

//----------------------------------------------------------------------------
