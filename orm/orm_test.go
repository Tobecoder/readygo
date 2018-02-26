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
	"testing"
	"os"
	"log"
	_ "github.com/go-sql-driver/mysql"
)

var config = map[string]interface{}{
	"default": "mysql_dev",
	"mysql_dev":map[string]string{
		"dsn": "root:123456@tcp(localhost:3306)/gotest?charset=utf8",
		"prefix": "",
		"driver": "mysql",
		"maxOpenConns": "200",
		"maxIdleConns": "10",
	},
	"mysql_dev2":map[string]string{
		"dsn": "root:123456@tcp(localhost:3306)/gotest2?charset=utf8",
		"prefix": "test_",
		"driver": "mysql",
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
	_, err := NewOrm("")
	if err != nil {
		t.Fatal(err)
	}
}
