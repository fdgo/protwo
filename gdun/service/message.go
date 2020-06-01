package service

import (
	"github.com/wangmhgo/go-project/gdun/common"
)

//----------------------------------------------------------------------------

type MessageService struct {
	cache  *common.Cache
	db     *common.Database
	umeng  *common.UMengMessageClient
	weixin *common.WeixinClient
}

func NewMessageService(cache *common.Cache, db *common.Database) (*MessageService, error) {
	r := new(MessageService)
	r.cache = cache
	r.db = db

	r.umeng = common.NewUMengMessageClient(
		&common.UMengAccount{"550a2e71fd98c53903001908", "c5uwn1snxbawtztinhl9opybzgherl1k"},
		&common.UMengAccount{"529932a456240b5723078682", "jfyf0fiwj2uomdcr7jnlab4r9pthi5np"})

	r.weixin = common.NewWeixinClient("wx1e93f7d48cb5bb00", "00f4d9b465b7d2878e93cbebb87fd14c")

	return r, nil
}

//----------------------------------------------------------------------------
