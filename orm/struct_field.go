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
	"reflect"
	"fmt"
	"database/sql"
	"time"
)

const (
	TypeBooleanField = 1 << iota
	TypeVarCharField
	TypeDateTimeField
	TypeBitField
	TypeSmallIntegerField
	TypeIntegerField
	TypeBigIntegerField
	TypePositiveBitField
	TypePositiveSmallIntegerField
	TypePositiveIntegerField
	TypePositiveBigIntegerField
	TypeFloatField
	RelReverseOne
	RelReverseMany
)

// Define some logic enum
const (
	IsIntegerField         = ^-TypePositiveBigIntegerField >> 2 << 3
	IsPositiveIntegerField = ^-TypePositiveBigIntegerField >> 6 << 7
	IsRelField             = ^-RelReverseMany >> 11 << 12
	IsFieldType            = ^-RelReverseMany << 1 + 1
)

type structField struct {
	name string
	field reflect.Value
	fieldType int
}

func newField(fieldName string, field reflect.Value) (fi *structField, err error) {
	fi = new(structField)
	fieldType, err := getFieldType(field)
	if err != nil {
		return
	}
	fi.name = fieldName
	fi.field = field
	fi.fieldType = fieldType
	return
}

// return field type as type constant from reflect.Value
func getFieldType(val reflect.Value) (ft int, err error) {
	switch val.Type() {
	case reflect.TypeOf(new(int8)):
		ft = TypeBitField
	case reflect.TypeOf(new(int16)):
		ft = TypeSmallIntegerField
	case reflect.TypeOf(new(int32)),
		reflect.TypeOf(new(int)):
		ft = TypeIntegerField
	case reflect.TypeOf(new(int64)):
		ft = TypeBigIntegerField
	case reflect.TypeOf(new(uint8)):
		ft = TypePositiveBitField
	case reflect.TypeOf(new(uint16)):
		ft = TypePositiveSmallIntegerField
	case reflect.TypeOf(new(uint32)),
		reflect.TypeOf(new(uint)):
		ft = TypePositiveIntegerField
	case reflect.TypeOf(new(uint64)):
		ft = TypePositiveBigIntegerField
	case reflect.TypeOf(new(float32)),
		reflect.TypeOf(new(float64)):
		ft = TypeFloatField
	case reflect.TypeOf(new(bool)):
		ft = TypeBooleanField
	case reflect.TypeOf(new(string)):
		ft = TypeVarCharField
	case reflect.TypeOf(new(time.Time)):
		ft = TypeDateTimeField
	default:
		elm := reflect.Indirect(val)
		switch elm.Kind() {
		case reflect.Int8:
			ft = TypeBitField
		case reflect.Int16:
			ft = TypeSmallIntegerField
		case reflect.Int32, reflect.Int:
			ft = TypeIntegerField
		case reflect.Int64:
			ft = TypeBigIntegerField
		case reflect.Uint8:
			ft = TypePositiveBitField
		case reflect.Uint16:
			ft = TypePositiveSmallIntegerField
		case reflect.Uint32, reflect.Uint:
			ft = TypePositiveIntegerField
		case reflect.Uint64:
			ft = TypePositiveBigIntegerField
		case reflect.Float32, reflect.Float64:
			ft = TypeFloatField
		case reflect.Bool:
			ft = TypeBooleanField
		case reflect.String:
			ft = TypeVarCharField
		default:
			if elm.Interface() == nil {
				err = fmt.Errorf("%s is nil pointer, may be miss setting tag", val)
				return
			}
			switch elm.Interface().(type) {
			case sql.NullInt64:
				ft = TypeBigIntegerField
			case sql.NullFloat64:
				ft = TypeFloatField
			case sql.NullBool:
				ft = TypeBooleanField
			case sql.NullString:
				ft = TypeVarCharField
			case time.Time:
				ft = TypeDateTimeField
			}
		}
	}
	if ft&IsFieldType == 0 {
		err = fmt.Errorf("unsupport field type %s, may be miss setting tag", val)
	}
	return
}