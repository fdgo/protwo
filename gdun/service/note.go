package service

import (
	"database/sql"
	"fmt"
	"github.com/wangmhgo/go-project/gdun/common"
	"strconv"
	"strings"
	"sync"
	"time"
)

//----------------------------------------------------------------------------

var noteCacheKeyPrefixs = []string{common.KEY_PREFIX_COURSEWARE, common.KEY_PREFIX_EXAM, common.KEY_PREFIX_MEETING, common.KEY_PREFIX_VIDEO}
var noteTableNames = []string{common.TABLE_COURSEWARE_NOTE, common.TABLE_EXAM_NOTE, common.TABLE_MEETING_NOTE, common.TABLE_VIDEO_NOTE}
var noteTableFields = []string{common.FIELD_COURSEWARE, common.FIELD_EXAM, common.FIELD_MEETING, common.FIELD_VIDEO}
var noteTableFieldTypes = []string{"VARCHAR(128)", "INT", "INT", "VARCHAR(64)"}
var noteTypeNum = 4

//----------------------------------------------------------------------------

type NoteService struct {
	db      *common.Database
	cache   *common.Cache
	cnt     int
	cntLock *sync.Mutex
	tasks   chan *NoteTask
}

func NewNoteService(db *common.Database, cache *common.Cache, backlog int) (*NoteService, error) {
	ns := new(NoteService)
	ns.db = db
	ns.cache = cache
	ns.cnt = 0
	ns.cntLock = new(sync.Mutex)

	err := ns.Init()
	if err != nil {
		return nil, err
	}

	ns.tasks = make(chan *NoteTask, backlog)
	go (func() {
		for {
			select {
			case task, okay := <-ns.tasks:
				if okay {
					if task.isAdd {
						// Save this note to database.
						if err := ns.addNoteToDatabase(task.n); err != nil {
							fmt.Println(err.Error())
						}
					} else {
						// Delete this note from database.
						if err := ns.deleteNoteFromDatabase(task.n); err != nil {
							fmt.Println(err.Error())
						}
					}
				}
			}
		}
	})()

	return ns, nil
}

//----------------------------------------------------------------------------

func (ns *NoteService) Init() error {
	if ns.db == nil {
		return common.ERR_NO_DATABASE
	}

	//----------------------------------------------------

	s := "CREATE TABLE IF NOT EXISTS `" + common.TABLE_USER_NOTE + "` ("
	s += " `" + common.FIELD_CLASS_ID + "` INT NOT NULL,"
	s += " `" + common.FIELD_USER_ID + "` INT NOT NULL,"
	s += " `" + common.FIELD_COURSEWARE + "` TEXT NOT NULL,"
	s += " `" + common.FIELD_EXAM + "` TEXT NOT NULL,"
	s += " `" + common.FIELD_MEETING + "` TEXT NOT NULL,"
	s += " `" + common.FIELD_VIDEO + "` TEXT NOT NULL,"
	s += " `" + common.FIELD_UPDATE_IP + "` VARCHAR(32) NOT NULL,"
	s += " `" + common.FIELD_UPDATE_TIME + "` BIGINT NOT NULL,"
	s += " KEY (`" + common.FIELD_CLASS_ID + "`),"
	s += " KEY (`" + common.FIELD_USER_ID + "`)"
	s += ") ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;"

	_, err := ns.db.Exec(s)
	if err != nil {
		return err
	}

	//----------------------------------------------------

	if err := (func() error {
		for i := 0; i < len(noteTableNames); i++ {
			s = "CREATE TABLE IF NOT EXISTS `" + noteTableNames[i] + "` ("
			s += " `" + noteTableFields[i] + "` " + noteTableFieldTypes[i] + " NOT NULL,"
			s += " `" + common.FIELD_BODY + "` TEXT NOT NULL,"
			s += " `" + common.FIELD_UPDATE_IP + "` VARCHAR(32) NOT NULL,"
			s += " `" + common.FIELD_UPDATE_TIME + "` BIGINT NOT NULL,"
			s += " `" + common.FIELD_UPDATER + "` INT NOT NULL,"
			s += " PRIMARY KEY (`" + noteTableFields[i] + "`)"
			s += ") ENGINE=InnoDB DEFAULT CHARSET=utf8;"

			_, err = ns.db.Exec(s)
			if err != nil {
				return err
			}
		}

		return nil

	})(); err != nil {
		return err
	}

	return nil
}

