package service

import (
	// "fmt"
	"github.com/wangmhgo/go-project/gdun/common"
	"strconv"
)

//----------------------------------------------------------------------------

type SubjectService struct {
	db    *common.Database
	cache *common.Cache
}

func NewSubjectService(db *common.Database, cache *common.Cache) (*SubjectService, error) {
	ss := new(SubjectService)
	ss.db = db
	ss.cache = cache

	if err := ss.Init(); err != nil {
		return ss, err
	}

	return ss, nil
}

//----------------------------------------------------------------------------

/*
SBJ:id
	{id}		{name}

G:{groupID}
	subjectList	{JSON of subjects}
*/

func (ss *SubjectService) Init() error {
	s := "CREATE TABLE IF NOT EXISTS `" + common.TABLE_SUBJECT + "` (" +
		"`" + common.FIELD_ID + "` INT NOT NULL AUTO_INCREMENT," +
		"`" + common.FIELD_NAME + "` VARCHAR(512) NOT NULL DEFAULT ''," +
		"`" + common.FIELD_UPDATE_IP + "` VARCHAR(32) NOT NULL," +
		"`" + common.FIELD_UPDATE_TIME + "` BIGINT NOT NULL," +
		"`" + common.FIELD_UPDATER + "` INT NOT NULL," +
		"PRIMARY KEY (`" + common.FIELD_ID + "`)" +
		") ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;"

	if _, err := ss.db.Exec(s); err != nil {
		return err
	}

	s = "CREATE TABLE IF NOT EXISTS `" + common.TABLE_GROUP_SUBJECT + "` (" +
		"`" + common.FIELD_GROUP_ID + "` INT NOT NULL," +
		"`" + common.FIELD_SUBJECT_LIST + "` TEXT NOT NULL," + // Comma separated integer list.
		"`" + common.FIELD_UPDATE_IP + "` VARCHAR(32) NOT NULL," +
		"`" + common.FIELD_UPDATE_TIME + "` BIGINT NOT NULL," +
		"`" + common.FIELD_UPDATER + "` INT NOT NULL," +
		"PRIMARY KEY (`" + common.FIELD_GROUP_ID + "`)" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8;"

	if _, err := ss.db.Exec(s); err != nil {
		return err
	}

	return nil
}

//----------------------------------------------------------------------------

func (ss *SubjectService) Preload() (int, int, error) {
	if ss.db == nil {
		return 0, 0, common.ERR_NO_DATABASE
	}
	if ss.cache == nil {
		return 0, 0, common.ERR_NO_CACHE
	}

	n1, err := ss.preloadSubjects()
	if err != nil {
		return 0, 0, err
	}

	n2, err := ss.preloadGroupedSubjects()
	if err != nil {
		return n1, 0, err
	}

	return n1, n2, nil
}

func (ss *SubjectService) preloadSubjects() (int, error) {
	s := "SELECT " +
		common.FIELD_ID + "," +
		common.FIELD_NAME +
		" FROM " +
		common.TABLE_SUBJECT + ";"

	rows, err := ss.db.Select(s)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	key := common.KEY_PREFIX_SUBJECT + common.FIELD_ID
	cnt := 0

	id := 0
	name := ""
	for rows.Next() {
		if err = rows.Scan(&id, &name); err != nil {
			return cnt, err
		}

		if err = ss.cache.SetField(key, strconv.Itoa(id), common.UnescapeForJSON(name)); err != nil {
			return cnt, err
		}

		cnt++
	}

	return cnt, nil
}

func (ss *SubjectService) preloadGroupedSubjects() (int, error) {
	s := "SELECT " +
		common.FIELD_GROUP_ID + "," +
		common.FIELD_SUBJECT_LIST +
		" FROM " +
		common.TABLE_GROUP_SUBJECT + ";"

	rows, err := ss.db.Select(s)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	cnt := 0

	groupID := 0
	ls := ""
	for rows.Next() {
		if err = rows.Scan(&groupID, &ls); err != nil {
			return cnt, err
		}

		if _, err = ss.groupSubjectListToJSON(groupID, ls); err != nil {
			return cnt, err
		}

		cnt++
	}

	return cnt, nil
}

