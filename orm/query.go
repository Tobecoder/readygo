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
	"strings"
	"strconv"
)

type BaseQuery struct {
	driver *driverAlias
	tableName string
	lastSql string
	builder Builder
	option Option
}

var _ QueryParser = new(BaseQuery)

type QueryClosure func (QueryParser)

// Connection retrieves current sql connection
func (q *BaseQuery) Connection() *driverAlias {
	return q.driver
}

// Connect change current database connection
// tableAlias must be registered when RegisterDataBase(config)
// otherwise Connect will not work
func (q *BaseQuery) Connect (alias string) QueryParser {
	driverAlias, ok := linkedCache.link[alias]
	if ok {
		q.driver = driverAlias
	}
	return q
}

// Builder return current sql builder
func (q *BaseQuery) Builder() Builder {
	return q.builder
}

// GetTable get current table name, which contains table prefix
func (q *BaseQuery) GetTable() string {
	return q.tableName
}

// Table set default table name, tableName should not contains table prefix
// usage:
// Table("example_user")
// Table("example_user user,example_role role")
// Table("example_user user")
// Table(map[string]string{"user":"u", "role":"r"})
// Table([]string{"example_user user","example_role role"})
func (q *BaseQuery) Table(tableName interface{}) QueryParser {
	table := make([]string, 0)
	prefix := q.driver.Prefix
	switch v := tableName.(type) {
	case string:
		if strings.Contains(v, ")"){
			//sub query
		}else if strings.Contains(v, ","){
			tables := strings.Split(v, ",")
			for _, t := range tables {
				alias := strings.SplitN(t, " ", 2)
				prefixTable := prefix + alias[0]
				if len(alias) == 2 {
					q.tableAlias(prefixTable, alias[1])
				}
				table = append(table, prefixTable)
			}
		}else if strings.Contains(v, " "){
			alias := strings.SplitN(v, " ", 2)
			prefixTable := prefix + alias[0]
			q.tableAlias(prefixTable, alias[1])
			table = append(table, prefixTable)
		}else{
			table = append(table, prefix + v)
		}
	case map[string]string:
		for t, alias := range v{
			q.tableAlias(t, alias)
			table = append(table, prefix + t)
		}
	case []string:
		for _, t := range v {
			alias := strings.SplitN(t, " ", 2)
			prefixTable := prefix + alias[0]
			if len(alias) == 2 {
				q.tableAlias(prefixTable, alias[1])
			}
			table = append(table, prefixTable)
		}
	}
	q.option.table = table
	return q
}

// tableAlias assemble table tableAlias to Option
func (q *BaseQuery) tableAlias(table, alias string) QueryParser{
	if q.option.tableAlias == nil {
		q.option.tableAlias = make(map[string]string)
	}
	q.option.tableAlias[table] = alias
	return q
}

