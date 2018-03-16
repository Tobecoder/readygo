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
	"database/sql"
	"fmt"
	"strconv"
	"sync"
	"time"
)

var (
	linkedCache = &connection{
		link: make(map[string]*driverAlias),
	}
)

// connection describe database connection stacks
type connection struct {
	sync.RWMutex
	link    map[string]*driverAlias
	Default string
}
// add add alias name of driver
func (conn *connection) add(aliasName string, driverAlias *driverAlias) (added bool) {
	conn.Lock()
	defer conn.Unlock()
	if _, ok := conn.link[aliasName]; !ok {
		conn.link[aliasName] = driverAlias
		added = true
	}
	return
}

// driverAlias describe database driver by alias name
type driverAlias struct {
	Name           string
	DriverName     string
	DataSourceName string
	Db             *sql.DB
	MaxIdleConns   int
	MaxOpenConns   int
	TimeZone       *time.Location
	Engine         string
	Prefix         string
}

//DbConfig describe db connection config
type DbConfig map[string]interface{}

// RegisterDataBase register database link
// var DbConfig = map[string]interface{}{
//	// Default database configuration
//	"default": "mysql_dev",
//	// Define the database configuration character "mysql_dev".
//	"mysql_dev": map[string]string{
//		"dsn":		"root:123456@tcp(localhost:3306)/gotest?charset=utf8"
//		"prefix":   "test_",
//		"driver":   "mysql",
//		"maxOpenConns": 300,
//		"maxIdleConns": 10,
//	},
//	......
//}
func RegisterDataBase(config DbConfig) error {
	for key, value := range config {
		if key == "default" && len(value.(string)) > 0 {
			linkedCache.Default = value.(string)
		} else if cfg, ok := value.(map[string]string); ok {
			db, err := sql.Open(cfg["driver"], cfg["dsn"])
			if err != nil {
				return fmt.Errorf("register database driver `%s` error : %s", key, err.Error())
			}

			alias, err := addAliasWithDB(key, cfg["driver"], db)
			if err != nil {
				db.Close()
				return err
			}
			alias.DataSourceName = cfg["dsn"]
			alias.Prefix = cfg["prefix"]

			detectTimeZone(alias)

			maxIdleConns, _ := strconv.Atoi(cfg["maxIdleConns"])
			if maxIdleConns > 0 {
				SetMaxIdleConns(alias.Name, maxIdleConns)
			}

			maxOpenConns, _ := strconv.Atoi(cfg["maxOpenConns"])
			if maxOpenConns > 0 {
				SetMaxOpenConns(alias.Name, maxOpenConns)
			}
		}
	}
	return nil
}

// SetMaxIdleConns Change the max idle connections for *sql.DB, use specify database alias name
func SetMaxIdleConns(aliasName string, maxIdleConns int) {
	alias := getDbWithAlias(aliasName)
	alias.MaxIdleConns = maxIdleConns
	alias.Db.SetMaxIdleConns(maxIdleConns)
}

// SetMaxOpenConns Change the max open conns for *sql.DB, use specify database alias name
func SetMaxOpenConns(aliasName string, maxOpenConns int) {
	alias := getDbWithAlias(aliasName)
	alias.MaxOpenConns = maxOpenConns
	alias.Db.SetMaxOpenConns(maxOpenConns)
}

// getDbWithAlias get table driver alias
func getDbWithAlias(aliasName string) *driverAlias {
	if al, ok := linkedCache.link[aliasName]; ok {
		return al
	}
	panic(fmt.Errorf("unknown database alias name %s", aliasName))
}

// AddAliasWithDB add a aliasName for the driver name
func addAliasWithDB(aliasName, driverName string, db *sql.DB) (*driverAlias, error) {
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("the connection to the database is unalive, err:%s", err.Error())
	}

	alias := new(driverAlias)
	alias.Name = aliasName
	alias.DriverName = driverName
	alias.Db = db

	if !linkedCache.add(aliasName, alias) {
		return nil, fmt.Errorf("database alias name `%s` already registered, cannot reuse", aliasName)
	}
	return alias, nil
}

// detectTimeZone detect database timezone
func detectTimeZone(alias *driverAlias) {
	alias.TimeZone = DefaultTimeZone

	switch alias.DriverName {
	case TypedMySQL:
		row := alias.Db.QueryRow("SELECT TIMEDIFF(NOW(), UTC_TIMESTAMP)")
		var tz string
		row.Scan(&tz)
		if len(tz) >= 8 {
			if tz[0] != '-' {
				tz = "+" + tz
			}
			t, err := time.Parse("-07:00:00", tz)
			if err == nil {
				if t.Location().String() != "" {
					alias.TimeZone = t.Location()
				}
			} else {
				DebugLog.Printf("Detect DB timezone: %s %s\n", tz, err.Error())
			}
		}

		// get default engine from current database
		row = alias.Db.QueryRow("SELECT ENGINE, TRANSACTIONS FROM information_schema.engines WHERE SUPPORT = 'DEFAULT'")
		var engine string
		var tx bool
		row.Scan(&engine, &tx)

		if engine != "" {
			alias.Engine = engine
		} else {
			alias.Engine = "INNODB"
		}
	}
}
