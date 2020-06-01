package service

import (
	"container/list"
	"errors"
	"github.com/wangmhgo/go-project/gdun/common"
	"strconv"
	"strings"
)

//----------------------------------------------------------------------------

type UserService struct {
	db                 *common.Database
	cache              *common.Cache
	ss                 *SessionService
	loginKey           string
	gdStudentIDKey string
}

func NewUserService(db *common.Database, cache *common.Cache, ss *SessionService) (*UserService, error) {
	us := new(UserService)
	us.db = db
	us.cache = cache
	us.ss = ss

	us.loginKey = common.KEY_PREFIX_USER + common.FIELD_NAME
	us.gdStudentIDKey = common.KEY_PREFIX_USER + common.FIELD_GAODUN_STUDENT_ID

	err := us.Init()
	if err != nil {
		return nil, err
	}

	return us, nil
}

//----------------------------------------------------------------------------

func (us *UserService) Init() error {
	if us.db == nil {
		return common.ERR_NO_DATABASE
	}

	sql := "CREATE TABLE IF NOT EXISTS `" + common.TABLE_USER + "` ("
	sql += " `" + common.FIELD_ID + "` INT NOT NULL AUTO_INCREMENT,"
	sql += " `" + common.FIELD_NAME + "` VARCHAR(512) NOT NULL,"
	sql += " `" + common.FIELD_PASSWORD + "` VARCHAR(512) NOT NULL,"
	sql += " `" + common.FIELD_NICKNAME + "` VARCHAR(512) NOT NULL,"
	sql += " `" + common.FIELD_REMARK + "` VARCHAR(512) NOT NULL,"
	sql += " `" + common.FIELD_GAODUN_STUDENT_ID + "` INT DEFAULT 0,"
	sql += " `" + common.FIELD_MAIL + "` VARCHAR(128) NOT NULL DEFAULT '',"
	sql += " `" + common.FIELD_PHONE + "` VARCHAR(32) NOT NULL DEFAULT '',"
	sql += " `" + common.FIELD_QQ + "` VARCHAR(32) NOT NULL DEFAULT '',"
	sql += " `" + common.FIELD_WEIXIN + "` VARCHAR(128) NOT NULL DEFAULT '',"
	sql += " `" + common.FIELD_WEIBO + "` VARCHAR(128) NOT NULL DEFAULT '',"
	// sql += " `" + common.FIELD_PRIVILEGE + "` INT NOT NULL DEFAULT 0,"
	sql += " `" + common.FIELD_GROUP_ID + "` INT NOT NULL,"
	sql += " `" + common.FIELD_UPDATE_IP + "` VARCHAR(32) NOT NULL,"
	sql += " `" + common.FIELD_UPDATE_TIME + "` BIGINT NOT NULL,"
	sql += " `" + common.FIELD_UPDATER + "` INT NOT NULL,"
	sql += " PRIMARY KEY (`" + common.FIELD_ID + "`),"
	sql += " KEY (`" + common.FIELD_NAME + "`),"
	sql += " KEY (`" + common.FIELD_NICKNAME + "`),"
	sql += " KEY (`" + common.FIELD_GROUP_ID + "`)"
	sql += ") ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;"

	_, err := us.db.Exec(sql)
	return err
}

//----------------------------------------------------------------------------

func (us *UserService) Preload() (int, error) {
	if us.db == nil {
		return 0, common.ERR_NO_DATABASE
	}
	if us.cache == nil {
		return 0, common.ERR_NO_CACHE
	}

	rest, err := us.db.Count(common.TABLE_USER)
	if err != nil {
		return 0, err
	}

	i := 0
	for i < rest {
		n, err := us.preload(i, common.DATABASE_PRELOAD_SIZE)
		if err != nil {
			return i + n, err
		}
		i += n
	}

	return i, nil
}

