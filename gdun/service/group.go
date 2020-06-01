package service

import (
	"errors"
	"github.com/wangmhgo/go-project/gdun/common"
	"sort"
	"strconv"
)

//----------------------------------------------------------------------------

type GroupService struct {
	db          *common.Database
	cache       *common.Cache
	cachedQuery string
}

func NewGroupService(db *common.Database, cache *common.Cache) (*GroupService, error) {
	gs := new(GroupService)
	gs.db = db
	gs.cache = cache
	gs.cachedQuery = common.KEY_PREFIX_GROUP + common.FIELD_TEXT

	err := gs.Init()
	if err != nil {
		return nil, err
	}

	return gs, nil
}

//----------------------------------------------------------------------------

func (gs *GroupService) Init() error {
	if gs.db == nil {
		return common.ERR_NO_DATABASE
	}

	sql := "CREATE TABLE IF NOT EXISTS `" + common.TABLE_GROUP + "` ("
	sql += " `" + common.FIELD_ID + "` INT NOT NULL AUTO_INCREMENT,"
	sql += " `" + common.FIELD_NAME + "` VARCHAR(512) NOT NULL,"
	// sql += " `" + common.FIELD_INSTITUTE_ID + "` INT NOT NULL DEFAULT 0,"
	sql += " `" + common.FIELD_UPDATE_IP + "` VARCHAR(32) NOT NULL,"
	sql += " `" + common.FIELD_UPDATE_TIME + "` BIGINT NOT NULL,"
	sql += " `" + common.FIELD_UPDATER + "` INT NOT NULL,"
	sql += " PRIMARY KEY (`" + common.FIELD_ID + "`)"
	// sql += " KEY (`" + common.FIELD_INSTITUTE_ID + "`)"
	sql += ") ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;"

	_, err := gs.db.Exec(sql)
	return err
}

//----------------------------------------------------------------------------

func (gs *GroupService) Preload() (int, error) {
	if gs.db == nil {
		return 0, common.ERR_NO_DATABASE
	}
	if gs.cache == nil {
		return 0, common.ERR_NO_CACHE
	}

	sql := "SELECT " +
		common.FIELD_ID + "," +
		common.FIELD_NAME +
		// common.FIELD_INSTITUTE_ID +
		" FROM " +
		common.TABLE_GROUP +
		" WHERE " +
		common.FIELD_ID + ">4;"

	rows, err := gs.db.Select(sql)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	key := common.KEY_PREFIX_GROUP + common.FIELD_ID

	cnt := 0

	id := 0
	name := ""
	// instituteID := 0
	for rows.Next() {
		err = rows.Scan(&id, &name)
		if err != nil {
			return cnt, err
		}

		// TODO:

		err = gs.cache.SetField(key, strconv.Itoa(id), name)
		if err != nil {
			return cnt, err
		}
		cnt++
	}

	return cnt, nil
}

//----------------------------------------------------------------------------
// Database: Required.
// Cache   : Compatible.

