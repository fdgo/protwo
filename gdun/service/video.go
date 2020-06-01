package service

import (
	"fmt"
	"github.com/wangmhgo/go-project/gdun/common"
	"strconv"
	"strings"
)

//----------------------------------------------------------------------------

type VideoService struct {
	db           *common.Database
	cache        *common.Cache
	oss          *common.ObjectStorage
	lines        []string
	tlsLines     []string
	resolutions  []string
	authorizeAPI string
}

func NewVideoService(db *common.Database, cache *common.Cache, oss *common.ObjectStorage, lines []string, tlsLines []string, resolutions []string, authorizeAPI string) (*VideoService, error) {

	vs := new(VideoService)
	vs.db = db
	vs.cache = cache
	vs.oss = oss
	vs.lines = lines
	vs.tlsLines = tlsLines
	vs.resolutions = resolutions
	vs.authorizeAPI = authorizeAPI

	if vs.cache != nil {
		key := common.KEY_PREFIX_CONFIG + "vod"

		if vs.db == nil {
			vs.db = (func() *common.Database {
				dataSourceName, err := vs.cache.GetField(key, "db")
				if err != nil {
					fmt.Println("No db.")
					return nil
				}

				s, err := vs.cache.GetField(key, "maxDBConns")
				if err != nil {
					fmt.Println("No maxConns.")
					return nil
				}
				maxConns, err := strconv.Atoi(s)
				if err != nil {
					fmt.Println(err.Error())
					maxConns = 10
				}

				db, err := common.NewDatabase(dataSourceName, maxConns)
				if err != nil {
					return nil
				}
				return db
			})()
		}

		if vs.oss == nil {
			vs.oss = (func() *common.ObjectStorage {
				endpoint, err := vs.cache.GetField(key, "OSSEndpoint")
				if err != nil {
					return nil
				}

				bucketName, err := vs.cache.GetField(key, "OSSBucketName")
				if err != nil {
					return nil
				}

				keyID, err := vs.cache.GetField(key, "OSSKeyID")
				if err != nil {
					return nil
				}

				keySec, err := vs.cache.GetField(key, "OSSKeySec")
				if err != nil {
					return nil
				}

				oss, err := common.NewObjectStorage(endpoint, bucketName, keyID, keySec)
				if err != nil {
					return nil
				}
				return oss
			})()
		}

		if vs.lines == nil || len(vs.lines) == 0 {
			s, err := vs.cache.GetField(key, "lines")
			if err == nil {
				vs.lines = strings.Split(s, ",")
			}
		}

		if vs.tlsLines == nil || len(vs.tlsLines) == 0 {
			s, err := vs.cache.GetField(key, "tlsLines")
			if err == nil {
				vs.tlsLines = strings.Split(s, ",")
			}
		}

		if vs.resolutions == nil || len(vs.resolutions) == 0 {
			s, err := vs.cache.GetField(key, "resolutions")
			if err == nil {
				vs.resolutions = strings.Split(s, ",")
			}
		}

		if len(vs.authorizeAPI) == 0 {
			s, err := vs.cache.GetField(key, "authorizeAPI")
			if err == nil {
				vs.authorizeAPI = s
			}
		}
	}

	return vs, nil
}

//----------------------------------------------------------------------------

func (vs *VideoService) Init() error {
	if vs.db == nil {
		return common.ERR_NO_DATABASE
	}

	return nil
}

//----------------------------------------------------------------------------

func (vs *VideoService) Preload() (int, int, error) {
	if vs.db == nil {
		return 0, 0, common.ERR_NO_DATABASE
	}
	if vs.cache == nil {
		return 0, 0, common.ERR_NO_CACHE
	}

	rest, err := vs.db.Count(common.TABLE_VIDEO_KEY)
	if err != nil {
		return 0, 0, err
	}

	i := 0
	for i < rest {
		n, err := vs.preloadKeyAndIV(i, common.DATABASE_PRELOAD_SIZE)
		if err != nil {
			return i + n, 0, err
		}
		i += n
	}
	n1 := i

	i = 0
	for i < rest {
		n, err := vs.preloadDuration(i, common.DATABASE_PRELOAD_SIZE)
		if err != nil {
			return n1, i + n, err
		}
		i += n
	}
	n2 := i

	return n1, n2, nil
}