func (us *UserService) preload(start int, length int) (int, error) {
	sql := "SELECT " +
		common.FIELD_ID + "," +
		common.FIELD_NAME + "," +
		common.FIELD_PASSWORD + "," +
		common.FIELD_NICKNAME + "," +
		common.FIELD_REMARK + "," +
		common.FIELD_WEIXIN + "," +
		// common.FIELD_PRIVILEGE + "," +
		common.FIELD_GROUP_ID + "," +
		common.FIELD_GAODUN_STUDENT_ID +
		" FROM " +
		common.TABLE_USER +
		" LIMIT " + strconv.Itoa(start) + "," + strconv.Itoa(length) + ";"
	rows, err := us.db.Select(sql)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	m := make(map[string]string)
	cnt := 0

	id := 0
	name := ""
	password := ""
	nickname := ""
	remark := ""
	weixin := ""
	// privilege := 0
	groupID := 0
	gdStudentID := 0
	for rows.Next() {
		err = rows.Scan(&id, &name, &password, &nickname, &remark, &weixin, &groupID, &gdStudentID)
		if err != nil {
			return cnt, err
		}

		// Index via name.
		if (name != common.VALUE_NOT_ALLOWED) && (password != common.VALUE_NOT_ALLOWED) {
			// This is a local user.
			err = us.cache.SetField(us.loginKey, common.Escape(common.Unescape(name)), strconv.Itoa(id))
			if err != nil {
				return cnt, err
			}
		}
		// Index via Gd student ID.
		if gdStudentID > 0 {
			err = us.cache.SetField(us.gdStudentIDKey, strconv.Itoa(gdStudentID), strconv.Itoa(id))
			if err != nil {
				return cnt, err
			}
		}

		// Save his info.
		m[common.FIELD_PASSWORD] = password
		m[common.FIELD_NICKNAME] = nickname
		m[common.FIELD_REMARK] = remark
		m[common.FIELD_WEIXIN] = weixin
		// m[common.FIELD_PRIVILEGE] = strconv.Itoa(privilege)
		m[common.FIELD_GROUP_ID] = strconv.Itoa(groupID)
		m[common.FIELD_GAODUN_STUDENT_ID] = strconv.Itoa(gdStudentID)

		err = us.cache.SetFields(common.KEY_PREFIX_USER+strconv.Itoa(id), m)
		if err != nil {
			return cnt, err
		}

		cnt++
	}

	return cnt, nil
}

//----------------------------------------------------------------------------

func (us *UserService) AddUser(name string, pwd string, nickname string, groupID int, ip string, updater int) (int, error) {
	// Save him to database.
	if us.db != nil {
		sName := common.Escape(name)
		sPwd := common.Escape(pwd)
		sNickname := common.Escape(nickname)
		sGroupID := strconv.Itoa(groupID)
		sUpdateTime := common.GetTimeString()
		sUpdateIP := ip
		sUpdater := strconv.Itoa(updater)

		sql := "INSERT INTO " +
			common.TABLE_USER +
			"(" +
			common.FIELD_NAME + "," +
			common.FIELD_PASSWORD + "," +
			common.FIELD_NICKNAME + "," +
			common.FIELD_REMARK + "," +
			// common.FIELD_PRIVILEGE + "," +
			common.FIELD_GROUP_ID + "," +
			common.FIELD_UPDATE_TIME + "," +
			common.FIELD_UPDATE_IP + "," +
			common.FIELD_UPDATER +
			") VALUES (" +
			"'" + sName + "'," +
			"'" + sPwd + "'," +
			"'" + sNickname + "'," +
			"''," +
			// "0," +
			sGroupID + "," +
			sUpdateTime + "," +
			"'" + sUpdateIP + "'," +
			sUpdater +
			");"

		userID, err := us.db.Insert(sql, 1)
		if err != nil {
			return 0, err
		}

		if us.cache != nil {
			sUserID := strconv.FormatInt(userID, 10)

			err = us.cache.SetField(us.loginKey, sName, sUserID)
			if err != nil {
				return 0, err
			}

			m := make(map[string]string)
			m[common.FIELD_PASSWORD] = sPwd
			m[common.FIELD_NICKNAME] = sNickname
			m[common.FIELD_REMARK] = ""
			// m[common.FIELD_PRIVILEGE] = "0"
			m[common.FIELD_GROUP_ID] = sGroupID

			err = us.cache.SetFields(common.KEY_PREFIX_USER+sUserID, m)
			if err != nil {
				return 0, err
			}
		}

		return int(userID), nil
	}

	return 0, common.ERR_NO_SERVICE
}

