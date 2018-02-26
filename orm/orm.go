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
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

var (
	Debug           = false
	DebugLog        = NewLog(os.Stdout)
	DefaultTimeZone = time.Local
	drivers         = make(map[string]QueryParser)
	TypedMySQL      = "mysql"
)

// Log describe database execute log operator
type Log struct {
	*log.Logger
}

// NewLog return the Log operator
func NewLog(out io.Writer) *Log {
	d := new(Log)
	d.Logger = log.New(out, "[ORM] ", log.LstdFlags)
	return d
}

// NewOrm return the query parser
func NewOrm(alias string) (QueryParser, error) {
	if alias == "" {
		alias = linkedCache.Default
	}
	aliasDriver, ok := linkedCache.link[alias]
	if !ok {
		return nil, fmt.Errorf("alias driver %s have not registered", alias)
	}

	parser, ok := drivers[aliasDriver.DriverName]
	if !ok {
		return nil, fmt.Errorf("query parser alias %s have not registered", aliasDriver.DriverName)
	}

	switch aliasDriver.DriverName {
	case TypedMySQL:
		parser.(*mysqlQuery).driver = aliasDriver
	default:
		return nil, fmt.Errorf("alias %s of parse query not found", aliasDriver.DriverName)
	}
	return parser, nil
}

// RegisterQuery register
func RegisterQuery(driverName string, parser QueryParser) {
	if parser == nil {
		panic("parser couldn't not be nil")
	}
	if _, ok := drivers[driverName]; ok {
		panic("couldn't register query " + driverName + " twice")
	}
	drivers[driverName] = parser
}
