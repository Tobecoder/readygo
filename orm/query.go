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
	"regexp"
	"container/list"
)

type BaseQuery struct {
	driver *driverAlias
	tableName string
	lastSql string
	builder Builder
	option Option
	ins QueryParser
	bindArgs []interface{}
}

var _ QueryParser = new(BaseQuery)

type QueryClosure func (QueryParser)

// Connection retrieves current sql connection
func (q *BaseQuery) connection() *driverAlias {
	return q.driver
}

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

// Builder return current sql builder
func (q *BaseQuery) Builder() Builder {
	return q.builder
}

// GetTable get current table name, which contains table prefix
func (q *BaseQuery) GetTable() string {
	return q.driver.Prefix + q.tableName
}

// Table set default table name, tableName should not contains table prefix
// usage:
// Table(subSql + " a") sub sql clause
// Table("example_user")
// Table("example_user user,example_role role")
// Table("example_user user")
// Table(map[string]string{"user":"u", "role":"r"})
// Table([]string{"example_user user","example_role role"})
func (q *BaseQuery) Table(tableName interface{}) QueryParser {
	table := make([]string, 0)
	switch v := tableName.(type) {
	case string:
		if strings.Contains(v, ")"){
			//sub query
			table = append(table, v)
		}else if strings.Contains(v, ","){
			tables := strings.Split(v, ",")
			for _, t := range tables {
				alias := strings.SplitN(t, " ", 2)
				if len(alias) == 2 {
					q.tableAlias(alias[0], alias[1])
				}
				table = append(table, alias[0])
			}
		}else if strings.Contains(v, " "){
			alias := strings.SplitN(v, " ", 2)
			q.tableAlias(alias[0], alias[1])
			table = append(table, alias[0])
		}else{
			table = append(table, v)
		}
	case map[string]string:
		for t, alias := range v{
			q.tableAlias(t, alias)
			table = append(table, t)
		}
	case []string:
		for _, t := range v {
			alias := strings.SplitN(t, " ", 2)
			if len(alias) == 2 {
				q.tableAlias(alias[0], alias[1])
			}
			table = append(table, alias[0])
		}
	}
	q.tableName = table[0]
	q.option.table = append(q.option.table, table...)
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
		q.tableAlias(table, aliasName)
	}
	entry := map[string]string{
		"table":table,
		"type":strings.ToUpper(joinType),
		"condition": joinCondition,
	}
	joins = append(joins, entry)
	q.option.join = append(q.option.join, joins...)
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

// WHere assemble where sql clause
// usage:
// Where("uid > ? and username = ?", []interface{}{1, "test"})->( uid > ? and username = ? )
// Where("uid") -> uis is NULL
// Where("uid", []interface{}{">", 1}, []interface{}{"<", 3}, "or")
//     generate sql as -> ( uid > ? OR uid < ? )
//     which support unlimited []interface{}
// Where("uid", "null") -> uid IS NULL
// Where("uid", 1) -> uid = ?
// Where("uid", "in", func (query QueryParser){
//     query.Table("userdetail").Field("uid")
// }) -> uid IN ( SELECT uid FROM test_userdetail  )
// Where("uid", "in", "1,2,3") -> uid IN (?,?,?)
// Where("uid", "in", []interface{}{1,2,3}) -> uid IN (?,?,?)
// Where("uid", "between", []interface{}{1,2}) -> uid BETWEEN ? AND ?
// Where("uid", "between", "1,10") -> uid BETWEEN ? AND ?
// Where("uid", "exists", func (query QueryParser){
//     query.Table("userdetail").Field("uid")
// })
// Where("uid", "exists", "select uid from test_userdetail")
// Where(func(parser QueryParser){
//     parser.Where("id", 1).Where("username", "hehe")
// })
// Where("created", ">= time", "2017-12-01 01:01:01")-> created >= ?
// Where("created", " between time ", "2006-01-02 15:04:05,2006-01-02 15:04:05")
// Where("created", " between time ", []string{"2006-01-02 15:04:05", "2006-01-02 15:04:05"})

func (q *BaseQuery) Where(args ...interface{}) QueryParser {
	field, op, condition, params := q.parseWhereArgs(args...)
	q.parseWhereExp("AND", field, op, condition, params)
	return q
}

func (q *BaseQuery) parseWhereArgs(args ...interface{}) (interface{}, interface{}, interface{}, []interface{}){
	var (
		field interface{}
		op interface{}
		params []interface{}
		condition interface{}
	)
	if args == nil {
		return field, op, condition, params
	}
	field = args[0]
	if len(args) > 1 {
		op = args[1]
		params = args[1:]
	}
	if len(args) > 2 {
		condition = args[2]
	}
	return field, op, condition, params
}

// WhereOr assemble where sql clause
func (q *BaseQuery) WhereOr(args ...interface{}) QueryParser{
	field, op, condition, params := q.parseWhereArgs(args...)
	q.parseWhereExp("OR", field, op, condition, params)
	return q
}

