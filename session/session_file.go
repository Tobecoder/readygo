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
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"
)

const (
	filePrefix = "session_"
)

var (
	fileProvider = &FileProvider{}
)

type FileProvider struct {
	maxLifeTime int64
	savePath    string
	lock        sync.RWMutex
}

// SessionInit configs the maxLifeTime and savePath
func (file *FileProvider) SessionInit(maxLifeTime int64, savePath string) error {
	file.maxLifeTime = maxLifeTime
	file.savePath = savePath
	return nil
}

// SessionRead get file session store by sid
func (file *FileProvider) SessionRead(sid string) (Store, error) {
	file.lock.Lock()
	defer file.lock.Unlock()

	sessionFile := file.getFilePath(sid)
	err := os.MkdirAll(filepath.Dir(sessionFile), 0777)
	if err != nil {
		Logger.Println(err.Error())
	}
	var f *os.File
	_, err = os.Stat(sessionFile)
	if err == nil {
		f, err = os.OpenFile(sessionFile, os.O_RDWR, 0777)
	} else if os.IsNotExist(err) {
		f, err = os.Create(sessionFile)
	} else {
		return nil, err
	}
	defer f.Close()

	os.Chtimes(sessionFile, time.Now(), time.Now())
	var kv map[interface{}]interface{}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		kv = make(map[interface{}]interface{})
	} else {
		kv, err = DecodeGob(b)
		if err != nil {
			return nil, err
		}
	}

	return &FileStore{sid: sid, value: kv}, nil
}

// SessionExist assert whether session id exist
func (file *FileProvider) SessionExist(sid string) bool {
	file.lock.Lock()
	defer file.lock.Unlock()

	sessionFile := file.getFilePath(sid)
	_, err := os.Stat(sessionFile)
	return err == nil
}

// SessionRegenerate regenerate the session id
// it also copy the oldSid content into the new sid at the same time
func (file *FileProvider) SessionRegenerate(oldSid, sid string) (Store, error) {
	file.lock.Lock()
	defer file.lock.Unlock()

	oldFile := file.getFilePath(oldSid)
	newFile := file.getFilePath(sid)

	if err := os.MkdirAll(filepath.Dir(newFile), 0777); err != nil {
		return nil, err
	}
	_, err := os.Stat(oldFile)
	if err == nil {
		var kv map[interface{}]interface{}
		b, err := ioutil.ReadFile(oldFile)
		if err != nil {
			return nil, err
		}
		if len(b) == 0 {
			kv = make(map[interface{}]interface{})
		} else {
			kv, err = DecodeGob(b)
			if err != nil {
				return nil, err
			}
		}
		ioutil.WriteFile(newFile, b, 0777)
		os.Chtimes(newFile, time.Now(), time.Now())
		os.Remove(oldFile)
		return &FileStore{sid: sid, value: kv}, nil
	}
	f, err := os.Create(newFile)
	if err != nil {
		return nil, err
	}
	f.Close()
	return &FileStore{sid: sid, value: make(map[interface{}]interface{})}, nil
}

// SessionDestroy destroy the session by sid
func (file *FileProvider) SessionDestroy(sid string) error {
	file.lock.Lock()
	defer file.lock.Unlock()
	sessionFile := file.getFilePath(sid)
	err := os.Remove(sessionFile)
	if err != nil {
		return err
	}
	return nil
}

// SessionAll gets number of all active session
func (file *FileProvider) SessionAll() int {
	stats := new(Stats)
	err := filepath.Walk(file.savePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		stats.total = stats.total + 1
		return nil
	})
	if err != nil {
		Logger.Printf("filepath.Walk() returned %v\n", err)
		return 0
	}
	return stats.total
}

// SessionGC recycle the invalid session
func (file *FileProvider) SessionGC() {
	file.lock.Lock()
	defer file.lock.Unlock()

	filepath.Walk(file.savePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if time.Now().Sub(info.ModTime()) > time.Duration(file.maxLifeTime)*time.Second {
			os.Remove(path)
		}
		return nil
	})
}

// getFilePath retrieves the session file path
func (file *FileProvider) getFilePath(sid string) string {
	subDir := path.Join(string(sid[0]), string(sid[1]))
	return filepath.Join(file.savePath, subDir, filePrefix+sid)
}

type Stats struct {
	total int
}

type FileStore struct {
	sid   string //session id
	value map[interface{}]interface{}
	lock  sync.RWMutex
}

// Set set session key's value
func (fs *FileStore) Set(key, value interface{}) error {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	fs.value[key] = value
	return nil
}

// Get gets the value of key
func (fs *FileStore) Get(key interface{}) interface{} {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	if v, ok := fs.value[key]; ok {
		return v
	}
	return nil
}

// Delete delete the value of key
func (fs *FileStore) Delete(key interface{}) error {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	delete(fs.value, key)
	return nil
}

// SessionID back current sessionID
func (fs *FileStore) SessionID() string {
	return fs.sid
}

// SessionRelease release the resource & save data to provider
func (fs *FileStore) SessionRelease(w http.ResponseWriter) {
	fs.lock.Lock()
	defer fs.lock.Unlock()

	sessionPath := fileProvider.getFilePath(fs.sid)
	var f *os.File
	_, err := os.Stat(sessionPath)
	if err == nil {
		f, err = os.OpenFile(sessionPath, os.O_RDWR, 0777)
		if err != nil {
			Logger.Printf("open file:%v\n", err)
			return
		}
	} else if os.IsNotExist(err) {
		f, err = os.Create(sessionPath)
		if err != nil {
			Logger.Printf("create file:%v\n", err)
			return
		}
	} else {
		Logger.Printf("stat file:%v\n", err)
		return
	}
	b, err := EncodeGob(fs.value)
	if err != nil {
		Logger.Printf("encode gob:%v\n", err)
		return
	}

	f.Truncate(0)
	f.Seek(0, 0)
	f.Write(b)
	f.Close()

	fs.sid = ""
	fs.value = make(map[interface{}]interface{})
}

// Flush delete all data
func (fs *FileStore) Flush() error {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	fs.value = make(map[interface{}]interface{})
	return nil
}

func init() {
	Register("file", fileProvider)
}