//----------------------------------------------------------------------------

func (ns *NoteService) Preload() (int, int, error) {
	// Check requirements.
	if ns.db == nil {
		return 0, 0, common.ERR_NO_DATABASE
	}
	if ns.cache == nil {
		return 0, 0, common.ERR_NO_CACHE
	}

	nUser, err := (func() (int, error) {
		rest, err := ns.db.Count(common.TABLE_USER_NOTE)
		if err != nil {
			return 0, err
		}

		valid := 0
		for i := 0; i < rest; {
			nLoaded, nRecord, err := ns.preloadUserClassNotes(i, common.DATABASE_PRELOAD_SIZE)
			if err != nil {
				return valid, err
			}
			valid += nLoaded
			i += nRecord
		}
		return valid, nil
	})()
	if err != nil {
		return nUser, 0, err
	}

	nTyped := 0
	for i := 0; i < noteTypeNum; i++ {
		n, err := (func() (int, error) {
			rest, err := ns.db.Count(noteTableNames[i])
			if err != nil {
				return 0, err
			}

			valid := 0
			for j := 0; j < rest; {
				nLoaded, nRecord, err := ns.preloadTypedNotes(i, j, common.DATABASE_PRELOAD_SIZE)
				if err != nil {
					return valid, err
				}
				valid += nLoaded
				j += nRecord
			}
			return valid, nil
		})()

		nTyped += n

		if err != nil {
			return nUser, nTyped, err
		}
	}

	return nUser, nTyped, nil
}

func (ns *NoteService) preloadUserClassNotes(start int, length int) (int, int, error) {
	s := "SELECT " +
		common.FIELD_CLASS_ID + "," +
		common.FIELD_USER_ID + "," +
		common.FIELD_COURSEWARE + "," +
		common.FIELD_EXAM + "," +
		common.FIELD_MEETING + "," +
		common.FIELD_VIDEO +
		" FROM " +
		common.TABLE_USER_NOTE +
		" LIMIT " + strconv.Itoa(start) + "," + strconv.Itoa(length) + ";"

	rows, err := ns.db.Select(s)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()

	all := 0
	cnt := 0

	classID := 0
	userID := 0
	l := []string{"", "", "", ""}

	for rows.Next() {
		all++

		if err = rows.Scan(&classID, &userID, &l[0], &l[1], &l[2], &l[3]); err != nil {
			return cnt, all, err
		}

		key := ns.getUserClassNoteKeyForCache(userID, classID)

		for i := 0; i < noteTypeNum; i++ {
			arr := strings.Split(l[i], ",")
			for j := 0; j < len(arr); j++ {
				if n := NewNoteFromString(arr[j]); n != nil {
					if err = ns.cache.SetField(key, strconv.Itoa(n.ID), n.ToJSON()); err != nil {
						return cnt, all, err
					}
				}
			}
		}
		cnt++
	}

	return cnt, all, nil
}