//----------------------------------------------------------------------------

func (us *UserService) RegisterViaInvitationToken(name string, pwd string, token string, ip string) error {
	if len(name) == 0 {
		return errors.New("Empty name.")
	}
	sName := common.Escape(name)
	if len(sName) == 0 || sName == common.VALUE_NOT_ALLOWED {
		return errors.New("Invalid name.")
	}

	if len(pwd) == 0 {
		return errors.New("Empty password.")
	}
	sPwd := common.Escape(pwd)
	if len(sPwd) == 0 || sPwd == common.VALUE_NOT_ALLOWED {
		return errors.New("Invalid password.")
	}

	// Check whether this user exists.
	if _, err := us.cache.GetField(us.loginKey, sName); err == nil {
		return common.ERR_DUPLICATED_USER
	}

	// Check the invitation token.
	groupID, err := us.UseInvitationToken(token)
	if err != nil {
		return err
	}

	_, err = us.AddUser(name, pwd, name, groupID, ip, 0)
	return err
}

//----------------------------------------------------------------------------

func (us *UserService) Login(name string, pwd string, ip string) (*UserInfo, error) {
	if len(name) == 0 {
		return nil, errors.New("Empty name.")
	}
	sName := common.Escape(name)
	if len(sName) == 0 {
		return nil, errors.New("Invalid name.")
	}
	if sName == common.VALUE_NOT_ALLOWED {
		return nil, errors.New("Login is not allowed.")
	}

	if len(pwd) == 0 {
		return nil, errors.New("Empty password.")
	}
	sPwd := common.Escape(pwd)
	if len(sPwd) == 0 {
		return nil, errors.New("Invalid password.")
	}
	if sPwd == common.VALUE_NOT_ALLOWED {
		return nil, errors.New("Login is not allowed.")
	}

	if us.cache != nil {
		// Get his user ID via cache.
		s, err := us.cache.GetField(us.loginKey, sName)
		if err != nil {
			return nil, errors.New("Invalid user name.")
		}
		userID, err := strconv.Atoi(s)
		if err != nil {
			return nil, errors.New("Invalid user ID.")
		}

		// Get his personal info via cache.
		m, err := us.cache.GetAllFields(common.KEY_PREFIX_USER + strconv.Itoa(userID))
		if err != nil {
			return nil, err
		}

		// Get his password.
		s, okay := m[common.FIELD_PASSWORD]
		if !okay {
			return nil, errors.New("No password existing.")
		}
		if sPwd != s {
			return nil, errors.New("Invalid password.")
		}

		// Get his group ID.
		s, okay = m[common.FIELD_GROUP_ID]
		if !okay {
			return nil, errors.New("Empty group ID.")
		}
		groupID, err := strconv.Atoi(s)
		if err != nil {
			return nil, errors.New("Invalid group ID.")
		}

		// Get his nickname.
		nickname, okay := m[common.FIELD_NICKNAME]
		if !okay {
			return nil, errors.New("Empty nickname.")
		}

		ui := new(UserInfo)
		ui.ID = userID
		ui.Nickname = nickname
		ui.GroupID = groupID

		return ui, nil
	}

	if us.db != nil {
		// Check whether this user exists.
		sql := "SELECT " +
			common.FIELD_ID + "," +
			common.FIELD_NICKNAME + "," +
			common.FIELD_GROUP_ID +
			" FROM " + common.TABLE_USER +
			" WHERE " +
			common.FIELD_NAME + "='" + sName + "' AND " +
			common.FIELD_PASSWORD + "='" + sPwd + "';"

		rows, err := us.db.Select(sql)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		if !rows.Next() {
			return nil, errors.New("Login failed.")
		}

		ui := new(UserInfo)
		err = rows.Scan(&ui.ID, &ui.Nickname, &ui.GroupID)
		if err != nil {
			return nil, err
		}

		ui.Nickname = common.Unescape(ui.Nickname)
		return ui, nil
	}

	return nil, common.ERR_NO_SERVICE
}

//----------------------------------------------------------------------------

