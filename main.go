package main

import (
	"encoding/json"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
	"github.com/golang-jwt/jwt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"oauthsso/config"
	"oauthsso/model"
	"oauthsso/pkg/session"
	"time"
	//"github.com/go-redis/redis"
)

var srv *server.Server
var mgr *manage.Manager

func main() {
	//config init
	config.SetUp()
	// init db connection
	// congifure db in app.yaml then uncomment
	session.SetUp()

	//manager config 默认OAuth 2.0管理器
	mgr = manage.NewDefaultManager()
	mgr.SetAuthorizeCodeTokenCfg(&manage.Config{
		AccessTokenExp:    time.Hour * time.Duration(config.Get().OAuth2.AccessTokenExp),
		RefreshTokenExp:   time.Hour * 24 * 3,
		IsGenerateRefresh: true,
	})

	//提供token的存储方式，直接存储在内存中，这里也可以选择用redis存储
	mgr.MustTokenStorage(store.NewMemoryTokenStore())
	// or use redis token store
	/*
		mgr.MapTokenStorage(oredis.NewRedisStore(&redis.Options{
			Addr: config.Get().Redis.Default.Addr,
			DB: config.Get().Redis.Default.DB,
		}))
	*/

	//access token generate method: jwt指定:token的生成方式-jwt
	mgr.MapAccessGenerate(generates.NewJWTAccessGenerate("", []byte(config.Get().OAuth2.JWTSignedKey), jwt.SigningMethodES512))

	// 提取Client配置
	clientStore := store.NewClientStore()
	for _, v := range config.Get().OAuth2.Client {
		clientStore.Set(v.ID, &models.Client{
			ID:     v.ID,
			Secret: v.Secret,
			Domain: v.Domain,
		})
	}

	//client注册到oauth2服务器中
	mgr.MapClientStorage(clientStore)
	//config oauth2 server
	srv = server.NewServer(server.NewConfig(), mgr)

	//http server
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/authorize", authorizeHandler) // 该接口用于获取授权code
	http.HandleFunc("/token", tokenHandler)
	http.HandleFunc("verify/", verifyHandler)
	// 专门用来支持SSO功能，主要是销毁浏览器的会话，退出登录状态，跳转到指定的链接
	http.HandleFunc("logout/", logoutHandler)]
	http.HandleFunc("/", notFoundHandler)

	log.Println("Server is running at 9096 port.")
	log.Fatal(http.ListenAndServe(":9096", nil))

}

type TplData struct {
	Client config.OAuth2Client
	//用户申请的合规scope
	Scope []config.Scope
	Error string
}

//首先进入执行
//如果之前有提交表单的话，直接使用之前的提交表单
func authorizeHandler(w http.ResponseWriter, r *http.Request) {
	var form url.Values
	if v, _ := session.Get(r, "RequestForm"); v != nil {
		r.ParseForm()
		if r.Form.Get("client_id") == "" {
			form = v.(url.Values)
		}
	}
	r.Form = form
	//该函数尝试从会话中删除之前保存的表单数据，以避免表单重复提交等问题。
	//如果删除过程中出现错误，则会报服务器500错误
	if err := session.Delete(w, r, "RequestForm"); err != nil {
		errorHandler(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := srv.HandleAuthorizeRequest(w, r); err != nil {
		errorHandler(w, err.Error(), http.StatusBadRequest)
		return
	}
}

// 自定义错误显示页面
// 以页面的形式展示大于400的错误
func errorHandler(w http.ResponseWriter, message string, status int) {
	w.WriteHeader(status)
	if status >= 400 {
		t, _ := template.ParseFiles("tpl/error.html")
		body := struct {
			Status  int
			Message string
		}{Status: status, Message: message}
		t.Execute(w, body)
	}
}

// 根据code获取token，此步骤中，客户端需要向授权服务器发送一个包含授权码的请求，并将客户端标识和密钥等信息用于身份验证
// 授权服务器发送一个包含授权码的请求，并将客户端标识和密钥等信息用于身份验证
func tokenHandler(w http.ResponseWriter, r *http.Request) {
	err := srv.HandleTokenRequest(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// 用来验证access_token, scope和domain
func verifyHandler(w http.ResponseWriter, r *http.Request) {
	token, err := srv.ValidationBearerToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cli, err := mgr.GetClient(r.Context(), token.GetClientID())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data := map[string]interface{}{
		"expires_in": int64(token.GetAccessCreateAt().Add(token.GetAccessExpiresIn()).Sub(time.Now()).Seconds()),
		"user_id":    token.GetUserID(),
		"client_id":  token.GetClientID(),
		"scope":      token.GetScope(),
		"domain":     cli.GetDomain(),
	}
	e := json.NewEncoder(w)
	e.SetIndent("", " ")
	e.Encode(data)
}

// 用户登录
func loginHandler(w http.ResponseWriter, r *http.Request) {
	form, _ := session.Get(r, "RequestForm")
	if form == nil {
		errorHandler(w, "无效的请求", http.StatusInternalServerError)
		return
	}

	clientID := form.(url.Values).Get("client_id")
	scope := form.(url.Values).Get("scope")

	//页面数据
	data := TplData{
		Client: config.GetOAuth2Client(clientID),
		Scope:  config.ScopeFilter(clientID, scope),
	}
	if data.Scope == nil {
		errorHandler(w, "无效的权限范围", http.StatusBadRequest)
		return
	}

	if r.Method == "POST" {
		var userID string
		var err error

		if r.Form == nil {
			err = r.ParseForm()
			if err != nil {
				errorHandler(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		//方式1：账号密码验证
		if r.Form.Get("type") == "password" {
			var user model.User
			userID, err = user.Authentication(r.Context(), r.Form.Get("username"), r.Form.Get("password"))
			if err != nil {
				data.Error = err.Error()
				t, _ := template.ParseFiles("tcp/login.html")
				t.Execute(w, data)
				return
			}
		}

		err = session.Set(w, r, "LoggedInUserID", userID)
		if err != nil {
			errorHandler(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Location", "/authorize") //重定向到authorize页面
		w.WriteHeader(http.StatusFound)
		return
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// 这里的为空是判断当前的表单数据还没有被解析
	if r.Form == nil {
		if err := r.ParseForm(); err != nil {
			errorHandler(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	//检查redirect_uri参数
	redirectURI := r.Form.Get("redirect_uri")
	if redirectURI == "" {
		errorHandler(w, "参数不能为空{redirect_uri}", http.StatusBadRequest)
		return
	}

	//解析redirect_uri参数，判断是否能够成功解析
	if _, err := url.Parse(redirectURI); err != nil {
		errorHandler(w, "参数无效{redirect_uri}", http.StatusBadRequest)
		return
	}

	//按redirectURI参数进行重定向
	w.Header().Set("Location", redirectURI)
	w.WriteHeader(http.StatusFound)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request){
	errorHandler(w, "无效的地址", http.StatusNotFound)
	return
}