func (ns *NoteService) preloadTypedNotes(t int, start int, length int) (int, int, error) {
	s := "SELECT " +
		noteTableFields[t] + "," +
		common.FIELD_BODY + "," +
		common.FIELD_UPDATE_IP + "," +
		common.FIELD_UPDATE_TIME + "," +
		common.FIELD_UPDATER +
		" FROM " +
		noteTableNames[t] +
		" LIMIT " + strconv.Itoa(start) + "," + strconv.Itoa(length) + ";"

	rows, err := ns.db.Select(s)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()

	all := 0
	cnt := 0

	var raw interface{}
	body := ""
	updateIP := ""
	updateTime := 0
	Updater := 0

	for rows.Next() {
		all++

		if err = rows.Scan(&raw, &body, &updateIP, &updateTime, &Updater); err != nil {
			return cnt, all, err
		}

		k := ""
		if noteTableFieldTypes[t] == "INT" {
			id, okay := raw.(int)
			if !okay {
				continue
			}
			k = strconv.Itoa(id)
		} else {
			s, okay := raw.(string)
			if !okay {
				continue
			}
			k = s
		}
		key := ns.getTypedNoteKeyForCache(t, k)

		arr := strings.Split(body, ",")
		for i := 0; i < len(arr); i++ {
			if n := NewNoteFromString(arr[i]); n != nil {
				if err = ns.cache.SetField(key, strconv.Itoa(n.ID), n.ToJSON()); err != nil {
					return cnt, all, err
				}
			}
		}
		cnt++
	}

	return cnt, all, nil
}

//----------------------------------------------------------------------------

func (ns *NoteService) getUserClassNoteKeyForCache(userID int, classID int) string {
	return common.KEY_PREFIX_NOTE +
		common.KEY_PREFIX_USER + strconv.Itoa(userID) + ":" +
		common.KEY_PREFIX_CLASS + strconv.Itoa(classID)
}

func (ns *NoteService) getTypedNoteKeyForCache(t int, k string) string {
	if t < 1 || t > noteTypeNum {
		return ""
	}

	return common.KEY_PREFIX_NOTE + noteCacheKeyPrefixs[t-1] + k
}

//----------------------------------------------------------------------------

func (ns *NoteService) assignNoteID() int {
	if ns.cache != nil {
		if id, err := ns.cache.Incr(common.COUNTER_NOTE); err == nil {
			return int(id)
		}
	}

	ns.cntLock.Lock()
	defer ns.cntLock.Unlock()

	id := ns.cnt
	ns.cnt++

	return id
}

//----------------------------------------------------------------------------

func (ns *NoteService) newNote(id int, classID int, meetingID int, t int, key string, subKey int, body string, session *Session) *Note {
	n := new(Note)

	if id <= 0 {
		n.ID = ns.assignNoteID()
	} else {
		n.ID = id
	}

	n.ClassID = classID
	n.MeetingID = meetingID
	n.UserID = session.UserID

	n.Type = t
	if t == common.TYPE_FOR_VIDEO {
		n.Key = common.EscapeForStr64(key)
	} else {
		n.Key = common.Escape(key)
	}

	n.SubKey = subKey

	n.Body = common.Escape(strings.TrimSpace(body))

	n.UpdateIP = session.IP
	n.UpdateTime = int(time.Now().Unix())

	return n
}

//----------------------------------------------------------------------------
// Database: Compatible.
// Cache   : Required.

func (ns *NoteService) AddNote(classID int, meetingID int, t int, key string, subKey int, body string, session *Session) (int, error) {
	if ns.cache == nil {
		return 0, common.ERR_NO_SERVICE
	}

	n := ns.newNote(0, classID, meetingID, t, key, subKey, body, session)
	if n == nil {
		return 0, common.ERR_INVALID_NOTE
	}

	sID := strconv.Itoa(n.ID)
	sContent := n.ToJSON()

	if err := (func() error {
		// Add it to user-class notes.
		if err := ns.cache.SetField(ns.getUserClassNoteKeyForCache(n.UserID, n.ClassID), sID, sContent); err != nil {
			return err
		}

		// Add it to typed notes.
		if err := ns.cache.SetField(ns.getTypedNoteKeyForCache(n.Type, n.Key), sID, sContent); err != nil {
			return err
		}

		return nil
	})(); err != nil {
		return 0, err
	}

	// Update database, respectively.
	if ns.db != nil {
		// Put it into the task queue.
		ns.tasks <- &NoteTask{true, n}
	}

	return n.ID, nil
}

