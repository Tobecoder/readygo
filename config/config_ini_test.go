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

package config

import (
	"log"
	"os"
	"testing"
)

var (
	container  Provider
	err        error
	configFile = "./test_files/php.ini"
	saveFile = "./test_files/php_test.ini"
)

func TestMain(m *testing.M) {
	container, err = NewConfig("ini", configFile)
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(m.Run())
}

func TestIni(t *testing.T) {
	// test Get
	engine := container.Get("php.engine")
	isEngine, _ := ParseBool(engine)
	if !isEngine {
		t.Fatal("get php.engine error")
	}
	// test Set
	err = container.Set("php.engine", "off")
	if err != nil {
		t.Fatal(err)
	}
	engine = container.Get("php.engine")
	isEngine, _ = ParseBool(engine)
	if isEngine {
		t.Fatal("set php.engine error")
	}
	// test Has
	if !container.Has("php.engine") {
		t.Fatal("file has php.engine setting")
	}
	// test SaveFile
	if err := container.SaveFile(saveFile); err != nil {
		t.Fatal(err)
	}
	// test GetSection
	section, err := container.GetSection("php")
	if err != nil {
		t.Fatal(err)
	}
	if section["engine"] != engine {
		t.Fatal("get section data error")
	}
}
