package session

import (
	"encoding/gob"
	"github.com/gprilla/sessions"
	"githun.com/twilight2017/oauthsso/config"
	"net/http"
	"net/url"
)

var store *sessions.CookieStore

// var store *redistore.RediStore