//----------------------------------------------------------------------------
// Database: Nope.
// Cache   : Required.

func (ns *NoteService) GetMyNote(classID int, session *Session) (*NoteMap, error) {
	if ns.cache == nil {
		return nil, common.ERR_NO_SERVICE
	}

	m, err := ns.cache.GetAllFields(ns.getUserClassNoteKeyForCache(session.UserID, classID))
	if err != nil {
		// Let the result be an empty map.
		m = make(map[string]string)
	}

	return (*NoteMap)(&m), nil
}

//----------------------------------------------------------------------------
// Database: Nope.
// Cache   : Required.

func (ns *NoteService) GetTypedNote(t int, key string, session *Session) (*NoteMap, error) {
	if ns.cache == nil {
		return nil, common.ERR_NO_SERVICE
	}

	m, err := ns.cache.GetAllFields(ns.getTypedNoteKeyForCache(t, key))
	if err != nil {
		// Let the result be an empty map.
		m = make(map[string]string)
	}

	return (*NoteMap)(&m), nil
}

//----------------------------------------------------------------------------
// Database: Compatible.
// Cache   : Required.

func (ns *NoteService) DeleteNote(id int, classID int, meetingID int, t int, key string, session *Session) error {
	if ns.cache == nil {
		return common.ERR_NO_SERVICE
	}

	sID := strconv.Itoa(id)

	if err := ns.cache.DelField(ns.getUserClassNoteKeyForCache(session.UserID, classID), sID); err != nil {
		// TODO:
	}

	if err := ns.cache.DelField(ns.getTypedNoteKeyForCache(t, key), sID); err != nil {
		// TODO:
	}

	// Update database, respectively.
	if ns.db != nil {
		if n := ns.newNote(id, classID, meetingID, t, key, 0, "", session); n != nil {
			// Put it into the task queue.
			ns.tasks <- &NoteTask{false, n}
		}
	}

	return nil
}

//----------------------------------------------------------------------------
// Database: Required.

