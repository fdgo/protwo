package service

import (
	"github.com/wangmhgo/go-project/gdun/common"
	"strconv"
)

//----------------------------------------------------------------------------

/*
1. Interfaces

/class/issue/get?class=xxx
/class/meeting/issue/ask?class=xxx&meeting=xxx&type=xxx&key=xxx&subKey=xxx&question=xxx
/class/meeting/issue/answer?class=xxx&issue=xxx&answer=xxx
/class/meeting/issue/answer/change?class=xxx&issue=xxx&answer=xxx

2. Structures in Cache

2.1 For Answered Questions
SET C:{classID}:issue:text

2.2 For Unanswered Questions
HSET C:{classID}:issue {issueID} {updateTime}

2.3 For Issue itself
HSET G:{groupID}:I:{issueID} ...

{
	"classID": {classID},
	"meetingID": {meetingID},

	"type": {type},				// 1: courseware,	2: exam,		3: , 4: video
	"key": "{key}",				// 1: coursewareID,	2: examID,		3: , 4: videoID
	"subKey": {subKey},			// 1: pageID,		2: questionID,	3: , 4: currentPosition

	"questionBody": "{body}",
	"questionUpdater": {userID},
	"questionUpdateTime": {updateTime},
	"questionUpdateIP": "{updateIP}",

	"answerBody": "{body}",
	"answerUpdater": {userID},
	"answerUpdateTime": {updateTIme},
	"answerUpdateIP": "{updateIP}"
}

3. Structures in Database

For All Issues

ID		 			INT
classID				INT
meetingID			INT
type 				INT
key					VARCHAR(128)
subKey 				INT
questionBody 		TEXT
questionUpdater 	INT
questionUpdateTime 	BIGINT
questionUpdateIP 	VARCHAR
answerBody 			TEXT
answerUpdater 		INT
answerUpdateTime 	BIGINT
answerUpdateIP 		VARCHAR(32)
*/

//----------------------------------------------------------------------------

type IssueInfo struct {
	ID                 int
	GroupID            int
	ClassID            int
	MeetingID          int
	Type               int
	Key                string
	SubKey             int
	QuestionBody       string
	QuestionUpdateIP   string
	QuestionUpdateTime int
	QuestionUpdater    int
	AnswerBody         string
	AnswerUpdateIP     string
	AnswerUpdateTime   int
	AnswerUpdater      int
}

func NewIssueInfoFromMap(m map[string]string, groupID int, id int) (*IssueInfo, int) {
	// groupID, err := strconv.Atoi(m[common.FIELD_GROUP_ID])
	// if err != nil {
	// 	return nil, -1
	// }
	classID, err := strconv.Atoi(m[common.FIELD_CLASS_ID])
	if err != nil {
		return nil, -2
	}
	meetingID, err := strconv.Atoi(m[common.FIELD_MEETING_ID])
	if err != nil {
		return nil, -3
	}
	t, err := strconv.Atoi(m[common.FIELD_TYPE])
	if err != nil {
		return nil, -4
	}
	key, okay := m[common.FIELD_KEY]
	if !okay || len(key) == 0 {
		return nil, -5
	}
	subKey, err := strconv.Atoi(m[common.FIELD_SUB_KEY])
	if err != nil {
		return nil, -6
	}

	qBody, okay := m[common.FIELD_QUESTION_BODY]
	if !okay {
		return nil, -7
	}
	qUpdateIP, okay := m[common.FIELD_QUESTION_UPDATE_IP]
	if !okay {
		return nil, -8
	}
	qUpdateTime, err := strconv.Atoi(m[common.FIELD_QUESTION_UPDATE_TIME])
	if err != nil {
		return nil, -9
	}
	qUpdater, err := strconv.Atoi(m[common.FIELD_QUESTION_UPDATER])
	if err != nil {
		return nil, -10
	}

	aBody, okay := m[common.FIELD_ANSWER_BODY]
	if !okay {
		return nil, -11
	}
	aUpdateIP, okay := m[common.FIELD_ANSWER_UPDATE_IP]
	if !okay {
		return nil, -12
	}
	aUpdateTime, err := strconv.Atoi(m[common.FIELD_ANSWER_UPDATE_TIME])
	if err != nil {
		return nil, -13
	}
	aUpdater, err := strconv.Atoi(m[common.FIELD_ANSWER_UPDATER])
	if err != nil {
		return nil, -14
	}

	ii := new(IssueInfo)
	ii.ID = id
	ii.GroupID = groupID
	ii.ClassID = classID
	ii.MeetingID = meetingID
	ii.Type = t
	ii.Key = key
	ii.SubKey = subKey
	ii.QuestionBody = qBody
	ii.QuestionUpdateIP = qUpdateIP
	ii.QuestionUpdateTime = qUpdateTime
	ii.QuestionUpdater = qUpdater
	ii.AnswerBody = aBody
	ii.AnswerUpdateIP = aUpdateIP
	ii.AnswerUpdateTime = aUpdateTime
	ii.AnswerUpdater = aUpdater

	return ii, 0
}

func (ii *IssueInfo) ToJSON() string {
	s := `{` +
		`"` + common.FIELD_ID + `":` + strconv.Itoa(ii.ID) + `,` +
		`"` + common.FIELD_CLASS_ID + `":` + strconv.Itoa(ii.ClassID) + `,` +
		`"` + common.FIELD_MEETING_ID + `":` + strconv.Itoa(ii.MeetingID) + `,` +
		`"` + common.FIELD_TYPE + `":` + strconv.Itoa(ii.Type) + `,` +
		`"` + common.FIELD_KEY + `":"` + common.ReplaceForJSON(ii.Key) + `",` +
		`"` + common.FIELD_SUB_KEY + `":` + strconv.Itoa(ii.SubKey) + `,` +
		`"` + common.FIELD_QUESTION_BODY + `":"` + common.ReplaceForJSON(ii.QuestionBody) + `",` +
		`"` + common.FIELD_QUESTION_UPDATE_IP + `":"` + ii.QuestionUpdateIP + `",` +
		`"` + common.FIELD_QUESTION_UPDATER + `":` + strconv.Itoa(ii.QuestionUpdater) + `,` +
		`"` + common.FIELD_QUESTION_UPDATE_TIME + `":` + strconv.Itoa(ii.QuestionUpdateTime*1000) + `,` +
		`"` + common.FIELD_ANSWER_BODY + `":"` + common.ReplaceForJSON(ii.AnswerBody) + `",` +
		`"` + common.FIELD_ANSWER_UPDATE_IP + `":"` + ii.AnswerUpdateIP + `",` +
		`"` + common.FIELD_ANSWER_UPDATER + `":` + strconv.Itoa(ii.AnswerUpdater) + `,` +
		`"` + common.FIELD_ANSWER_UPDATE_TIME + `":` + strconv.Itoa(ii.AnswerUpdateTime*1000) +
		`}`

	return s
}

//----------------------------------------------------------------------------
