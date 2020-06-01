package elementary

import (
	"github.com/wangmhgo/go-project/gdun/common"
	"net/http"
)

//----------------------------------------------------------------------------

func (sv *Server) onHttpGetAppConfig(w http.ResponseWriter, r *http.Request) {
	if sv.cache == nil {
		sv.Send(w, r, -1, common.S_NO_SERVICE, "")
		return
	}

	result, err := sv.cache.GetKey(common.KEY_PREFIX_CONFIG + common.FIELD_APP)
	if err != nil {
		sv.Send(w, r, -2, err.Error(), "")
		return
	}

	sv.Send(w, r, 0, "", result)
}

//----------------------------------------------------------------------------
