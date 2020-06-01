package service

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	// "fmt"
	"github.com/wangmhgo/go-project/gdun/common"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

//----------------------------------------------------------------------------

type GdPass struct {
	cs  *ClassService
	us  *UserService
	key []byte
}

func NewGdPass(cs *ClassService, us *UserService, key string) *GdPass {
	gdp := new(GdPass)
	gdp.cs = cs
	gdp.us = us
	gdp.key = []byte(key)

	return gdp
}

//----------------------------------------------------------------------------

func (gdp *GdPass) Pass(password string, ip string) (*UserInfo, int, int, int, error) {
	// Check requirements.
	if gdp.cs == nil || gdp.us == nil {
		return nil, 0, 0, -1, common.ERR_NO_SERVICE
	}

	// Check inputs.
	if len(password) == 0 {
		return nil, 0, 0, -2, errors.New("Empty token.")
	}

	// Decrypt the encrypted string.
	s, err := (func() (string, error) {
		encrypted, err := base64.StdEncoding.DecodeString(password)
		if err != nil {
			return "", err
		}
		if len(encrypted)%16 != 0 {
			return "", errors.New("Invalid string length.")
		}

		plain, err := common.Decrypt(encrypted, gdp.key, gdp.key)
		if err != nil {
			return "", err
		}

		return string(plain), nil
	})()
	if err != nil {
		return nil, 0, 0, -3, err
	}

	// Get parameters from the decrypted string.
	arr := strings.Split(s, ",")
	if len(arr) < 4 {
		return nil, 0, 0, -4, errors.New("Invalid parameter length.")
	}

	_, err = strconv.Atoi(arr[1]) // timestamp
	if err != nil {
		return nil, 0, 0, -5, errors.New("Invalid timestamp.")
	}

	// Get user ID.
	gdStudentID, err := strconv.Atoi(arr[3])
	if err != nil {
		return nil, 0, 0, -7, errors.New("Invalid Gd student ID.")
	}
	userID, nickname, _, err := gdp.us.GetOrAddGdStudent(gdStudentID, ip, false)
	if err != nil {
		return nil, 0, 0, -8, err
	}

	// Get class ID.
	classID := 0
	showDate := 1
	if len(arr) == 5 {
		classID, err = strconv.Atoi(arr[2])
		if err != nil {
			return nil, 0, 0, -8, errors.New("Invalid class ID.")
		}

		showDate, err = strconv.Atoi(arr[4])
		if err != nil {
			showDate = 1
		}
	} else {
		gdCourseID, err := strconv.Atoi(arr[2])
		if err != nil {
			return nil, 0, 0, -9, errors.New("Invalid Gd course ID.")
		}
		classID, err = gdp.cs.GetClassIDViaGdCourseID(gdCourseID)
		if err != nil {
			return nil, 0, 0, -10, err
		}
	}

	// Get potential child class ID.
	classID, status, err := gdp.cs.GetClassIDViaUserID(classID, userID)
	if err != nil {
		return nil, 0, 0, status - 10, err
	}

	// Prepare results.
	ui := new(UserInfo)
	ui.ID = userID
	ui.Nickname = nickname
	ui.GroupID = common.GROUP_ID_FOR_STUDENT
	ui.GdStudentID = gdStudentID

	return ui, classID, showDate, 0, nil
}

// func (gdp *GdPass) PassA(password string, ip string) (*UserInfo, int, int, error) {
// 	if gdp.cs == nil || gdp.us == nil {
// 		return nil, 0, 0, common.ERR_NO_SERVICE
// 	}

// 	if len(password) == 0 {
// 		return nil, 0, 0, errors.New("Empty token.")
// 	}

// 	// Decrypt the encrypted string.
// 	s, err := (func() (string, error) {
// 		encrypted, err := base64.StdEncoding.DecodeString(password)
// 		if err != nil {
// 			return "", err
// 		}
// 		if len(encrypted)%16 != 0 {
// 			return "", errors.New("Invalid string length.")
// 		}

// 		plain, err := common.Decrypt(encrypted, gdp.key, gdp.key)
// 		if err != nil {
// 			return "", err
// 		}

// 		return string(plain), nil
// 	})()
// 	if err != nil {
// 		return nil, 0, 0, err
// 	}

// 	//----------------------------------------------------
// 	// Get parameters from the decrypted string.

// 	arr := strings.Split(s, ",")
// 	if len(arr) != 4 {
// 		return nil, 0, 0, errors.New("Invalid parameter length.")
// 	}

// 	_, err = strconv.Atoi(arr[1]) // timestamp
// 	if err != nil {
// 		return nil, 0, 0, err
// 	}

