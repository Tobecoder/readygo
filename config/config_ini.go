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
	"bufio"
	"bytes"
	"container/list"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
)

var (
	defaultSection   = "common"
	byteEmpty        = []byte{}
	byteWellNumber   = []byte{'#'} // comment
	byteSemicolon    = []byte{';'} // comment
	byteAssign       = []byte{'='} // assign
	byteQuote        = []byte{'"'} // quote start sign
	byteSectionStart = []byte{'['} // section start sign
	byteSectionEnd   = []byte{']'} // section end sign
)

var (
	sectionDivision   = "."
	attributeDivision = "."
	lineBreak         = "\n"
)

type IniConfig struct {
}

//Parse parse ini file
func (ini *IniConfig) Parse(fileName string) (Provider, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	return ini.ParseData(data)
}

// ParseData parse ini bytes data
func (ini *IniConfig) ParseData(data []byte) (Provider, error) {
	c := &Container{
		data:             make(map[string]map[string]string),
		sectionComment:   make(map[string]string),
		attributeComment: make(map[string]string),
		RWMutex:          sync.RWMutex{},
		list:             list.New(),
	}
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()

	buf := bufio.NewReader(bytes.NewReader(data))
	// check file bom
	bom, err := buf.Peek(3)
	if err == nil && bom[0] == 239 && bom[1] == 187 && bom[2] == 191 {
		for i := 1; i <= 3; i++ {
			buf.ReadByte()
		}
	}
	// read by lines
	var comment bytes.Buffer
	section := defaultSection
	for {
		line, _, err := buf.ReadLine()
		if err == io.EOF {
			break
		}
		// trim space
		line = bytes.TrimSpace(line)
		// skip empty line
		if bytes.Equal(line, byteEmpty) {
			continue
		}
		// parse comment
		if bytes.HasPrefix(line, byteWellNumber) || bytes.HasPrefix(line, byteSemicolon) {
			line = line[1:]
			if comment.Len() > 0 {
				comment.WriteByte('\n')
			}
			comment.Write(line)
			continue
		}
		// parse section
		if bytes.HasPrefix(line, byteSectionStart) && bytes.HasSuffix(line, byteSectionEnd) {
			section = strings.ToLower(string(line[1 : len(line)-1]))
			if comment.Len() > 0 {
				c.sectionComment[section] = comment.String()
				comment.Reset()
			}
			continue
		}
		// parse attribute
		if split := bytes.Split(line, byteAssign); split != nil {
			// ensure original sort
			listMap := make(map[string]*list.List)
			listMap[section] = list.New()
			c.list.PushBack(listMap)
			// if section is not set, init it
			if _, ok := c.data[section]; !ok {
				c.data[section] = make(map[string]string)
			}

			// support attribute's value appear byteAssign, which must has prefix byteQuote
			if len(split) > 2 && !bytes.HasPrefix(bytes.TrimSpace(split[1]), byteQuote) {
				return nil, fmt.Errorf("read content err:the \"%s\" in %s should appear only once", byteAssign, string(line))
			}

			key := strings.ToLower(string(bytes.TrimSpace(split[0])))
			keyValue := bytes.TrimSpace(split[1])
			keyValue = bytes.Replace(keyValue, byteQuote, byteEmpty, -1)
			// support comment likes below
			// extension=php_exif.dll      ; Must be after mbstring as it depends on it
			valSplit := bytes.SplitN(keyValue, byteSemicolon, 2)
			if len(valSplit) >= 2 {
				comment.WriteByte('\n')
				comment.Write(valSplit[1])
			}
			keyValue = valSplit[0]
			c.data[section][key] = string(keyValue)
			// ensure original sort
			listMap[section].PushBack(key)
			if comment.Len() > 0 {
				c.attributeComment[section+attributeDivision+key] = comment.String()
				comment.Reset()
			}
			continue
		}
	}
	return c, nil
}

type Container struct {
	sync.RWMutex
	data             map[string]map[string]string
	list             *list.List
	sectionComment   map[string]string
	attributeComment map[string]string
}

