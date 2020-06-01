package service

import (
	"github.com/wangmhgo/go-project/gdun/common"
	"strconv"
)

//----------------------------------------------------------------------------
// Database: Required.

func (us *UserService) GetOrAddGdStudent(gdStudentID int, ip string, addIfNotExist bool) (int, string, bool, error) {
	sGdStudentID := strconv.Itoa(gdStudentID)

	// Try to get user information from cache.
	if us.cache != nil {
		if id, name, err := (func() (int, string, error) {
			s, err := us.cache.GetField(us.gdStudentIDKey, sGdStudentID)
			if err != nil {
				return 0, "", err
			}

			userID, err := strconv.Atoi(s)
			if err != nil {
				return 0, "", err
			}

			s, err = us.cache.GetField(common.KEY_PREFIX_USER+s, common.FIELD_NICKNAME)
			if err != nil {
				s = common.VALUE_NOT_ALLOWED
			}

			return userID, s, nil
		})(); err == nil {
			return id, name, false, nil
		}
	}

	if us.db == nil {
		return 0, "", false, common.ERR_NO_SERVICE
	}

	// Try to get and/or add user information via database.
	id, name, err := (func() (int, string, error) {
		s := "SELECT " +
			common.FIELD_ID + "," +
			common.FIELD_NICKNAME +
			" FROM " +
			common.TABLE_USER +
			" WHERE " +
			common.FIELD_GAODUN_STUDENT_ID + "=" + sGdStudentID + ";"

		rows, err := us.db.Select(s)
		if err != nil {
			return 0, "", err
		}
		defer rows.Close()

		if !rows.Next() {
			return 0, "", nil
		}

		id := 0
		name := ""
		if err = rows.Scan(&id, &name); err != nil {
			return 0, "", err
		}

		if name == common.VALUE_NOT_ALLOWED {
			name = ""
		}
		return id, name, nil
	})()
	if err != nil {
		return 0, "", false, err
	}

	// If it exists.
	if id > 0 {
		// Fill cache.
		if us.cache != nil {
			if err = us.cache.SetField(us.gdStudentIDKey, sGdStudentID, strconv.Itoa(id)); err != nil {
				// TODO:
			}
		}
		return id, name, false, nil
	}

	// If does not exist.
	if !addIfNotExist {
		return 0, "", false, common.ERR_NO_USER
	}

	// Add his information to database.
	id, name, err = (func() (int, string, error) {
		s := "INSERT INTO " +
			common.TABLE_USER +
			"(" +
			common.FIELD_NAME + "," +
			common.FIELD_PASSWORD + "," +
			common.FIELD_NICKNAME + "," +
			common.FIELD_GROUP_ID + "," +
			common.FIELD_GAODUN_STUDENT_ID + "," +
			common.FIELD_UPDATE_TIME + "," +
			common.FIELD_UPDATE_IP + "," +
			common.FIELD_UPDATER +
			") VALUES (" +
			"'" + common.VALUE_NOT_ALLOWED + "'," +
			"'" + common.VALUE_NOT_ALLOWED + "'," +
			"'" + common.VALUE_NOT_ALLOWED + "'," +
			strconv.Itoa(common.GROUP_ID_FOR_STUDENT) + "," +
			sGdStudentID + "," +
			common.GetTimeString() + "," +
			"'" + ip + "'," +
			strconv.Itoa(0) +
			");"
		id, err := us.db.Insert(s, 1)
		if err != nil {
			return 0, "", err
		}

		if us.cache != nil {
			sID := strconv.FormatInt(id, 10)

			if err = us.cache.SetField(us.gdStudentIDKey, sGdStudentID, sID); err != nil {
				return int(id), "", err
			}

			m := make(map[string]string)
			m[common.FIELD_PASSWORD] = common.VALUE_NOT_ALLOWED
			m[common.FIELD_NICKNAME] = common.VALUE_NOT_ALLOWED
			m[common.FIELD_REMARK] = ""
			// m[common.FIELD_PRIVILEGE] = "0"
			m[common.FIELD_GROUP_ID] = strconv.Itoa(common.GROUP_ID_FOR_STUDENT)
			m[common.FIELD_GAODUN_STUDENT_ID] = sGdStudentID

			if err = us.cache.SetFields(common.KEY_PREFIX_USER+sID, m); err != nil {
				return int(id), "", err
			}
		}

		return int(id), "", nil
	})()

	return id, name, true, err
}

//----------------------------------------------------------------------------

func (us *UserService) QueryUserID(gdStudentID int) (int, error) {
	sGdStudentID := strconv.Itoa(gdStudentID)

	if us.cache != nil {
		if s, err := us.cache.GetField(us.gdStudentIDKey, sGdStudentID); err == nil {
			if userID, err := strconv.Atoi(s); err == nil {
				return userID, nil
			}
		}

	}

	if us.db != nil {
		s := "SELECT " +
			common.FIELD_ID +
			" FROM " +
			common.TABLE_USER +
			" WHERE " +
			common.FIELD_GAODUN_STUDENT_ID + "=" + sGdStudentID + ";"

		rows, err := us.db.Select(s)
		if err != nil {
			return 0, err
		}
		defer rows.Close()

		if !rows.Next() {
			return 0, common.ERR_NO_USER
		}

		userID := 0
		if err = rows.Scan(&userID); err != nil {
			return 0, err
		}

		if us.cache != nil {
			if err = us.cache.SetField(us.gdStudentIDKey, sGdStudentID, strconv.Itoa(userID)); err != nil {
				// TODO:
			}
		}

		return userID, nil
	}

	return 0, common.ERR_NO_SERVICE
}

//----------------------------------------------------------------------------

func (us *UserService) QueryGdStudentID(userID int) (int, error) {
	sUserID := strconv.Itoa(userID)

	if us.cache != nil {
		if s, err := us.cache.GetField(common.KEY_PREFIX_USER+sUserID, common.FIELD_GAODUN_STUDENT_ID); err == nil {
			if gdStudentID, err := strconv.Atoi(s); err == nil {
				return gdStudentID, nil
			}
		}
	}

	if us.db != nil {
		s := "SELECT " +
			common.FIELD_GAODUN_STUDENT_ID +
			" FROM " +
			common.TABLE_USER +
			" WHERE " +
			common.FIELD_ID + "=" + sUserID + ";"

		rows, err := us.db.Select(s)
		if err != nil {
			return 0, err
		}
		defer rows.Close()

		if !rows.Next() {
			return 0, common.ERR_NO_USER
		}

		gdStudentID := 0
		if err = rows.Scan(&gdStudentID); err != nil {
			return 0, err
		}

		if us.cache != nil {
			if err = us.cache.SetField(common.KEY_PREFIX_USER+sUserID, common.FIELD_GAODUN_STUDENT_ID, strconv.Itoa(gdStudentID)); err != nil {
				// TODO:
			}
		}

		return gdStudentID, nil
	}

	return 0, common.ERR_NO_SERVICE
}

//----------------------------------------------------------------------------