// 	gdCourseID, err := strconv.Atoi(arr[2])
// 	if err != nil {
// 		return nil, 0, 0, err
// 	}

// 	gdStudentID, err := strconv.Atoi(arr[3])
// 	if err != nil {
// 		return nil, 0, 0, err
// 	}

// 	//----------------------------------------------------

// 	classID, groupID, err := gdp.cs.GetClassIDViaGdCourseID(gdCourseID)
// 	if err != nil {
// 		return nil, 0, 0, err
// 	}

// 	userID, nickname, _, err := gdp.us.GetOrAddGdStudent(gdStudentID, ip, false)
// 	if err != nil {
// 		return nil, 0, 0, err
// 	}

// 	// if isNew {
// 	// 	// Construct a session structure, if this request originates from a back-end server.
// 	// 	session := new(Session)
// 	// 	session.UserID = 0
// 	// 	session.GroupID = common.GROUP_ID_FOR_SYSTEM
// 	// 	session.IP = ip

// 	// 	err = gdp.cs.ChangeUser(userID, classID, false, true, session)
// 	// 	if err != nil {
// 	// 		return nil, 0, 0, err
// 	// 	}
// 	// }

// 	//----------------------------------------------------

// 	ui := new(UserInfo)
// 	ui.ID = userID
// 	ui.Nickname = nickname
// 	ui.GroupID = common.GROUP_ID_FOR_STUDENT
// 	ui.GdStudentID = gdStudentID

// 	return ui, classID, groupID, nil
// }

//----------------------------------------------------------------------------

func (gdp *GdPass) get(uri string) ([]byte, error) {
	// fmt.Println(uri)

	c := &http.Client{}
	r, err := http.NewRequest("GET", uri, strings.NewReader(""))
	if err != nil {
		return nil, err
	}

	r.Header.Add("Origin", "http://live-hz.gitlab.hfjy.com")

	resp, err := c.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(buf))
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func (gdp *GdPass) post(uri string, params string) ([]byte, error) {
	// fmt.Println(uri)

	c := &http.Client{}
	r, err := http.NewRequest("POST", uri, strings.NewReader(params))
	if err != nil {
		return nil, err
	}

	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Origin", "http://live-hz.gitlab.hfjy.com")

	resp, err := c.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(buf))
	if err != nil {
		return nil, err
	}

	return buf, nil
}

//----------------------------------------------------------------------------

func (gdp *GdPass) Login(name string, password string) (int, error) {
	// Create an HTTP request.
	buf, err := gdp.post("http://sso.gitlab.hfjy.com/gdun/login", "user="+common.Escape(name)+"&password="+common.Escape(password))
	if err != nil {
		return 0, err
	}

	// Parse the HTTP response.
	gdStudentID := (func() int {
		var resp map[string]interface{}
		err := json.Unmarshal(buf, &resp)
		if err != nil {
			return 0
		}

		status, okay := resp["status"].(float64)
		if !okay || status != 0 {
			return 0
		}

		data, okay := resp["data"].(map[string]interface{})
		if !okay {
			return 0
		}

		userData, okay := data["UserData"].(map[string]interface{})
		if !okay {
			return 0
		}

		studentID, okay := userData["StudentId"].(float64)
		if !okay {
			return 0
		}

		return int(studentID)
	})()
	if gdStudentID == 0 {
		// fmt.Println(string(buf))
		return 0, common.ERR_INVALID_RESPONSE
	}

	return gdStudentID, nil
}

//----------------------------------------------------------------------------

func (gdp *GdPass) CheckLogin(session string) (int, error) {
	// Create an HTTP request.
	buf, err := gdp.get("http://sso.gitlab.hfjy.com/gdun/api/v1/checklogin?GDSSID=" + common.Escape(session))
	if err != nil {
		return 0, err
	}
	// fmt.Println(string(buf))

	// Parse the HTTP response.
	gdStudentID := (func() int {
		var resp map[string]interface{}
		err := json.Unmarshal(buf, &resp)
		if err != nil {
			return 0
		}

		status, okay := resp["status"].(float64)
		if !okay || status != 0 {
			return 0
		}

		data, okay := resp["data"].(map[string]interface{})
		if !okay {
			return 0
		}

		studentID, okay := data["studentId"].(float64)
		if !okay {
			return 0
		}

		return int(studentID)
	})()
	if gdStudentID == 0 {
		return 0, common.ERR_INVALID_RESPONSE
	}

	// fmt.Println(gdStudentID)
	return gdStudentID, nil
}

//----------------------------------------------------------------------------