func (ns *NoteService) addNoteToDatabase(n *Note) error {
	// Check requirements.
	if ns.db == nil {
		return common.ERR_NO_SERVICE
	}

	// Check inputs.
	if n == nil {
		return common.ERR_INVALID_NOTE
	}

	index := n.Type - 1
	if index < 0 || index >= noteTypeNum {
		return common.ERR_INVALID_TYPE
	}

	sTable := noteTableNames[index]
	sField := noteTableFields[index]
	sUserID := strconv.Itoa(n.UserID)
	sClassID := strconv.Itoa(n.ClassID)
	sContent := n.ToString()
	sUpdateTime := strconv.Itoa(n.UpdateTime)

	// Append it to user-class notes.
	if err := (func() error {
		// Create a new transaction.
		tx, err := ns.db.Transaction()
		if err != nil {
			return err
		}

		// Get existing value.
		s := "SELECT " +
			sField +
			" FROM " +
			common.TABLE_USER_NOTE +
			" WHERE " +
			common.FIELD_USER_ID + "=" + sUserID + " AND " +
			common.FIELD_CLASS_ID + "=" + sClassID + " FOR UPDATE;"

		row := tx.QueryRow(s)

		l := ""
		existing := true
		if err = row.Scan(&l); err != nil {
			if err == sql.ErrNoRows {
				existing = false
			} else {
				tx.Rollback()
				return err
			}
		}

		// Update database.
		if existing {
			s = "UPDATE " +
				common.TABLE_USER_NOTE +
				" SET " +
				sField + "='"
			if len(l) > 0 {
				s += l + ","
			}
			s += sContent + "'," +
				common.FIELD_UPDATE_IP + "='" + n.UpdateIP + "'," +
				common.FIELD_UPDATE_TIME + "=" + sUpdateTime +
				" WHERE " +
				common.FIELD_USER_ID + "=" + sUserID + " AND " +
				common.FIELD_CLASS_ID + "=" + sClassID + ";"

			if _, err = tx.Exec(s); err != nil {
				tx.Rollback()
				return err
			}
		} else {
			s = "INSERT INTO " +
				common.TABLE_USER_NOTE +
				" (" +
				common.FIELD_USER_ID + "," +
				common.FIELD_CLASS_ID + "," +
				common.FIELD_COURSEWARE + "," +
				common.FIELD_EXAM + "," +
				common.FIELD_MEETING + "," +
				common.FIELD_VIDEO + "," +
				common.FIELD_UPDATE_IP + "," +
				common.FIELD_UPDATE_TIME +
				") VALUES (" +
				sUserID + "," +
				sClassID + ","

			for i := 0; i < noteTypeNum; i++ {
				if i == index {
					s += "'" + sContent + "',"
				} else {
					s += "'',"
				}
			}

			s += "'" + n.UpdateIP + "'," +
				sUpdateTime +
				");"

			if _, err = tx.Exec(s); err != nil {
				tx.Rollback()
				return err
			}
		}

		// Commit it.
		if err = tx.Commit(); err != nil {
			tx.Rollback()
			return err
		}

		return nil
	})(); err != nil {
		return err
	}

	// Append it to typed noted.
	if err := (func() error {
		// Create a new transaction.
		tx, err := ns.db.Transaction()
		if err != nil {
			return err
		}

		// Get existing values.
		s := "SELECT " +
			common.FIELD_BODY +
			" FROM " +
			sTable +
			" WHERE " +
			sField + "="
		if noteTableFieldTypes[index] == "INT" {
			s += n.Key
		} else {
			s += "'" + n.Key + "'"
		}
		s += " FOR UPDATE;"

		row := tx.QueryRow(s)

		existing := true
		l := ""
		if err = row.Scan(&l); err != nil {
			if err == sql.ErrNoRows {
				existing = false
			} else {
				tx.Rollback()
				return err
			}
		}

		if existing {
			s = "UPDATE " +
				sTable +
				" SET " +
				common.FIELD_BODY + "='"
			if len(l) > 0 {
				s += l + ","
			}
			s += sContent + "'," +
				common.FIELD_UPDATE_IP + "='" + n.UpdateIP + "'," +
				common.FIELD_UPDATE_TIME + "=" + sUpdateTime + "," +
				common.FIELD_UPDATER + "=" + sUserID +
				" WHERE " +
				sField + "="
			if noteTableFieldTypes[index] == "INT" {
				s += n.Key
			} else {
				s += "'" + n.Key + "'"
			}
			s += ";"

			if _, err = tx.Exec(s); err != nil {
				tx.Rollback()
				return err
			}
		} else {
			s = "INSERT INTO " +
				sTable +
				" (" +
				sField + "," +
				common.FIELD_BODY + "," +
				common.FIELD_UPDATE_TIME + "," +
				common.FIELD_UPDATE_IP + "," +
				common.FIELD_UPDATER +
				") VALUES ("
			if noteTableFieldTypes[index] == "INT" {
				s += n.Key + ","
			} else {
				s += "'" + n.Key + "',"
			}
			s += "'" + sContent + "'," +
				sUpdateTime + "," +
				"'" + n.UpdateIP + "'," +
				sUserID +
				");"

			if _, err = tx.Exec(s); err != nil {
				tx.Rollback()
				return err
			}
		}

		// Commit it.
		if err = tx.Commit(); err != nil {
			tx.Rollback()
			return err
		}

		return nil
	})(); err != nil {
		return err
	}

	return nil
}

