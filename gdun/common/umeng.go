package common

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//----------------------------------------------------------------------------

type UMengAccount struct {
	AppKey       string
	MasterSecret string
}

type UMengMessageClient struct {
	AndroidAppKey       string
	AndroidMasterSecret string

	IOSAppKey       string
	IOSMasterSecret string

	API string
}

func NewUMengMessageClient(android *UMengAccount, ios *UMengAccount) *UMengMessageClient {
	r := new(UMengMessageClient)
	r.API = "http://msg.umeng.com/api/send"

	r.AndroidAppKey = android.AppKey
	r.AndroidMasterSecret = android.MasterSecret

	r.IOSAppKey = ios.AppKey
	r.IOSMasterSecret = ios.MasterSecret

	return r
}

//----------------------------------------------------------------------------

type UMengMessageAndroid struct {
	AppKey       string `json:"appkey"`
	Timestamp    string `json:"timestamp"`
	Type         string `json:"type"` // unicast, listcast
	DeviceTokens string `json:"device_tokens"`

	Payload struct {
		DisplayType string `json:"display_type"` // notification, message

		Body struct {
			Ticker    string `json:"ticker"`     // 通知栏提示文字
			Title     string `json:"title"`      // 通知标题
			Text      string `json:"text"`       // 通知文字描述
			AfterOpen string `json:"after_open"` // go_app, go_url, go_activity, go_custom
		} `json:"body"`
	} `json:"payload"`
}

func NewUMengMessageAndroid(title string, content string, deviceTokens string, key string) *UMengMessageAndroid {
	r := new(UMengMessageAndroid)

	r.AppKey = key
	r.Timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	r.Type = "listcast"
	r.DeviceTokens = deviceTokens

	r.Payload.DisplayType = "notification"

	r.Payload.Body.Ticker = content
	r.Payload.Body.Title = title
	r.Payload.Body.Text = content + " " + title
	r.Payload.Body.AfterOpen = "go_app"

	return r
}

//----------------------------------------------------------------------------

type UMengMessageIOS struct {
	AppKey       string `json:"appkey"`
	Timestamp    string `json:"timestamp"`
	Type         string `json:"type"` // unicast, listcast
	DeviceTokens string `json:"device_tokens"`

	Payload struct {
		APS struct {
			Alert string `json:"alert"`
			Badge int    `json:"badge"`
		} `json:"aps"`
	} `json:"payload"`
}

func NewUMengMessageIOS(title string, content string, deviceTokens string, key string) *UMengMessageIOS {
	r := new(UMengMessageIOS)

	r.AppKey = key
	r.Timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	r.Type = "listcast"
	r.DeviceTokens = deviceTokens

	r.Payload.APS.Alert = title + ":" + content
	r.Payload.APS.Badge = 1

	return r
}

//----------------------------------------------------------------------------

func (um *UMengMessageClient) computeSign(body string, android bool) string {
	s := ""
	if android {
		s = "POST" + um.API + body + um.AndroidMasterSecret
	} else {
		s = "POST" + um.API + body + um.IOSMasterSecret
	}
	buf := md5.Sum(([]byte)(s))

	return fmt.Sprintf("%x", buf)
}

//----------------------------------------------------------------------------

func (um *UMengMessageClient) Send(title string, content string, deviceTokens string, android bool) error {
	body, err := (func() (string, error) {
		if android {
			m := NewUMengMessageAndroid(title, content, deviceTokens, um.AndroidAppKey)

			buf, err := json.Marshal(m)
			if err != nil {
				return "", err
			}

			return string(buf), nil
		} else {
			m := NewUMengMessageIOS(title, content, deviceTokens, um.IOSAppKey)

			buf, err := json.Marshal(m)
			if err != nil {
				return "", err
			}

			return string(buf), nil
		}
	})()
	if err != nil {
		return err
	}

	err = (func() error {
		resp, err := http.Post(um.API+"?sign="+um.computeSign(body, android), "application/x-www-form-urlencoded", strings.NewReader(body))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Say everything is fine.
		if resp.StatusCode == 200 {
			return nil
		}

		// Load the response body, and then retrieve the error code.
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var obj map[string]interface{}
		err = json.Unmarshal(buf, &obj)
		if err != nil {
			return err
		}

		data, okay := obj["data"].(map[string]interface{})
		if !okay {
			return errors.New("Could not find the data segment in response body.")
		}

		errCode, okay := data["error_code"].(string)
		if !okay {
			return errors.New("Could not find the error code in data segment.")
		}

		return errors.New(errCode)
	})()
	if err != nil {
		return err
	}

	return nil
}

//----------------------------------------------------------------------------
