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

package config

import (
	"errors"
	"fmt"
)

// ParseBool convert value to bool
func ParseBool(value interface{}) (bool, error) {
	if value == nil {
		return false, errors.New("value is nil")
	}

	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		switch v {
		case "1", "t", "T", "true", "TRUE", "True", "On", "ON", "on":
			return true, nil
		case "0", "f", "F", "false", "FALSE", "False", "Off", "OFF", "off":
			return false, nil
		}
	case int8, int32, int64:
		strVal := fmt.Sprintf("%d", v)
		if strVal == "1" {
			return true, nil
		} else if strVal == "0" {
			return false, nil
		}
	case float64:
		if v == 1.0 {
			return true, nil
		} else if v == 0.0 {
			return false, nil
		}
	}
	return false, fmt.Errorf("parsing %q: invalid syntax", value)
}