func (us *UserService) ChangePassword(userID int, oldPwd string, pwd string) error {
	if userID < 1 {
		return common.ERR_INVALID_USER
	}

	if len(oldPwd) == 0 {
		return errors.New("Empty old password.")
	}
	sOldPwd := common.Escape(oldPwd)
	if len(sOldPwd) == 0 {
		return errors.New("Empty old password.")
	}
	if sOldPwd == common.VALUE_NOT_ALLOWED {
		return common.ERR_NO_AUTHORITY
	}

	if len(pwd) == 0 {
		return errors.New("Empty new password.")
	}
	sPwd := common.Escape(pwd)
	if len(sPwd) == 0 {
		return errors.New("Empty new password.")
	}
	if sPwd == common.VALUE_NOT_ALLOWED {
		return errors.New("Invalid new password.")
	}

	okay := false

	if us.cache != nil {
		key := common.KEY_PREFIX_USER + strconv.Itoa(userID)

		s, err := us.cache.GetField(key, common.FIELD_PASSWORD)
		if err != nil {
			return err
		}

		if s != sOldPwd {
			return errors.New("Invalid old password.")
		}

		err = us.cache.SetField(key, common.FIELD_PASSWORD, sPwd)
		if err != nil {
			return err
		}

		okay = true
	}

	if us.db != nil {
		sql := "UPDATE " +
			common.TABLE_USER +
			" SET " +
			common.FIELD_PASSWORD + "='" + sPwd +
			"' WHERE " +
			common.FIELD_ID + "=" + strconv.Itoa(userID) +
			" AND " + common.FIELD_PASSWORD + "='" + sOldPwd + "'" + ";"

		_, err := us.db.Exec(sql)
		if err != nil {
			return err
		}

		okay = true
	}

	if okay {
		return nil
	} else {
		return common.ERR_NO_SERVICE
	}
}

//----------------------------------------------------------------------------

func (us *UserService) ChangeProfile(userID int, nickname string, mail string, phone string, qq string, weixin string, weibo string, session *Session) error {
	if len(nickname) == 0 {
		return common.ERR_INVALID_NICKNAME
	}

	okay := false

	sNickname := common.Escape(nickname)

	if us.cache != nil {
		err := us.cache.SetField(common.KEY_PREFIX_USER+strconv.Itoa(userID), common.FIELD_NICKNAME, sNickname)
		if err != nil {
			return err
		}

		okay = true
	}

	if us.db != nil {
		sMail := common.Escape(mail)
		if len(sMail) >= 124 {
			return errors.New("The mail is too long.")
		}

		sPhone := common.Escape(phone)
		if len(sPhone) >= 28 {
			return errors.New("The phone number is too long.")
		}

		sQQ := common.Escape(qq)
		if len(sQQ) >= 28 {
			return errors.New("The QQ number is too long.")
		}

		sWeixin := common.Escape(weixin)
		if len(sWeixin) >= 124 {
			return errors.New("The Weixin is too long.")
		}

		sWeibo := common.Escape(weibo)
		if len(sWeibo) >= 124 {
			return errors.New("The Weibo is too long.")
		}

		sql := "UPDATE " +
			common.TABLE_USER +
			" SET " +
			common.FIELD_NICKNAME + "='" + sNickname + "'," +
			common.FIELD_MAIL + "='" + sMail + "'," +
			common.FIELD_PHONE + "='" + sPhone + "'," +
			common.FIELD_QQ + "='" + sQQ + "'," +
			common.FIELD_WEIXIN + "='" + sWeixin + "'," +
			common.FIELD_WEIBO + "='" + sWeibo + "'," +
			common.FIELD_UPDATE_IP + "='" + session.IP + "'," +
			common.FIELD_UPDATE_TIME + "=" + common.GetTimeString() +
			" WHERE " +
			common.FIELD_ID + "=" + strconv.Itoa(userID) + ";"

		_, err := us.db.Exec(sql)
		if err != nil {
			return err
		}

		okay = true
	}

	if err := us.ss.UpdateSessionNickname(session, sNickname); err != nil {
		return err
	}

	if okay {
		return nil
	} else {
		return common.ERR_NO_SERVICE
	}
}

//----------------------------------------------------------------------------