// Set writes a new value for key.
// if write to one section, the key need be "section::key", otherwise write to default section.
func (c *Container) Set(key, value string) error {
	c.Lock()
	defer c.Unlock()

	if key == "" {
		return errors.New("key is empty")
	}
	section, k := c.parseSectionKey(key)
	if _, ok := c.data[section]; !ok {
		c.data[section] = make(map[string]string)
	}
	c.data[section][k] = value
	return nil
}

// Get retrieves the raw value by a given key
// if get one section key, the key need be "section::key", otherwise write to default section.
func (c *Container) Get(key string) string {
	c.RLock()
	defer c.RUnlock()
	if key == "" {
		return ""
	}
	section, k := c.parseSectionKey(key)
	val, ok := c.data[section][k]
	if !ok {
		return ""
	}
	return val
}

// Has retrieves whether the key exist.
// for section, the key need to be "section::key", otherwise retrieves the default section
func (c *Container) Has(key string) bool {
	c.RLock()
	defer c.RUnlock()
	if key == "" {
		return false
	}
	section, k := c.parseSectionKey(key)
	if _, ok := c.data[section]; !ok {
		return false
	}
	if _, ok := c.data[section][k]; !ok {
		return false
	}
	return true
}

// SaveFile save the config into file.
func (c *Container) SaveFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	parseSectionComment := func(section, key string) string {
		var (
			comment string
			ok      bool
		)
		if key == "" {
			comment, ok = c.sectionComment[section]
		} else {
			comment, ok = c.attributeComment[section+attributeDivision+key]
		}
		if ok {
			if len(comment) == 0 || len(strings.TrimSpace(comment)) == 0 {
				return string(byteWellNumber)
			}
			prefix := string(byteWellNumber)
			return prefix + strings.Replace(comment, lineBreak, lineBreak+prefix, -1)
		}
		return ""
	}

	buf := bytes.NewBuffer(nil)
	// Save default section at first place
	if data, ok := c.data[defaultSection]; ok {
		for key, val := range data {
			if key != "" {
				// save comment
				if v := parseSectionComment(defaultSection, key); v != "" {
					if _, err := buf.WriteString(v + lineBreak); err != nil {
						return err
					}
				}
				// Write key and value.
				if _, err := buf.WriteString(key + string(byteAssign) + val + lineBreak); err != nil {
					return err
				}
			}
		}
		// Put a line between sections.
		if _, err = buf.WriteString(lineBreak); err != nil {
			return err
		}
	}
	// Save named sections
	for {
		element := c.list.Front()
		if element == nil {
			break
		}

		sectionList := element.Value.(map[string]*list.List)
	first:
		for section, keyList := range sectionList {
			if section == defaultSection {
				c.list.Remove(element)
				break first
			}
			// write section comment
			if comment := parseSectionComment(section, ""); comment != "" {
				if _, err := buf.WriteString(comment + lineBreak); err != nil {
					return err
				}
			}
			// write section name
			if _, err := buf.WriteString(string(byteSectionStart) + section + string(byteSectionEnd) + lineBreak); err != nil {
				return err
			}
		second:
			for {
				keyElement := keyList.Front()
				if keyElement == nil {
					break second
				}
				k := keyElement.Value.(string)
				if k != "" {
					val := c.data[section][k]
					// write attribute comment
					if comment := parseSectionComment(section, k); comment != "" {
						if _, err := buf.WriteString(comment + lineBreak); err != nil {
							return err
						}
					}
					// write key and value
					if _, err := buf.WriteString(k + string(byteAssign) + val + lineBreak); err != nil {
						return err
					}
				}
				keyList.Remove(keyElement)
			}
			// Put a line between sections.
			if _, err = buf.WriteString(lineBreak); err != nil {
				return err
			}
			c.list.Remove(element)
		}
	}
	//for section, data := range c.data {
	//	if section != defaultSection {
	//		// write section comment
	//		if comment := parseSectionComment(section, ""); comment != "" {
	//			if _, err := buf.WriteString(comment + lineBreak); err != nil {
	//				return err
	//			}
	//		}
	//		// write section name
	//		if _, err := buf.WriteString(string(byteSectionStart) + section + string(byteSectionEnd) + lineBreak); err != nil {
	//			return err
	//		}
	//
	//		for k, val := range data {
	//			if k != "" {
	//				// write attribute comment
	//				if comment := parseSectionComment(section, k); comment != "" {
	//					if _, err := buf.WriteString(comment + lineBreak); err != nil {
	//						return err
	//					}
	//				}
	//				// write key and value
	//				if _, err := buf.WriteString(k + string(byteAssign) + val + lineBreak); err != nil {
	//					return err
	//				}
	//			}
	//		}
	//	}
	//	// Put a line between sections.
	//	if _, err = buf.WriteString(lineBreak); err != nil {
	//		return err
	//	}
	//}
	_, err = buf.WriteTo(f)
	return err
}

