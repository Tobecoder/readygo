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
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"testing"
)

var config = map[string]interface{}{
	"default": "mysql_dev",
	"mysql_dev": map[string]string{
		"dsn":          "root:123456@tcp(localhost:3306)/gotest?charset=utf8",
		"prefix":       "test_",
		"driver":       "mysql",
		"maxOpenConns": "200",
		"maxIdleConns": "10",
	},
	"mysql_dev2": map[string]string{
		"dsn":          "root:123456@tcp(localhost:3306)/gotest2?charset=utf8",
		"prefix":       "test_",
		"driver":       "mysql",
		"maxOpenConns": "200",
		"maxIdleConns": "10",
	},
}

func TestMain(m *testing.M) {
	err := RegisterDataBase(config)
	if err != nil {
		log.Fatal(err)
	}
	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestOrm(t *testing.T) {
	orm, err := NewOrm("")
	if err != nil {
		t.Fatal(err)
	}
	dataSet, _ := orm.Query("SELECT * FROM test_userinfo")
	data, _ := json.Marshal(dataSet)
	fmt.Println(string(data))
	rows, err := orm.Exec("UPDATE test_userinfo set username = ? where uid = ?", "houhou", "06 and 1=1")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(rows)
	fmt.Println(orm.LastSql())

	//orm.Connect("mysql_dev2")
	orm.Table("userinfo u").
		Where("uid > ? and username = ?", []interface{}{1, "test"}).
		Field("uid").
		Find()
	orm.Table("userinfo u").
		Where("uid").
		Field("uid").
		Find()
	orm.Table("userinfo u").
		Where("uid", []interface{}{">", "1"}, []interface{}{"<", 3}, "or").
		Field("uid").
		Find()
	orm.Table("userinfo u").
		Where("uid", "null").
		Field("uid").
		Find()
	orm.Table("userinfo u").
		Where("uid", 1).
		Field("uid").
		Find()
	orm.Table("userinfo u").
		Where("uid", "in", func (query QueryParser){
			query.Table("userdetail").Field("uid")
		}).
		Field("uid").
		Find()
	orm.Table("userinfo u").
		Where("uid", "in", "1,2,3").
		Field("uid").
		Find()
	orm.Table("userinfo u").
		Where("uid", "in", []interface{}{1,2,3}).
		Field("uid").
		Find()
	t.Fatal("test done")
}
