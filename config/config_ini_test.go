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
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
)

var (
	container  Provider
	err        error
	configFile = "./test_files/php.ini"
	saveFile   = "./test_files/php_test.ini"
)

func TestMain(m *testing.M) {
	container, err = NewConfig("ini", configFile)
	if err != nil {
		log.Fatal(err)
	}
	exitCode := m.Run()
	os.Remove(saveFile)
	os.Exit(exitCode)
}

func TestIni(t *testing.T) {
	// test Get
	container.Get("")
	engine := container.Get("php.engine")
	isEngine, _ := ParseBool(engine)
	if !isEngine {
		t.Fatal("get php.engine error")
	}
	// test Set
	container.Set("", "")
	err = container.Set("php.engine", "off")
	if err != nil {
		t.Fatal(err)
	}
	engine = container.Get("php.engine")
	isEngine, _ = ParseBool(engine)
	if isEngine {
		t.Fatal("set php.engine error")
	}
	container.Set("test.aaa", "test1")
	if container.Get("test.aaa") != "test1" {
		t.Fatal("set test1 error")
	}
	// test Has
	if !container.Has("php.engine") {
		t.Fatal("file has php.engine setting")
	}
	if container.Has("") {
		t.Fatal("has should return \"\"")
	}
	if container.Has("test.bbb") {
		t.Fatal("has should return \"\"")
	}
	if container.Has("test11.bbb") {
		t.Fatal("has should return \"\"")
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
	if c, _ := container.GetSection(""); c["test"] != "123456" {
		t.Fatal("get section value error")
	}
	if _, err = container.GetSection("hehe"); err == nil {
		t.Fatal("section hehe shouldn't exist")
	}
}
func TestHelper(t *testing.T) {
	// test String
	if tags := container.String("session.url_rewriter.tags"); tags != "a=href,area=href,frame=src,input=src,form=fakeentry" {
		t.Fatal("String failed")
	}
	// test Strings
	if s := container.Strings("php.asp_tags"); strings.Join(s, "") != strings.Join([]string{"Off"}, "") {
		t.Fatal("Strings failed")
	}
	// test Int
	_, err = container.Int("php.precision")
	if err != nil {
		t.Fatal(err)
	}
	// test Int64
	_, err = container.Int64("php.precision")
	if err != nil {
		t.Fatal(err)
	}
	// test Bool
	b, err := container.Bool("php.engine")
	if err != nil || b {
		t.Fatal(err)
	}
	// test Float
	if _, err = container.Float("php.precision"); err != nil {
		t.Fatal(err)
	}
	// test DefaultString
	if v := container.DefaultString("php.aaa", ""); v != "" {
		t.Fatal("DefaultString value isn't empty")
	}
	// test DefaultStrings
	if v := container.DefaultStrings("php.aaa", []string{}); strings.Join(v, "") != "" {
		t.Fatal("DefaultStrings value isn't empty")
	}
	// test DefaultInt
	if v := container.DefaultInt("php.aaa", 0); v != 0 {
		t.Fatal("php.aaa value is 0")
	}
	// test DefaultInt64
	if v := container.DefaultInt64("php.aaa", 0); v != 0 {
		t.Fatal("php.aaa value is 0")
	}
	// test DefaultBool
	if v := container.DefaultBool("php.aaa", false); v {
		t.Fatal("php.aaa isn't false")
	}
	// test DefaultFloat
	if v := container.DefaultFloat("php.aaa", 0.00); v < 0.00 {
		t.Fatal("php.aaa is less than 0.00")
	}
}

func TestExpose(t *testing.T) {
	// test NewXXX
	b, _ := ioutil.ReadFile(configFile)
	_, err := NewConfigData("ini", b)
	if err != nil {
		t.Fatal(err)
	}
	NewConfig("aaa", configFile)
	NewConfigData("aaa", []byte{})

	// test parseSectionKey
	c := new(Container)
	section, key := c.parseSectionKey("")
	if section != "" || key != "" {
		t.Fatal("parse section error")
	}
	section, key = c.parseSectionKey("aaa")
	if section != defaultSection || key != "aaa" {
		t.Fatal("parse section error")
	}
}
func TestParseBool(t *testing.T) {
	if _, err := ParseBool(nil); err == nil {
		t.Fatal("err shouldn't be nil")
	}
	if b, _ := ParseBool(false); b {
		t.Fatal("bool should be false")
	}
	if b, _ := ParseBool(int64(1)); !b {
		t.Fatal("bool should be true")
	}
	if b, _ := ParseBool(int64(0)); b {
		t.Fatal("bool should be false")
	}
	if b, _ := ParseBool(1.00); !b {
		t.Fatal("bool should be true")
	}
	if b, _ := ParseBool(0.00); b {
		t.Fatal("bool should be false")
	}
}
