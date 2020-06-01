package service

import (
	// "container/list"
	"github.com/wangmhgo/go-project/gdun/common"
	"strconv"
)

//----------------------------------------------------------------------------

func (cs *ClassService) AddSubClass(classID int, subClassID int, session *Session) error {
	// Check requirements.
	if cs.db == nil {
		return common.ERR_NO_SERVICE
	}

	// Check authority.
	ci, err := cs.GetClass(classID, session)
	if err != nil {
		return err
	}
	if ci.PlatformID != common.PLATFORM_ID_FOR_PACKAGE {
		return common.ERR_INVALID_CLASS
	}

	// Check whether the relationship exists or not.
	sSubClassID := strconv.Itoa(subClassID)
	if common.InList(sSubClassID, common.Unescape(ci.PlatformData)) {
		return nil
	}

	// Check authority.
	sci, err := cs.GetClass(subClassID, session)
	if err != nil {
		return err
	}

	// Prepare parameters.
	sClassID := strconv.Itoa(classID)

	sUpdater := strconv.Itoa(session.UserID)
	sUpdateIP := session.IP
	sUpdateTime := common.GetTimeString()

	if err = (func() error {
		// Create a transaction.
		tx, err := cs.db.Transaction()
		if err != nil {
			return err
		}

		// Get existing value.
		s := "SELECT " +
			common.FIELD_PLATFORM_DATA +
			" FROM " +
			common.TABLE_CLASS +
			" WHERE " +
			common.FIELD_ID + "=" + sClassID + " FOR UPDATE;"
		row := tx.QueryRow(s)

		pd := ""
		if err = row.Scan(&pd); err != nil {
			tx.Rollback()
			return err
		}

		// Compute new value.
		changed := false

		pd = common.Unescape(pd)
		pd, changed = common.AddToList(sSubClassID, pd)
		pd = common.Escape(pd)

		if !changed {
			tx.Rollback()

			if cs.cache != nil {
				m := make(map[string]string)
				m[common.FIELD_PLATFORM_DATA] = pd
				m[common.FIELD_UPDATER] = sUpdater
				m[common.FIELD_UPDATE_IP] = sUpdateIP
				m[common.FIELD_UPDATE_TIME] = sUpdateTime

				if err = cs.cache.SetFields(common.KEY_PREFIX_CLASS+sClassID, m); err != nil {
					return err
				}

				// Remove it from group's class list.
				if err = cs.cache.DelField(common.KEY_PREFIX_CLASS+common.KEY_PREFIX_GROUP+strconv.Itoa(sci.GroupID), strconv.Itoa(subClassID)); err != nil {
					// TODO:
				}
			}

			return nil
		}

		// Update database.
		s = "UPDATE " +
			common.TABLE_CLASS +
			" SET " +
			common.FIELD_PLATFORM_DATA + "='" + pd + "'," +
			common.FIELD_UPDATER + "=" + sUpdater + "," +
			common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
			common.FIELD_UPDATE_TIME + "=" + sUpdateTime +
			" WHERE " +
			common.FIELD_ID + "=" + sClassID + ";"

		if _, err = tx.Exec(s); err != nil {
			tx.Rollback()
			return err
		}

		if cs.cache != nil {
			m := make(map[string]string)
			m[common.FIELD_PLATFORM_DATA] = pd
			m[common.FIELD_UPDATER] = sUpdater
			m[common.FIELD_UPDATE_IP] = sUpdateIP
			m[common.FIELD_UPDATE_TIME] = sUpdateTime

			if err = cs.cache.SetFields(common.KEY_PREFIX_CLASS+sClassID, m); err != nil {
				tx.Rollback()
				return err
			}

			// Remove it from group's class list.
			if err = cs.cache.DelField(common.KEY_PREFIX_CLASS+common.KEY_PREFIX_GROUP+strconv.Itoa(sci.GroupID), sSubClassID); err != nil {
				// TODO:
			}
		}

		// Commit the transaction.
		if err = tx.Commit(); err != nil {
			tx.Rollback()
			return err
		}

		return nil
	})(); err != nil {
		return err
	}

	if err = (func() error {
		// Create a transaction.
		tx, err := cs.db.Transaction()
		if err != nil {
			return err
		}

		// Get existing value.
		s := "SELECT " +
			common.FIELD_PARENT +
			" FROM " +
			common.TABLE_CLASS +
			" WHERE " +
			common.FIELD_ID + "=" + sSubClassID + " FOR UPDATE;"

		row := tx.QueryRow(s)

		parent := ""
		if err = row.Scan(&parent); err != nil {
			tx.Rollback()
			return err
		}

		// Compute new value.
		changed := false
		parent, changed = common.AddToList(sClassID, parent)
		if !changed {
			tx.Rollback()

			if cs.cache != nil {
				m := make(map[string]string)
				m[common.FIELD_PARENT] = parent
				m[common.FIELD_UPDATER] = sUpdater
				m[common.FIELD_UPDATE_IP] = sUpdateIP
				m[common.FIELD_UPDATE_TIME] = sUpdateTime

				if err = cs.cache.SetFields(common.KEY_PREFIX_CLASS+sSubClassID, m); err != nil {
					// return err
					// TODO:
				}
			}

			return nil
		}

		// Update database.
		s = "UPDATE " +
			common.TABLE_CLASS +
			" SET " +
			common.FIELD_PARENT + "='" + parent + "'," +
			common.FIELD_UPDATER + "=" + sUpdater + "," +
			common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
			common.FIELD_UPDATE_TIME + "=" + sUpdateTime +
			" WHERE " +
			common.FIELD_ID + "=" + sSubClassID + ";"

		if _, err = tx.Exec(s); err != nil {
			tx.Rollback()
			return err
		}

		// Update cache.
		if cs.cache != nil {
			m := make(map[string]string)
			m[common.FIELD_PARENT] = parent
			m[common.FIELD_UPDATER] = sUpdater
			m[common.FIELD_UPDATE_IP] = sUpdateIP
			m[common.FIELD_UPDATE_TIME] = sUpdateTime

			if err = cs.cache.SetFields(common.KEY_PREFIX_CLASS+sSubClassID, m); err != nil {
				tx.Rollback()
				return err
			}
		}

		// Commit the transaction.
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

func (cs *ClassService) DeleteSubClass(classID int, subClassID int, session *Session) error {
	// Check requirements.
	if cs.db == nil {
		return common.ERR_NO_SERVICE
	}

	// Check authority.
	ci, err := cs.GetClass(classID, session)
	if err != nil {
		return err
	}
	if ci.PlatformID != common.PLATFORM_ID_FOR_PACKAGE {
		return common.ERR_INVALID_CLASS
	}

	// Check whether the relationship exists or not.
	sSubClassID := strconv.Itoa(subClassID)
	if !common.InList(sSubClassID, common.Unescape(ci.PlatformData)) {
		return nil
	}
	sci, err := cs.GetClass(subClassID, session)
	if err != nil {
		return err
	}

	// Prepare parameters.
	sClassID := strconv.Itoa(classID)

	sUpdater := strconv.Itoa(session.UserID)
	sUpdateIP := session.IP
	sUpdateTime := common.GetTimeString()

	if err = (func() error {
		// Create a new transaction.
		tx, err := cs.db.Transaction()
		if err != nil {
			return err
		}

		// Get existing value.
		s := "SELECT " +
			common.FIELD_PLATFORM_DATA +
			" FROM " +
			common.TABLE_CLASS +
			" WHERE " +
			common.FIELD_ID + "=" + sClassID + " FOR UPDATE;"
		row := tx.QueryRow(s)

		pd := ""
		if err = row.Scan(&pd); err != nil {
			tx.Rollback()
			return err
		}

		// Compute new value.
		changed := false

		pd = common.Unescape(pd)
		pd, changed = common.DeleteFromList(sSubClassID, pd)
		pd = common.Escape(pd)

		if !changed {
			tx.Rollback()

			if cs.cache != nil {
				m := make(map[string]string)
				m[common.FIELD_PLATFORM_DATA] = pd
				m[common.FIELD_UPDATER] = sUpdater
				m[common.FIELD_UPDATE_IP] = sUpdateIP
				m[common.FIELD_UPDATE_TIME] = sUpdateTime

				if err = cs.cache.SetFields(common.KEY_PREFIX_CLASS+sClassID, m); err != nil {
					return err
				}

				// Add it to group's class list.
				if err = cs.cache.SetField(common.KEY_PREFIX_CLASS+common.KEY_PREFIX_GROUP+strconv.Itoa(sci.GroupID), strconv.Itoa(subClassID), strconv.Itoa(sci.EndTime)); err != nil {
					// TODO:
				}
			}

			return nil
		}

		// Update database.
		s = "UPDATE " +
			common.TABLE_CLASS +
			" SET " +
			common.FIELD_PLATFORM_DATA + "='" + pd + "'," +
			common.FIELD_UPDATER + "=" + sUpdater + "," +
			common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
			common.FIELD_UPDATE_TIME + "=" + sUpdateTime +
			" WHERE " +
			common.FIELD_ID + "=" + sClassID + ";"

		if _, err = tx.Exec(s); err != nil {
			tx.Rollback()
			return err
		}

		if cs.cache != nil {
			m := make(map[string]string)
			m[common.FIELD_PLATFORM_DATA] = pd
			m[common.FIELD_UPDATER] = sUpdater
			m[common.FIELD_UPDATE_IP] = sUpdateIP
			m[common.FIELD_UPDATE_TIME] = sUpdateTime

			if err = cs.cache.SetFields(common.KEY_PREFIX_CLASS+sClassID, m); err != nil {
				tx.Rollback()
				return err
			}

			// Add it to group's class list.
			// if err = cs.cache.SetField(common.KEY_PREFIX_CLASS+common.KEY_PREFIX_GROUP+strconv.Itoa(sci.GroupID), strconv.Itoa(subClassID), strconv.Itoa(sci.EndTime)); err != nil {
			// TODO:
			// }
		}

		// Commit the transaction.
		if err = tx.Commit(); err != nil {
			tx.Rollback()
			return err
		}

		return nil
	})(); err != nil {
		return err
	}

	if err = (func() error {
		// Create a new transaction.
		tx, err := cs.db.Transaction()
		if err != nil {
			return err
		}

		// Get existing value from database.
		s := "SELECT " +
			common.FIELD_PARENT +
			" FROM " +
			common.TABLE_CLASS +
			" WHERE " +
			common.FIELD_ID + "=" + sSubClassID + " FOR UPDATE;"

		row := tx.QueryRow(s)

		parent := ""
		if err = row.Scan(&parent); err != nil {
			tx.Rollback()
			return err
		}

		// Compute new value.
		changed := false
		parent, changed = common.DeleteFromList(sClassID, parent)
		if !changed {
			tx.Rollback()

			if cs.cache != nil {
				m := make(map[string]string)
				m[common.FIELD_PARENT] = parent
				m[common.FIELD_UPDATER] = sUpdater
				m[common.FIELD_UPDATE_IP] = sUpdateIP
				m[common.FIELD_UPDATE_TIME] = sUpdateTime

				if err = cs.cache.SetFields(common.KEY_PREFIX_CLASS+sSubClassID, m); err != nil {
					return err
				}

				if len(parent) == 0 {
					// Add it to group's class list.
					if err = cs.cache.SetField(common.KEY_PREFIX_CLASS+common.KEY_PREFIX_GROUP+strconv.Itoa(sci.GroupID), sSubClassID, strconv.Itoa(sci.EndTime)); err != nil {
						// TODO:
					}
				}
			}

			return err
		}

		// Update database.
		s = "UPDATE " +
			common.TABLE_CLASS +
			" SET " +
			common.FIELD_PARENT + "='" + parent + "'," +
			common.FIELD_UPDATER + "=" + sUpdater + "," +
			common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
			common.FIELD_UPDATE_TIME + "=" + sUpdateTime +
			" WHERE " +
			common.FIELD_ID + "=" + sSubClassID + ";"

		if _, err = tx.Exec(s); err != nil {
			tx.Rollback()
			return err
		}

		// Update cache.
		if cs.cache != nil {
			m := make(map[string]string)
			m[common.FIELD_PARENT] = parent
			m[common.FIELD_UPDATER] = sUpdater
			m[common.FIELD_UPDATE_IP] = sUpdateIP
			m[common.FIELD_UPDATE_TIME] = sUpdateTime

			if err = cs.cache.SetFields(common.KEY_PREFIX_CLASS+sSubClassID, m); err != nil {
				tx.Rollback()
				return err
			}

			if len(parent) == 0 {
				// Add it to group's class list.
				if err = cs.cache.SetField(common.KEY_PREFIX_CLASS+common.KEY_PREFIX_GROUP+strconv.Itoa(sci.GroupID), sSubClassID, strconv.Itoa(sci.EndTime)); err != nil {
					// TODO:
				}
			}
		}

		// Commit the transaction.
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

func (cs *ClassService) QuerySubClasses(classID int, session *Session) (ClassInfoSlice, error) {
	// Check authority.
	ci, err := cs.GetClass(classID, session)
	if err != nil {
		return nil, err
	}
	if ci.PlatformID != common.PLATFORM_ID_FOR_PACKAGE {
		return nil, common.ERR_INVALID_CLASS
	}

	classIDs := common.StringToIntArray(common.Unescape(ci.PlatformData))
	if len(classIDs) == 0 {
		return ClassInfoSlice([]*ClassInfo{}), nil
	}

	result := make([]*ClassInfo, len(classIDs))
	for i := 0; i < len(classIDs); i++ {
		sci, err := cs.GetClass(classIDs[i], session)
		if err != nil {
			result[i] = nil
			continue
		}

		result[i] = sci
	}

	return ClassInfoSlice(result), nil
}

//----------------------------------------------------------------------------