func (gs *GroupService) AddGroup(name string, session *Session) error {
	if gs.db == nil {
		return common.ERR_NO_SERVICE
	}

	if !session.IsSystem() {
		return common.ERR_NO_AUTHORITY
	}

	if len(name) == 0 {
		return errors.New("Empty group name.")
	}
	sName := common.Escape(name)
	if len(sName) == 0 {
		return errors.New("Invalid group name.")
	}

	sUpdateIP := session.IP
	sUpdateTime := common.GetTimeString()
	sUpdater := strconv.Itoa(session.UserID)

	// Check via cache.
	if gs.cache != nil {
		m, err := gs.cache.GetAllFields(common.KEY_PREFIX_GROUP + common.FIELD_ID)
		if err != nil {
			return err
		}

		for _, v := range m {
			if v == sName {
				return common.ERR_DUPLICATED_GROUP
			}
		}
	}

	if err := (func() error {
		tx, err := gs.db.Transaction()
		if err != nil {
			return err
		}

		// Check via database.
		sql := "SELECT " +
			common.FIELD_ID +
			" FROM " +
			common.TABLE_GROUP +
			" WHERE " +
			common.FIELD_NAME + "='" + sName + "'" +
			" FOR UPDATE;"

		rows, err := tx.Query(sql)
		if err != nil {
			tx.Rollback()
			return err
		}

		if rows.Next() {
			rows.Close()
			tx.Rollback()
			return common.ERR_DUPLICATED_GROUP
		}
		rows.Close()

		// Update database.
		sql = "INSERT INTO " +
			common.TABLE_GROUP +
			" (" +
			common.FIELD_NAME + "," +
			common.FIELD_UPDATE_IP + "," +
			common.FIELD_UPDATE_TIME + "," +
			common.FIELD_UPDATER +
			") VALUES (" +
			"'" + sName + "'," +
			"'" + sUpdateIP + "'," +
			sUpdateTime + "," +
			sUpdater +
			");"

		r, err := tx.Exec(sql)
		if err != nil {
			tx.Rollback()
			return err
		}

		id, err := r.LastInsertId()
		if err != nil {
			tx.Rollback()
			return err
		}
		sGroupID := strconv.FormatInt(id, 10)

		// Update cache.
		if gs.cache != nil {
			err = gs.cache.SetField(common.KEY_PREFIX_GROUP+common.FIELD_ID, sGroupID, sName)
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

	_, err := gs.SetCachedQuery()
	return err
}

//----------------------------------------------------------------------------
// Database: Required.
// Cache   : Compatible.

func (gs *GroupService) DeleteGroup(groupID int, session *Session) error {
	if gs.db == nil {
		return common.ERR_NO_SERVICE
	}

	if !session.IsSystem() {
		return common.ERR_NO_AUTHORITY
	}

	sGroupID := strconv.Itoa(groupID)

	if err := (func() error {
		// Create a transaction.
		tx, err := gs.db.Transaction()
		if err != nil {
			return err
		}

		sql := "SELECT COUNT(*) FROM " + common.TABLE_CLASS + " WHERE " + common.FIELD_GROUP_ID + "=" + sGroupID + ";"
		row := tx.QueryRow(sql)

		cnt := 0
		if err = row.Scan(&cnt); err != nil {
			tx.Rollback()
			return err
		}
		if cnt > 0 {
			tx.Rollback()
			return common.ERR_GROUP_IS_NOT_EMPTY
		}

		sql = "DELETE FROM " +
			common.TABLE_GROUP +
			" WHERE " +
			common.FIELD_ID + "=" + sGroupID + ";"

		_, err = tx.Exec(sql)
		if err != nil {
			tx.Rollback()
			return err
		}

		// Delete from cache.
		if gs.cache != nil {
			err := gs.cache.DelField(common.KEY_PREFIX_GROUP+common.FIELD_ID, sGroupID)
			if err != nil {
				tx.Rollback()
				return err
			}

			err = gs.cache.Del(common.KEY_PREFIX_GROUP + sGroupID)
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

	_, err := gs.SetCachedQuery()
	return err
}

//----------------------------------------------------------------------------
// Database: Compatible.
// Cache   : Compatible.

func (gs *GroupService) QueryGroup() (string, error) {

	// Get them via cache.
	if gs.cache != nil {
		if s, err := gs.cache.GetKey(gs.cachedQuery); err == nil {
			return s, nil
		}
	}

	// Okay, get them via database.
	if gs.db != nil {
		sql := "SELECT " +
			common.FIELD_ID + "," +
			common.FIELD_NAME +
			" FROM " +
			common.TABLE_GROUP +
			" WHERE " +
			common.FIELD_ID + ">4;"

		rows, err := gs.db.Select(sql)
		if err != nil {
			return "", err
		}
		defer rows.Close()

		//------------------------------------------------

		first := true
		gi := new(GroupInfo)

		s := `"` + common.FIELD_GROUP + `":[`
		for rows.Next() {
			err := rows.Scan(&gi.ID, &gi.Name)
			if err != nil {
				return "", errors.New("Failed to get group info.")
			}

			if first {
				first = false
			} else {
				s += `,`
			}

			s += `{` + gi.ToJSON(true) + `}`
		}
		s += `]`

		// Update cache.
		if gs.cache != nil {
			if err = gs.cache.SetKey(gs.cachedQuery, s); err != nil {
				// TODO:
			}
		}

		return s, nil
	}

	return "", common.ERR_NO_SERVICE
}

//----------------------------------------------------------------------------
// Database: None.
// Cache   : Required.

func (gs *GroupService) SetCachedQuery() (string, error) {
	// Check requirements.
	if gs.cache == nil {
		return "", common.ERR_NO_CACHE
	}

	// Get all groups from cache.
	m, err := gs.cache.GetAllFields(common.KEY_PREFIX_GROUP + common.FIELD_ID)
	if err != nil {
		return "", err
	}

	// Sort the groups according to their IDs.
	i := 0
	arr := make([]int, len(m))
	for k, _ := range m {
		arr[i], err = strconv.Atoi(k)
		if err == nil {
			i++
		}
	}
	sort.Ints(arr)

	gi := new(GroupInfo)
	first := true

	// Construct the result string.
	s := `"` + common.FIELD_GROUP + `":[`
	for i := 0; i < len(arr); i++ {
		name, okay := m[strconv.Itoa(arr[i])]
		if !okay {
			// This scenario should never happen.
			continue
		}

		gi.ID = arr[i]
		gi.Name = name

		if first {
			first = false
		} else {
			s += `,`
		}

		s += `{` + gi.ToJSON(true) + `}`
	}
	s += `]`

	// Cache the result string.
	return s, gs.cache.SetKey(gs.cachedQuery, s)
}

//----------------------------------------------------------------------------

func (gs *GroupService) Count() (int, error) {
	if gs.db == nil {
		return 0, common.ERR_NO_DATABASE
	}

	return gs.db.Count(common.TABLE_GROUP)
}

//----------------------------------------------------------------------------
