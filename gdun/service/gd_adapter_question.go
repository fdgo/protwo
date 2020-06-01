package service

import (
	"container/list"
	"database/sql"
	// "fmt"
	"github.com/wangmhgo/go-project/gdun/common"
	"golang.org/x/net/html"
	"strconv"
	"strings"
)

//----------------------------------------------------------------------------
// Database: Required.
// Cache   : Compatible.

func (ga *GdAdapter) GetQuestion(id int) (interface{}, error) {
	if ga.db == nil {
		return nil, common.ERR_NO_SERVICE
	}

	s := "SELECT `type`,`title`,`option`,`answer`,`analysis` FROM `gd_item` WHERE `id`=" + strconv.Itoa(id) + ";"
	rows, err := ga.db.Select(s)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, common.ERR_NO_QUESTION
	}

	var t sql.NullInt64
	var body sql.NullString
	var choice sql.NullString
	var answer sql.NullString
	var analysis sql.NullString
	err = rows.Scan(&t, &body, &choice, &answer, &analysis)
	if err != nil {
		return nil, err
	}

	switch t.Int64 {
	case 1, 2, 3, 4, 6, 7:
		return ga.getQuestion(id, int(t.Int64), body.String, choice.String, answer.String, analysis.String)
	case 5, 10:
		return ga.getComposedQuestion(id)
	}

	return nil, common.ERR_INVALID_QUESTION_TYPE
}

//----------------------------------------------------------------------------
// Database: Required.
// Cache   : Compatible.

func (ga *GdAdapter) getComposedQuestion(id int) (*GdTikuComposedItem, error) {
	if ga.db == nil {
		return nil, common.ERR_NO_SERVICE
	}

	// Get its body.
	body, err := (func() (string, error) {
		s := "SELECT `title` FROM `gd_item` WHERE `id`=" + strconv.Itoa(id) + ";"
		rows, err := ga.db.Select(s)
		if err != nil {
			return "", err
		}
		defer rows.Close()

		if !rows.Next() {
			return "", common.ERR_NO_QUESTION
		}

		var body sql.NullString
		err = rows.Scan(&body)
		if err != nil {
			return "", err
		}

		return body.String, nil
	})()
	if err != nil {
		return nil, err
	}

	gtci := NewGdTikuComposedItem()
	gtci.GdQuestionID = id
	if gtci.Body, err = ga.Parse(body); err != nil {
		return nil, err
	}

	// Get its children.
	err = (func() error {
		s := "SELECT `id`,`type`,`title`,`option`,`answer`,`analysis` FROM `gd_item` WHERE `pid`=" + strconv.Itoa(id) + ";"
		rows, err := ga.db.Select(s)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var childID sql.NullInt64
			var t sql.NullInt64
			var body sql.NullString
			var choice sql.NullString
			var answer sql.NullString
			var analysis sql.NullString

			err = rows.Scan(&childID, &t, &body, &choice, &answer, &analysis)
			if err != nil {
				return err
			}

			item, err := ga.getQuestion(int(childID.Int64), int(t.Int64), body.String, choice.String, answer.String, analysis.String)
			if err != nil {
				return err
			}

			gtci.Items.PushBack(item)
		}

		return nil
	})()
	if err != nil {
		return nil, err
	}

	return gtci, nil
}

//----------------------------------------------------------------------------

func (ga *GdAdapter) getQuestion(id int, t int, body string, choice string, answer string, analysis string) (*GdTikuItem, error) {
	var err error

	item := new(GdTikuItem)
	item.GdQuestionID = id

	if item.Body, err = ga.Parse(body); err != nil {
		return nil, err
	}
	if item.Answer, err = ga.Parse(answer); err != nil {
		return nil, err
	}
	if item.Analysis, err = ga.Parse(analysis); err != nil {
		return nil, err
	}
	item.Type = t

	if t == 3 {
		// True-false question.
		if ga.chinese.MatchString(body) {
			item.Choice = "对\\n错"
		} else {
			item.Choice = "True\\nFalse"
		}

		if item.Answer == "1" {
			item.Answer = "10"
		} else {
			item.Answer = "01"
		}
	} else if t == 4 {
		// Blank-filling question.
		item.Choice = ""
	} else if t == 6 {
		// Essay question.
		item.Choice = ""
	} else {
		if item.Choice, err = ga.Parse(choice); err != nil {
			return nil, err
		}
		item.Choice, item.Answer = ga.verifyChoicesAndAnswer(item.Choice, item.Answer)
	}

	return item, nil
}

func (ga *GdAdapter) isChar(n rune) bool {
	if (n >= 'a') && (n <= 'z') {
		return true
	}
	if (n >= 'A') && (n <= 'Z') {
		return true
	}

	return false
}

