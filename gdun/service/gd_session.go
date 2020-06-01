package service

import (
	"github.com/go-redis/redis"
	"github.com/wangmhgo/go-project/gdun/common"
	"gopkg.in/redis.v5"
	"strconv"
	"strings"
)

//----------------------------------------------------------------------------

type GdSessionService struct {
	cache *redis.Client
}

//----------------------------------------------------------------------------

func NewGdSessionService() *GdSessionService {
	opt := new(redis.Options)
	opt.Addr = "master.redis.gdwangxiao.com:6379"
	opt.Password = "Zjreg9CLterPZyJVFJVW"

	gss := new(GdSessionService)
	gss.cache = redis.NewClient(opt)

	return gss
}

//----------------------------------------------------------------------------

func (gss *GdSessionService) GetStudentID(sessionID string) (int, error) {
	s, err := gss.cache.Get("PHPREDIS_SESSION:" + sessionID).Result()
	if err != nil {
		return 0, err
	}

	arr := strings.Split(s, `;`)
	for i := 0; i < len(arr); i++ {
		brr := strings.Split(arr[i], `|`)
		if len(brr) != 2 || brr[0] != `studentID` {
			continue
		}

		crr := strings.Split(brr[1], `:`)
		switch len(crr) {
		case 2:
			if crr[0] == `i` {
				return strconv.Atoi(crr[1])
			}
		case 3:
			if crr[0] == `s` {
				return strconv.Atoi((crr[2])[2 : len(crr[2])-2])
			}
		}
		break
	}

	return 0, common.ERR_INVALID_SESSION
}

//----------------------------------------------------------------------------
