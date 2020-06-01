package service

import (
	"github.com/wangmhgo/go-project/gdun/common"
)

//----------------------------------------------------------------------------

type InstituteService struct {
	db    *common.Database
	cache *common.Cache
}

func NewInstituteService(db *common.Database, cache *common.Cache) (*InstituteService, error) {
	is := new(InstituteService)
	is.db = db
	is.cache = cache

	if err := is.Init(); err != nil {
		return is, err
	}

	return is, nil
}

//----------------------------------------------------------------------------

func (is *InstituteService) Init() error {
	if is.db == nil {
		return common.ERR_NO_DATABASE
	}

	// TODO:
	return nil

	sql := "CREATE TABLE IF NOT EXISTS `" + common.TABLE_INSTITUTE + "` ("
	sql += " `" + common.FIELD_ID + "` INT NOT NULL AUTO_INCREMENT,"
	sql += " `" + common.FIELD_NAME + "` VARCHAR(512) NOT NULL,"
	sql += " `" + common.FIELD_GROUP_LIST + "` TEXT,"
	sql += " `" + common.FIELD_USER_LIST + "` TEXT,"
	sql += " `" + common.FIELD_UPDATE_IP + "` VARCHAR(32) NOT NULL,"
	sql += " `" + common.FIELD_UPDATE_TIME + "` BIGINT NOT NULL,"
	sql += " `" + common.FIELD_UPDATER + "` INT NOT NULL,"
	sql += " PRIMARY KEY (`" + common.FIELD_ID + "`)"
	sql += ") ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;"

	if _, err := is.db.Exec(sql); err != nil {
		return err
	}

	return nil
}

//----------------------------------------------------------------------------

func (is *InstituteService) Preload() (int, error) {
	if is.db == nil {
		return 0, common.ERR_NO_DATABASE
	}
	if is.cache == nil {
		return 0, common.ERR_NO_CACHE
	}

	s := "SELECT " +
		common.FIELD_ID + "," +
		common.FIELD_NAME + "," +
		common.FIELD_GROUP_LIST + "," +
		common.FIELD_USER_LIST +
		" FROM " +
		common.TABLE_INSTITUTE + ";"

	rows, err := is.db.Select(s)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	cnt := 0

	id := 0
	name := ""
	gl := ""
	ul := ""

	for rows.Next() {
		if err = rows.Scan(&id, &name, &gl, &ul); err != nil {
			return cnt, err
		}

		// TODO:

		cnt++
	}

	return cnt, nil
}

//----------------------------------------------------------------------------