// WhereXor assemble where sql clause
func (q *BaseQuery) WhereXor(args ...interface{}) QueryParser{
	field, op, condition, params := q.parseWhereArgs(args...)
	q.parseWhereExp("XOR", field, op, condition, params)
	return q
}

// WhereTime support convenient time query
// usage:
// WhereTime("created", ">", "2006-01-02 15:04:05")
// WhereTime("created", ">=", "2006-01-02 15:04:05")
// WhereTime("created", "<", "2006-01-02 15:04:05")
// WhereTime("created", "<=", "2006-01-02 15:04:05")
// WhereTime("created", "between", "2006-01-02 15:04:05, 2006-01-02 15:04:05")
// WhereTime("created", "not between", []string{"2006-01-02 15:04:05", "2006-01-02 15:04:05"})
func (q *BaseQuery) WhereTime(field, operator string, value interface{}) QueryParser{
	var operatorMap = map[string]int{">=":1,">":1,"<":1,"<=":1,"between":1,"not between":1}
	if _, ok := operatorMap[operator]; !ok{
		return q
	}
	return q.Where(field, operator + " time", value)
}

// parseWhereExp assemble where options
func (q *BaseQuery) parseWhereExp(logic string, field interface{}, op interface{}, condition interface{}, params []interface{}) {
	if q.option.where.whereMap == nil {
		q.option.where.whereMap = make(whereType)
		// init where list
		q.option.where.list = list.New()
	}
	if q.option.where.whereMap[logic] == nil {
		q.option.where.whereMap[logic] = make(map[string][]interface{})
		// ensure where sort
		listNew := make(whereList)
		listNew[logic] = list.New()
		q.option.where.list.PushBack(listNew)
	}

	var (
		where = make(map[string][]interface{})
		fieldName string
	)

	if v, ok := field.(func (QueryParser)); ok {
		fieldName = "_closure"
		q.option.where.whereMap[logic][fieldName] = append(q.option.where.whereMap[logic][fieldName], v)
		q.option.where.list.Back().Value.(whereList)[logic].PushBack(fieldName)
		return
	}

	regex, err := regexp.Compile(`[,=><'"(\s]`)
	//express where style
	if v, ok := field.(string); ok && err == nil && regex.MatchString(v) {
		//eg:Where("uid > ? and username = ?", []interface{}{1, "test"})
		fieldName = "_exp"
		q.option.where.whereMap[logic][fieldName] = append(q.option.where.whereMap[logic][fieldName], "exp", v)
		if op != nil{
			if bindArgs, ok := op.([]interface{}); ok{
				q.bind(bindArgs)
			}
		}
	}else if op == nil && condition == nil{
		//eg:Where("uid")->uis is NULL
		if fieldName, ok = field.(string); ok && len(fieldName) > 0 {
			q.option.where.whereMap[logic][fieldName] = append(where[fieldName], "null", "")
		}
	}else if _, ok := op.([]interface{}); ok {
		//eg:Where("uid", []interface{}{">", 1}, []interface{}{"<", 3}, "or")
		//support unlimited []interface{}
		if fieldName, ok = field.(string); ok && len(fieldName) > 0 {
			q.option.where.whereMap[logic][fieldName] = append(where[fieldName], params...)
		}
	}else if condition == nil {// equal
		if fieldName, ok = field.(string); ok && len(fieldName) > 0 {
			q.option.where.whereMap[logic][fieldName] = append(where[fieldName], "eq", op)
		}
	}else if v, ok := op.(string); ok {
		nullMap := map[string]int{"null":1,"notnull":1,"not null":1}
		//eg:Where("uid", "null")
		if _, ok := nullMap[strings.ToLower(v)]; ok{
			if fieldName, ok = field.(string); ok && len(fieldName) > 0 {
				q.option.where.whereMap[logic][fieldName] = append(where[fieldName], v, "")
			}
		}else if condition != nil{
			// default operation
			if fieldName, ok = field.(string); ok && len(fieldName) > 0 {
				q.option.where.whereMap[logic][fieldName] = append(where[fieldName], op, condition)
			}
		}
	}
	q.option.where.list.Back().Value.(whereList)[logic].PushBack(fieldName)
}

// bind bind sql args
func (q *BaseQuery) bind(args interface{}) {
	bind := make([]interface{}, 0)
	switch v := args.(type){
	case []interface{}:
		bind = append(bind, v...)
	case interface{}:
		bind = append(bind, v)
	}
	q.bindArgs = append(q.bindArgs, bind...)
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
	bindArgs := q.getBind()
	fmt.Println(bindArgs)
	fmt.Println(sql)
	return nil, nil
}

// getBind returns previous bind args
func (q *BaseQuery) getBind() []interface{}{
	args := q.bindArgs
	q.bindArgs = make([]interface{}, 0)
	return args
}

// BuildSql assemble query sql
func (q *BaseQuery) BuildSql(sub ...bool) string {
	q.option.fetchSql = true
	options := q.parseOptions()
	var isSub = false
	if len(sub) > 0 {
		isSub = sub[0]
	}
	if isSub {
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

func (q *BaseQuery) getOption() Option {
	return q.option
}