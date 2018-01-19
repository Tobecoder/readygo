// Copyright readygo Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package session

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"time"
)

// Store contains all data for one session process with specific id.
// The idea comes from "github.com/astaxie/beego/session"
type Store interface {
	Set(key, value interface{}) error     //set session value
	Get(key interface{}) interface{}      //get session value
	Delete(key interface{}) error         //delete session value
	SessionID() string                    //back current sessionID
	SessionRelease(w http.ResponseWriter) //release the resource & save data to provider
	Flush() error                         //delete all data
}

// Provider contains global session methods and saved SessionStores.
// it can operate a SessionStore by its id.
// The idea comes from "github.com/astaxie/beego/session"
type Provider interface {
	SessionInit(maxlifetime int64, config string) error
	SessionRead(sid string) (Store, error)
	SessionExist(sid string) bool
	SessionRegenerate(oldsid, sid string) (Store, error)
	SessionDestroy(sid string) error
	SessionAll() int //get all active session
	SessionGC()
}

var (
	providers = make(map[string]Provider)
	Logger    = NewSessionLog(os.Stderr)
)

// Register makes a session provide available by the provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, provider Provider) {
	if provider == nil {
		panic("session: Register provide is nil")
	}
	if _, ok := providers[name]; ok {
		panic("session: Register provide is already exists")
	}
	providers[name] = provider
}

// ManagerConfig define the session config
// The idea comes from "github.com/astaxie/beego/session"
type ManagerConfig struct {
	CookieName              string `json:"cookieName"`
	EnableSetCookie         bool   `json:"enableSetCookie,omitempty"`
	Gclifetime              int64  `json:"gclifetime"`
	Maxlifetime             int64  `json:"maxLifetime"`
	DisableHTTPOnly         bool   `json:"disableHTTPOnly"`
	Secure                  bool   `json:"secure"`
	CookieLifeTime          int    `json:"cookieLifeTime"`
	Domain                  string `json:"domain"`
	ProviderConfig          string `json:"providerConfig"`
	SessionIDLength         int64  `json:"sessionIDLength"`
	EnableSidInHTTPHeader   bool   `json:"EnableSidInHTTPHeader"`
	SessionNameInHTTPHeader string `json:"SessionNameInHTTPHeader"`
	EnableSidInURLQuery     bool   `json:"EnableSidInURLQuery"`
}

// Manager contains Provider and its configuration.
// The idea comes from "github.com/astaxie/beego/session"
type Manager struct {
	provider Provider
	config   *ManagerConfig
}

// NewManager Create new Manager with provider name and json config string.
// providerName lists here:
// 1. memory
func NewManager(providerName string, cf *ManagerConfig) (*Manager, error) {
	provider, ok := providers[providerName]
	if !ok {
		return nil, fmt.Errorf("provider %v is not registered", providerName)
	}
	if cf.Maxlifetime == 0 {
		cf.Maxlifetime = cf.Gclifetime
	}

	if cf.EnableSidInHTTPHeader {
		if cf.SessionNameInHTTPHeader == "" {
			panic(errors.New("SessionNameInHTTPHeader is empty"))
		}

		strMimeHeader := textproto.CanonicalMIMEHeaderKey(cf.SessionNameInHTTPHeader)
		if cf.SessionNameInHTTPHeader != strMimeHeader {
			strErrMsg := "SessionNameInHTTPHeader (" + cf.SessionNameInHTTPHeader + ") has the wrong format, it should be like this : " + strMimeHeader
			panic(errors.New(strErrMsg))
		}
	}

	err := provider.SessionInit(cf.Maxlifetime, cf.ProviderConfig)
	if err != nil {
		return nil, err
	}

	if cf.SessionIDLength == 0 {
		cf.SessionIDLength = 16
	}

	return &Manager{
		provider: provider,
		config:   cf,
	}, nil
}

// getSid retrieves session identifier from HTTP Request.
// First try to retrieve id by reading from cookie, session-cookie name is configurable,
// if not exist, then retrieve id from querying parameters.
//
// error is not nil when there is anything wrong.
// sid is empty when need to generate a new session id
// otherwise return an valid session id.
func (manager *Manager) getSid(r *http.Request) (string, error) {
	cookie, err := r.Cookie(manager.config.CookieName)
	if err != nil || cookie.Value == "" {
		var sid string
		if manager.config.EnableSidInURLQuery {
			if err := r.ParseForm(); err != nil {
				return "", err
			}
			sid = r.Form.Get(manager.config.CookieName)
		}
		if manager.config.EnableSidInHTTPHeader && sid == "" {
			sid = r.Header.Get(manager.config.CookieName)
		}
		return sid, nil
	}
	return url.QueryUnescape(cookie.Value)
}

