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
)

const (
	filename  = "username"
	filevalue = "to be a coder"
)

func TestFile(t *testing.T) {
	config := `{"cookieName":"GOSESSID","providerConfig":"./tmp","enableSetCookie":true,"EnableSidInHTTPHeader":true,"SessionNameInHTTPHeader":"Gosessid","EnableSidInURLQuery":true,"gclifetime":10,"disableHTTPOnly":false,"secure":true,"cookieLifeTime":10,"domain":""}`
	conf := new(ManagerConfig)
	if err := json.Unmarshal([]byte(config), conf); err != nil {
		t.Fatal("json decode err:", err)
	}
	// test NewManager
	globalSessions, err := NewManager("file", conf)
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
	// repeat session start
	session, err = globalSessions.SessionStart(w, r)
	if err != nil {
		t.Fatalf("session start:%v", err)
	}
	// test session Set
	err = session.Set(filename, filevalue)
	if err != nil {
		t.Fatalf("set session %s:%v", filename, err)
	}
	// test SessionDestroy
	globalSessions.SessionDestroy(w, r)
	session.SessionRelease(w)
	if username := session.Get(filename); username != nil {
		t.Fatal("session destroy faild: value not destroy")
	}

	// test SessionGC
	session, _ = globalSessions.SessionStart(w, r)
	session.Set(filename, filevalue)
	session.SessionRelease(w)
	go globalSessions.SessionGC()

	//test session regenerate
	session, err = globalSessions.SessionRegenerate(w, r)
	if err != nil {
		t.Fatal("session regenerate falied", err)
	}
	defer session.SessionRelease(w)

	// test session Get
	session.Set(filename, filevalue)
	if username := session.Get(filename); username != filevalue {
		t.Fatalf("get session %s's value %v failed", filename, filevalue)
	}

	// test session Delete
	if err = session.Delete(filename); err != nil {
		t.Fatalf("delete session %s failed", filename)
	}
	if username := session.Get(filename); username != nil {
		t.Fatalf("delete session %s's value %v failed", filename, filevalue)
	}

	// test get SessionID
	session.SessionID()

	// test session Flush
	session.Set(filename, filevalue)
	if err = session.Flush(); err != nil {
		t.Fatal("session flush failed")
	}

	if username := session.Get(filename); username != nil {
		t.Fatalf("session flush failed, the %s:%v exist", filename, filevalue)
	}

	// test SessionInit
	globalSessions.provider.SessionInit(conf.Maxlifetime, conf.ProviderConfig)
	sid, err := globalSessions.getSid(r)
	if err != nil {
		t.Fatal(err)
	}

	// test SessionRead
	session, err = globalSessions.provider.SessionRead(sid)
	if err != nil {
		t.Fatal(err)
	}

	// test SessionAll
	globalSessions.provider.SessionAll()

	if exist := globalSessions.provider.SessionExist(sid); !exist {
		t.Fatal("session exist failed")
	}

	//test SessionRegenerate
	session.Set(filename, filevalue)
	session.SessionRelease(w)
	session, err = globalSessions.SessionRegenerate(w, r)
	if username := session.Get(filename); username != filevalue {
		t.Fatal("session regenerate falied")
	}

	newRequest, _ := http.NewRequest("GET", "/", nil)
	session, err = globalSessions.SessionRegenerate(w, newRequest)
	if username := session.Get(filename); username != nil {
		t.Fatal("session regenerate falied")
	}

	oldSid, _ := globalSessions.sessionId()
	newSid, _ := globalSessions.sessionId()
	session, err = globalSessions.provider.SessionRegenerate(oldSid, newSid)

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
