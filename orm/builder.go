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
)

var (
	selectSql    = `SELECT%DISTINCT% %FIELD% FROM %TABLE%%FORCE%%JOIN%%WHERE%%GROUP%%HAVING%%ORDER%%LIMIT% %UNION%%LOCK%%COMMENT%`
	insertSql    = `%INSERT% INTO %TABLE% (%FIELD%) VALUES (%DATA%) %COMMENT%`
	insertAllSql = `%INSERT% INTO %TABLE% (%FIELD%) %DATA% %COMMENT%`
	updateSql    = `UPDATE %TABLE% SET %SET% %JOIN% %WHERE% %ORDER%%LIMIT% %LOCK%%COMMENT%`
	deleteSql    = `DELETE FROM %TABLE% %USING% %JOIN% %WHERE% %ORDER%%LIMIT% %LOCK%%COMMENT%`

	expMap = map[string]string{
		"eq":               "=",
		"neq":              "<>",
		"gt":               ">",
		"egt":              ">=",
		"lt":               "<",
		"elt":              "<=",
		"notlike":          "NOT LIKE",
		"like":             "LIKE",
		"in":               "IN",
		"exp":              "EXP",
		"notin":            "NOT IN",
		"not in":           "NOT IN",
		"between":          "BETWEEN",
		"not between":      "NOT BETWEEN",
		"notbetween":       "NOT BETWEEN",
		"exists":           "EXISTS",
		"notexists":        "NOT EXISTS",
		"not exists":       "NOT EXISTS",
		"null":             "NULL",
		"notnull":          "NOT NULL",
		"not null":         "NOT NULL",
		"> time":           "> TIME",
		"< time":           "< TIME",
		">= time":          ">= TIME",
		"<= time":          "<= TIME",
		"between time":     "BETWEEN TIME",
		"not between time": "NOT BETWEEN TIME",
		"notbetween time":  "NOT BETWEEN TIME",
	}
)

type BaseBuilder struct {
	query QueryParser
	ins Builder
}

var _ Builder = new(BaseBuilder)

// selects build select sql
func (b *BaseBuilder) selects(option Option) string {
	var replace = []string{
		"%TABLE%",
		b.parseTable(option.table, &option),
		"%DISTINCT%",
		b.parseDistinct(option.distinct),
		"%FIELD%",
		b.parseField(&option),
		"%JOIN%",
		b.parseJoin(&option),
		"%WHERE%",
		b.parseWhere(option.where, &option),
		"%GROUP%",
		b.parseGroup(option.group),
		"%HAVING%",
		b.parseHaving(option.having),
		"%ORDER%",
		b.parseOrder(option.order),
		"%LIMIT%",
		b.parseLimit(option.limit),
		"%UNION%",
		b.parseUnion(option.union, option.unionType),
		"%LOCK%",
		b.parseLock(option.lock),
		"%COMMENT%",
		b.parseComment(option.comment),
		"%FORCE%",
		b.parseForce(option.force),
	}
	r := strings.NewReplacer(replace...)
	return r.Replace(selectSql)
}

// parseTable parse sql query tables
func (b *BaseBuilder) parseTable(table []string, option *Option) string {
	tables := make([]string, 0)
	for _, t := range table {
		if alias, ok := option.tableAlias[t]; ok {
			tables = append(tables, t+" "+alias)
		} else {
			tables = append(tables, t)
		}
	}
	return strings.Join(tables, ",")
}

// parseDistinct assemble distinct to query sql
func (b *BaseBuilder) parseDistinct(distinct bool) string {
	var str string
	if distinct {
		str = " DISTINCT "
	}
	return str
}

// parseField parse sql query fields
func (b *BaseBuilder) parseField(option *Option) string {
	var field = make([]string, 0)
	if option.field == nil {
		field = append(field, "*")
	} else {
		for _, f := range option.field {
			if alias, ok := option.fieldAlias[f]; ok {
				field = append(field, f+" AS "+alias)
			} else {
				field = append(field, f)
			}
		}
	}
	return strings.Join(field, ",")
}

// parseJoin parse table join clause
func (b *BaseBuilder) parseJoin(option *Option) string {
	var joinStr string
	if option.join == nil {
		return joinStr
	}
	for _, v := range option.join {
		table := b.parseTable([]string{v["table"]}, option)
		joinStr += " " + v["type"] + " JOIN " + table
		if c, ok := v["condition"]; ok && len(c) > 0 {
			joinStr += " ON " + c
		}
	}
	return joinStr
}

// parseWhere parse where sql clause
func (b *BaseBuilder) parseWhere(where whereType, option *Option) string {
	var whereStr string
	if where == nil {
		return whereStr
	}
	for logic, whereOptions := range where {
		var str = make([]string, 0)
		for field, value := range whereOptions {
			if false {
			} else {
				str = append(str, " "+logic+" "+b.parseWhereItem(field, value, logic, option))
			}
		}
		if len(str) > 0 {
			s := strings.Join(str, " ")
			if len(whereStr) == 0 {
				s = s[len(logic)+1:]
			}
			whereStr += s
		}
	}
	if len(whereStr) > 0 {
		whereStr = " WHERE " + whereStr
	}
	return whereStr
}

