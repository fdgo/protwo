package service

import (
	"github.com/wangmhgo/go-project/gdun/common"
	"strconv"
)

//----------------------------------------------------------------------------
// Database: Required.
// Cache   : Compatible.

func (ms *MeetingService) AddReplay(videoID string, meetingID int, session *Session) error {
	// Check requirements.
	if ms.db == nil {
		return common.ERR_NO_SERVICE
	}

	// Check inputs.
	if len(videoID) == 0 {
		return common.ERR_INVALID_VIDEO
	}

	sVideoID := ms.getVideo(videoID)
	if sVideoID == "" {
		sVideoID = videoID
	}
	sVideoID = common.EscapeForStr64(sVideoID)

	// Check authority.
	mi, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return err
	}

	// Check via cache.
	existing := false
	for i := 0; i < len(mi.Replays); i++ {
		if mi.Replays[i] == sVideoID {
			existing = true
			break
		}
	}
	if existing {
		return nil
	}

	sMeetingID := strconv.Itoa(meetingID)

	sUpdateTime := common.GetTimeString()
	sUpdateIP := session.IP
	sUpdater := strconv.Itoa(session.UserID)

	err = (func() error {
		// Create a transaction.
		tx, err := ms.db.Transaction()
		if err != nil {
			return err
		}

		// Get existing value.
		sql := "SELECT " +
			common.FIELD_REPLAY_LIST +
			" FROM " +
			common.TABLE_MEETING +
			" WHERE " +
			common.FIELD_ID + "=" + sMeetingID +
			" FOR UPDATE;"
		row := tx.QueryRow(sql)

		rl := ""
		err = row.Scan(&rl)
		if err != nil {
			tx.Rollback()
			return err
		}

		// Compute new value.
		rl, _ = common.AddToList(sVideoID, rl)

		sql = "UPDATE " +
			common.TABLE_MEETING +
			" SET " +
			common.FIELD_REPLAY_LIST + "='" + rl + "'," +
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

			m[common.FIELD_REPLAY_LIST] = rl
			m[common.FIELD_UPDATE_IP] = sUpdateIP
			m[common.FIELD_UPDATE_TIME] = sUpdateTime
			m[common.FIELD_UPDATER] = sUpdater

			err = ms.cache.SetFields(common.KEY_PREFIX_MEETING+sMeetingID, m)
			if err != nil {
				tx.Rollback()
				return err
			}
		}

		// Commit it.
		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			return err
		}

		return nil
	})()

	return err
}

//----------------------------------------------------------------------------
// Database: Required.
// Cache   : Compatible.

func (ms *MeetingService) DeleteReplay(videoID string, meetingID int, session *Session) error {
	// Check requirements.
	if ms.db == nil {
		return common.ERR_NO_SERVICE
	}

	// Check inputs.
	if len(videoID) == 0 {
		return common.ERR_INVALID_VIDEO
	}
	sVideoID := common.EscapeForStr64(videoID)

	// Check authority.
	mi, err := ms.GetMeeting(meetingID, session, true)
	if err != nil {
		return err
	}

	// Delete this video from the video list.
	existing := false
	for i := 0; i < len(mi.Replays); i++ {
		if mi.Replays[i] == sVideoID {
			existing = true
			break
		}
	}
	if !existing {
		return nil
	}

	sMeetingID := strconv.Itoa(meetingID)

	sUpdateTime := common.GetTimeString()
	sUpdateIP := session.IP
	sUpdater := strconv.Itoa(session.UserID)

	err = (func() error {
		// Create a transaction.
		tx, err := ms.db.Transaction()
		if err != nil {
			return err
		}

		// Get existing value.
		sql := "SELECT " +
			common.FIELD_REPLAY_LIST +
			" FROM " +
			common.TABLE_MEETING +
			" WHERE " +
			common.FIELD_ID + "=" + sMeetingID +
			" FOR UPDATE;"
		row := tx.QueryRow(sql)

		rl := ""
		err = row.Scan(&rl)
		if err != nil {
			tx.Rollback()
			return err
		}

		// Compute new value.
		rl, _ = common.DeleteFromList(sVideoID, rl)

		sql = "UPDATE " +
			common.TABLE_MEETING +
			" SET " +
			common.FIELD_REPLAY_LIST + "='" + rl + "'," +
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

			m[common.FIELD_REPLAY_LIST] = rl
			m[common.FIELD_UPDATE_IP] = sUpdateIP
			m[common.FIELD_UPDATE_TIME] = sUpdateTime
			m[common.FIELD_UPDATER] = sUpdater

			err = ms.cache.SetFields(common.KEY_PREFIX_MEETING+sMeetingID, m)
			if err != nil {
				tx.Rollback()
				return err
			}
		}

		// Commit it.
		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			return err
		}

		return nil
	})()

	return err
}

//----------------------------------------------------------------------------