func (ss *SubjectService) groupSubjectListToJSON(groupID int, ls string) (string, error) {
	key := common.KEY_PREFIX_SUBJECT + common.FIELD_ID
	arr := common.StringToIntArray(ls)

	s := ``
	first := true
	for i := 0; i < len(arr); i++ {
		sID := strconv.Itoa(arr[i])

		name, err := ss.cache.GetField(key, sID)
		if err != nil {
			return "", err
		}

		if first {
			first = false
		} else {
			s += `,`
		}

		s += `"` + sID + `":"` + name + `"`
	}

	if ss.cache != nil {
		if err := ss.cache.SetField(common.KEY_PREFIX_GROUP+strconv.Itoa(groupID), common.FIELD_SUBJECT_LIST, s); err != nil {
			return s, err
		}
	}

	return s, nil
}

func (ss *SubjectService) reloadSubjects(groupID int) (string, error) {
	if ss.db == nil {
		return "", common.ERR_NO_SERVICE
	}

	sGroupID := strconv.Itoa(groupID)

	s := "SELECT " +
		common.FIELD_SUBJECT_LIST +
		" FROM " +
		common.TABLE_GROUP_SUBJECT +
		" WHERE " +
		common.FIELD_GROUP_ID + "=" + sGroupID + ";"

	rows, err := ss.db.Select(s)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	if !rows.Next() {
		return "", nil
	}

	ls := ""
	if err = rows.Scan(&ls); err != nil {
		return "", err
	}

	return ss.groupSubjectListToJSON(groupID, ls)
}

//----------------------------------------------------------------------------

func (ss *SubjectService) AddSubject(name string, session *Session) (int, error) {
	if ss.db == nil {
		return 0, common.ERR_NO_SERVICE
	}

	sName := common.Escape(name)
	if len(sName) == 0 {
		return 0, common.ERR_INVALID_NAME
	}

	sUpdater := strconv.Itoa(session.UserID)
	sUpdateIP := session.IP
	sUpdateTime := common.GetTimeString()

	s := "INSERT INTO " +
		common.TABLE_SUBJECT +
		"(" +
		common.FIELD_NAME + "," +
		common.FIELD_UPDATE_TIME + "," +
		common.FIELD_UPDATE_IP + "," +
		common.FIELD_UPDATER +
		") VALUES (" +
		"'" + sName + "'," +
		sUpdateTime + "," +
		"'" + sUpdateIP + "'," +
		sUpdater +
		");"

	id, err := ss.db.Insert(s, 1)
	if err != nil {
		return 0, err
	}

	if ss.cache != nil {
		if err = ss.cache.SetField(common.KEY_PREFIX_SUBJECT+common.FIELD_ID, strconv.FormatInt(id, 10), common.UnescapeForJSON(sName)); err != nil {
			return int(id), err
		}
	}

	return int(id), nil
}

//----------------------------------------------------------------------------

func (ss *SubjectService) ChangeSubject(subjectID int, name string, session *Session) error {
	if ss.db == nil {
		return common.ERR_NO_SERVICE
	}

	if subjectID < 1 {
		return common.ERR_INVALID_ID
	}
	sID := strconv.Itoa(subjectID)

	sName := common.Escape(name)
	if len(sName) == 0 {
		return common.ERR_INVALID_NAME
	}

	sUpdater := strconv.Itoa(session.UserID)
	sUpdateIP := session.IP
	sUpdateTime := common.GetTimeString()

	s := "UPDATE " +
		common.TABLE_SUBJECT +
		" SET " +
		common.FIELD_NAME + "='" + sName + "'," +
		common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
		common.FIELD_UPDATE_TIME + "=" + sUpdateTime + "," +
		common.FIELD_UPDATER + "=" + sUpdater +
		" WHERE " +
		common.FIELD_ID + "=" + sID + ";"

	if _, err := ss.db.Exec(s); err != nil {
		return err
	}

	if ss.cache != nil {
		if err := ss.cache.SetField(common.KEY_PREFIX_SUBJECT+common.FIELD_ID, sID, common.UnescapeForJSON(sName)); err != nil {
			return err
		}
	}

	return nil
}

//----------------------------------------------------------------------------

