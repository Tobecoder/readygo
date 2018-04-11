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
	"fmt"
	"time"
)

type queryRows struct {
	*sql.Rows
	queryIns QueryParser
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
	tagValue := make(map[string]*structField)
	e := item.Elem()
	eType := e.Type()
	fieldNumber := e.NumField()
	for i := 0;i < fieldNumber; i++{
		f := eType.Field(i)
		tagMap[f.Name] = strings.ToLower(f.Name)
		tagValue[f.Name], err = newField(f.Name, e.Field(i))
		if err != nil {
			return err
		}
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
			// 2，链接查询字段嵌套解析
			var v interface{}
			val := values[key]
			if b, ok := val.([]byte); ok {
				v = string(b)
			} else {
				v = val
			}
			v, err = qr.convertValueFromDB(tagValue[field], v)
			if err != nil {
				return err
			}
			qr.setFieldVal(tagValue[field], v)
		}
	}
	return nil
}

// convertValueFromDB convert value form db to suite struct
func (qr *queryRows) convertValueFromDB(field *structField, val interface{}) (interface{}, error){
	if val == nil {
		return nil, nil
	}

	var value interface{}
	var tErr error

	var str *StrTo
	switch v := val.(type) {
	case []byte:
		s := StrTo(string(v))
		str = &s
	case string:
		s := StrTo(v)
		str = &s
	}

	fieldType := field.fieldType
setValue:
	switch {
	case fieldType == TypeBooleanField:
		if str == nil {
			switch v := val.(type) {
			case int64:
				b := v == 1
				value = b
			default:
				s := StrTo(ToStr(v))
				str = &s
			}
		}
		if str != nil {
			b, err := str.Bool()
			if err != nil {
				tErr = err
				goto end
			}
			value = b
		}
	case fieldType == TypeVarCharField:
		if str == nil {
			value = ToStr(val)
		} else {
			value = str.String()
		}
	case fieldType == TypeDateTimeField:
		tz := qr.queryIns.connection().TimeZone
		if str == nil {
			switch t := val.(type) {
			case time.Time:
				tmpT := &t
				value = tmpT.In(tz)
			default:
				s := StrTo(ToStr(t))
				str = &s
			}
		}
		if str != nil {
			s := str.String()
			var (
				t   time.Time
				err error
			)
			if len(s) >= 19 {
				s = s[:19]
				t, err = time.ParseInLocation(formatDateTime, s, tz)
			} else if len(s) >= 10 {
				if len(s) > 10 {
					s = s[:10]
				}
				t, err = time.ParseInLocation(formatDate, s, tz)
			} else if len(s) >= 8 {
				if len(s) > 8 {
					s = s[:8]
				}
				t, err = time.ParseInLocation(formatTime, s, tz)
			}
			t = t.In(DefaultTimeZone)

			if err != nil && s != "00:00:00" && s != "0000-00-00" && s != "0000-00-00 00:00:00" {
				tErr = err
				goto end
			}
			value = t
		}
	case fieldType&IsIntegerField > 0:
		if str == nil {
			s := StrTo(ToStr(val))
			str = &s
		}
		if str != nil {
			var err error
			switch fieldType {
			case TypeBitField:
				_, err = str.Int8()
			case TypeSmallIntegerField:
				_, err = str.Int16()
			case TypeIntegerField:
				_, err = str.Int32()
			case TypeBigIntegerField:
				_, err = str.Int64()
			case TypePositiveBitField:
				_, err = str.Uint8()
			case TypePositiveSmallIntegerField:
				_, err = str.Uint16()
			case TypePositiveIntegerField:
				_, err = str.Uint32()
			case TypePositiveBigIntegerField:
				_, err = str.Uint64()
			}
			if err != nil {
				tErr = err
				goto end
			}
			if fieldType&IsPositiveIntegerField > 0 {
				v, _ := str.Uint64()
				value = v
			} else {
				v, _ := str.Int64()
				value = v
			}
		}
	case fieldType == TypeFloatField:
		if str == nil {
			switch v := val.(type) {
			case float64:
				value = v
			default:
				s := StrTo(ToStr(v))
				str = &s
			}
		}
		if str != nil {
			v, err := str.Float64()
			if err != nil {
				tErr = err
				goto end
			}
			value = v
		}
	case fieldType&IsRelField > 0:
		//todo:处理嵌套结构体的字段值转换
		//fi = fi.relModelInfo.fields.pk
		//fieldType = fi.fieldType
		goto setValue
	}

end:
	if tErr != nil {
		err := fmt.Errorf("convert to `%s` failed, field: %s err: %s", field.field.Type(), field.name, tErr)
		return nil, err
	}

	return value, nil
}