func (vs *VideoService) preloadKeyAndIV(start int, length int) (int, error) {
	sql := "SELECT source_id,aes_key,aes_iv FROM " +
		common.TABLE_VIDEO_KEY +
		" LIMIT " + strconv.Itoa(start) + "," + strconv.Itoa(length) + ";"

	rows, err := vs.db.Select(sql)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	m := make(map[string]string)

	cnt := 0

	id := ""
	key := ""
	iv := ""
	for rows.Next() {
		err = rows.Scan(&id, &key, &iv)
		if err != nil {
			return cnt, err
		}

		m[common.FIELD_KEY] = key
		m[common.FIELD_IV] = iv

		err = vs.cache.SetFields(common.KEY_PREFIX_VIDEO+id, m)
		if err != nil {
			return cnt, err
		}

		cnt++
	}

	return cnt, nil
}

func (vs *VideoService) preloadDuration(start int, length int) (int, error) {
	sql := "SELECT source_id,duration FROM " +
		common.TABLE_VIDEO_KEY +
		" LIMIT " + strconv.Itoa(start) + "," + strconv.Itoa(length) + ";"

	rows, err := vs.db.Select(sql)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	cnt := 0

	id := ""
	d := 0
	for rows.Next() {
		err = rows.Scan(&id, &d)
		if err != nil {
			return cnt, err
		}

		err = vs.cache.SetField(common.KEY_PREFIX_VIDEO+id, common.FIELD_DURATION, strconv.Itoa(d))
		if err != nil {
			return cnt, err
		}

		cnt++
	}

	return cnt, nil
}

//----------------------------------------------------------------------------

func (vs *VideoService) GetVideoAuthorizeInfo(id string) (*VideoAuthorizeInfo, error) {
	sID := common.EscapeForStr64(id)

	n := len(sID)
	if (n > 6) && (sID[n-6] == '0') {
		vi := new(VideoAuthorizeInfo)
		vi.ID = sID
		vi.AESKey = ""
		vi.AESIV = ""

		return vi, nil
	}

	k := common.KEY_PREFIX_VIDEO + sID

	if vs.cache != nil {
		if m, err := vs.cache.GetAllFields(k); err == nil {
			if vi := NewVideoAuthorizeInfoFromMap(m, sID); vi != nil {
				return vi, nil
			}
		}
	}

	if vs.db != nil {
		aesKey, aesIV, err := (func() (string, string, error) {
			s := "SELECT aes_key,aes_iv FROM gd_vod_key WHERE source_id='" + sID + "';"
			rows, err := vs.db.Select(s)
			if err != nil {
				return "", "", err
			}
			defer rows.Close()

			if !rows.Next() {
				return "", "", common.ERR_NO_VIDEO
			}

			key := ""
			iv := ""
			err = rows.Scan(&key, &iv)
			if err != nil {
				return "", "", err
			}

			return key, iv, nil
		})()
		if err != nil {
			return nil, err
		}

		d, err := (func() (int, error) {
			s := "SELECT duration FROM gd_vod_video WHERE source_id='" + sID + "';"
			rows, err := vs.db.Select(s)
			if err != nil {
				return 0, err
			}
			defer rows.Close()

			if !rows.Next() {
				return 0, common.ERR_NO_VIDEO
			}

			d := 0
			err = rows.Scan(&d)
			if err != nil {
				return 0, err
			}

			return d, nil
		})()
		if err != nil {
			d = 0
		}

		if vs.cache != nil {
			m := make(map[string]string)
			m[common.FIELD_KEY] = aesKey
			m[common.FIELD_IV] = aesIV
			m[common.FIELD_DURATION] = strconv.Itoa(d)

			err = vs.cache.SetFields(k, m)
			if err != nil {
				// Do something here.
			}
		}

		vi := new(VideoAuthorizeInfo)
		vi.ID = sID
		vi.AESKey = aesKey
		vi.AESIV = aesIV
		vi.Duration = d

		return vi, nil
	}

	return nil, common.ERR_NO_SERVICE
}

//----------------------------------------------------------------------------
