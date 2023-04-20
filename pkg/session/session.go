package session

import (
	"encoding/gob"
	"github.com/gorilla/sessions"
	//"gopkg.in/boj/redistore.v1"
	"net/http"
	"net/url"
	"oauthsso/config"
)

//用cookie对session进行存储
var store *sessions.CookieStore

// var store *redistore.RediStore
func SetUp() {
	/*url.Values是一个用于存储HTTP请求参数的类型，实现了map[string][]string类型，而map类型是不能被序列化的
	所以需要在这里进行注册，之后对url.Values()进行序列化和反序列化就可以正常操作了*/
	gob.Register(url.Values{}) //注册url.Values{}类型，让它可以序列化和反序列化

	//创建一个新的cookie存储器，NewCookieStore会创建一个基于cookie的session存储器
	//该方法返回了一个实现了session.Store接口的结构体，可用于在后续的代码中进行session的存储和管理
	store = sessions.NewCookieStore([]byte(config.Get().Session.SecretKey))

	store.Options = &sessions.Options{
		Path: "/",
		// session有效期
		//单位秒
		MaxAge:   config.Get().Session.MaxAge,
		HttpOnly: true,
	}
	/*
			store_redis, _ = redistore.NewRediStore(yaml.Cfg.Redis.Defaule.Db, "tcp", yaml.Cfg.Redis.Default.Addr, "", []byte("secret-key"))
		    if err != nil {
			log.Fatal(err)
		}
	*/
}

func Get(r *http.Request, name string) (val interface{}, err error) {
	// Get a session
	session, err := store.Get(r, config.Get().Session.Name)
	if err != nil {
		return
	}
	val = session.Values[name]
	return
}

func Set(w http.ResponseWriter, r *http.Request, name string, val interface{}) (err error) {
	// Get a session
	session, err := store.Get(r, config.Get().Session.Name)
	if err != nil {
		return
	}
	session.Values[name] = val
	err = session.Save(r, w)
	return
}

func Delete(w http.ResponseWriter, r *http.Request, name string) (err error) {
	//Get a session
	session, err := store.Get(r, config.Get().Session.Name)
	if err != nil {
		return
	}
	delete(session.Values, name)
	err = session.Save(r, w)

	return
}