func (gdp *GdPass) LoginAs3rd(id string, t int) (int, error) {
	// Create an HTTP request.
	buf, err := gdp.post("http://sso.gitlab.hfjy.com/gdun/wxopenidlogin", "openid="+common.Escape(id)+"&source="+strconv.Itoa(t))
	if err != nil {
		return 0, err
	}

	// Parse the HTTP response.
	gdStudentID := (func() int {
		var resp map[string]interface{}
		err := json.Unmarshal(buf, &resp)
		if err != nil {
			return 0
		}

		status, okay := resp["status"].(float64)
		if !okay || status != 0 {
			return 0
		}

		data, okay := resp["data"].(map[string]interface{})
		if !okay {
			return 0
		}

		userData, okay := data["UserData"].([]interface{})
		if !okay {
			return 0
		}
		if len(userData) == 0 {
			return 0
		}

		user, okay := userData[0].(map[string]interface{})
		if !okay {
			return 0
		}

		studentID, okay := user["StudentId"].(float64)
		if !okay {
			return 0
		}

		return int(studentID)
	})()
	if gdStudentID == 0 {
		return 0, common.ERR_INVALID_RESPONSE
	}

	return gdStudentID, nil
}

//----------------------------------------------------------------------------

func (gdp *GdPass) GetStudentID(account string) int {
	buf, err := gdp.get("http://sso.gitlab.hfjy.com/gdun/getdatabyaccount/" + account)
	if err != nil {
		return 0
	}
	// fmt.Println(string(buf))

	gdStudentID := (func() int {
		var resp map[string]interface{}
		err := json.Unmarshal(buf, &resp)
		if err != nil {
			return 0
		}

		status, okay := resp["status"].(float64)
		if !okay || status != 0 {
			return 0
		}

		data, okay := resp["data"].(map[string]interface{})
		if !okay {
			return 0
		}

		studentID, okay := data["StudentId"].(float64)
		if !okay {
			return 0
		}

		return int(studentID)
	})()

	return gdStudentID
}

//----------------------------------------------------------------------------

func (gdp *GdPass) GetSutdentInfo(id int) (string, string, string) {
	buf, err := gdp.get("http://sso.gitlab.hfjy.com/gdun/getbaseuserinfo/" + strconv.Itoa(id))
	if err != nil {
		// fmt.Println(err.Error())
		return "", "", ""
	}
	// fmt.Println(string(buf))

	return (func() (string, string, string) {
		var resp map[string]interface{}
		err := json.Unmarshal(buf, &resp)
		if err != nil {
			// fmt.Println("1:" + err.Error())
			return "", "", ""
		}

		status, okay := resp["status"].(float64)
		if !okay || status != 0 {
			// fmt.Println("2")
			return "", "", ""
		}

		data, okay := resp["data"].(map[string]interface{})
		if !okay {
			// fmt.Println("3")
			return "", "", ""
		}

		userData, okay := data["UserData"].([]interface{})
		if !okay || len(userData) == 0 {
			// fmt.Println("4")
			return "", "", ""
		}

		user, okay := userData[0].(map[string]interface{})
		if !okay {
			// fmt.Println("4.5")
			return "", "", ""
		}

		status, okay = user["Status"].(float64)
		if !okay || status != 0 {
			// fmt.Println("5")
			return "", "", ""
		}

		name, okay := user["Status"].(string)
		if !okay {
			// fmt.Println("6")
			name = ""
		}
		phone, okay := user["Phone"].(string)
		if !okay {
			// fmt.Println("7")
			phone = ""
		}
		email, okay := user["Email"].(string)
		if !okay {
			// fmt.Println("8")
			email = ""
		}

		return name, phone, email
	})()
}

//----------------------------------------------------------------------------

func (gdp *GdPass) GetStudentName(id int) (string, error) {
	r, err := http.Get("http://ssm.gitlab.hfjy.com/gdun/api/InnerMesh/getRealname/studentId/" + strconv.Itoa(id))
	if err != nil {
		return "", err
	}
	defer r.Body.Close()

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	var obj map[string]string
	if err = json.Unmarshal(buf, &obj); err != nil {
		return "", err
	}

	s, okay := obj["realname"]
	if !okay {
		return "", common.ERR_NO_RECORD
	}

	s, err = url.QueryUnescape(s)
	if err != nil {
		return s, err
	}

	return s, nil
}

//----------------------------------------------------------------------------

func (gdp *GdPass) GetClassListOfStudent(id int) string {
	return ""
}

//----------------------------------------------------------------------------

func (gdp *GdPass) HasClass(studentID int, courseID int) bool {
	return true
}

//----------------------------------------------------------------------------
