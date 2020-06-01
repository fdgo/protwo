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

func (ms *MeetingService) getVideo(s string) string {
	n := len(s)
	if n == 0 {
		return ""
	}

	target := "player/loader.js?"
	pos := strings.Index(s, target)
	if pos < 0 {
		return ""
	}

	pos += len(target)
	i := pos
	for i < n {
		if s[i] == '-' || s[i] == '"' {
			return s[pos:i]
		}
		i++
	}

	return ""
}

func (ms *MeetingService) AddVideo(videoID string, videoName string, meetingID int, preparation int, necessary int, session *Session) error {
	// Check requirements.
	if ms.db == nil {
		return common.ERR_NO_SERVICE
	}

	// Check inputs.
	if len(videoID) == 0 {
		return common.ERR_INVALID_VIDEO
	}

	// Check authority.
	mi, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return err
	}

	key := ms.getVideo(videoID)
	if key == "" {
		key = videoID
	}
	key = common.EscapeForStr64(key)

	value := common.Escape(videoName) + "_" + strconv.Itoa(preparation)
	if necessary == 0 {
		value += "_0"
	} else {
		value += "_1"
	}

	target := key + ":" + value

	// Check whether this video resides in the meeting via cache.
	existing := false
	for i := 0; i < len(mi.Videos); i++ {
		if mi.Videos[i] == target {
			existing = true
			break
		}
	}
	if existing {
		return nil
	}

	sMeetingID := strconv.Itoa(meetingID)

	sUpdateTime := common.GetTimeString()
	sUpdater := strconv.Itoa(session.UserID)
	sUpdateIP := session.IP

	if err := (func() error {
		// Create a transaction.
		tx, err := ms.db.Transaction()
		if err != nil {
			return err
		}

		sql := "SELECT " +
			common.FIELD_VIDEO_LIST +
			" FROM " +
			common.TABLE_MEETING +
			" WHERE " +
			common.FIELD_ID + "=" + sMeetingID +
			" FOR UPDATE;"

		row := tx.QueryRow(sql)

		vl := ""
		err = row.Scan(&vl)
		if err != nil {
			tx.Rollback()
			return err
		}

		// Check again.
		changed := false
		vl, changed = common.AddResourceToMap(key, value, vl)
		if !changed {
			tx.Rollback()

			if ms.cache != nil {
				err = ms.cache.SetField(common.KEY_PREFIX_MEETING+sMeetingID, common.FIELD_VIDEO_LIST, vl)
				if err != nil {
					return err
				}
			}
			return nil
		}

		// Update database.
		sql = "UPDATE " +
			common.TABLE_MEETING +
			" SET " +
			common.FIELD_VIDEO_LIST + "='" + vl + "'," +
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

			m[common.FIELD_VIDEO_LIST] = vl
			m[common.FIELD_UPDATE_IP] = sUpdateIP
			m[common.FIELD_UPDATE_TIME] = sUpdateTime
			m[common.FIELD_UPDATER] = sUpdater

			err = ms.cache.SetFields(common.KEY_PREFIX_MEETING+sMeetingID, m)
			if err != nil {
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

func (ms *MeetingService) DeleteVideo(videoID string, meetingID int, session *Session) error {
	// Check requirements.
	if ms.db == nil {
		return common.ERR_NO_SERVICE
	}

	// Check inputs.
	if len(videoID) == 0 {
		return errors.New("Empty video ID.")
	}
	sVideoID := common.EscapeForStr64(videoID)

	// Check authority.
	mi, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return err
	}

	// Check via cache.
	existing := false
	for i := 0; i < len(mi.Videos); i++ {
		if strings.HasPrefix(mi.Videos[i], sVideoID+":") {
			existing = true
			break
		}
	}
	if !existing {
		return nil
	}

	sMeetingID := strconv.Itoa(meetingID)
	sUpdater := strconv.Itoa(session.UserID)
	sUpdateTime := common.GetTimeString()
	sUpdateIP := session.IP

	if err := (func() error {
		// Create a new transaction.
		tx, err := ms.db.Transaction()
		if err != nil {
			return err
		}

		sql := "SELECT " +
			common.FIELD_VIDEO_LIST +
			" FROM " +
			common.TABLE_MEETING +
			" WHERE " +
			common.FIELD_ID + "=" + sMeetingID +
			" FOR UPDATE;"

		row := tx.QueryRow(sql)

		vl := ""
		err = row.Scan(&vl)
		if err != nil {
			tx.Rollback()
			return err
		}

		// Check again.
		okay := false
		vl, okay = common.DeleteFromMap(sVideoID, vl)
		if !okay {
			tx.Rollback()

			if ms.cache != nil {
				err = ms.cache.SetField(common.KEY_PREFIX_MEETING+sMeetingID, common.FIELD_VIDEO_LIST, vl)
				if err != nil {
					return err
				}
			}
			return nil
		}

		// Update database.
		sql = "UPDATE " +
			common.TABLE_MEETING +
			" SET " +
			common.FIELD_VIDEO_LIST + "='" + vl + "'," +
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

			m[common.FIELD_VIDEO_LIST] = vl
			m[common.FIELD_UPDATE_IP] = sUpdateIP
			m[common.FIELD_UPDATE_TIME] = sUpdateTime
			m[common.FIELD_UPDATER] = sUpdater

			err = ms.cache.SetFields(common.KEY_PREFIX_MEETING+sMeetingID, m)
			if err != nil {
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