func (us *UserService) Remark(userID int, name string, remark string, session *Session) error {
	if us.db == nil {
		return common.ERR_NO_SERVICE
	}

	sName := common.Escape(name)
	sRemark := common.Escape(remark)
	if (len(sName) == 0) && (len(sRemark) == 0) {
		return nil
	}

	sUserID := strconv.Itoa(userID)
	sUpdateIP := session.IP
	sUpdater := strconv.Itoa(session.UserID)
	sUpdateTime := common.GetTimeString()

	s := "UPDATE " +
		common.TABLE_USER +
		" SET "
	if len(sName) > 0 {
		s += common.FIELD_NICKNAME + "='" + sName + "',"
	}
	if len(sRemark) > 0 {
		s += common.FIELD_REMARK + "='" + sRemark + "',"
	}
	s += common.FIELD_UPDATE_IP + "='" + sUpdateIP + "'," +
		common.FIELD_UPDATE_TIME + "=" + sUpdateTime + "," +
		common.FIELD_UPDATER + "=" + sUpdater +
		" WHERE " +
		common.FIELD_ID + "=" + sUserID + ";"

	if _, err := us.db.Exec(s); err != nil {
		return err
	}

	if us.cache != nil {
		m := make(map[string]string)
		if len(sName) > 0 {
			m[common.FIELD_NICKNAME] = sName
		}
		if len(sRemark) > 0 {
			m[common.FIELD_REMARK] = sRemark
		}
		m[common.FIELD_UPDATE_IP] = sUpdateIP
		m[common.FIELD_UPDATE_TIME] = sUpdateTime
		m[common.FIELD_UPDATER] = sUpdater

		if err := us.cache.SetFields(common.KEY_PREFIX_USER+sUserID, m); err != nil {
			return err
		}
	}

	return nil
}

//----------------------------------------------------------------------------

func (us *UserService) QueryUsersByNickname(keywords string, groupID int, session *Session) (*UserInfoArray, error) {
	if us.db == nil {
		return nil, common.ERR_NO_SERVICE
	}

	arr := strings.Split(keywords, " ")

	sql := "SELECT " +
		common.FIELD_ID + "," +
		common.FIELD_NICKNAME + "," +
		common.FIELD_GROUP_ID +
		" FROM " +
		common.TABLE_USER +
		" WHERE " +
		common.FIELD_GROUP_ID + "="

	if session.IsSystem() {
		sql += strconv.Itoa(groupID)
	} else if session.IsAssistant() {
		switch groupID {
		case common.GROUP_ID_FOR_STUDENT:
			sql += strconv.Itoa(common.GROUP_ID_FOR_STUDENT)
		case common.GROUP_ID_FOR_TEACHER:
			sql += strconv.Itoa(common.GROUP_ID_FOR_TEACHER)
		case common.GROUP_ID_FOR_KEEPER:
			sql += strconv.Itoa(common.GROUP_ID_FOR_KEEPER)
		default:
			sql += strconv.Itoa(session.GroupID)
		}
	} else {
		return nil, common.ERR_NO_AUTHORITY
	}

	cnt := 0
	for i := 0; i < len(arr); i++ {
		if len(arr[i]) == 0 {
			continue
		}
		sql += " AND " + common.FIELD_NICKNAME + " LIKE '%" + common.EscapeForSQL(arr[i]) + "%'"
		cnt++
	}
	if cnt == 0 {
		if groupID == common.GROUP_ID_FOR_STUDENT || groupID == common.GROUP_ID_FOR_TEACHER {
			return nil, errors.New("No valid keyword specified.")
		}
	}
	sql += ";"

	rows, err := us.db.Select(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	uia := new(UserInfoArray)
	uia.Users = list.New()

	for rows.Next() {
		ui := new(UserInfo)
		err := rows.Scan(&ui.ID, &ui.Nickname, &ui.GroupID)
		if err != nil {
			return nil, err
		}
		uia.Users.PushBack(ui)
	}

	return uia, nil
}

//----------------------------------------------------------------------------

func (us *UserService) Count() (int, error) {
	if us.db == nil {
		return 0, common.ERR_NO_DATABASE
	}

	return us.db.Count(common.TABLE_USER)
}

//----------------------------------------------------------------------------