// parseWhereItem parse where item
func (b *BaseBuilder) parseWhereItem(field string, value []interface{}, logic string, option *Option) string {
	var (
		whereStr string
		exp      string
		val      interface{}
	)

	if len(value) == 1 {
		exp = "="
		val = value[0]
	} else {
		switch v := value[0].(type) {
		case string:
			exp = v
		case []interface{}:
			item := value[len(value)-1]
			if s, ok := item.(string); ok {
				var andOr = map[string]int{"AND":1, "OR":1}
				s = strings.ToUpper(s)
				if _, ok := andOr[s]; ok {
					logic = s
				}
			}
			var str []string
			val = value[0:len(value)-1]
			if v, ok := val.([]interface{}); ok {
				for _, vItem := range v {
					str = append(str, b.parseWhereItem(field, vItem.([]interface{}), logic, option))
				}
			}
			return "( " + strings.Join(str, " "+logic+" ") + " )"

		}
		val = value[1]
	}

	// check express operator
	if !checkOperator(exp) {
		return whereStr
	}
	var (
		isNull         = map[string]int{"NOT NULL": 1, "NULL": 1}
		compareAndLike = map[string]int{"=":1, "<>":1, ">":1, ">=":1, "<":1, "<=":1, "LIKE":1, "NOT LIKE":1}
	)
	exp = strings.ToUpper(exp)
	if _, ok := compareAndLike[exp]; ok {
		whereStr += field + " " + exp + " " + b.parseStringValue(val, field)
	} else if exp == "EXP" {
		s, ok := val.(string)
		if ok {
			whereStr += "( " + s + " )"
		}
	} else if _, ok := isNull[exp]; ok {
		whereStr += field + " IS " + exp
	}
	return whereStr
}

// parseStringValue parse value interface{} to string
func (b *BaseBuilder) parseStringValue(value interface{}, field string) string {
	var str string
	switch v := value.(type) {
	case string:
		str = string(b.escapeStringQuotes([]byte{}, v))
	case int:
		str = strconv.Itoa(v)
	}
	return str
}

// parseGroup assemble field group clause
func (b *BaseBuilder) parseGroup(group string) string {
	var groupStr string
	if len(group) > 0 {
		groupStr = " GROUP BY " + group
	}
	return groupStr
}

// parseHaving assemble having sql clause
func (b *BaseBuilder) parseHaving(having string) string {
	var havingStr string
	if len(havingStr) > 0 {
		havingStr = " HAVING " + havingStr
	}
	return havingStr
}

// parseOrder assemble order sql clause
func (b *BaseBuilder) parseOrder(order []string) string {
	var (
		orderStr string
		orders   []string
	)
	if order != nil {
		for _, v := range order {
			if !strings.Contains(v, "(") {
				orders = append(orders, v)
			}
		}
		orderStr += " ORDER BY " + strings.Join(orders, ",")
	}
	return orderStr
}

// parseLimit parse sql query limit
func (b *BaseBuilder) parseLimit(limit string) string {
	if len(limit) == 0 || strings.Contains(limit, "(") {
		return ""
	}
	return " LIMIT " + limit + " "
}

// parseUnion parse union sql clause
func (b *BaseBuilder) parseUnion(union []interface{}, unionType unionType) string {
	var unionStr string
	if union == nil {
		return unionStr
	}
	unions := make([]string, len(union))
	for _, v := range union {
		switch u := v.(type) {
		case string:
			unions = append(unions, string(unionType)+u)
		case func(QueryParser):
			unions = append(unions, string(unionType)+b.parseClosure(u, false))
		case []string:
			for _, entry := range u {
				unions = append(unions, string(unionType)+entry)
			}
		}
	}
	unionStr = strings.Join(unions, " ")
	return unionStr
}

// parseClosure parse closure call, which return assemble sql
func (b *BaseBuilder) parseClosure(call QueryClosure, sub bool) string {
	query := newMysqlQuery(b.query.Connection())
	call(query)
	return query.BuildSql(sub)
}

// parseLock assemble for update sql clause
func (b *BaseBuilder) parseLock(lock bool) string {
	var lockStr string
	if lock {
		lockStr = " FOR UPDATE "
	}
	return lockStr
}

// parseComment assemble comment sql clause
func (b *BaseBuilder) parseComment(comment string) string {
	var commentStr string
	if len(comment) > 0 {
		commentStr = " /* " + comment + " */"
	}
	return commentStr
}

// parseForce assemble force index sql clause
func (b *BaseBuilder) parseForce(force string) string {
	var forceStr string
	if len(force) > 0 {
		forceStr = fmt.Sprintf(" FORCE INDEX ( %s ) ", force)
	}
	return forceStr
}

// checkOperator check whether operator in map expMap
func checkOperator(exp string) bool{
	var exist bool
	for _, v := range expMap{
		if v == exp{
			exist = true
			break
		}
	}
	if exist == false {
		exp := strings.ToLower(exp)
		if _, ok := expMap[exp]; ok {
			exist = true
		}
	}
	return exist
}

// escapeStringQuotes is similar to escapeBytesQuotes but for string.
func (b *BaseBuilder)escapeStringQuotes(buf []byte, v string) []byte {
	return b.ins.escapeStringQuotes(buf, v)
}