func (ga *GdAdapter) verifyChoicesAndAnswer(choices string, answer string) (string, string) {
	arr := strings.Split(choices, "\\n")

	rChoices := ""

	cnt := 0
	first := true
	for i := 0; i < len(arr); i++ {
		s := ([]rune)(arr[i])

		n := len(s)
		if n == 1 {
			if ga.isChar(s[0]) {
				continue
			}
			// if ((s[0] >= 'a') && (s[0] <= 'z')) || ((s[0] >= 'A') && (s[0] <= 'Z')) {
			// continue
			// }
		} else if n == 2 {
			if ga.isChar(s[0]) && (!ga.isChar(s[1])) {
				continue
			}
			// if s[1] == '.' || s[1] == ')' || s[1] == '、' || s[1] == ' ' || s[1] == '\t' {
			// 	if ((s[0] >= 'a') && (s[0] <= 'z')) || ((s[0] >= 'A') && (s[0] <= 'Z')) {
			// 		continue
			// 	}
			// }
		}

		if first {
			first = false
		} else {
			rChoices += "\\n"
		}

		// Save this choice.
		if n > 0 {
			// Remove unnecessary prefix such as "A." and "B.".
			if n > 2 {
				if ga.isChar(s[0]) && (!ga.isChar(s[1])) {
					s = s[2:]
				}
				// if s[1] == '.' || s[1] == ')' || s[1] == '、' {
				// 	if ((s[0] >= 'a') && (s[0] <= 'z')) || ((s[0] >= 'A') && (s[0] <= 'Z')) {
				// 		s = s[2:]
				// 	}
				// }
			}

			// Make sure the first character is a upper one.
			if (s[0] >= 'a') && (s[0] <= 'z') {
				s[0] -= 32
			}

			rChoices += strings.TrimSpace(string(s))
		}

		cnt++
	}

	//----------------------------------------------------

	ansArr := make([]int, cnt)
	for i := 0; i < cnt; i++ {
		ansArr[i] = 0
	}

	s := strings.ToUpper(answer)
	for i := 0; i < len(s); i++ {
		if (s[i] >= 'A') && (s[i] <= 'Z') {
			pos := int(s[i]) - int('A')
			if pos < cnt {
				ansArr[pos] = 1
			}
		}
	}

	rAnswer := ""
	for i := 0; i < len(ansArr); i++ {
		if ansArr[i] == 0 {
			rAnswer += "0"
		} else {
			rAnswer += "1"
		}
	}

	return rChoices, rAnswer
}

//----------------------------------------------------------------------------

func (ga *GdAdapter) Parse(s string) (string, error) {
	// fmt.Println(s)

	q := strings.Replace(s, "\r", "", -1)
	q = strings.Replace(q, "\n", "", -1)
	q = strings.Replace(q, "\t", " ", -1)
	q = strings.Replace(q, "+", "%2B", -1)

	node, err := html.Parse(strings.NewReader(q))
	if err != nil {
		return "", err
	}

	tmp := list.New()
	final := list.New()
	ga.parse(node, tmp, final)
	ga.saveTmpList(tmp, final)

	r := ``
	first := true
	for e := final.Front(); e != nil; e = e.Next() {
		line, okay := e.Value.(string)
		if !okay {
			continue
		}

		if len(line) == 0 {
			continue
		}

		if first {
			first = false
		} else {
			r += "\\n"
		}
		r += line
	}

	return r, nil
}

func (ga *GdAdapter) parse(node *html.Node, tmp *list.List, final *list.List) {

	// Parse this node.
	switch node.Type {

	case html.TextNode:
		s := strings.TrimSpace(node.Data)
		if len(s) > 0 {
			tmp.PushBack(s)
		}

	case html.ElementNode:
		name := strings.ToLower(node.Data)
		switch name {
		case "img":
			// Combine temporary lines and save the result to the list.
			ga.saveTmpList(tmp, final)

			// Get its source.
			for i := 0; i < len(node.Attr); i++ {
				if node.Attr[i].Key == "src" {
					url := node.Attr[i].Val
					n := len(url)
					if n > 0 {
						if url[0] == '/' {
							if (n > 1) && (url[1] == '/') {
								url = "https:" + url
							} else {
								url = "https://www.gitlab.hfjy.com" + url
							}
						}
						final.PushBack(url)
					}
					break
				}
			}

		case "br":
			// Combine temporary lines and save the result to the list.
			ga.saveTmpList(tmp, final)

		case "p", "table", "thead", "tr":
			// Combine temporary lines and save the result to the list.
			ga.saveTmpList(tmp, final)

			// Parse its children.
			ga.parseChildren(node, tmp, final)
			ga.saveTmpList(tmp, final)

		default:
			ga.parseChildren(node, tmp, final)
		}

	case html.DocumentNode:
		ga.parseChildren(node, tmp, final)
	}
}

func (ga *GdAdapter) saveTmpList(tmp *list.List, final *list.List) {
	r := ""
	for e := tmp.Front(); e != nil; e = e.Next() {
		s, okay := e.Value.(string)
		if !okay {
			continue
		}

		if len(s) == 0 {
			continue
		}

		if len(r) == 0 {
			r = s
		} else {
			r += " " + s
		}
	}

	tmp.Init()

	if len(r) > 0 {
		final.PushBack(r)
	}
}

func (ga *GdAdapter) parseChildren(node *html.Node, tmp *list.List, final *list.List) {
	for p := node.FirstChild; p != nil; p = p.NextSibling {
		ga.parse(p, tmp, final)
	}
}

func (ga *GdAdapter) parseAppendLine(s string, ls *list.List) {
	t := strings.TrimSpace(s)
	if len(t) > 0 {
		ls.PushBack(t)
	}
}

func (ga *GdAdapter) parseConcatLine(s string, ls *list.List) {
	t := strings.TrimSpace(s)
	if len(t) == 0 {
		return
	}

	if ls.Len() == 0 {
		ls.PushBack(t)
		return
	}

	// Get the last line.
	e := ls.Back()
	line, okay := e.Value.(string)
	if !okay {
		return
	}
	ls.Remove(e)

	line += " " + t
	ls.PushBack(line)
}

//----------------------------------------------------------------------------
