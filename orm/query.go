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

import "fmt"

type BaseQuery struct {
	driver *driverAlias
	TableInfo
	lastSql string
}

var _ QueryParser = new(BaseQuery)

// Connect change current database connection
// alias must be registered when RegisterDataBase(config)
// otherwise Connect will not work
func (q *BaseQuery) Connect (alias string) QueryParser {
	driverAlias, ok := linkedCache.link[alias]
	if ok {
		q.driver = driverAlias
	}
	return q
}

// Table set default table name, tableName should not contains table prefix
func (q *BaseQuery) Table(tableName string) QueryParser {
	q.tableName = q.driver.Prefix + tableName
	return q
}

// GetTable get current table name, which contains table prefix
func (q *BaseQuery) GetTable() string {
	return q.TableInfo.tableName
}

// Query retrieves data set by sql and bind args
func (q *BaseQuery) Query(query string, args ...interface{}) ([]map[string]interface{}, error) {
	items := make([]map[string]interface{}, 0)
	rows, err := q.driver.Db.Query(query, args...)
	if err != nil {
		DebugLog.log(err)
		return items, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		DebugLog.log(err)
		return items, err
	}
	count := len(columns)
	values := make([]interface{}, count)
	scanArgs := make([]interface{}, count)
	for i := 0; i < count; i++ {
		scanArgs[i] = &values[i]
	}
	for rows.Next(){
		rows.Scan(scanArgs...)
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
		items = append(items, entry)
	}
	if err := rows.Err(); err != nil {
		DebugLog.log(err)
		return make([]map[string]interface{}, 0), err
	}
	return items, nil
}

// Exec execute sql with bind args, which retrieves rows affected numbers
func (q *BaseQuery) Exec(query string, args ...interface{}) (RowsAffected int64, err error) {
	q.lastSql = fmt.Sprintf(query, args...)

	result, err := q.driver.Db.Exec(query, args...)
	if err != nil {
		DebugLog.log(err)
		return
	}
	RowsAffected, err = result.RowsAffected()
	return
}

// LastSql retrieves last execute sql
func (q *BaseQuery) LastSql() string {
	return q.lastSql
}