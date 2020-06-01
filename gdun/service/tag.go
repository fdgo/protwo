package service

import (
	"github.com/wangmhgo/go-project/gdun/common"
	"strconv"
)

//----------------------------------------------------------------------------

type TagService struct {
	db    *common.Database
	cache *common.Cache
}

func NewTagService(db *common.Database, cache *common.Cache) (*TagService, error) {
	ts := new(TagService)
	ts.db = db
	ts.cache = cache

	if err := ts.Init(); err != nil {
		return ts, err
	}

	return ts, nil
}

//----------------------------------------------------------------------------

/*
TAG:G:{groupID}
	{id}	{name}

TAG:G:{groupID}:text
*/

func (ts *TagService) Init() error {
	s := "CREATE TABLE IF NOT EXISTS `" + common.TABLE_TAG + "` (" +
		"`" + common.FIELD_ID + "` INT NOT NULL AUTO_INCREMENT," +
		"`" + common.FIELD_NAME + "` VARCHAR(512) NOT NULL DEFAULT ''," +
		"`" + common.FIELD_GROUP_ID + "` INT NOT NULL DEFAULT 0," +
		"`" + common.FIELD_UPDATE_IP + "` VARCHAR(32) NOT NULL," +
		"`" + common.FIELD_UPDATE_TIME + "` BIGINT NOT NULL," +
		"`" + common.FIELD_UPDATER + "` INT NOT NULL," +
		"PRIMARY KEY (`" + common.FIELD_ID + "`)," +
		"KEY (`" + common.FIELD_GROUP_ID + "`)" +
		") ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8;"

	if _, err := ts.db.Exec(s); err != nil {
		return err
	}
	return nil
}

//----------------------------------------------------------------------------

func (ts *TagService) refreshCache(groupID int) (int, error) {
	if ts.cache == nil {
		return -1, common.ERR_NO_SERVICE
	}

	key := common.KEY_PREFIX_TAG + common.KEY_PREFIX_GROUP + strconv.Itoa(groupID)

	m, err := ts.cache.GetAllFields(key)
	if err != nil {
		return -2, err
	}

	r := ``
	first := true
	for sID, sName := range m {
		if len(sID) == 0 || len(sName) == 0 {
			continue
		}

		if first {
			first = false
		} else {
			r += `,`
		}
		r += `"` + common.UnescapeForJSON(sID) + `":"` + common.UnescapeForJSON(sName) + `"`
	}

	if err = ts.cache.SetKey(key+":"+common.FIELD_TEXT, r); err != nil {
		return -3, err
	}

	return 0, nil
}

func (ts *TagService) checkAuthority(groupID int, session *Session) bool {
	if session.IsSystem() {
		if groupID > common.GROUP_ID_FOR_KEEPER {
			return true
		}
	}

	if session.IsAssistant() {
		if groupID == session.GroupID {
			return true
		}
	}

	return false
}

//----------------------------------------------------------------------------

func (ts *TagService) AddTag(name string, groupID int, session *Session) (int, int, error) {
	if ts.db == nil {
		return 0, -1, common.ERR_NO_SERVICE
	}

	// Check authority.
	if !ts.checkAuthority(groupID, session) {
		return 0, -2, common.ERR_NO_AUTHORITY
	}

	sName := common.Escape(name)
	if len(sName) == 0 {
		return 0, -3, common.ERR_INVALID_NAME
	}

	sGroupID := strconv.Itoa(groupID)

	sUpdater := strconv.Itoa(session.UserID)
	sUpdateIP := session.IP
	sUpdateTime := common.GetTimeString()

	s := "INSERT INTO " +
		common.TABLE_TAG +
		"(" +
		common.FIELD_NAME + "," +
		common.FIELD_GROUP_ID + "," +
		common.FIELD_UPDATE_TIME + "," +
		common.FIELD_UPDATE_IP + "," +
		common.FIELD_UPDATER +
		") VALUES (" +
		"'" + sName + "'," +
		sGroupID + "," +
		sUpdateTime + "," +
		"'" + sUpdateIP + "'," +
		sUpdater +
		");"

	id, err := ts.db.Insert(s, 1)
	if err != nil {
		return 0, -4, err
	}

	if ts.cache != nil {
		if err = ts.cache.SetField(common.KEY_PREFIX_TAG+common.KEY_PREFIX_GROUP+sGroupID, strconv.FormatInt(id, 10), sName); err != nil {
			return 0, -5, err
		}
		if status, err := ts.refreshCache(groupID); err != nil {
			return 0, status - 5, err
		}
	}

	return int(id), 0, nil
}