// Query retrieves data set by sql and bind args
func (q *BaseQuery) Query(query string, args ...interface{}) ([]map[string]interface{}, error) {
	q.lastSql = fmt.Sprintf(query, args...)

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

// Value retrieves field value
func (q *BaseQuery) Value (fieldName string) (interface{}, error){
	q.option.field = []string{fieldName}
	result, err := q.Find()
	if err != nil {
		return nil, err
	}
	// clean the option
	q.option = Option{}
	return result, err
}

// Join assemble table join to sql query, table name prefix is filled auto
// usage
// Join("work w")
// Join("work w", "w.id = s.id")
// Join("work w", "w.id = s.id", "INNER")
func (q *BaseQuery) Join(join ...string) QueryParser {
	if join == nil {
		return q
	}
	joins := make([]map[string]string, 0)
	var (
		aliasName string
		table string
		joinCondition string
		joinType = "INNER"
	)
	if len(join) >= 3 {
		joinType = join[2]
	}
	if len(join) >= 2 {
		joinCondition = join[1]
	}
	joinTable := strings.TrimSpace(join[0])
	if strings.Contains(joinTable, ")"){
		table = joinTable
	}else{
		if strings.Contains(joinTable, " ") {
			alias := strings.SplitN(joinTable, " ", 2)
			table = alias[0]
			aliasName = alias[1]
		}else{
			table = joinTable
			aliasName = joinTable
		}
	}
	if len(aliasName) > 0 {
		q.tableAlias(q.driver.Prefix + table, aliasName)
	}
	entry := map[string]string{
		"table":q.driver.Prefix + table,
		"type":strings.ToUpper(joinType),
		"condition": joinCondition,
	}
	joins = append(joins, entry)
	q.option.join = joins
	return q
}

// Union assemble union sql clause
// usage:
// Union("SELECT name FROM example_user_1")
// Union(func (query QueryParser){
// 	query.Table("test").Field("name")
// })
// Union([]string{"SELECT name FROM example_user_1", "SELECT name FROM example_user_2"})
func (q *BaseQuery) Union(union interface{}) QueryParser {
	if union == nil {
		return q
	}
	q.option.unionType = "UNION "
	q.option.union = append(q.option.union, union)
	return q
}

// UnionAll assemble union all sql clause
// usage: the same as Union
func (q *BaseQuery) UnionAll(union interface{}) QueryParser {
	if union == nil {
		return q
	}
	q.option.unionType = "UNION ALL "
	q.option.union = append(q.option.union, union)
	return q
}

// Field assemble query field to Option, default field is *
// usage
// Field("id,title,content")
// Field("id,sum(num) total")
// Field("id,sum(num) as total")
// Field([]string{"id", "title", "content"})
// Field([]string{"id", "title t", "content"})
// Field([]string{"id", `concat(name,"-",id) truename`, "LEFT(title,7) subtitle"})
func (q *BaseQuery) Field(field interface{}) QueryParser {
	if field == nil {
		return q
	}
	fields := make([]string, 0)
	var fieldSlice []string
	switch v := field.(type){
	case string:
		fieldSlice = strings.Split(v, ",")
	case []string:
		fieldSlice = v
	}
	for _, f := range fieldSlice{
		if strings.Contains(strings.ToUpper(f), "AS") {
			fields = append(fields, f)
		}else{
			alias := strings.SplitN(strings.TrimSpace(f), " ", 2)
			if len(alias) == 2 {
				q.fieldAlias(alias[0], alias[1])
			}
			fields = append(fields, alias[0])
		}
	}
	q.option.field = fields
	return q
}

// fieldAlias assemble field's alias
func (q *BaseQuery) fieldAlias(field, alias string) QueryParser{
	if q.option.fieldAlias == nil {
		q.option.fieldAlias = make(map[string]string)
	}
	q.option.fieldAlias[field] = alias
	return q
}

// Comment assemble sql comment
func (q *BaseQuery) Comment(comment string) QueryParser {
	q.option.comment = comment
	return q
}

// Order assemble order sql clause
// usage:
// Order("id desc,username")
// Order([]string{"id desc","username"})
func (q *BaseQuery) Order(order interface{}) QueryParser {
	if order == nil {
		return q
	}
	orders := make([]string, 0)
	switch v := order.(type) {
	case string:
		orders = append(orders, v)
	case []string:
		orders = append(orders, v...)
	}
	q.option.order = orders
	return q
}

// Group assemble group sql clause
func (q *BaseQuery) Group (group string) QueryParser {
	q.option.group = group
	return q
}

// Having assemble having sql clause
func (q *BaseQuery) Having(having string) QueryParser {
	q.option.having = having
	return q
}

// Force assemble force index sql clause
// usage
// Force
func (q *BaseQuery) Force(index string) QueryParser {
	q.option.force = index
	return q
}

// Lock assemble for update sql clause
func (q *BaseQuery) Lock () QueryParser {
	q.option.lock = true
	return q
}

// Distinct assemble distinct query field
func (q *BaseQuery) Distinct () QueryParser{
	q.option.distinct = true
	return q
}

// Find retrieves one data set
func (q *BaseQuery) Find() (interface{}, error){
	q.option.limit = "1"
	options := q.parseOptions()
	sql := q.builder.selects(options)
	fmt.Println(sql)
	return nil, nil
}

// BuildSql assemble query sql
func (q *BaseQuery) BuildSql(sub bool) string {
	q.option.fetchSql = true
	options := q.parseOptions()
	if sub {
		return "( " + q.builder.selects(options) + " )"
	}else{
		return q.builder.selects(options)
	}
}

// parseOptions parse query options
func (q *BaseQuery) parseOptions() Option{
	options := q.option
	if len(options.table) == 0 {
		options.table = []string{q.GetTable()}
	}
	if options.where == nil {
		options.where = make([]string, 0)
	}
	if len(options.page) > 0 {
		pages := strings.Split(options.page, ",")
		var (
			page int
			rows int
			err error
		)
		page, err = strconv.Atoi(pages[0])
		if err != nil {
			page = 1
		}
		if len(pages) == 2 {
			rows, err = strconv.Atoi(pages[1])
			if err != nil {
				rows = 10
			}
		}
		if page <= 0 {
			page = 1
		}
		if rows <= 0 {
			rows = 10
		}
		offset := rows * (page - 1)
		options.limit = strconv.Itoa(offset) + "," + strconv.Itoa(rows)
	}
	q.option = Option{}
	return options
}