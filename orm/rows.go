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

package orm

import (
	"database/sql"
	"reflect"
	"strings"
)

type queryRows struct {
	*sql.Rows
}

// scanSliceMap scan query result to slice of map
func (qr *queryRows) scanSliceMap(items *reflect.Value) error{
	columns, err := qr.Columns()
	if err != nil {
		DebugLog.log(err)
		return err
	}
	count := len(columns)
	values := make([]interface{}, count)
	scanArgs := make([]interface{}, count)
	for i := 0; i < count; i++ {
		scanArgs[i] = &values[i]
	}
	var data = make([]map[string]interface{}, 0)
	for qr.Next(){
		qr.Scan(scanArgs...)
		entry := make(map[string]interface{})
		for key, col := range columns {
			var v interface{}
			val := values[key]
			if b, ok := val.([]byte); ok {
				v = string(b)
			} else {
				v = val
			}
			entry[col] = v
		}
		data = append(data, entry)
	}
	e := items.Elem()
	if e.CanSet(){
		e.Set(reflect.ValueOf(data))
	}else{
		return ErrScan
	}
	err = qr.Err()
	if err != nil {
		DebugLog.log(err)
	}
	return err
}

// scanMap scan query result to map
func (qr *queryRows) scanMap(item *reflect.Value) error {
	columns, err := qr.Columns()
	if err != nil {
		DebugLog.log(err)
		return err
	}
	count := len(columns)
	values := make([]interface{}, count)
	scanArgs := make([]interface{}, count)
	for i := 0; i < count; i++ {
		scanArgs[i] = &values[i]
	}
	if qr.Next(){
		qr.Scan(scanArgs...)
		entry := make(map[string]interface{})
		for key, col := range columns {
			var v interface{}
			val := values[key]
			if b, ok := val.([]byte); ok {
				v = string(b)
			} else {
				v = val
			}
			entry[col] = v
		}
		e := item.Elem()
		if e.CanSet(){
			e.Set(reflect.ValueOf(entry))
		}else{
			return ErrScan
		}
	}
	err = qr.Err()
	if err != nil {
		DebugLog.log(err)
	}
	return nil
}

// scanStruct scan query result to struct
func (qr *queryRows) scanStruct(item *reflect.Value) error{
	columns, err := qr.Columns()
	if err != nil {
		DebugLog.log(err)
		return err
	}
	count := len(columns)
	values := make([]interface{}, count)
	scanArgs := make([]interface{}, count)
	for i := 0; i < count; i++ {
		scanArgs[i] = &values[i]
	}

	tagMap := make(map[string]string)
	tagValue := make(map[string]reflect.Value)
	e := item.Elem()
	eType := e.Type()
	fieldNumber := e.NumField()
	for i := 0;i < fieldNumber; i++{
		f := eType.Field(i)
		tName := f.Tag.Get(tagName)
		if len(tName) > 0 {
			tagMap[f.Name] = tName
		}else{
			tagMap[f.Name] = strings.ToLower(f.Name)
		}
		tagValue[f.Name] = e.Field(i)
	}

	if qr.Next(){
		qr.Scan(scanArgs...)
		for key, col := range columns {
			var field = col
			for name, tagCol := range tagMap {
				if tagCol == col {
					field = name
					break
				}
			}
			_, ok := tagMap[field]
			if !ok {
				continue
			}
			//todo:
			// 1，解析字段对应正确值，目前类型转换报错，
			// 2，链接查询字段嵌套解析
			var v interface{}
			val := values[key]
			if b, ok := val.([]byte); ok {
				v = string(b)
			} else {
				v = val
			}
			tagValue[field].Set(reflect.ValueOf(v))
		}
	}
	return nil
}
