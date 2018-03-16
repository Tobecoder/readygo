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
	Debug           = true
	DebugLog        = NewLog(os.Stdout)
	DefaultTimeZone = time.Local
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

// log logs err
func (l *Log) log(err error) {
	if Debug {
		l.Println(err)
	}
}

// NewOrm return the query parser
// if tableAlias is empty, default tableAlias will be assembled to QueryParser
func NewOrm(alias string) (QueryParser, error) {
	if alias == "" {
		alias = linkedCache.Default
	}
	aliasDriver, ok := linkedCache.link[alias]
	if !ok {
		return nil, fmt.Errorf("tableAlias driver %s have not registered", alias)
	}
	var parser QueryParser

	switch aliasDriver.DriverName {
	case TypedMySQL:
		parser = newMysqlQuery(aliasDriver)
	default:
		return nil, fmt.Errorf("tableAlias %s of parse query not found", aliasDriver.DriverName)
	}
	return parser, nil
}