func (ns *NoteService) deleteNoteFromDatabase(n *Note) error {
	// Check requirements.
	if ns.db == nil {
		return common.ERR_NO_SERVICE
	}

	// Check inputs.
	if n == nil {
		return common.ERR_INVALID_NOTE
	}

	index := n.Type - 1
	if index < 0 || index >= noteTypeNum {
		return common.ERR_INVALID_TYPE
	}

	sTable := noteTableNames[index]
	sField := noteTableFields[index]
	sPrefix := n.GetStringPrefix()
	sUserID := strconv.Itoa(n.UserID)
	sClassID := strconv.Itoa(n.ClassID)
	sUpdateTime := strconv.Itoa(n.UpdateTime)

	// Delete it from user-class notes.
	if err := (func() error {
		// Create a new transaction.
		tx, err := ns.db.Transaction()
		if err != nil {
			return err
		}

		// Get existing values.

		s := "SELECT " +
			sField +
			" FROM " +
			common.TABLE_USER_NOTE +
			" WHERE " +
			common.FIELD_USER_ID + "=" + sUserID + " AND " +
			common.FIELD_CLASS_ID + "=" + sClassID + " FOR UPDATE;"

		row := tx.QueryRow(s)

		l := ""
		if err = row.Scan(&l); err != nil {
			tx.Rollback()

			if err == sql.ErrNoRows {
				return nil
			} else {
				return err
			}
		}

		// Compute new values.

		if len(l) == 0 {
			tx.Rollback()
			return nil
		}
		arr := strings.Split(s, ",")
		size := len(arr)
		if size == 0 {
			tx.Rollback()
			return nil
		}

		l = ""
		found := false
		first := true
		for i := 0; i < size; i++ {
			if strings.HasPrefix(arr[i], sPrefix) {
				found = true
				continue
			}

			if first {
				first = false
			} else {
				l += ","
			}
			l += arr[i]
		}

		if !found {
			tx.Rollback()
			return nil
		}

		// Update database.
		s = "UPDATE " +
			common.TABLE_USER_NOTE +
			" SET " +
			sField + "='" + l + "'," +
			common.FIELD_UPDATE_IP + "='" + n.UpdateIP + "'," +
			common.FIELD_UPDATE_TIME + "=" + sUpdateTime +
			" WHERE " +
			common.FIELD_USER_ID + "=" + sUserID + " AND " +
			common.FIELD_CLASS_ID + "=" + sClassID + ";"

		if _, err = tx.Exec(s); err != nil {
			tx.Rollback()
			return err
		}

		// Commit it.
		if err = tx.Commit(); err != nil {
			tx.Rollback()
			return err
		}

		return nil
	})(); err != nil {
		return err
	}

	// Delete it from typed notes.
	if err := (func() error {
		// Create a new transaction.
		tx, err := ns.db.Transaction()
		if err != nil {
			return err
		}

		// Get existing values.

		s := "SELECT " +
			sField +
			" FROM " +
			sTable +
			" WHERE " +
			common.FIELD_KEY + "=" + n.Key + " FOR UPDATE;"

		row := tx.QueryRow(s)

		l := ""
		if err = row.Scan(&l); err != nil {
			tx.Rollback()

			if err == sql.ErrNoRows {
				return nil
			} else {
				return err
			}
		}

		// Compute new values.

		if len(l) == 0 {
			tx.Rollback()
			return nil
		}
		arr := strings.Split(s, ",")
		size := len(arr)
		if size == 0 {
			tx.Rollback()
			return nil
		}

		l = ""
		found := false
		first := true
		for i := 0; i < size; i++ {
			if strings.HasPrefix(arr[i], sPrefix) {
				found = true
				continue
			}

			if first {
				first = false
			} else {
				l += ","
			}
			l += arr[i]
		}

		if !found {
			tx.Rollback()
			return nil
		}

		// Update database.
		s = "UPDATE " +
			sTable +
			" SET " +
			sField + "='" + l + "'," +
			common.FIELD_UPDATE_IP + "='" + n.UpdateIP + "'," +
			common.FIELD_UPDATE_TIME + "=" + sUpdateTime +
			" WHERE " +
			common.FIELD_KEY + "=" + n.Key + ";"

		if _, err = tx.Exec(s); err != nil {
			tx.Rollback()
			return err
		}

		// Commit it.
		if err = tx.Commit(); err != nil {
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
