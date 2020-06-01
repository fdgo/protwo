package handler

import (
	"fmt"
	dblayer "github.com/fdgo/distributed-cloud/db"
	"github.com/fdgo/distributed-cloud/util"
	"io/ioutil"
	"net/http"
	"time"
)
const(
	pwd_salt = "*#890"
)
func SignupHandler(w http.ResponseWriter,r *http.Request)  {
	if r.Method == http.MethodGet{
		data, err := ioutil.ReadFile("./static/view/signup.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)
		//http.Redirect(w, r, "./static/view/signup.html", http.StatusFound)
		return
	}
	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	if len(username) <3 || len(password) < 5{
		w.Write([]byte("invalid param!"))
		return
	}
	enc_passwd := util.Sha1([]byte(password+pwd_salt))
	suc := dblayer.UserSignup(username,enc_passwd)
	if suc{
		w.Write([]byte("SUCCESS"))
	}else {
		w.Write([]byte("FAIL"))
	}
}

func SignInHandler(w http.ResponseWriter, r *http.Request)  {
	if r.Method == http.MethodGet {
		data, err := ioutil.ReadFile("./static/view/signin.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)
		//http.Redirect(w, r, "./static/view/signin.html", http.StatusFound)
		return
	}
	r.ParseForm()
	username:=r.Form.Get("username")
	password:=r.Form.Get("password")
	enc_passwd := util.Sha1([]byte(password+pwd_salt))
	pwdChecked :=  dblayer.UserSignin(username,enc_passwd)
	if !pwdChecked{
		w.Write([]byte("FAILED"))
		return
	}
	token := GenToken(username)
	upRes :=dblayer.UpdateToken(username,token)
	if !upRes{
		w.Write([]byte("FAILED"))
		return
	}
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			Location: "http://" + r.Host + "/static/view/home.html",
			Username: username,
			Token:    token,
		},
	}
	w.Write(resp.JSONBytes())
}

func GenToken(username string) string {
	//md5(username+timestamp+token_salt)+timestamp[:8]
	ts:=fmt.Sprintf("%x",time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username+ts+"_tokensalt"))
	return  tokenPrefix+ts[:8]
}