// SessionStart generate or read the session id from http request.
// if session id exists, return SessionStore with this id.
func (manager *Manager) SessionStart(w http.ResponseWriter, r *http.Request) (session Store, err error) {
	sid, err := manager.getSid(r)
	if err != nil {
		return nil, err
	}

	if sid != "" && manager.provider.SessionExist(sid) {
		session, err = manager.provider.SessionRead(sid)
		if err != nil {
			return nil, err
		}
		manager.setCookie(sid, w, r)
		return
	}
	sid, err = manager.sessionId()
	if err != nil {
		return nil, err
	}
	session, err = manager.provider.SessionRead(sid)
	if err != nil {
		return nil, err
	}
	manager.setCookie(sid, w, r)
	return
}

// SessionDestroy set the session destroy
func (manager *Manager) SessionDestroy(w http.ResponseWriter, r *http.Request) {
	sid, _ := manager.getSid(r)

	if manager.config.EnableSidInHTTPHeader {
		r.Header.Del(manager.config.SessionNameInHTTPHeader)
		w.Header().Del(manager.config.SessionNameInHTTPHeader)
	}

	manager.provider.SessionDestroy(sid)

	if manager.config.EnableSetCookie {
		http.SetCookie(w, &http.Cookie{
			Name:     manager.config.CookieName,
			Path:     "/",
			HttpOnly: !manager.config.DisableHTTPOnly,
			MaxAge:   -1,
			Expires:  time.Now(),
			Secure:   manager.isSecure(r)})
	}
}

// SessionGC recycle expires sessions
func (manager *Manager) SessionGC() {
	manager.provider.SessionGC()
	time.AfterFunc(time.Duration(manager.config.Gclifetime)*time.Second, func() { manager.SessionGC() })
}

// SessionRegenerate regenerate session id for old session id,
// the old content will be hold to the new session
func (manager *Manager) SessionRegenerate(w http.ResponseWriter, r *http.Request) (session Store, err error) {
	sid, err := manager.sessionId()
	if err != nil {
		return nil, err
	}
	oldSid, err := manager.getSid(r)
	if err != nil || oldSid == "" {
		session, err = manager.provider.SessionRead(sid)
		if err != nil {
			return nil, err
		}
	} else {
		session, err = manager.provider.SessionRegenerate(oldSid, sid)
	}
	manager.setCookie(sid, w, r)
	return
}

// setCookie set the http cookie of session
func (manager *Manager) setCookie(sid string, w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:     manager.config.CookieName,
		Value:    url.QueryEscape(sid),
		Path:     "/",
		HttpOnly: !manager.config.DisableHTTPOnly,
		Secure:   manager.isSecure(r),
		Domain:   manager.config.Domain}

	if manager.config.CookieLifeTime > 0 {
		cookie.MaxAge = manager.config.CookieLifeTime
		cookie.Expires = time.Now().Add(time.Duration(manager.config.CookieLifeTime) * time.Second)
	}

	if manager.config.EnableSetCookie {
		http.SetCookie(w, cookie)
	}

	r.AddCookie(cookie)

	if manager.config.EnableSidInHTTPHeader {
		r.Header.Set(manager.config.SessionNameInHTTPHeader, sid)
		w.Header().Set(manager.config.SessionNameInHTTPHeader, sid)
	}
}

// isSecure returns whether cookie is secure
func (manager *Manager) isSecure(r *http.Request) bool {
	if !manager.config.Secure {
		return false
	}
	if r.URL.Scheme != "" {
		return r.URL.Scheme == "https"
	}
	if r.TLS != nil {
		return true
	}
	return true
}

// sessionId generate session-id
// return any error
func (manager *Manager) sessionId() (string, error) {
	var b = make([]byte, manager.config.SessionIDLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Log supports the log handler of session
type Log struct {
	*log.Logger
}

// NewSessionLog retrieves the Log
func NewSessionLog(w io.Writer) *Log {
	sl := new(Log)
	sl.Logger = log.New(w, "[SESSION]", 1e9)
	return sl
}
