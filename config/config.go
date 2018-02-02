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

import "fmt"

// Provider defines how to get and set value from configuration raw data.
type Provider interface {
	Set(key, value string) error // set config data
	Get(key string) string
	Has(key string) bool                                  // check config exists
	SaveFile(filename string) error                       // save config data
	GetSection(section string) (map[string]string, error) //

	String(key string) string
	Strings(key string) []string
	Int(key string) (int, error)
	Int64(key string) (int64, error)
	Bool(key string) (bool, error)
	Float(key string) (float64, error)
	DefaultString(key, defaultVal string) string
	DefaultStrings(key string, defaultVal []string) []string //get string slice
	DefaultInt(key string, defaultVal int) int
	DefaultInt64(key string, defaultVal int64) int64
	DefaultBool(key string, defaultVal bool) bool
	DefaultFloat(key string, defaultVal float64) float64
}

// Config defines how to parse value from configuration file and bytes data
type Config interface {
	Parse(key string) (Provider, error)      // parse config data from file
	ParseData(data []byte) (Provider, error) // parse config data from byte
}

var adapters = make(map[string]Config)

// NewConfig adapterName is ini/json/xml/yaml.
// fileName is the config file path.
func NewConfig(adapterName, fileName string) (Provider, error) {
	adapter, ok := adapters[adapterName]
	if !ok {
		return nil, fmt.Errorf("new config: unknown adapter %s, register it first please", adapterName)
	}
	return adapter.Parse(fileName)
}

// NewConfig adapterName is ini/json/xml/yaml.
// data is the config byte data.
func NewConfigData(adapterName string, data []byte) (Provider, error) {
	adapter, ok := adapters[adapterName]
	if !ok {
		return nil, fmt.Errorf("new config: unknown adapter %s, register it first please", adapterName)
	}
	return adapter.ParseData(data)
}

// Register register adaptor for config
func Register(name string, adapter Config) {
	if adapter == nil {
		panic("adapter is nil")
	}
	if _, ok := adapters[name]; ok {
		panic("adapter " + name + " existed")
	}
	adapters[name] = adapter
}
