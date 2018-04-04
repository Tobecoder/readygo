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

import "container/list"

type unionType string

// Hierarchy is
// 1st,logic operator.such as and,or and so on
// 2nd,field name,such as uid, username and so on
// 3rd,condition presents where condition
type whereType map[string]map[string][]interface{}

type whereList map[string]*list.List

// where
// linked map ensure where options FIFO
type where struct {
	whereMap 	whereType
	list		*list.List
}

// Option query option
type Option struct {
	table      []string
	tableAlias map[string]string
	field      []string
	fieldAlias map[string]string
	where      where
	page       string
	limit      string
	lock       bool
	fetchSql   bool
	distinct   bool
	join       []map[string]string
	union      []interface{}
	unionType  unionType
	group      string
	having     string
	order      []string
	force      string
	comment    string
}

type Builder interface {
	selects(option Option) string // build select sql
	escapeStringQuotes(buf []byte, v string) []byte //escape string quotes
	parseKey(field string, option *Option) string
}

type QueryParser interface {
	connection() *driverAlias // get query current sql connection

	Connect(alias string) QueryParser                                          // change current database connection
	Builder() Builder                                                          // get current builder
	GetTable() string                                                          // get current table name, which contains table prefix
	Table(tableName interface{}) QueryParser                                   // set default table name, tableName should not contains table prefix
	Query(query string, args ...interface{}) ([]map[string]interface{}, error) // execute sql query, return data set
	Exec(query string, args ...interface{}) (int64, error)                     // execute sql query
	//LastInsID()                      // get last insert id
	LastSql() string // get last execute sql
	//Transaction()                    // execute sql database transaction
	//StartTrans()                     // start transaction
	//Commit()                         // commit transaction
	//Rollback()                       // rollback transaction
	Value(fieldName string) (interface{}, error) // retrieves field value
	//PartitionTableName()             // retrieves table partition name
	//Partition()                      // set table partition name's rule
	//Column()                         // retrieves column data set
	//Count()                          // retrieves data set count numbers
	//Sum()                            // retrieves sum value
	//Min()                            // retrieves min value
	//Max()                            // retrieves max value
	//Avg()                            // retrieves avg value
	//SetField()                       // set field's value
	//SetInc()                         // set field's increment step
	//SetDec()                         // set field's decrement step
	Join(join ...string) QueryParser // assemble join clause
	Union(union interface{}) QueryParser // assemble union clause
	UnionAll(union interface{}) QueryParser // assemble union clause
	Field(field interface{}) QueryParser // assemble query fields
	Where(args ...interface{}) QueryParser // assemble query condition
	WhereOr(args ...interface{}) QueryParser                        // assemble or query condition
	WhereXor(args ...interface{}) QueryParser                       // assemble xor query condition
	//WhereNull()                      // assemble null query condition
	//WhereNotNull()                   // assemble not null query condition
	//WhereExists()                    // assemble exist query condition
	//WhereNotExists()                 // assemble not exist query condition
	//WhereIn()                        // assemble in query condition
	//WhereNotIn()                     // assemble not in query condition
	//WhereLike()                      // assemble like query condition
	//WhereNotLike()                   // assemble not like query condition
	//WhereBetween()                   // assemble between query condition
	//WhereNotBetween()                // assemble noe between query condition
	//WhereExp()                       // assemble query express condition
	WhereTime(field, operator string, value interface{}) QueryParser   // assemble time query condition
	//Limit()                          // assemble limit clause
	//Page()                           // assemble page query options
	Comment(comment string) QueryParser // assemble sql comment
	Order(order interface{}) QueryParser // assemble order clause
	Group(group string) QueryParser      // assemble group clause
	Having(having string) QueryParser    // assemble having clause
	Lock() QueryParser                 // assemble for update clause
	Force(index string) QueryParser    // assemble for update clause
	Distinct() QueryParser // assemble distinct clause
	//SetPK()                          // set table primary key
	//TableInfo()                      // retrieves table's information, which contains fields、type、bind、pk
	//Insert()                         // insert data
	//InsertGetId()                    // insert data and retrieves last insert id
	//InsertBatch()                    // batch insert data set
	//SelectInsert()                   // select and insert
	//Update()                         // update data set
	//UpdateBatch()                    // batch update data set
	//Select()                         // select multiple data set
	Find() (interface{}, error) // get one data set
	BuildSql(sub ...bool) string                       // retrieves query sql, don't execute sql actually
	//Delete()                         // delete query

	bind(args interface{}) // bind sql args
	getBind() []interface{} // get bind sql args
	getOption() Option
}
