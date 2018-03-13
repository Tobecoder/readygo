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

type TableInfo struct {
	tableName string
}

type Builder interface {
}

type QueryParser interface {
	Connect(alias string) QueryParser // change current database connection
	//SetBuilder()                     // set current builder
	//Builder() Builder                // get current builder
	GetTable() string      // get current table name, which contains table prefix
	Table(tableName string) QueryParser // set default table name, tableName should not contains table prefix
	//ParseSqlTable(sql string) string // replace __TABLE_NAME__ in sql with table name in lowercase, which contains table prefix
	Query(query string, args ...interface{}) ([]map[string]interface{}, error)  // execute sql query, return data set
	Exec(query string, args ...interface{}) (int64, error)   // execute sql query
	//LastInsID()                      // get last insert id
	LastSql() string                 // get last execute sql
	//Transaction()                    // execute sql database transaction
	//StartTrans()                     // start transaction
	//Commit()                         // commit transaction
	//Rollback()                       // rollback transaction
	//Value()                          // retrieves field value
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
	//Join()                           // assemble join clause
	//Union()                          // assemble union clause
	//Field()                          // assemble query fields
	//Where()                          // assemble query condition
	//WhereOr()                        // assemble or query condition
	//WhereXor()                       // assemble xor query condition
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
	//WhereTime()                      // assemble time query condition
	//Limit()                          // assemble limit clause
	//Page()                           // assemble page query options
	//Order()                          // assemble order clause
	//Group()                          // assemble group clause
	//Having()                         // assemble having clause
	//Lock()                           // assemble for update clause
	//Distinct()                       // assemble distinct clause
	//SetPK()                          // set table primary key
	//TableInfo()                      // retrieves table's information, which contains fields、type、bind、pk
	//Insert()                         // insert data
	//InsertGetId()                    // insert data and retrieves last insert id
	//InsertBatch()                    // batch insert data set
	//SelectInsert()                   // select and insert
	//Update()                         // update data set
	//UpdateBatch()                    // batch update data set
	//Select()                         // select multiple data set
	//Find()                           // get one data set
	//BuildSql()                       // retrieves query sql, don't execute sql actually
	//Delete()                         // delete query
}
