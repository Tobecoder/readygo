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
	"bytes"
	"encoding/gob"
)

func init() {
	gob.Register(map[string]interface{}{})
	gob.Register(map[string]string{})
	gob.Register(map[int]interface{}{})
	gob.Register(map[int]string{})
	gob.Register(map[int]int{})
	gob.Register(map[int]int64{})
	gob.Register(map[interface{}]interface{}{})
	gob.Register([]interface{}{})
}

// DecodeGob decode the []byte b to map
func DecodeGob(b []byte) (map[interface{}]interface{}, error) {
	buf := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buf)
	var output map[interface{}]interface{}
	if err := decoder.Decode(&output); err != nil {
		return nil, err
	}
	return output, nil
}

// EncodeGob encode the mp to gob
func EncodeGob(mp map[interface{}]interface{}) ([]byte, error) {
	for _, v := range mp {
		gob.Register(v)
	}
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(mp)
	if err != nil {
		return []byte(""), err
	}
	return buf.Bytes(), nil
}
