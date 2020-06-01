package service

import (
	"container/list"
	"database/sql"
	// "fmt"
	"github.com/wangmhgo/go-project/gdun/common"
	// "io/ioutil"
	// "net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

//----------------------------------------------------------------------------

type GdAdapter struct {
	db         *common.Database
	//serializer *common.PHPSerializer
	chinese    *regexp.Regexp
}

func NewGdAdapter(db *common.Database) *GdAdapter {
	a := new(GdAdapter)
	a.db = db
	//a.serializer = new(common.PHPSerializer)
	a.chinese = regexp.MustCompile("[\\p{Han}]+")

	return a
}

//----------------------------------------------------------------------------

func (ga *GdAdapter) GetCourse(courseID int) (*GdCourse, error) {
	if ga.db == nil {
		return nil, common.ERR_NO_SERVICE
	}

	// Get courseware IDs and their positions within the specified course.
	IDArr, err := (func(courseID int) ([]*IDPositionPair, error) {
		s := "SELECT ware_id,sortid FROM gd_course_ware_relative WHERE course_id=" + strconv.Itoa(courseID) + ";"
		rows, err := ga.db.Select(s)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		// Get courseware IDs and their positions.
		ls := list.New()
		var id sql.NullInt64
		var pos sql.NullInt64
		for rows.Next() {
			err = rows.Scan(&id, &pos)
			if err != nil {
				return nil, err
			}

			cwID := int(id.Int64)
			if cwID == 0 {
				continue
			}

			pair := &IDPositionPair{cwID, int(pos.Int64)}
			ls.PushBack(pair)
		}

		// Build an array.
		arr := make([]*IDPositionPair, ls.Len())
		i := 0
		for e := ls.Front(); e != nil; e = e.Next() {
			pair, okay := e.Value.(*IDPositionPair)
			if !okay {
				continue
			}
			arr[i] = pair
			i++
		}

		// Clear the list.
		ls.Init()

		// Sort them.
		sort.Sort(IDPositionSlice(arr))

		return arr, nil
	})(courseID)
	if err != nil {
		return nil, err
	}

	if len(IDArr) == 0 {
		return NewGdCourse(courseID), nil
	}

	// Get the courseware names.
	nameMap, err := (func(cws []*IDPositionPair) (map[int]string, error) {
		// Construct a string containing every courseware ID.
		sIDs := ""
		first := true
		for i := 0; i < len(cws); i++ {
			if first {
				first = false
			} else {
				sIDs += `,`
			}
			sIDs += strconv.Itoa(cws[i].ID)
		}

		// Get courseware names.
		s := "SELECT `id`,`name` FROM `gd_courseware` WHERE `id` IN (" + sIDs + ");"
		rows, err := ga.db.Select(s)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		m := make(map[int]string)

		var id sql.NullInt64
		var name sql.NullString
		for rows.Next() {
			err = rows.Scan(&id, &name)
			if err != nil {
				return nil, err
			}

			cwID := int(id.Int64)
			if cwID == 0 {
				continue
			}

			m[cwID] = name.String
		}

		return m, nil
	})(IDArr)
	if err != nil {
		return nil, err
	}

	// Get resources who belong to the coursewares.
	c := NewGdCourse(courseID)
	for i := 0; i < len(IDArr); i++ {
		cw, err := ga.GetCourseware(IDArr[i].ID)
		if err != nil {
			return nil, err
		}
		cw.Name, _ = nameMap[cw.ID]

		c.Coursewares.PushBack(cw)
	}

	return c, nil
}

//----------------------------------------------------------------------------

func (ga *GdAdapter) GetCourseware(coursewareID int) (*GdCourseware, error) {
	if ga.db == nil {
		return nil, common.ERR_NO_SERVICE
	}

	s := "SELECT `id`,`name`,`type`,`partpath`,`partpath_new` FROM `gd_courseware_part` WHERE `courseware_id`=" + strconv.Itoa(coursewareID) + ";"
	rows, err := ga.db.Select(s)
	if err != nil {
		return nil, err
	}
	defer rows.Next()

	gc := NewGdCourseware(coursewareID, "")

	var id sql.NullInt64
	var name sql.NullString
	var t sql.NullInt64
	var data1 sql.NullString
	var data2 sql.NullString
	for rows.Next() {
		err = rows.Scan(&id, &name, &t, &data1, &data2)
		if err != nil {
			return nil, err
		}

		switch t.Int64 {
		case 1, 5:
			// Is a video.
			gv := new(GdVideoResource)
			gv.Name = name.String
			gv.ID = ga.getVideo(data2.String)

			gc.Resources.PushBack(gv)

		case 2, 3, 9:
			// Is an exam.
			id, err := strconv.Atoi(data1.String)
			if err != nil {
				return nil, err
			}
			ge, err := ga.GetExam(id)
			if err != nil {
				return nil, err
			}
			ge.Name = name.String

			gc.Resources.PushBack(ge)
		}
	}

	return gc, nil
}

//----------------------------------------------------------------------------

func (ga *GdAdapter) getVideo(s string) string {
	n := len(s)
	if n == 0 {
		return ""
	}

	target := "player/loader.js?"
	pos := strings.Index(s, target)
	if pos < 0 {
		return ""
	}

	pos += len(target)
	i := pos
	for i < n {
		if s[i] == '-' || s[i] == '"' {
			return s[pos:i]
		}
		i++
	}

	return ""
}

//----------------------------------------------------------------------------