// setFieldVal set struct field value
func (qr *queryRows) setFieldVal(fieldInfo *structField, value interface{}){
	field := fieldInfo.field
	fieldType := fieldInfo.fieldType

setValue:
	switch {
	case fieldType == TypeBooleanField:
		if nb, ok := field.Interface().(sql.NullBool); ok {
			if value == nil {
				nb.Valid = false
			} else {
				nb.Bool = value.(bool)
				nb.Valid = true
			}
			field.Set(reflect.ValueOf(nb))
		} else if field.Kind() == reflect.Ptr {
			if value != nil {
				v := value.(bool)
				field.Set(reflect.ValueOf(&v))
			}
		} else {
			if value == nil {
				value = false
			}
			field.SetBool(value.(bool))
		}
	case fieldType == TypeVarCharField:
		if ns, ok := field.Interface().(sql.NullString); ok {
			if value == nil {
				ns.Valid = false
			} else {
				ns.String = value.(string)
				ns.Valid = true
			}
			field.Set(reflect.ValueOf(ns))
		} else if field.Kind() == reflect.Ptr {
			if value != nil {
				v := value.(string)
				field.Set(reflect.ValueOf(&v))
			}
		} else {
			if value == nil {
				value = ""
			}
			field.SetString(value.(string))
		}
	case fieldType == TypeDateTimeField:
			if value == nil {
				value = time.Time{}
			} else if field.Kind() == reflect.Ptr {
				if value != nil {
					v := value.(time.Time)
					field.Set(reflect.ValueOf(&v))
				}
			} else {
				field.Set(reflect.ValueOf(value))
			}
	case fieldType == TypePositiveBitField && field.Kind() == reflect.Ptr:
		if value != nil {
			v := uint8(value.(uint64))
			field.Set(reflect.ValueOf(&v))
		}
	case fieldType == TypePositiveSmallIntegerField && field.Kind() == reflect.Ptr:
		if value != nil {
			v := uint16(value.(uint64))
			field.Set(reflect.ValueOf(&v))
		}
	case fieldType == TypePositiveIntegerField && field.Kind() == reflect.Ptr:
		if value != nil {
			if field.Type() == reflect.TypeOf(new(uint)) {
				v := uint(value.(uint64))
				field.Set(reflect.ValueOf(&v))
			} else {
				v := uint32(value.(uint64))
				field.Set(reflect.ValueOf(&v))
			}
		}
	case fieldType == TypePositiveBigIntegerField && field.Kind() == reflect.Ptr:
		if value != nil {
			v := value.(uint64)
			field.Set(reflect.ValueOf(&v))
		}
	case fieldType == TypeBitField && field.Kind() == reflect.Ptr:
		if value != nil {
			v := int8(value.(int64))
			field.Set(reflect.ValueOf(&v))
		}
	case fieldType == TypeSmallIntegerField && field.Kind() == reflect.Ptr:
		if value != nil {
			v := int16(value.(int64))
			field.Set(reflect.ValueOf(&v))
		}
	case fieldType == TypeIntegerField && field.Kind() == reflect.Ptr:
		if value != nil {
			if field.Type() == reflect.TypeOf(new(int)) {
				v := int(value.(int64))
				field.Set(reflect.ValueOf(&v))
			} else {
				v := int32(value.(int64))
				field.Set(reflect.ValueOf(&v))
			}
		}
	case fieldType == TypeBigIntegerField && field.Kind() == reflect.Ptr:
		if value != nil {
			v := value.(int64)
			field.Set(reflect.ValueOf(&v))
		}
	case fieldType&IsIntegerField > 0:
		if fieldType&IsPositiveIntegerField > 0 {
			if value == nil {
				value = uint64(0)
			}
			field.SetUint(value.(uint64))
		} else {
			if ni, ok := field.Interface().(sql.NullInt64); ok {
				if value == nil {
					ni.Valid = false
				} else {
					ni.Int64 = value.(int64)
					ni.Valid = true
				}
				field.Set(reflect.ValueOf(ni))
			} else {
				if value == nil {
					value = int64(0)
				}
				field.SetInt(value.(int64))
			}
		}
	case fieldType == TypeFloatField:
		if nf, ok := field.Interface().(sql.NullFloat64); ok {
			if value == nil {
				nf.Valid = false
			} else {
				nf.Float64 = value.(float64)
				nf.Valid = true
			}
			field.Set(reflect.ValueOf(nf))
		} else if field.Kind() == reflect.Ptr {
			if value != nil {
				if field.Type() == reflect.TypeOf(new(float32)) {
					v := float32(value.(float64))
					field.Set(reflect.ValueOf(&v))
				} else {
					v := value.(float64)
					field.Set(reflect.ValueOf(&v))
				}
			}
		} else {

			if value == nil {
				value = float64(0)
			}
			field.SetFloat(value.(float64))
		}
	case fieldType&IsRelField > 0:
		if value != nil {
			//todo:处理嵌套结构体的字段值转换
			//fieldType = fi.relModelInfo.fields.pk.fieldType
			//mf := reflect.New(fi.relModelInfo.addrField.Elem().Type())
			//field.Set(mf)
			//f := mf.Elem().FieldByIndex(fi.relModelInfo.fields.pk.fieldIndex)
			//field = f
			goto setValue
		}
	}
}
