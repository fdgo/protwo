package common

import (
	"strconv"
)

//----------------------------------------------------------------------------

type WebMessage struct {
	Status int
	Info   string
	Result string
}

func (msg *WebMessage) ToJSON() []byte {
	return PlainJSONMessage(msg.Status, msg.Info, msg.Result)
}

func PlainJSONMessage(status int, info string, result string) []byte {
	switch status {
	case 0:
		return ([]byte)(`{"status":0,"info":"Okay.","result":{` + result + `}}`)
	default:
		return ([]byte)(`{"status":` + strconv.Itoa(status) + `,"info":"` + UnescapeForJSON(info) + `","result":null}`)
	}
}

//----------------------------------------------------------------------------

type SocketMessage struct {
	Status     int
	Info       string
	Command    int
	SubCommand int
	From       int
	Result     string
}

func (msg *SocketMessage) ToJSON() []byte {
	return PlainCommandJSONMessage(msg.Status, msg.Info, msg.Command, msg.SubCommand, msg.From, msg.Result)
}

func PlainCommandJSONMessage(status int, info string, cmd int, subCmd int, from int, result string) []byte {
	timestamp := GetMillisecondString()

	if status == 0 {
		return ([]byte)(`{"status":0,"info":"Okay.","command":` + strconv.Itoa(cmd) + `,"subCommand":` + strconv.Itoa(subCmd) + `,"from":` + strconv.Itoa(from) + `,"timestamp":` + timestamp + `,"result":{` + result + `}}`)
	} else {
		sInfo := ""
		switch status {
		case -1:
			sInfo = "Invalid command package."
		case -2:
			sInfo = "Unknown command."
		case -3:
			sInfo = "Unknown sub-command."
		case -4:
			sInfo = "No authority."
		default:
			sInfo = UnescapeForJSON(info)
		}
		return ([]byte)(`{"status":` + strconv.Itoa(status) + `,"info":"` + sInfo + `","command":` + strconv.Itoa(cmd) + `,"subCommand":` + strconv.Itoa(subCmd) + `,"from":` + strconv.Itoa(from) + `,"timestamp":` + timestamp + `,"result":null}`)
	}
}

//----------------------------------------------------------------------------