// GetSection retrieves section data
// if section is empty, default section data will back
func (c *Container) GetSection(section string) (map[string]string, error) {
	if section == "" {
		section = defaultSection
	}
	if data, ok := c.data[section]; ok {
		return data, nil
	}
	return nil, fmt.Errorf("section %s not find", section)
}

// String retrieves key's value, which format is string
func (c *Container) String(key string) string {
	return c.Get(key)
}

// Strings retrieves key's slice value, which format is []string
func (c *Container) Strings(key string) []string {
	v := c.String(key)
	if v == "" {
		return nil
	}
	return strings.Split(v, ";")
}

// Int return Int value of given key
func (c *Container) Int(key string) (int, error) {
	return strconv.Atoi(c.Get(key))
}

// Int64 return Int64 value of given key
func (c *Container) Int64(key string) (int64, error) {
	return strconv.ParseInt(c.Get(key), 10, 64)
}

// Bool return bool value of given key
func (c *Container) Bool(key string) (bool, error) {
	return ParseBool(c.Get(key))
}

// Float return Float value of given key
func (c *Container) Float(key string) (float64, error) {
	return strconv.ParseFloat(c.Get(key), 64)
}

// DefaultString returns the string value for a given key.
// if err != nil return defaultVal
func (c *Container) DefaultString(key, defaultVal string) string {
	value := c.Get(key)
	if value == "" {
		value = defaultVal
	}
	return value
}

// DefaultStrings returns the []string value for a given key.
// if err != nil return defaultVal
func (c *Container) DefaultStrings(key string, defaultVal []string) []string {
	value := c.Strings(key)
	if value == nil {
		return defaultVal
	}
	return value
}

// DefaultInt returns the integer value for a given key.
// if err != nil return defaultVal
func (c *Container) DefaultInt(key string, defaultVal int) int {
	value, err := c.Int(key)
	if err != nil {
		return defaultVal
	}
	return value
}

// DefaultInt64 returns the int64 value for a given key.
// if err != nil return defaultVal
func (c *Container) DefaultInt64(key string, defaultVal int64) int64 {
	value, err := c.Int64(key)
	if err != nil {
		return defaultVal
	}
	return value
}

// DefaultBool returns the boolean value for a given key.
// if err != nil return defaultVal
func (c *Container) DefaultBool(key string, defaultVal bool) bool {
	value, err := c.Bool(key)
	if err != nil {
		return defaultVal
	}
	return value
}

// DefaultFloat returns the float64 value for a given key.
// if err != nil return defaultVal
func (c *Container) DefaultFloat(key string, defaultVal float64) float64 {
	value, err := c.Float(key)
	if err != nil {
		return defaultVal
	}
	return value
}

// parseSectionKey retrieves the key
// for section key, the key need to be "section::key", otherwise retrieves the default section
func (c *Container) parseSectionKey(key string) (section, k string) {
	if key == "" {
		return
	}
	keys := strings.Split(strings.ToLower(key), sectionDivision)
	if len(keys) >= 2 {
		section = keys[0]
		k = strings.Join(keys[1:], attributeDivision)
	} else {
		section = defaultSection
		k = keys[0]
	}
	return
}

func init() {
	Register("ini", &IniConfig{})
}
