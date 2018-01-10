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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const (
	name  = "username"
	value = "to be a coder"
)

func TestEnableCookie(t *testing.T) {
	config := `{"cookieName":"GOSESSID","enableSetCookie":true,"gclifetime":10,"disableHTTPOnly":false,"secure":true,"cookieLifeTime":0,"domain":""}`
	conf := new(ManagerConfig)
	if err := json.Unmarshal([]byte(config), conf); err != nil {
		t.Fatal("json decode err:", err)
	}
	// test NewManager
	globalSessions, err := NewManager("memory", conf)
	if err != nil {
		t.Fatal(err)
	}
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal("build new Request failed:", err)
	}
	w := httptest.NewRecorder()

	// test SessionStart
	session, err := globalSessions.SessionStart(w, r)
	if err != nil {
		t.Fatalf("session start:%v", err)
	}
	defer session.SessionRelease(w)
	// test session Set
	err = session.Set(name, value)
	if err != nil {
		t.Fatalf("set session %s:%v", name, err)
	}
	// test SessionDestroy
	globalSessions.SessionDestroy(w, r)
	if username := session.Get(name); username != nil {
		t.Fatal("session destroy faild: value not destroy")
	}

	// test SessionGC
	session, _ = globalSessions.SessionStart(w, r)
	session.Set(name, value)
	go globalSessions.SessionGC()
	time.AfterFunc(time.Duration(conf.Gclifetime)*time.Second, func() {
		if username := session.Get(name); username != nil {
			t.Fatal("session gc failed")
		}
	})

	// test session Get
	session.Set(name, value)
	if username := session.Get(name); username != value {
		t.Fatalf("get session %s's value %v failed", name, value)
	}

	// test session Delete
	if err = session.Delete(name); err != nil {
		t.Fatalf("delete session %s failed", name)
	}
	if username := session.Get(name); username != nil {
		t.Fatalf("delete session %s's value %v failed", name, value)
	}

	// test get SessionID
	session.SessionID()

	// test session Flush
	session.Set(name, value)
	if err = session.Flush(); err != nil {
		t.Fatal("session flush failed")
	}

	if username := session.Get(name); username != nil {
		t.Fatalf("session flush failed, the %s:%v exist", name, value)
	}

	// test cookie
	if cookieStr := w.Header().Get("Set-Cookie"); cookieStr == "" {
		t.Fatal("set cookie error")
	} else {
		splits := strings.Split(cookieStr, ";")
		for k, v := range splits {
			value := strings.Split(v, "=")
			if k == 0 && value[0] != conf.CookieName {
				t.Fatal("cookie name error")
			}
		}
	}

}
