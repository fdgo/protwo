package elementary

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

//----------------------------------------------------------------------------

type Config struct {
	LogDir    string `json:"log"`
	UploadDir string `json:"upload"`
	TmpDir    string `json:"tmp"`

	GhostScript      string `json:"ghostScript"`
	DB               string `json:"db"`
	MaxDBConns       int    `json:"maxDBConns"`
	GdDB         string `json:"gdDB"`
	MaxGdDBConns int    `json:"maxGdDBConns"`
	VoDDB            string `json:"vodDB"`
	MaxVoDDBConns    int    `json:"maxVoDDBConns"`
	Cache            string `json:"cache"`
	OSSEndpoint      string `json:"ossEndpoint"`
	OSSBucket        string `json:"ossBucketName"`
	OSSKeyID         string `json:"ossKeyID"`
	OSSKeySec        string `json:"ossKeySec"`

	FrontEndDomain      string   `json:"frontEndDomain"`
	LanIPPrefix         []string `json:"lanIPPrefix"`
	LiveServer          string   `json:"liveServer"`
	UrlPrefix           string   `json:"urlPrefix"`
	CoursewareKeyPrefix string   `json:"coursewareKeyPrefix"`
	ExamKeyPrefix       string   `json:"examKeyPrefix"`
	GdEncryptKey    string   `json:"gdEncryptKey"`
}

func LoadConfig(fileName string) (*Config, error) {
	fin, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer fin.Close()

	buf, err := ioutil.ReadAll(fin)
	if err != nil {
		return nil, err
	}

	cfg := new(Config)
	err = json.Unmarshal(buf, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

//----------------------------------------------------------------------------