func (ss *SubjectService) GetSubject() (string, error) {
	if ss.cache != nil {
		s, err := (func() (string, error) {
			m, err := ss.cache.GetAllFields(common.KEY_PREFIX_SUBJECT + common.FIELD_ID)
			if err != nil {
				return "", err
			}

			s := ``
			first := true
			for sID, sName := range m {
				if first {
					first = false
				} else {
					s += `,`
				}
				s += `"` + sID + `":"` + sName + `"`
			}

			return s, nil
		})()
		if err == nil {
			return s, nil
		}
	}

	if ss.db != nil {
		s := "SELECT " +
			common.FIELD_ID + "," +
			common.FIELD_NAME +
			" FROM " +
			common.TABLE_SUBJECT + ";"

		rows, err := ss.db.Select(s)
		if err != nil {
			return "", err
		}
		defer rows.Close()

		key := common.KEY_PREFIX_SUBJECT + common.FIELD_ID

		r := ``
		first := true

		id := 0
		name := ""
		for rows.Next() {
			if err = rows.Scan(&id, &name); err != nil {
				return r, err
			}

			sID := strconv.Itoa(id)
			sName := common.UnescapeForJSON(name)

			if first {
				first = false
			} else {
				r += `,`
			}
			r += `"` + sID + `":"` + sName + `"`

			if ss.cache != nil {
				if err = ss.cache.SetField(key, sID, sName); err != nil {
					return r, err
				}
			}
		}

		return r, nil
	}

	return "", common.ERR_NO_SERVICE
}

//----------------------------------------------------------------------------

func (ss *SubjectService) ChangeSubjectList(groupID int, subjectID int, isAdd bool, session *Session) error {
	if ss.db == nil {
		return common.ERR_NO_SERVICE
	}

	sGroupID := strconv.Itoa(groupID)
	sSubjectID := strconv.Itoa(subjectID)

	sUpdater := strconv.Itoa(session.UserID)
	sUpdateIP := session.IP
	sUpdateTime := common.GetTimeString()

	existing, ls, err := (func() (bool, string, error) {
		s := "SELECT " +
			common.FIELD_SUBJECT_LIST +
			" FROM " +
			common.TABLE_GROUP_SUBJECT +
			" WHERE " +
			common.FIELD_GROUP_ID + "=" + sGroupID + ";"

		rows, err := ss.db.Select(s)
		if err != nil {
			// Encounter an error.
			return false, "", err
		}
		defer rows.Close()

		if !rows.Next() {
			// Does not exist.
			return false, "", nil
		}

		r := ""
		if err = rows.Scan(&r); err != nil {
			// Exists, but encounter an error.
			return true, "", err
		}

		return true, r, nil
	})()
	if err != nil {
		return err
	}

	// fmt.Println(ls)

	if existing {
		changed := false
		if isAdd {
			ls, changed = common.AddToList(sSubjectID, ls)
		} else {
			ls, changed = common.DeleteFromList(sSubjectID, ls)
		}
		// fmt.Println(ls)
		if !changed {
			// fmt.Println("!changed")
			return nil
		}

		// Append this subject ID to the existing list.
		err = (func() error {
			s := "UPDATE " +
				common.TABLE_GROUP_SUBJECT +
				" SET " +
				common.FIELD_SUBJECT_LIST + "='" + ls + "'," +
				common.FIELD_UPDATER + "=" + sUpdater + "," +
				common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
				common.FIELD_UPDATE_TIME + "=" + sUpdateTime +
				" WHERE " +
				common.FIELD_GROUP_ID + "=" + sGroupID + ";"

			// fmt.Println(s)

			if _, err := ss.db.Exec(s); err != nil {
				return err
			}
			return nil
		})()
	} else {
		if isAdd {
			// Insert a new record.
			err = (func() error {
				s := "INSERT INTO " +
					common.TABLE_GROUP_SUBJECT +
					"(" +
					common.FIELD_GROUP_ID + "," +
					common.FIELD_SUBJECT_LIST + "," +
					common.FIELD_UPDATER + "," +
					common.FIELD_UPDATE_IP + "," +
					common.FIELD_UPDATE_TIME +
					") VALUES (" +
					sGroupID + "," +
					"'" + sSubjectID + "'," +
					sUpdater + "," +
					"'" + sUpdateIP + "'," +
					sUpdateTime +
					");"

				// fmt.Println(s)

				if _, err := ss.db.Exec(s); err != nil {
					return err
				}
				return nil
			})()
		} else {
			return nil
		}
	}
	if err != nil {
		return err
	}

	if _, err = ss.reloadSubjects(groupID); err != nil {
		return err
	}

	return nil
}

//----------------------------------------------------------------------------

func (ss *SubjectService) GetSubjectList(groupID int) (string, error) {
	sGroupID := strconv.Itoa(groupID)

	if ss.cache != nil {
		if s, err := ss.cache.GetField(common.KEY_PREFIX_GROUP+sGroupID, common.FIELD_SUBJECT_LIST); err == nil {
			return s, nil
		}
	}

	if ss.db != nil {
		return ss.reloadSubjects(groupID)
	}

	return "", common.ERR_NO_SERVICE
}

//----------------------------------------------------------------------------
