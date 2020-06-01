package service

import (
	// "fmt"
	"github.com/wangmhgo/go-project/gdun/common"
	"strconv"
)

//----------------------------------------------------------------------------

func (vs *VideoService) QueryVideo(start int, groupID int, keywords []string, IDs []string, session *Session) (string, error) {
	// Check requirements.
	if vs.cache == nil {
		return "", common.ERR_NO_CACHE
	}
	if vs.db == nil {
		return "", common.ERR_NO_DATABASE
	}

	// Check inputs.
	sStart := "0"
	if start > 0 {
		sStart = strconv.Itoa(start)
	}

	// Check authority.
	okay := (func() bool {
		if session.IsSystem() {
			return true
		}
		if session.IsAssistant() && (session.GroupID == groupID) {
			return true
		}
		return false
	})()
	if !okay {
		return "", common.ERR_NO_AUTHORITY
	}

	// Get video categories.
	sCategories, err := vs.cache.GetField(common.KEY_PREFIX_GROUP+strconv.Itoa(groupID), common.FIELD_CATEGORY)
	if err != nil || len(sCategories) == 0 {
		return "", common.ERR_NO_VIDEO
	}

	sub := (func() string {
		s := " FROM " +
			common.TABLE_VIDEO_INFO +
			" WHERE " +
			common.FIELD_DELETED + "=0" +
			" AND " + common.FIELD_FINISHED + "=1" +
			" AND category_id IN (" + sCategories + ")"

		if keywords != nil {
			for i := 0; i < len(keywords); i++ {
				k := common.Prune(common.Unescape(keywords[i]))
				if len(k) == 0 {
					continue
				}
				k = common.ReplaceForSQL(k)

				s += " AND " + common.FIELD_TITLE + " LIKE '%" + k + "%'"
			}
		}
		if IDs != nil {
			for i := 0; i < len(IDs); i++ {
				k := common.Prune(common.Unescape(IDs[i]))
				if len(k) == 0 {
					continue
				}
				k = common.ReplaceForSQL(k)

				s += " AND source_id LIKE '%" + k + "%'"
			}
		}

		return s
	})()

	cnt, err := (func() (int, error) {
		s := "SELECT COUNT(*)" + sub + ";"
		// fmt.Println(s)
		rows, err := vs.db.Select(s)
		if err != nil {
			return 0, err
		}
		defer rows.Close()

		n := 0
		for rows.Next() {
			if err = rows.Scan(&n); err != nil {
				return 0, err
			}
			break
		}

		return n, nil
	})()
	if err != nil {
		return "", err
	}
	if cnt <= 0 {
		return `"` + common.FIELD_VIDEO + `":[],"` + common.FIELD_COUNT + `":0`, nil
	}

	return (func() (string, error) {
		s := "SELECT " +
			"source_id," +
			common.FIELD_TITLE + "," +
			common.FIELD_DURATION + "," +
			common.FIELD_WIDTH + "," +
			common.FIELD_HEIGHT + "," +
			common.FIELD_ENCRYPTION + "," +
			"update_time" +
			sub +
			" ORDER BY update_time DESC" +
			" LIMIT " + sStart + ",100;"

		rows, err := vs.db.Select(s)
		if err != nil {
			return "", err
		}
		defer rows.Close()

		r := `"` + common.FIELD_VIDEO + `":[`
		first := true
		vi := new(VideoInfo)
		for rows.Next() {
			err = rows.Scan(&vi.ID, &vi.Title, &vi.Duration, &vi.Width, &vi.Height, &vi.Encryption, &vi.UpdateTime)
			if err != nil {
				return r, err
			}

			if first {
				first = false
			} else {
				r += `,`
			}
			r += `{` + vi.ToJSON() + `}`
		}
		r += `],"` + common.FIELD_COUNT + `":` + strconv.Itoa(cnt)

		return r, nil
	})()
}

//----------------------------------------------------------------------------
