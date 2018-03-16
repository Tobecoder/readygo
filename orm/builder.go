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
	"strings"
	"fmt"
)

var (
	selectSql    = `SELECT%DISTINCT% %FIELD% FROM %TABLE%%FORCE%%JOIN%%WHERE%%GROUP%%HAVING%%ORDER%%LIMIT% %UNION%%LOCK%%COMMENT%`
	insertSql    = `%INSERT% INTO %TABLE% (%FIELD%) VALUES (%DATA%) %COMMENT%`
	insertAllSql = `%INSERT% INTO %TABLE% (%FIELD%) %DATA% %COMMENT%`
	updateSql    = `UPDATE %TABLE% SET %SET% %JOIN% %WHERE% %ORDER%%LIMIT% %LOCK%%COMMENT%`
	deleteSql    = `DELETE FROM %TABLE% %USING% %JOIN% %WHERE% %ORDER%%LIMIT% %LOCK%%COMMENT%`
)

type BaseBuilder struct {
	query QueryParser
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
		"",
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
			tables = append(tables, t + " " + alias)
		}else{
			tables = append(tables, t)
		}
	}
	return strings.Join(tables, ",")
}

// parseDistinct assemble distinct to query sql
func (b *BaseBuilder) parseDistinct (distinct bool) string {
	var str string
	if distinct {
		str = " DISTINCT "
	}
	return str
}

// parseField parse sql query fields
func (b *BaseBuilder) parseField (option *Option) string {
	var field = make([]string, 0)
	if option.field == nil {
		field = append(field, "*")
	} else {
		for _, f := range option.field {
			if alias, ok := option.fieldAlias[f]; ok {
				field = append(field, f + " AS " + alias)
			}else{
				field = append(field, f)
			}
		}
	}
	return strings.Join(field, ",")
}

// parseJoin parse table join clause
func (b *BaseBuilder) parseJoin (option *Option) string{
	var joinStr string
	if option.join == nil {
		return joinStr
	}
	for _, v := range option.join{
		table := b.parseTable([]string{v["table"]}, option)
		fmt.Println(table, v["table"])
		joinStr += " " + v["type"] + " JOIN " + table
		if c, ok := v["condition"]; ok && len(c) > 0 {
			joinStr += " ON " + c
		}
	}
	return joinStr
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
func (b *BaseBuilder) parseOrder (order []string) string {
	var (
		orderStr string
		orders []string
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
func (b *BaseBuilder) parseLimit(limit string) string{
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
		switch u := v.(type){
		case string:
			unions = append(unions, string(unionType) + u)
		case func (QueryParser):
			unions = append(unions, string(unionType) + b.parseClosure(u, false))
		case []string:
			for _, entry := range u {
				unions = append(unions, string(unionType) + entry)
			}
		}
	}
	unionStr = strings.Join(unions, " ")
	return unionStr
}

// parseClosure parse closure call, which return assemble sql
func (b *BaseBuilder) parseClosure (call QueryClosure, sub bool) string{
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
func (b *BaseBuilder) parseForce (force string) string {
	var forceStr string
	if len(force) > 0 {
		forceStr = fmt.Sprintf(" FORCE INDEX ( %s ) ", force)
	}
	return forceStr
}