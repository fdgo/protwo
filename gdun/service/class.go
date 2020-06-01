package service

import (
	// "container/list"
	"encoding/json"
	"errors"
	// "fmt"
	"github.com/wangmhgo/go-project/gdun/common"
	//"gitlab.hfjy.com/gdun/vender/bokecc"
	"strconv"
	"strings"
)

//----------------------------------------------------------------------------

type ClassService struct {
	db                *common.Database
	cache             *common.Cache
	oss               *common.ObjectStorage
	ms                *MeetingService
	gdCourseIDKey string
}

func NewClassService(db *common.Database, cache *common.Cache, oss *common.ObjectStorage, ms *MeetingService) (*ClassService, error) {
	cs := new(ClassService)
	cs.db = db
	cs.cache = cache
	cs.oss = oss
	cs.ms = ms
	cs.gdCourseIDKey = common.KEY_PREFIX_CLASS + common.FIELD_GAODUN_COURSE_ID

	err := cs.Init()
	if err != nil {
		return nil, err
	}

	return cs, nil
}

//----------------------------------------------------------------------------

func (cs *ClassService) Init() error {
	if cs.db == nil {
		return common.ERR_NO_DATABASE
	}

	sql := "CREATE TABLE IF NOT EXISTS `" + common.TABLE_CLASS + "` ("
	sql += " `" + common.FIELD_ID + "` INT NOT NULL AUTO_INCREMENT,"
	sql += " `" + common.FIELD_NAME + "` VARCHAR(512) NOT NULL,"
	sql += " `" + common.FIELD_SUBJECT_LIST + "` TEXT,"
	sql += " `" + common.FIELD_TEACHER_LIST + "` TEXT,"
	sql += " `" + common.FIELD_KEEPER_LIST + "` TEXT,"
	sql += " `" + common.FIELD_STUDENT_LIST + "` TEXT,"
	sql += " `" + common.FIELD_MEETING_LIST + "` TEXT,"
	sql += " `" + common.FIELD_DELETED + "` TEXT,"
	sql += " `" + common.FIELD_NUMBER_OF_FINISHED_MEETING + "` INT NOT NULL DEFAULT 0,"
	sql += " `" + common.FIELD_START_TIME + "` BIGINT NOT NULL,"
	sql += " `" + common.FIELD_END_TIME + "` BIGINT NOT NULL DEFAULT 0,"
	sql += " `" + common.FIELD_PARENT + "` TEXT NOT NULL,"
	sql += " `" + common.FIELD_GROUP_ID + "` INT NOT NULL,"
	sql += " `" + common.FIELD_GAODUN_COURSE_ID + "` INT NOT NULL DEFAULT 0,"
	sql += " `" + common.FIELD_TEMPLATE + "` INT NOT NULL DEFAULT 0,"
	sql += " `" + common.FIELD_PLATFORM_ID + "` INT NOT NULL DEFAULT 0,"
	sql += " `" + common.FIELD_PLATFORM_DATA + "` VARCHAR(512) DEFAULT '',"
	sql += " `" + common.FIELD_ALLY + "` INT NOT NULL DEFAULT 0,"
	sql += " `" + common.FIELD_UPDATE_IP + "` VARCHAR(32) NOT NULL,"
	sql += " `" + common.FIELD_UPDATE_TIME + "` BIGINT NOT NULL,"
	sql += " `" + common.FIELD_UPDATER + "` INT NOT NULL,"
	sql += " PRIMARY KEY (`" + common.FIELD_ID + "`),"
	sql += " KEY (`" + common.FIELD_GROUP_ID + "`)"
	sql += ") ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;"

	_, err := cs.db.Exec(sql)
	if err != nil {
		return err
	}

	sql = "CREATE TABLE IF NOT EXISTS `" + common.TABLE_USER_CLASS + "` ("
	sql += " `" + common.FIELD_USER_ID + "` INT NOT NULL,"
	sql += " `" + common.FIELD_CLASS_LIST + "` TEXT,"
	sql += " `" + common.FIELD_UPDATE_IP + "` VARCHAR(32) NOT NULL,"
	sql += " `" + common.FIELD_UPDATE_TIME + "` BIGINT NOT NULL,"
	sql += " `" + common.FIELD_UPDATER + "` INT NOT NULL,"
	sql += " PRIMARY KEY (`" + common.FIELD_USER_ID + "`)"
	sql += ") ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;"

	_, err = cs.db.Exec(sql)
	if err != nil {
		return err
	}

	sql = "CREATE TABLE IF NOT EXISTS `" + common.TABLE_KEEPER_CLASS + "` ("
	sql += " `" + common.FIELD_USER_ID + "` INT NOT NULL,"
	sql += " `" + common.FIELD_CLASS_ID + "` INT NOT NULL,"
	sql += " `" + common.FIELD_STUDENT_LIST + "` TEXT NOT NULL,"
	sql += " KEY (`" + common.FIELD_USER_ID + "`),"
	sql += " KEY (`" + common.FIELD_CLASS_ID + "`)"
	sql += ") ENGINE=InnoDB DEFAULT CHARSET=utf8;"

	_, err = cs.db.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

//----------------------------------------------------------------------------
// This routine yields:
//
// 1. A map from Gd course ID to class ID.
// 2. A map from class name to class ID.
// 3. Class Hash tables.
// 4. Class lists appended to user Hash tables.
// 5. Statistics appended to group Hash tables.

func (cs *ClassService) Preload() (int, int, int, error) {
	// Check requirements.
	if cs.db == nil {
		return 0, 0, 0, common.ERR_NO_DATABASE
	}
	if cs.cache == nil {
		return 0, 0, 0, common.ERR_NO_CACHE
	}

	//----------------------------------------------------
	// Step 1.

	n1, err := cs.preloadClasses()
	if err != nil {
		return n1, 0, 0, err
	}

	//----------------------------------------------------
	// Step 2.

	n2, err := (func() (int, error) {
		rest, err := cs.db.Count(common.TABLE_USER_CLASS)
		if err != nil {
			return 0, err
		}

		i := 0
		for i < rest {
			n, err := cs.preloadStudentClassRelation(i, common.DATABASE_PRELOAD_SIZE)
			if err != nil {
				return i, err
			}
			i += n
		}

		return i, nil
	})()
	if err != nil {
		return n1, n2, 0, err
	}

	//----------------------------------------------------
	// Step 3.

	n3, err := (func() (int, error) {
		rest, err := cs.db.Count(common.TABLE_USER_CLASS)
		if err != nil {
			return 0, err
		}

		i := 0
		for i < rest {
			n, err := cs.preloadKeeperClassRelation(i, common.DATABASE_PRELOAD_SIZE)
			if err != nil {
				return i, err
			}
			i += n
		}

		return i, nil
	})()
	if err != nil {
		return n1, n2, n3, err
	}

	return n1, n2, n3, err
}

func (cs *ClassService) preloadClasses() (int, error) {
	sql := "SELECT " +
		common.FIELD_ID + "," +
		common.FIELD_NAME + "," +
		common.FIELD_SUBJECT_LIST + "," +
		common.FIELD_TEACHER_LIST + "," +
		common.FIELD_KEEPER_LIST + "," +
		common.FIELD_STUDENT_LIST + "," +
		common.FIELD_MEETING_LIST + "," +
		common.FIELD_DELETED + "," +
		common.FIELD_NUMBER_OF_FINISHED_MEETING + "," +
		common.FIELD_START_TIME + "," +
		common.FIELD_END_TIME + "," +
		common.FIELD_PARENT + "," +
		common.FIELD_GROUP_ID + "," +
		common.FIELD_GAODUN_COURSE_ID + "," +
		common.FIELD_TEMPLATE + "," +
		common.FIELD_PLATFORM_ID + "," +
		common.FIELD_PLATFORM_DATA + "," +
		common.FIELD_ALLY + "," +
		common.FIELD_UPDATE_IP + "," +
		common.FIELD_UPDATE_TIME + "," +
		common.FIELD_UPDATER +
		" FROM " +
		common.TABLE_CLASS + ";"

	rows, err := cs.db.Select(sql)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	groupClasses := make(map[string]string)
	m := make(map[string]string)

	cnt := 0

	id := 0
	name := ""
	sbl := ""
	tl := ""
	kl := ""
	sl := ""
	ml := ""
	dl := ""
	nfm := 0
	st := 0
	et := 0
	parent := ""
	groupID := 0
	gdCourseID := 0
	template := 0
	platformID := 0
	platformData := ""
	ally := 0
	updateIP := ""
	ut := 0
	updater := 0

	subClassMap := make(map[int]int)

	for rows.Next() {
		err = rows.Scan(&id, &name, &sbl, &tl, &kl, &sl, &ml, &dl, &nfm, &st, &et, &parent, &groupID, &gdCourseID, &template, &platformID, &platformData, &ally, &updateIP, &ut, &updater)
		if err != nil {
			return cnt, err
		}

		if platformID == common.PLATFORM_ID_FOR_PACKAGE {
			arr := common.StringToIntArray(common.Unescape(platformData))
			for i := 0; i < len(arr); i++ {
				subClassMap[arr[i]] = groupID
			}
		}

		sID := strconv.Itoa(id)
		sGdCourseID := strconv.Itoa(gdCourseID)

		//------------------------------------------------
		// A map from Gd course ID to class ID.

		if gdCourseID > 0 {
			err = cs.cache.SetField(cs.gdCourseIDKey, sGdCourseID, sID)
			if err != nil {
				return cnt, err
			}
		}

		//------------------------------------------------
		// Append the class ID of this class to its group.

		sGroupID := strconv.Itoa(groupID)
		cl, okay := groupClasses[sGroupID]
		if !okay {
			groupClasses[sGroupID] = strconv.Itoa(id)
		} else {
			groupClasses[sGroupID] = cl + "," + strconv.Itoa(id)
		}

		//------------------------------------------------
		// The class itself.

		m[common.FIELD_NAME] = name
		m[common.FIELD_SUBJECT_LIST] = sbl
		m[common.FIELD_TEACHER_LIST] = tl
		m[common.FIELD_KEEPER_LIST] = kl
		m[common.FIELD_STUDENT_LIST] = sl
		m[common.FIELD_MEETING_LIST] = ml
		m[common.FIELD_DELETED] = dl
		m[common.FIELD_NUMBER_OF_FINISHED_MEETING] = strconv.Itoa(nfm)
		m[common.FIELD_START_TIME] = strconv.Itoa(st)
		m[common.FIELD_END_TIME] = strconv.Itoa(et)
		m[common.FIELD_PARENT] = parent
		m[common.FIELD_GROUP_ID] = strconv.Itoa(groupID)
		m[common.FIELD_GAODUN_COURSE_ID] = sGdCourseID
		m[common.FIELD_TEMPLATE] = strconv.Itoa(template)
		m[common.FIELD_PLATFORM_ID] = strconv.Itoa(platformID)
		m[common.FIELD_PLATFORM_DATA] = platformData
		m[common.FIELD_ALLY] = strconv.Itoa(ally)
		m[common.FIELD_UPDATE_IP] = updateIP
		m[common.FIELD_UPDATE_TIME] = strconv.Itoa(ut)
		m[common.FIELD_UPDATER] = strconv.Itoa(updater)

		err = cs.cache.SetFields(common.KEY_PREFIX_CLASS+strconv.Itoa(id), m)
		if err != nil {
			return cnt, err
		}

		// A map from group ID to class ID.
		err = cs.cache.SetField(common.KEY_PREFIX_CLASS+common.KEY_PREFIX_GROUP+strconv.Itoa(groupID), strconv.Itoa(id), strconv.Itoa(et))
		if err != nil {
			return cnt, err
		}

		cnt++
	}

	// Remove sub-classes from the group's class list.
	for cID, gID := range subClassMap {
		if err = cs.cache.DelField(common.KEY_PREFIX_CLASS+common.KEY_PREFIX_GROUP+strconv.Itoa(gID), strconv.Itoa(cID)); err != nil {
			// TODO:
		}

		cnt--
	}

	return cnt, nil
}

func (cs *ClassService) preloadStudentClassRelation(start int, length int) (int, error) {
	sql := "SELECT " +
		common.FIELD_USER_ID + "," +
		common.FIELD_CLASS_LIST +
		" FROM " +
		common.TABLE_USER_CLASS +
		" LIMIT " + strconv.Itoa(start) + "," + strconv.Itoa(length) + ";"

	rows, err := cs.db.Select(sql)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	cnt := 0

	id := 0
	cl := ""
	for rows.Next() {
		err = rows.Scan(&id, &cl)
		if err != nil {
			return cnt, err
		}

		// Append this field to the user Hash table.
		err = cs.cache.SetField(common.KEY_PREFIX_USER+strconv.Itoa(id), common.FIELD_CLASS_LIST, cl)
		if err != nil {
			return cnt, err
		}

		cnt++
	}

	return cnt, nil
}

func (cs *ClassService) preloadKeeperClassRelation(start int, length int) (int, error) {
	sql := "SELECT " +
		common.FIELD_USER_ID + "," +
		common.FIELD_CLASS_ID + "," +
		common.FIELD_STUDENT_LIST +
		" FROM " +
		common.TABLE_KEEPER_CLASS +
		" LIMIT " + strconv.Itoa(start) + "," + strconv.Itoa(length) + ";"

	rows, err := cs.db.Select(sql)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	cnt := 0

	userID := 0
	classID := 0
	sl := ""
	for rows.Next() {
		err = rows.Scan(&userID, &classID, &sl)
		if err != nil {
			return cnt, err
		}

		// Append this field to the user Hash table.
		err = cs.cache.SetField(common.KEY_PREFIX_CLASS+strconv.Itoa(classID)+":"+common.FIELD_KEEPER, strconv.Itoa(userID), sl)
		if err != nil {
			return cnt, err
		}

		cnt++
	}

	return cnt, nil
}

//----------------------------------------------------------------------------
// Database: Required
// Cache   : Compatible

func (cs *ClassService) AddClass(name string, subjects []int, groupID int, gdCourseID int, template int, platformID int, platformData string, session *Session) (int, error) {
	// Check requirements.
	if cs.db == nil {
		return 0, common.ERR_NO_SERVICE
	}

	//----------------------------------------------------
	// Check inputs.

	gID := session.GroupID
	if gID == common.GROUP_ID_FOR_SYSTEM {
		gID = groupID
	}
	if gID < 4 {
		return 0, errors.New("Invalid group ID.")
	}

	if len(name) == 0 {
		return 0, errors.New("Empty class name.")
	}
	sName := common.Escape(name)
	if len(sName) == 0 {
		return 0, errors.New("Invalid class name.")
	}

	sSubjectList := ""
	if subjects != nil {
		sSubjectList = common.IntArrayToString(subjects)
	}

	if gdCourseID > 0 {
		// Check whether this Gd course ID exists.
		if _, err := cs.GetClassIDViaGdCourseID(gdCourseID); err == nil {
			return 0, common.ERR_DUPLICATED_GAODUN_COURSE_ID
		}
	}

	sGroupID := strconv.Itoa(gID)
	sGdCourseID := strconv.Itoa(gdCourseID)

	sTemplate := strconv.Itoa(template)

	sPlatformID := strconv.Itoa(platformID)
	sPlatformData := common.Escape(platformData)

	if (platformID == common.PLATFORM_ID_FOR_SALESMEN) && (len(sPlatformData) == 0) {
		//n, roomID := bokecc.CreateSaleRoom(name)
		//if n < 0 {
		//	return 0, errors.New("Failed to create a new meeting room for salesmen.")
		//}
		//sPlatformData = common.Escape(roomID)
	}

	//----------------------------------------------------
	// Step 1. Update database.

	timestamp := common.GetTimeString()
	sUpdater := strconv.Itoa(session.UserID)

	// Create a new class.
	sql := "INSERT INTO " + common.TABLE_CLASS + " (" +
		common.FIELD_NAME + "," +
		common.FIELD_SUBJECT_LIST + "," +
		common.FIELD_TEACHER_LIST + "," +
		common.FIELD_KEEPER_LIST + "," +
		common.FIELD_STUDENT_LIST + "," +
		common.FIELD_MEETING_LIST + "," +
		common.FIELD_DELETED + "," +
		common.FIELD_NUMBER_OF_FINISHED_MEETING + "," +
		common.FIELD_START_TIME + "," +
		common.FIELD_END_TIME + "," +
		common.FIELD_PARENT + "," +
		common.FIELD_GROUP_ID + "," +
		common.FIELD_GAODUN_COURSE_ID + "," +
		common.FIELD_TEMPLATE + "," +
		common.FIELD_PLATFORM_ID + "," +
		common.FIELD_PLATFORM_DATA + "," +
		common.FIELD_ALLY + "," +
		common.FIELD_UPDATE_IP + "," +
		common.FIELD_UPDATE_TIME + "," +
		common.FIELD_UPDATER +
		") VALUES ('" +
		sName + "','" + sSubjectList + "','','','','','',0," +
		timestamp + ",0,''," +
		sGroupID + "," +
		sGdCourseID + "," +
		sTemplate + "," +
		sPlatformID + ",'" +
		sPlatformData + "',0,'" +
		session.IP + "'," +
		timestamp + "," +
		sUpdater + ");"

	id, err := cs.db.Insert(sql, 1)
	if err != nil {
		return 0, err
	}

	//----------------------------------------------------
	// Step 2. Update cache.

	if cs.cache != nil {

		sID := strconv.FormatInt(id, 10)
		sGroupID := strconv.Itoa(gID)

		// Insert a new class into cache.
		key := common.KEY_PREFIX_CLASS + sID
		m := make(map[string]string)

		m[common.FIELD_NAME] = sName
		m[common.FIELD_SUBJECT_LIST] = sSubjectList
		m[common.FIELD_TEACHER_LIST] = ""
		m[common.FIELD_KEEPER_LIST] = ""
		m[common.FIELD_STUDENT_LIST] = ""
		m[common.FIELD_MEETING_LIST] = ""
		m[common.FIELD_DELETED] = ""
		m[common.FIELD_NUMBER_OF_FINISHED_MEETING] = "0"
		m[common.FIELD_START_TIME] = timestamp
		m[common.FIELD_END_TIME] = "0"
		m[common.FIELD_PARENT] = ""
		m[common.FIELD_GROUP_ID] = sGroupID
		m[common.FIELD_GAODUN_COURSE_ID] = sGdCourseID
		m[common.FIELD_TEMPLATE] = sTemplate
		m[common.FIELD_PLATFORM_ID] = sPlatformID
		m[common.FIELD_PLATFORM_DATA] = sPlatformData
		m[common.FIELD_ALLY] = "0"
		m[common.FIELD_UPDATE_IP] = session.IP
		m[common.FIELD_UPDATE_TIME] = timestamp
		m[common.FIELD_UPDATER] = sUpdater

		if err = cs.cache.SetFields(key, m); err != nil {
			return 0, err
		}

		if gdCourseID > 0 {
			if err = cs.cache.SetField(cs.gdCourseIDKey, sGdCourseID, sID); err != nil {
				// TODO:
			}
		}

		// Append the class to the group-class Hash table.
		if cs.cache.SetField(common.KEY_PREFIX_CLASS+common.KEY_PREFIX_GROUP+sGroupID, sID, "0"); err != nil {
			// TODO:
		}
	}

	return int(id), nil
}

//----------------------------------------------------------------------------
// Database: Required
// Cache   : Compatible

func (cs *ClassService) CloneClass(classID int, name string, platformID int, platformData string, session *Session) (int, error) {
	// Check requirements.
	if cs.db == nil {
		return 0, common.ERR_NO_SERVICE
	}

	// Check authority.
	ci, err := cs.GetClass(classID, session)
	if err != nil {
		return 0, err
	}

	// Create a new class.
	destClassID, err := cs.AddClass(name, ci.Subjects, ci.GroupID, 0, ci.Template, platformID, platformData, session)
	if err != nil {
		return destClassID, err
	}

	// Copy meetings to this class.
	ml := ""
	first := true
	for i := 0; i < len(ci.Meetings); i++ {
		meetingID, err := cs.ms.CloneMeeting(ci.Meetings[i], destClassID, session)
		if err != nil {
			return 0, err
		}

		if first {
			first = false
		} else {
			ml += ","
		}
		ml += strconv.Itoa(meetingID)
	}

	if err = (func() error {
		sDestClassID := strconv.Itoa(destClassID)

		// Update class info.
		sql := "UPDATE " +
			common.TABLE_CLASS +
			" SET " +
			common.FIELD_MEETING_LIST + "='" + ml + "'" +
			" WHERE " +
			common.FIELD_ID + "=" + sDestClassID + ";"

		if _, err = cs.db.Exec(sql); err != nil {
			return err
		}

		if cs.cache != nil {
			if err = cs.cache.SetField(common.KEY_PREFIX_CLASS+sDestClassID, common.FIELD_MEETING_LIST, ml); err != nil {
				return err
			}
		}

		return nil
	})(); err != nil {
		return destClassID, err
	}

	return destClassID, nil
}

//----------------------------------------------------------------------------
// Database: Required
// Cache   : Compatible

func (cs *ClassService) ChangeClass(classID int, name string, subjects []int, gdCourseID int, template int, platformID int, platformData string, session *Session) error {
	// Check requirements.
	if cs.db == nil {
		return common.ERR_NO_SERVICE
	}

	ci, err := cs.GetClass(classID, session)
	if err != nil {
		return err
	}
	if ci.EndTime > 0 {
		return common.ERR_CLASS_CLOSED
	}

	sClassID := strconv.Itoa(classID)

	// Check class name.
	changeName := false
	sName := common.Escape(name)
	if ci.Name != sName {
		changeName = true
	}

	// Check subjects.
	changeSubject := false
	sSubjectList := ""
	if subjects != nil {
		changeSubject = true
		sSubjectList = common.IntArrayToString(subjects)
	}

	// Check Gd course ID.
	changeGdCourseID := false
	sGdCourseID := ""
	if (gdCourseID >= 0) && (ci.GdCourseID != gdCourseID) {
		if gdCourseID > 0 {
			// Check whether this Gd course ID exists.
			if n, err := cs.GetClassIDViaGdCourseID(gdCourseID); (err == nil) && (n != classID) {
				return common.ERR_DUPLICATED_GAODUN_COURSE_ID
			}
		}
		changeGdCourseID = true
		sGdCourseID = strconv.Itoa(gdCourseID)
	}

	// Check template.
	changeTemplate := false
	sTemplate := ""
	if (template >= 0) && (ci.Template != template) {
		changeTemplate = true
		sTemplate = strconv.Itoa(template)
	}

	// Check platform.
	changePlatform := false
	sPlatformID := ""
	sPlatformData := common.Escape(platformData)
	if (platformID >= 0) && (ci.PlatformID != platformID || ci.PlatformData != sPlatformData) {
		changePlatform = true
		sPlatformID = strconv.Itoa(platformID)
	}

	if (!changeName) && (!changeSubject) && (!changeGdCourseID) && (!changePlatform) {
		// Nothing needs to be changed.
		return nil
	}

	sUpdateTime := common.GetTimeString()
	sUpdateIP := session.IP
	sUpdater := strconv.Itoa(session.UserID)

	// Update database.
	sql := "UPDATE " +
		common.TABLE_CLASS +
		" SET "
	if changeName {
		sql += common.FIELD_NAME + "='" + sName + "',"
	}
	if changeSubject {
		sql += common.FIELD_SUBJECT_LIST + "='" + sSubjectList + "',"
	}
	if changeGdCourseID {
		sql += common.FIELD_GAODUN_COURSE_ID + "=" + sGdCourseID + ","
	}
	if changeTemplate {
		sql += common.FIELD_TEMPLATE + "=" + sTemplate + ","
	}
	if changePlatform {
		sql += common.FIELD_PLATFORM_ID + "=" + sPlatformID + "," +
			common.FIELD_PLATFORM_DATA + "='" + sPlatformData + "',"
	}
	sql += common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
		common.FIELD_UPDATE_TIME + "=" + sUpdateTime + "," +
		common.FIELD_UPDATER + "=" + sUpdater +
		" WHERE " +
		common.FIELD_ID + "=" + sClassID + ";"

	_, err = cs.db.Exec(sql)
	if err != nil {
		return err
	}

	// Update cache.
	if cs.cache != nil {
		// Step 1.
		m := make(map[string]string)
		if changeName {
			m[common.FIELD_NAME] = sName
		}
		if changeSubject {
			m[common.FIELD_SUBJECT_LIST] = sSubjectList
		}
		if changeGdCourseID {
			m[common.FIELD_GAODUN_COURSE_ID] = sGdCourseID
		}
		if changeTemplate {
			m[common.FIELD_TEMPLATE] = sTemplate
		}
		if changePlatform {
			m[common.FIELD_PLATFORM_ID] = sPlatformID
			m[common.FIELD_PLATFORM_DATA] = sPlatformData
		}
		m[common.FIELD_UPDATE_IP] = sUpdateIP
		m[common.FIELD_UPDATE_TIME] = sUpdateTime
		m[common.FIELD_UPDATER] = sUpdater

		err = cs.cache.SetFields(common.KEY_PREFIX_CLASS+sClassID, m)
		if err != nil {
			return err
		}
	}

	if changeName {
		//if platformID == common.PLATFORM_ID_FOR_SALESMEN {
		//	// Notify the vendor.
		//	bokecc.UpdateSaleRoom(platformData, name)
		//}
	}
	if changeGdCourseID {
		// Remove original index.
		if ci.GdCourseID > 0 {
			if err = cs.cache.DelField(cs.gdCourseIDKey, strconv.Itoa(ci.GdCourseID)); err != nil {
				// TODO:
			}
			if err = cs.cache.Del(cs.gdCourseIDKey + ":" + strconv.Itoa(ci.GdCourseID)); err != nil {
				// TODO:
			}
		}
		// Initiate current index.
		if gdCourseID > 0 {
			if err = cs.cache.SetField(cs.gdCourseIDKey, sGdCourseID, sClassID); err != nil {
				// TODO:
			}
		}
	}

	return nil
}

//----------------------------------------------------------------------------
// Database: Required
// Cache   : Compatible

func (cs *ClassService) EndClass(classID int, session *Session) error {
	// Check requirements.
	if cs.db == nil {
		return common.ERR_NO_SERVICE
	}

	// Check authority.
	ci, err := cs.GetClass(classID, session)
	if err != nil {
		return common.ERR_NO_AUTHORITY
	}
	if ci.EndTime > 0 {
		return common.ERR_CLASS_CLOSED
	}

	//----------------------------------------------------
	// Step 1. Update database.

	timestamp := common.GetTimeString()
	sGroupID := strconv.Itoa(ci.GroupID)
	sClassID := strconv.Itoa(classID)

	// Set the end time of this class.
	sql := "UPDATE " +
		common.TABLE_CLASS +
		" SET " +
		common.FIELD_END_TIME + "=" + timestamp + "," +
		common.FIELD_UPDATE_TIME + "=" + timestamp + "," +
		common.FIELD_UPDATE_IP + "='" + session.IP + "'," +
		common.FIELD_UPDATER + "=" + strconv.Itoa(session.UserID) +
		" WHERE " +
		common.FIELD_ID + "=" + sClassID +
		" AND " + common.FIELD_GROUP_ID + "=" + sGroupID + ";"

	_, err = cs.db.Exec(sql)
	if err != nil {
		return err
	}

	//----------------------------------------------------
	// Step 2. Update cache.

	if cs.cache != nil {

		// Set the end time of this class.

		m := make(map[string]string)
		m[common.FIELD_END_TIME] = timestamp
		m[common.FIELD_UPDATE_IP] = session.IP
		m[common.FIELD_UPDATE_TIME] = timestamp
		m[common.FIELD_UPDATER] = strconv.Itoa(session.UserID)

		key := common.KEY_PREFIX_CLASS + sClassID
		err = cs.cache.SetFields(key, m)
		if err != nil {
			return err
		}
	}

	return nil
}

//----------------------------------------------------------------------------
// Database: Required
// Cache   : Compatible

func (cs *ClassService) DeleteClass(classID int, session *Session) error {
	// Check requirements.
	if cs.db == nil {
		return common.ERR_NO_SERVICE
	}

	ci, err := cs.GetClass(classID, session)
	if err != nil {
		return err
	}
	if ci.PlatformID == common.PLATFORM_ID_FOR_PACKAGE {
		if len(ci.PlatformData) > 0 {
			return common.ERR_CLASS_IS_NOT_EMPTY
		}
	} else {
		if len(ci.Meetings) > 0 {
			return common.ERR_CLASS_IS_NOT_EMPTY
		}
	}

	sClassID := strconv.Itoa(classID)
	sGroupID := strconv.Itoa(ci.GroupID)

	// Delete the class itself.
	if err = (func() error {
		s := "DELETE FROM " +
			common.TABLE_CLASS +
			" WHERE " +
			common.FIELD_ID + "=" + sClassID + ";"

		if _, err := cs.db.Exec(s); err != nil {
			return err
		}

		// Delete this class from cache.
		if cs.cache != nil {
			if err = cs.cache.Del(common.KEY_PREFIX_CLASS + sClassID); err != nil {
				return err
			}
		}

		return nil
	})(); err != nil {
		return err
	}

	// Delete this class from group's class list.
	if cs.cache != nil {
		if err = (func() error {
			if err = cs.cache.DelField(common.KEY_PREFIX_CLASS+common.KEY_PREFIX_GROUP+sGroupID, sClassID); err != nil {
				// TODO:
			}

			return nil
		})(); err != nil {
			return err
		}
	}

	// Delete this class from users' class lists.
	users := make([]int, len(ci.Students)+len(ci.Teachers))
	for i := 0; i < len(ci.Students); i++ {
		users[i] = ci.Students[i]
	}
	for i := 0; i < len(ci.Teachers); i++ {
		users[len(ci.Students)+i] = ci.Teachers[i]
	}
	for i := 0; i < len(users); i++ {
		sUserID := strconv.Itoa(users[i])

		// Get existing value.
		cl, err := (func() (string, error) {
			s := "SELECT " +
				common.FIELD_CLASS_LIST +
				" FROM " +
				common.TABLE_USER_CLASS +
				" WHERE " +
				common.FIELD_USER_ID + "=" + sUserID + ";"

			rows, err := cs.db.Select(s)
			if err != nil {
				return "", err
			}
			defer rows.Close()

			if !rows.Next() {
				return "", nil
			}

			cl := ""
			if err = rows.Scan(&cl); err != nil {
				return "", err
			}

			return cl, nil
		})()
		if err != nil {
			continue
		}

		// Compute new value.
		cl, changed := common.DeleteFromList(sClassID, cl)
		if !changed {
			continue
		}

		// Update database.
		s := "UPDATE " +
			common.TABLE_USER_CLASS +
			" SET " +
			common.FIELD_CLASS_LIST + "='" + cl + "'" +
			" WHERE " +
			common.FIELD_USER_ID + "=" + sUserID + ";"
		if _, err = cs.db.Exec(s); err != nil {
			continue
		}

		// Update cache.
		if cs.cache != nil {
			if err = cs.cache.SetField(common.KEY_PREFIX_USER+sUserID, common.FIELD_CLASS_LIST, cl); err != nil {
				continue
			}
		}
	}

	// Permanently delete meetings.
	for i := 0; i < len(ci.Deleted); i++ {
		cs.ms.DeleteMeeting(ci.Deleted[i], session)
	}

	return nil
}

//----------------------------------------------------------------------------

func (cs *ClassService) PublishClass(classID int, session *Session) error {
	if cs.oss == nil {
		return common.ERR_NO_SERVICE
	}

	ci, err := cs.GetClass(classID, session)
	if err != nil {
		return err
	}

	s := `{` +
		`"` + common.FIELD_NAME + `":"` + common.UnescapeForJSON(ci.Name) + `",`
	// `"` + common.FIELD_PLATFORM_ID + `":` + strconv.Itoa(ci.PlatformID) + `,` +
	// `"` + common.FIELD_PLATFORM_DATA + `":"` + common.UnescapeForJSON(ci.PlatformData) + `",` +
	// `"` + common.FIELD_TEMPLATE + `":` + strconv.Itoa(ci.Template) + `,`

	sTeachers := ``
	teacherCnt := 0

	sMeetings := ``
	firstMeeting := true

	for i := 0; i < len(ci.Meetings); i++ {
		// Get the meeting.
		mi, err := cs.ms.GetMeeting(ci.Meetings[i], session, false)
		if err != nil {
			return err
		}

		// Get its configuration.
		cfg, err := cs.cache.GetField(common.KEY_PREFIX_MEETING+strconv.Itoa(mi.ID), common.FIELD_CONFIG)
		if err != nil || len(cfg) == 0 {
			continue
		}
		// fmt.Println(mi.ID)
		// fmt.Println(cfg)

		cfg = common.Unescape(cfg)
		// cfg = strings.Replace(cfg, `/`, `-`, -1)
		cfg, err = common.DecompressFromEncodedUriComponent(cfg)
		if err != nil {
			continue
		}
		// fmt.Println(cfg)

		// Save teacher info.
		tl, err := (func() (string, error) {
			obj := make(map[string]interface{})
			if err := json.Unmarshal(([]byte)(cfg), &obj); err != nil {
				return "", err
			}
			teachers, okay := obj["teacher"].([]interface{})
			if !okay {
				return "", common.ERR_INVALID_USER
			}

			ls := ``
			first := true
			for i := 0; i < len(teachers); i++ {
				teacher, okay := teachers[i].(map[string]interface{})
				if !okay {
					continue
				}

				name, okay := teacher["name"].(string)
				if !okay {
					continue
				}
				brief, okay := teacher["desc"].(string)
				if !okay {
					continue
				}
				avantar, okay := teacher["headLink"].(string)
				if !okay {
					continue
				}
				if strings.HasPrefix(avantar, "http:") {
					avantar = avantar[5:]
				} else if strings.HasPrefix(avantar, "https:") {
					avantar = avantar[6:]
				}
				weixin, okay := teacher["weChat"].(string)
				if !okay {
					continue
				}

				teacherCnt++
				if teacherCnt > 1 {
					sTeachers += `,`
				}
				sTeachers += `"` + strconv.Itoa(teacherCnt) + `":{` +
					`"` + common.FIELD_NAME + `":"` + common.ReplaceForJSON(name) + `",` +
					`"` + common.FIELD_BRIEF + `":"` + common.ReplaceForJSON(brief) + `",` +
					`"` + common.FIELD_AVANTAR + `":"` + common.ReplaceForJSON(avantar) + `",` +
					`"` + common.FIELD_WEIXIN + `":"` + common.ReplaceForJSON(weixin) + `"` +
					`}`

				if first {
					first = false
				} else {
					ls += `,`
				}
				ls += strconv.Itoa(teacherCnt)
			}

			return ls, nil
		})()
		if err != nil {
			return err
		}

		// Save meeting info.
		if firstMeeting {
			firstMeeting = false
		} else {
			sMeetings += `,`
		}
		sMeetings += `{` +
			`"` + common.FIELD_ID + `":` + strconv.Itoa(mi.ID) + `,` +
			`"` + common.FIELD_NAME + `":"` + common.UnescapeForJSON(mi.Name) + `",` +
			`"` + common.FIELD_START_TIME + `":` + strconv.Itoa(mi.StartTime*1000) + `,` +
			`"` + common.FIELD_DURATION + `":` + strconv.Itoa(mi.Duration) + `,` +
			`"` + common.FIELD_END_TIME + `":` + strconv.Itoa(mi.EndTime*1000) + `,` +
			`"` + common.FIELD_DATA + `":"` + common.UnescapeForJSON(mi.Data) + `",` +
			`"` + common.FIELD_TEACHER + `":[` + tl + `],` +
			`"` + common.FIELD_REPLAY + `":` + common.StringArrayToJSON(mi.Replays) +
			`}`
	}

	s += `"` + common.FIELD_TEACHER + `":{` + sTeachers + `},` +
		`"` + common.FIELD_MEETING + `":[` + sMeetings + `]}`

	// fmt.Println(s)
	if err = cs.oss.UploadString("class/"+strconv.Itoa(ci.ID)+".js", s); err != nil {
		return err
	}

	return nil
}

//----------------------------------------------------------------------------

func (cs *ClassService) PackageClass(classID int, session *Session) error {
	if cs.oss == nil {
		return common.ERR_NO_SERVICE
	}

	s, err := cs.QueryUserProgresses(classID, session)
	if err != nil {
		return err
	}

	if err = cs.oss.UploadString("class/progress/"+strconv.Itoa(classID)+".js", "{"+s+"}"); err != nil {
		return err
	}

	return nil
}

//----------------------------------------------------------------------------
// Database: Compatible
// Cache   : Compatible

func (cs *ClassService) GetClass(classID int, session *Session) (*ClassInfo, error) {
	sClassID := strconv.Itoa(classID)

	var ci *ClassInfo = nil

	if cs.cache != nil {
		if m, err := cs.cache.GetAllFields(common.KEY_PREFIX_CLASS + sClassID); err == nil {
			ci = NewClassInfoFromMap(m, classID)
		}
	}

	if (ci == nil) && (cs.db != nil) {
		sql := "SELECT " +
			common.FIELD_NAME + "," +
			common.FIELD_SUBJECT_LIST + "," +
			common.FIELD_TEACHER_LIST + "," +
			common.FIELD_KEEPER_LIST + "," +
			common.FIELD_STUDENT_LIST + "," +
			common.FIELD_MEETING_LIST + "," +
			common.FIELD_DELETED + "," +
			common.FIELD_NUMBER_OF_FINISHED_MEETING + "," +
			common.FIELD_START_TIME + "," +
			common.FIELD_END_TIME + "," +
			common.FIELD_PARENT + "," +
			common.FIELD_GROUP_ID + "," +
			common.FIELD_GAODUN_COURSE_ID + "," +
			common.FIELD_TEMPLATE + "," +
			common.FIELD_PLATFORM_ID + "," +
			common.FIELD_PLATFORM_DATA + "," +
			common.FIELD_ALLY + "," +
			common.FIELD_UPDATE_IP + "," +
			common.FIELD_UPDATE_TIME + "," +
			common.FIELD_UPDATER +
			" FROM " +
			common.TABLE_CLASS +
			" WHERE " +
			common.FIELD_ID + "=" + sClassID + ";"

		rows, err := cs.db.Select(sql)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		if !rows.Next() {
			return nil, common.ERR_NO_CLASS
		}

		ci = new(ClassInfo)
		ci.ID = classID

		sbl := ""
		tl := ""
		kl := ""
		sl := ""
		ml := ""
		dl := ""
		parent := ""
		err = rows.Scan(
			&ci.Name,
			&sbl,
			&tl, &kl, &sl,
			&ml, &dl,
			&ci.NumberOfFinishedMeeting,
			&ci.StartTime, &ci.EndTime,
			&parent,
			&ci.GroupID,
			&ci.GdCourseID,
			&ci.Template,
			&ci.PlatformID, &ci.PlatformData,
			&ci.Ally,
			&ci.UpdateIP, &ci.UpdateTime, &ci.Updater)

		if err != nil {
			return nil, err
		}

		ci.Subjects = common.StringToIntArray(sbl)
		ci.Teachers = common.StringToIntArray(tl)
		ci.Keepers = kl
		ci.Students = common.StringToIntArray(sl)
		ci.Meetings = common.StringToIntArray(ml)
		ci.Deleted = common.StringToIntArray(dl)
		ci.Parent = common.StringToIntArray(parent)

		if cs.cache != nil {
			m := make(map[string]string)

			m[common.FIELD_NAME] = ci.Name
			m[common.FIELD_SUBJECT_LIST] = sbl
			m[common.FIELD_TEACHER_LIST] = tl
			m[common.FIELD_KEEPER_LIST] = kl
			m[common.FIELD_STUDENT_LIST] = sl
			m[common.FIELD_MEETING_LIST] = ml
			m[common.FIELD_DELETED] = dl
			m[common.FIELD_NUMBER_OF_FINISHED_MEETING] = strconv.Itoa(ci.NumberOfFinishedMeeting)
			m[common.FIELD_PARENT] = parent
			m[common.FIELD_GROUP_ID] = strconv.Itoa(ci.GroupID)
			m[common.FIELD_GAODUN_COURSE_ID] = strconv.Itoa(ci.GdCourseID)
			m[common.FIELD_TEMPLATE] = strconv.Itoa(ci.Template)
			m[common.FIELD_PLATFORM_ID] = strconv.Itoa(ci.PlatformID)
			m[common.FIELD_PLATFORM_DATA] = ci.PlatformData
			m[common.FIELD_ALLY] = strconv.Itoa(ci.Ally)
			m[common.FIELD_START_TIME] = strconv.Itoa(ci.StartTime)
			m[common.FIELD_END_TIME] = strconv.Itoa(ci.EndTime)
			m[common.FIELD_UPDATE_TIME] = strconv.Itoa(ci.UpdateTime)
			m[common.FIELD_UPDATE_IP] = ci.UpdateIP
			m[common.FIELD_UPDATER] = strconv.Itoa(ci.Updater)

			err = cs.cache.SetFields(common.KEY_PREFIX_CLASS+sClassID, m)
			if err != nil {
				return nil, err
			}
		}
	}

	// Check result.
	if ci == nil {
		return nil, common.ERR_NO_SERVICE
	}

	// Check authority.
	okay := (func() bool {
		if session.IsSystem() {
			return true
		} else if session.IsAssistant() {
			if session.GroupID == ci.GroupID {
				return true
			}
		} else {
			if session.UserID >= common.VALUE_MINIMAL_TEMPERARY_USER_ID {
				return true
			}

			cl, err := cs.GetUserClassRelation(session.UserID)
			if err == nil {
				if common.InList(strconv.Itoa(ci.ID), cl) {
					return true
				}
			}
		}

		return false
	})()
	if !okay {
		return nil, common.ERR_NO_AUTHORITY
	}

	return ci, nil
}

//----------------------------------------------------------------------------
// Database: Compatible
// Cache   : Compatible

func (cs *ClassService) GetClasses(groupID int, session *Session) (ClassInfoSlice, int, error) {
	// Get class IDs.
	var arr []int = nil
	if session.IsStudent() || session.IsTeacher() || session.IsKeeper() {
		if session.UserID >= common.VALUE_MINIMAL_TEMPERARY_USER_ID {
			arr = cs.queryClassListForExperienceUser(session.UserID)
		} else {
			cl, err := cs.GetUserClassRelation(session.UserID)
			if err == nil {
				// return nil, err
				arr = common.StringToIntArray(cl)
			}
		}
	} else if session.IsAssistantOrAbove() {
		if cs.cache == nil {
			return nil, -1, common.ERR_NO_SERVICE
		}

		m, err := cs.cache.GetAllFields(common.KEY_PREFIX_CLASS + common.KEY_PREFIX_GROUP + strconv.Itoa(groupID))
		if err == nil {
			arr = make([]int, len(m))
			i := 0
			for k, _ := range m {
				arr[i], _ = strconv.Atoi(k)
				i++
			}
		}
	} else {
		return nil, -2, common.ERR_NO_AUTHORITY
	}

	if arr == nil {
		return ClassInfoSlice([]*ClassInfo{}), 0, nil
	}

	// Get classes themselves.
	result := make([]*ClassInfo, len(arr))
	for i := 0; i < len(arr); i++ {
		ci, err := cs.GetClass(arr[i], session)
		if err != nil {
			result[i] = nil
			continue
		}

		result[i] = ci
	}

	return ClassInfoSlice(result), 0, nil
}

//----------------------------------------------------------------------------
// Database: Compatible
// Cache   : Compatible

func (cs *ClassService) GetUserClassRelation(userID int) (string, error) {
	sUserID := strconv.Itoa(userID)

	if cs.cache != nil {
		if cl, err := cs.cache.GetField(common.KEY_PREFIX_USER+sUserID, common.FIELD_CLASS_LIST); err == nil {
			return cl, nil
		}
	}

	if cs.db != nil {
		sql := "SELECT " +
			common.FIELD_CLASS_LIST +
			" FROM " +
			common.TABLE_USER_CLASS +
			" WHERE " +
			common.FIELD_USER_ID + "=" + sUserID + ";"

		rows, err := cs.db.Select(sql)
		if err != nil {
			return "", err
		}
		defer rows.Close()

		if !rows.Next() {
			return "", common.ERR_NO_USER
		}

		cl := ""
		if err = rows.Scan(&cl); err != nil {
			return "", err
		}

		if cs.cache != nil {
			if err = cs.cache.SetField(common.KEY_PREFIX_USER+sUserID, common.FIELD_CLASS_LIST, cl); err != nil {
				// TODO:
			}
		}

		return cl, nil
	}

	return "", common.ERR_NO_SERVICE
}

//----------------------------------------------------------------------------

func (cs *ClassService) getPlatformInfo(classID int) (int, string, error) {
	sClassID := strconv.Itoa(classID)

	s, err := cs.cache.GetField(common.KEY_PREFIX_CLASS+sClassID, common.FIELD_PLATFORM_ID)
	if err != nil {
		return 0, "", err
	}
	id, err := strconv.Atoi(s)
	if err != nil {
		return 0, "", err
	}

	data, err := cs.cache.GetField(common.KEY_PREFIX_CLASS+sClassID, common.FIELD_PLATFORM_DATA)
	if err != nil {
		return 0, "", err
	}

	return id, data, nil
}

//----------------------------------------------------------------------------

func (cs *ClassService) SetAlly(classID int, ally int, session *Session) error {
	if cs.db == nil {
		return common.ERR_NO_SERVICE
	}

	ci, err := cs.GetClass(classID, session)
	if err != nil {
		return err
	}
	if ally > 0 {
		_, err = cs.GetClass(ally, session)
		if err != nil {
			return err
		}
	}
	if ci.Ally == ally {
		return nil
	}

	sClassID := strconv.Itoa(classID)
	sAlly := strconv.Itoa(ally)
	sUpdater := strconv.Itoa(session.UserID)
	sUpdateTime := common.GetTimeString()
	sUpdateIP := session.IP

	err = (func() error {
		s := "UPDATE " +
			common.TABLE_CLASS +
			" SET " +
			common.FIELD_ALLY + "=" + sAlly + "," +
			common.FIELD_UPDATER + "=" + sUpdater + "," +
			common.FIELD_UPDATE_TIME + "=" + sUpdateTime + "," +
			common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'" +
			" WHERE " +
			common.FIELD_ID + "=" + sClassID + ";"

		_, err := cs.db.Exec(s)
		if err != nil {
			return err
		}

		if cs.cache != nil {
			m := make(map[string]string)

			m[common.FIELD_ALLY] = sAlly
			m[common.FIELD_UPDATER] = sUpdater
			m[common.FIELD_UPDATE_IP] = sUpdateIP
			m[common.FIELD_UPDATE_TIME] = sUpdateTime

			err = cs.cache.SetFields(common.KEY_PREFIX_CLASS+sClassID, m)
			if err != nil {
				return err
			}
		}

		return nil
	})()
	if err != nil {
		return err
	}

	return nil
}

//----------------------------------------------------------------------------
// Database: Compatible.
// Cache   : Compatible.

func (cs *ClassService) GetClassIDViaGdCourseID(gdCourseID int) (int, error) {
	sGdCourseID := strconv.Itoa(gdCourseID)

	if cs.cache != nil {
		classID, err := (func() (int, error) {
			s, err := cs.cache.GetField(cs.gdCourseIDKey, sGdCourseID)
			if err != nil {
				return 0, err
			}

			id, err := strconv.Atoi(s)
			if err != nil {
				return 0, err
			}

			return id, nil
		})()
		if err == nil {
			return classID, nil
		}
	}

	if cs.db != nil {
		s := "SELECT " +
			common.FIELD_ID +
			" FROM " +
			common.TABLE_CLASS +
			" WHERE " +
			common.FIELD_GAODUN_COURSE_ID + "=" + sGdCourseID + ";"

		rows, err := cs.db.Select(s)
		if err != nil {
			return 0, err
		}
		defer rows.Close()

		if !rows.Next() {
			return 0, common.ERR_NO_RECORD
		}

		n := 0
		if err = rows.Scan(&n); err != nil {
			return 0, err
		}

		if cs.cache != nil {
			if err = cs.cache.SetField(cs.gdCourseIDKey, sGdCourseID, strconv.Itoa(n)); err != nil {
				// TODO:
			}
		}

		return n, nil
	}

	return 0, common.ERR_NO_SERVICE
}

//----------------------------------------------------------------------------
// Database: Compatible.
// Cache   : Compatible.

func (cs *ClassService) GetClassIDViaUserID(potentialParentClassID int, userID int) (int, int, error) {
	sParentClassID := strconv.Itoa(potentialParentClassID)
	sUserID := strconv.Itoa(userID)

	sChildren := ""
	sClassList := ""
	okay := false

	if cs.cache != nil {
		// Check whether it is really a parent class.
		sChildren, okay = (func() (string, bool) {
			s, err := cs.cache.GetField(common.KEY_PREFIX_CLASS+sParentClassID, common.FIELD_PLATFORM_ID)
			if err != nil {
				return "", false
			}

			n, err := strconv.Atoi(s)
			if err != nil {
				return "", false
			}

			if n != common.PLATFORM_ID_FOR_PACKAGE {
				return sParentClassID, true
			}

			// Get children of this parent class.
			s, err = cs.cache.GetField(common.KEY_PREFIX_CLASS+sParentClassID, common.FIELD_PLATFORM_DATA)
			if err != nil {
				return "", false
			}

			return s, true
		})()

		if okay {
			// Get class list of this user.
			sClassList, okay = (func() (string, bool) {
				s, err := cs.cache.GetField(common.KEY_PREFIX_USER+sUserID, common.FIELD_CLASS_LIST)
				if err != nil {
					return "", false
				}

				return s, true
			})()
		}
	}

	if (!okay) && (cs.db != nil) {
		s, status, err := (func() (string, int, error) {
			s := "SELECT " +
				common.FIELD_PLATFORM_ID + "," +
				common.FIELD_PLATFORM_DATA +
				" FROM " +
				common.TABLE_CLASS +
				" WHERE " +
				common.FIELD_ID + "=" + sParentClassID + ";"

			rows, err := cs.db.Select(s)
			if err != nil {
				return "", -1, err
			}
			defer rows.Close()

			if !rows.Next() {
				return "", -2, common.ERR_NO_RECORD
			}

			n := 0
			r := ""
			if err = rows.Scan(&n, &r); err != nil {
				return "", -3, err
			}

			if cs.cache != nil {
				if err = cs.cache.SetField(common.KEY_PREFIX_CLASS+sParentClassID, common.FIELD_PLATFORM_ID, strconv.Itoa(n)); err != nil {
					// TODO:
				}
				if err = cs.cache.SetField(common.KEY_PREFIX_CLASS+sParentClassID, common.FIELD_PLATFORM_DATA, r); err != nil {
					// TODO:
				}
			}

			if n == common.PLATFORM_ID_FOR_PACKAGE {
				return r, 0, nil
			}

			// It is not a parent class, return itself instead.
			return sParentClassID, 0, nil
		})()
		if err != nil {
			return 0, status, err
		} else {
			sChildren = s
		}

		s, status, err = (func() (string, int, error) {
			s := "SELECT " +
				common.FIELD_CLASS_LIST +
				" FROM " +
				common.TABLE_USER +
				" WHERE " +
				common.FIELD_ID + "=" + sUserID + ";"

			rows, err := cs.db.Select(s)
			if err != nil {
				return "", -5, err
			}
			defer rows.Close()

			if !rows.Next() {
				return "", -6, common.ERR_NO_RECORD
			}

			r := ""
			if err = rows.Scan(&r); err != nil {
				return "", -7, err
			}

			if cs.cache != nil {
				if err = cs.cache.SetField(common.KEY_PREFIX_USER+sUserID, common.FIELD_CLASS_LIST, r); err != nil {
					// TODO:
				}
			}

			return r, 0, nil
		})()
		if err != nil {
			return 0, status, err
		} else {
			sClassList = s
		}
	}

	sChildren = common.Unescape(sChildren)
	children := common.StringToIntArray(sChildren)

	classes := common.StringToIntArray(sClassList)

	// Find the first class within the intersection of children and classes.
	for i := 0; i < len(children); i++ {
		for j := 0; j < len(classes); j++ {
			if children[i] == classes[j] {
				return children[i], 0, nil
			}
		}
	}

	return 0, -8, common.ERR_NO_AUTHORITY
}

//----------------------------------------------------------------------------

func (cs *ClassService) CacheMeetingProgress(meetingID int, userID int) (int, error) {
	// cs.ms.GetMeetingProgresses(meetingID, session)
	return 0, nil
}

func (cs *ClassService) CacheClassProgress(classID int, userID int) (int, error) {
	return 0, nil
}

//----------------------------------------------------------------------------

func (cs *ClassService) GetClassOverallProgress(classID int, userID int) (int, error) {
	sClassID := strconv.Itoa(classID)
	sUserID := strconv.Itoa(userID)

	if cs.cache != nil {
		n, err := (func() (int, error) {
			// Get meeting IDs within this class.
			s, err := cs.cache.GetField(common.KEY_PREFIX_CLASS+sClassID, common.FIELD_MEETING_LIST)
			if err != nil {
				return 0, err
			}
			meetingIDs := common.StringToIntArray(s)

			examCnt := 0
			examTotal := 0
			videoCnt := 0
			videoTotal := 0

			for i := 0; i < len(meetingIDs); i++ {
				sMeetingID := strconv.Itoa(meetingIDs[i])

				// Get exam IDs within this meeting.
				examIDs, err := (func() (map[string]string, error) {
					s, err := cs.cache.GetField(common.KEY_PREFIX_MEETING+sMeetingID, common.FIELD_EXAM_LIST)
					if err != nil {
						return nil, err
					}

					m := make(map[string]string)

					arr := strings.Split(s, ",")
					for j := 0; j < len(arr); j++ {
						kvs := strings.Split(arr[j], ":")
						if len(kvs) != 2 {
							continue
						}

						m[kvs[0]] = kvs[1]
					}

					return m, nil
				})()
				if err == nil {
					examTotal += len(examIDs)

					// Get exam progresses for the specified user.
					n, err := (func() (int, error) {
						s, err := cs.cache.GetField(common.KEY_PREFIX_MEETING+sMeetingID+":"+sUserID, common.FIELD_EXAM)
						if err != nil {
							return 0, err
						}

						cnt := 0
						arr := strings.Split(s, ",")
						for j := 0; j < len(arr); j++ {
							kvs := strings.Split(arr[j], ":")
							if len(kvs) != 2 {
								continue
							}

							if _, okay := examIDs[kvs[0]]; okay {
								cnt++
							}
						}

						return cnt, nil
					})()
					if err == nil {
						examCnt += n
					}
				}

				// Get video IDs within this meeting.
				videoIDs, err := (func() (map[string]string, error) {
					s, err := cs.cache.GetField(common.KEY_PREFIX_MEETING+sMeetingID, common.FIELD_VIDEO_LIST)
					if err != nil {
						return nil, err
					}

					m := make(map[string]string)

					arr := strings.Split(s, ",")
					for j := 0; j < len(arr); j++ {
						kvs := strings.Split(arr[j], ":")
						if len(kvs) != 2 {
							continue
						}

						m[kvs[0]] = kvs[1]
					}

					return m, nil
				})()
				if err == nil {
					videoTotal += len(videoIDs)

					// Get video progresses for the specified user.
					n, err := (func() (int, error) {
						s, err := cs.cache.GetField(common.KEY_PREFIX_MEETING+sMeetingID+":"+sUserID, common.FIELD_VIDEO_A)
						if err != nil {
							return 0, err
						}

						cnt := 0
						arr := strings.Split(s, ",")
						for j := 0; j < len(arr); j++ {
							kvs := strings.Split(arr[j], ":")
							if len(kvs) != 2 {
								continue
							}

							if _, okay := videoIDs[kvs[0]]; okay {
								cnt++
							}
						}

						return cnt, nil
					})()
					if err == nil {
						videoCnt += n
					}
				}
			}

			if examTotal == 0 {
				if videoTotal == 0 {
					return 0, nil
				} else {
					return int(videoCnt * 100 / videoTotal), nil
				}
			} else {
				if videoTotal == 0 {
					return int(examCnt * 100 / examTotal), nil
				} else {
					return int(examCnt*20/examTotal) + int(videoCnt*80/videoTotal), nil
				}
			}
		})()
		if err == nil {
			return n, nil
		}
	}

	if cs.db != nil {
		// TODO:
	}

	return 0, nil
}

//----------------------------------------------------------------------------
// Database: Required

func (cs *ClassService) Count() (int, error) {
	if cs.db == nil {
		return 0, common.ERR_NO_DATABASE
	}

	return cs.db.Count(common.TABLE_CLASS)
}

//----------------------------------------------------------------------------
