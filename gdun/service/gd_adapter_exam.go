package service

import (
	"container/list"
	"database/sql"
	"errors"

	"strconv"
)

//----------------------------------------------------------------------------

func (ga *GdAdapter) GetExam(id int) (*GdTikuExam, error) {
	if id <= 0 {
		return nil, errors.New("Invalid exam ID.")
	}

	s := "SELECT `title`,`items` FROM `gd_paper_item` WHERE `pid`=" + strconv.Itoa(id) + " AND `isDel`=0;"
	rows, err := ga.db.Select(s)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	//exam := NewGdTikuExam(id)

	var subjectName sql.NullString
	var subjectItems sql.NullString

	for rows.Next() {
		err = rows.Scan(&subjectName, &subjectItems)
		if err != nil {
			return nil, err
		}
	}
	//	ls, err := ga.serializer.Unserialize(subjectItems.String)
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	// Save the exam subject.
	//	subject, err := ga.getSubExam(ls)
	//	if err != nil {
	//		return nil, err
	//	}
	//	// subject.Name = common.Escape(subjectName.String)
	//	subject.Name = url.QueryEscape(subjectName.String)
	//
	//	exam.Subjects.PushBack(subject)
	//}
	//
	//exam.Sort()

	return nil, nil
}

func (ga *GdAdapter) getSubExam(ls *list.List) (*GdTikuSubExam, error) {
	if ls.Len() == 0 {
		return nil, errors.New("Empty exam subject.")
	}

	// Visit the array representing the question IDs within this exam subject.
	m, okay := ls.Front().Value.(map[interface{}]interface{})
	if !okay {
		return nil, errors.New("Invalid exam subject.")
	}

	gtes := NewGdTikuSubExam()

	// Visit each question, respectively.
	for i := 0; i < len(m); i++ {
		v, okay := m[i]
		if !okay {
			continue
		}

		// Get the question object.
		vm, okay := v.(map[interface{}]interface{})
		if !okay {
			continue
		}

		// Get its ID.
		s, okay := vm["id"].(string)
		if !okay {
			continue
		}
		id, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}

		// Get its parent ID.
		if s, okay = vm["pid"].(string); okay {
			if pid, err := strconv.Atoi(s); (err == nil) && (pid > 0) {
				// Say it owns a parent, thus we do not handle it again.
				continue
			}
		}

		item, err := ga.GetQuestion(id)
		if err != nil {
			// TODO: Ignore unsupported question types.
			// return nil, err
		} else {
			gtes.Items.PushBack(item)
		}
	}

	return gtes, nil
}

//----------------------------------------------------------------------------
