package service

import (
	"github.com/wangmhgo/go-project/gdun/common"
	"strconv"
	"strings"
)

//----------------------------------------------------------------------------

func (vs *VideoService) GetDowngradedM3U8(id string, resolution string, lineID int, isTLS bool, userID int, userToken string) (string, error) {
	s, resolutionID, err := vs.getDowngradedM3U8(id, resolution)
	if err != nil {
		return "", err
	}

	nLineID := lineID - 1
	if nLineID < 0 {
		nLineID = 0
	}

	line := ""
	if isTLS {
		if nLineID >= len(vs.tlsLines) {
			nLineID = 0
		}
		line = vs.tlsLines[nLineID]
	} else {
		if nLineID >= len(vs.lines) {
			nLineID = 0
		}
		line = vs.lines[nLineID]
	}

	result := ""
	ls := strings.Split(s, "\n")
	for i := 0; i < len(ls); i++ {
		if len(ls[i]) == 0 {
			continue
		}

		if strings.HasPrefix(ls[i], "#EXT-X-KEY") {
			vi, err := vs.GetVideoAuthorizeInfo(id)
			if err != nil {
				return "", err
			}

			result += "#EXT-X-KEY:METHOD=AES-128,URI=\""
			if isTLS {
				result += "https:"
			} else {
				result += "http:"
			}
			result += vs.authorizeAPI + "?" + common.FIELD_ID + "=" + id
			if (userID > 0) && (len(userToken) > 0) {
				result += "&" + common.FIELD_SESSION + "=" + strconv.Itoa(userID) + "&" + common.FIELD_TOKEN + "=" + userToken
			}
			result += "\"" + ",IV=0x" + vi.AESIV
		} else if strings.HasPrefix(ls[i], "#") {
			result += ls[i]
		} else {
			result += line + "/pub/" + id + "/" + vs.resolutions[resolutionID] + "/" + ls[i]
		}
		result += "\n"
	}

	return result, nil
}

func (vs *VideoService) getDowngradedM3U8(id string, resolution string) (string, int, error) {
	// Find the index of current resolution.
	i := len(vs.resolutions) - 1
	for i >= 0 {
		if vs.resolutions[i] == resolution {
			break
		}
		i--
	}
	// Not found, then use the lowest resolution.
	if i < 0 {
		i = 0
	}

	// Get the first available M3U8 file.
	for i >= 0 {
		s, err := vs.getM3U8(id, vs.resolutions[i])
		if err == nil {
			return s, i, nil
		}
		i--
	}

	// Not found.
	return "", 0, common.ERR_NO_VIDEO
}

func (vs *VideoService) getM3U8(id string, resolution string) (string, error) {
	key := common.KEY_PREFIX_VIDEO + id + ":" + resolution

	// Get it via cache.
	if vs.cache != nil {
		s, err := vs.cache.GetKey(key)
		if err == nil {
			return s, nil
		}
	}

	// Get it via OSS.
	if vs.oss != nil {
		buf, err := vs.oss.Get("pub/" + id + "/" + resolution + "/index.m3u8")
		if err != nil {
			return "", err
		}

		s := string(buf)

		// Save it to cache.
		if vs.cache != nil {
			err := vs.cache.SetKey(key, s)
			if err != nil {
				// TODO: log it.
			}
		}

		return s, nil
	}

	return "", common.ERR_NO_SERVICE
}

//----------------------------------------------------------------------------

func (vs *VideoService) GetVideoKey(id string) ([]byte, error) {
	vi, err := vs.GetVideoAuthorizeInfo(id)
	if err != nil {
		return nil, err
	}

	return common.HexStringToBytes(vi.AESKey), nil
}

//----------------------------------------------------------------------------

func (vs *VideoService) CheckInternalIP(id string, ip string) bool {
	if vs.cache == nil {
		return true
	}

	existing, err := vs.cache.FieldExist("vod:internal:id", id)
	if err != nil || !existing {
		return true
	}

	existing, err = vs.cache.FieldExist("bokecc:internal:ip", (strings.Split(ip, ":"))[0])
	if err != nil || existing {
		return true
	}

	return false
}

//----------------------------------------------------------------------------
