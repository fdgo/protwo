package wss

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

//----------------------------------------------------------------------------

type Config struct {
	Addr      string `json:"addr"`
	TLSAddr   string `json:"tlsAddr"`
	Cached    bool   `json:"cached"`
	Hostname  string `json:"hostname`
	HtDocDir  string `json:"htdoc"`
	LogDir    string `json:"log"`
	CacheAddr string `json:"cacheAddr"`
}

func LoadConfig(filename string) (*Config, error) {
	fin, err := os.Open(filename)
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

	if len(cfg.Addr) == 0 {
		cfg.Addr = ":80"
	}
	if len(cfg.TLSAddr) == 0 {
		cfg.TLSAddr = ":443"
	}
	if len(cfg.HtDocDir) == 0 {
		cfg.HtDocDir = "htdoc"
	}
	if len(cfg.LogDir) == 0 {
		cfg.LogDir = "log"
	}

	return cfg, nil
}

//----------------------------------------------------------------------------
