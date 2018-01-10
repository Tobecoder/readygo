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
	"container/list"
	"net/http"
	"sync"
	"time"
)

type MemProvider struct {
	maxLifeTime int64
	savePath    string
	list        *list.List
	sessions    map[string]*list.Element
	lock        sync.RWMutex
}

// SessionInit configs the maxLifeTime and savePath
func (mem *MemProvider) SessionInit(maxLifeTime int64, savePath string) error {
	mem.maxLifeTime = maxLifeTime
	mem.savePath = savePath
	return nil
}

// SessionRead get memory session store by sid
func (mem *MemProvider) SessionRead(sid string) (Store, error) {
	mem.lock.RLock()
	if element, ok := mem.sessions[sid]; ok {
		go mem.SessionUpdate(sid)
		mem.lock.RUnlock()
		return element.Value.(*MemStore), nil
	}
	mem.lock.RUnlock()

	mem.lock.Lock()
	defer mem.lock.Unlock()

	newMem := &MemStore{
		accessTime: time.Now(),
		sid:        sid,
		value:      make(map[interface{}]interface{})}

	element := mem.list.PushFront(newMem)
	mem.sessions[sid] = element

	return newMem, nil
}

// SessionExist assert whether session id exist
func (mem *MemProvider) SessionExist(sid string) bool {
	mem.lock.RLock()
	defer mem.lock.RUnlock()
	if _, ok := mem.sessions[sid]; ok {
		return true
	}
	return false
}

// SessionRegenerate regenerate the session id
// it also copy the oldSid content into the new sid at the same time
func (mem *MemProvider) SessionRegenerate(oldSid, sid string) (Store, error) {
	mem.lock.RLock()
	if element, ok := mem.sessions[oldSid]; ok {
		go mem.SessionUpdate(oldSid)
		mem.lock.RUnlock()

		mem.lock.Lock()
		defer mem.lock.Unlock()
		element.Value.(*MemStore).sid = sid
		mem.sessions[sid] = element
		delete(mem.sessions, oldSid)
		return element.Value.(*MemStore), nil
	}

	mem.lock.RUnlock()

	mem.lock.Lock()
	defer mem.lock.Unlock()

	newMem := &MemStore{
		accessTime: time.Now(),
		sid:        sid,
		value:      make(map[interface{}]interface{})}

	element := mem.list.PushFront(newMem)
	mem.sessions[sid] = element

	return newMem, nil
}

// SessionDestroy destroy the session by sid
func (mem *MemProvider) SessionDestroy(sid string) error {
	mem.lock.Lock()
	defer mem.lock.Unlock()
	if element, ok := mem.sessions[sid]; ok {
		mem.list.Remove(element)
		delete(mem.sessions, sid)
		// at the same time, set the session storage value nil
		// avoid the previous session storage get values
		element.Value.(*MemStore).value = nil
		element.Value.(*MemStore).sid = ""
	}
	return nil
}

// SessionAll gets number of all active session
func (mem *MemProvider) SessionAll() int {
	return mem.list.Len()
}

// SessionGC recycle the invalid session
func (mem *MemProvider) SessionGC() {
	mem.lock.RLock()
	for {
		element := mem.list.Back()
		if element == nil {
			break
		}

		// since method SessionUpdate push the latest access session to the front of list,
		// so if the last element is not expires, do break
		if time.Now().Sub(element.Value.(*MemStore).accessTime) > time.Duration(mem.maxLifeTime)*time.Second {
			mem.lock.RUnlock()
			mem.lock.Lock()
			mem.list.Remove(element)
			delete(mem.sessions, element.Value.(*MemStore).sid)
			mem.lock.Unlock()
			mem.lock.RLock()
		} else {
			break
		}
	}
	mem.lock.RUnlock()
}

// SessionUpdate expand time of session store by id in memory session
func (mem *MemProvider) SessionUpdate(sid string) error {
	mem.lock.Lock()
	defer mem.lock.Unlock()
	if element, ok := mem.sessions[sid]; ok {
		element.Value.(*MemStore).accessTime = time.Now()
		mem.list.MoveToFront(element)
	}
	return nil
}

type MemStore struct {
	accessTime time.Time
	sid        string                      //session id
	value      map[interface{}]interface{} //session store
	lock       sync.RWMutex
}

// Set set session key's value
func (ms *MemStore) Set(key, value interface{}) error {
	ms.lock.Lock()
	defer ms.lock.Unlock()
	ms.value[key] = value
	return nil
}

// Get gets the value of key
func (ms *MemStore) Get(key interface{}) interface{} {
	ms.lock.RLock()
	defer ms.lock.RUnlock()
	if v, ok := ms.value[key]; ok {
		return v
	}
	return nil
}

// Delete delete the value of key
func (ms *MemStore) Delete(key interface{}) error {
	ms.lock.Lock()
	defer ms.lock.Unlock()
	delete(ms.value, key)
	return nil
}

// SessionID back current sessionID
func (ms *MemStore) SessionID() string {
	return ms.sid
}

// SessionRelease release the resource & save data to provider & return the data
func (ms *MemStore) SessionRelease(w http.ResponseWriter) {

}

// Flush delete all data
func (ms *MemStore) Flush() error {
	ms.lock.Lock()
	defer ms.lock.Lock()
	ms.value = make(map[interface{}]interface{})
	return nil
}

func init() {
	Register("memory", &MemProvider{
		list:     list.New(),
		sessions: make(map[string]*list.Element)})
}