//----------------------------------------------------------------------------

func (ts *TagService) ChangeTag(id int, name string, groupID int, session *Session) (int, error) {
	if ts.db == nil {
		return -1, common.ERR_NO_SERVICE
	}

	// Check authority.
	if !ts.checkAuthority(groupID, session) {
		return -2, common.ERR_NO_AUTHORITY
	}

	sID := strconv.Itoa(id)
	sGroupID := strconv.Itoa(groupID)
	sName := common.Escape(name)
	if len(sName) == 0 {
		return -3, common.ERR_INVALID_NAME
	}

	sUpdater := strconv.Itoa(session.UserID)
	sUpdateIP := session.IP
	sUpdateTime := common.GetTimeString()

	s := "UPDATE " +
		common.TABLE_TAG +
		" SET " +
		common.FIELD_NAME + "='" + sName + "'," +
		common.FIELD_UPDATE_TIME + "=" + sUpdateTime + "," +
		common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
		common.FIELD_UPDATER + "=" + sUpdater +
		" WHERE " +
		common.FIELD_ID + "=" + sID +
		" AND " + common.FIELD_GROUP_ID + "=" + sGroupID + ";"

	if _, err := ts.db.Exec(s); err != nil {
		return -4, err
	}

	if ts.cache != nil {
		if err := ts.cache.SetField(common.KEY_PREFIX_TAG+common.KEY_PREFIX_GROUP+sGroupID, sID, sName); err != nil {
			return -5, err
		}
		if status, err := ts.refreshCache(groupID); err != nil {
			return status - 5, err
		}
	}

	return 0, nil
}

//----------------------------------------------------------------------------

func (ts *TagService) QueryTags(groupID int) (string, int, error) {
	prefix := `"1":"前导",` + `"2":"知识",` + `"3":"复习",` + `"4":"冲刺押题"`

	sGroupID := strconv.Itoa(groupID)
	key := common.KEY_PREFIX_TAG + common.KEY_PREFIX_GROUP + sGroupID

	if ts.cache != nil {
		s, err := ts.cache.GetKey(key + ":" + common.FIELD_TEXT)
		if err == nil {
			if len(s) == 0 {
				return prefix, 0, nil
			}

			return prefix + `,` + s, 0, nil
		}
	}

	if ts.db != nil {
		s := "SELECT " +
			common.FIELD_ID + "," +
			common.FIELD_NAME +
			" FROM " +
			common.TABLE_TAG +
			" WHERE " +
			common.FIELD_GROUP_ID + "=" + sGroupID + ";"

		rows, err := ts.db.Select(s)
		if err != nil {
			return "", -1, err
		}
		defer rows.Close()

		r := ``
		first := true
		for rows.Next() {
			id := 0
			name := ""
			if err = rows.Scan(&id, &name); err != nil {
				return r, -2, err
			}

			if first {
				first = false
			} else {
				r += `,`
			}
			r += `"` + strconv.Itoa(id) + `":"` + common.UnescapeForJSON(name) + `"`

			if ts.cache != nil {
				if err = ts.cache.SetField(key, strconv.Itoa(id), name); err != nil {
					return r, -3, err
				}
			}
		}

		if ts.cache != nil {
			if err = ts.cache.SetKey(key+":"+common.FIELD_TEXT, r); err != nil {
				return r, -4, err
			}
		}

		//------------------------------------------------

		if len(r) == 0 {
			return prefix, 0, nil
		}

		return prefix + `,` + r, 0, nil
	}

	return "", -5, common.ERR_NO_SERVICE
}

//----------------------------------------------------------------------------
