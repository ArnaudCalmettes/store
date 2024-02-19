// Copyright (c) 2024 nohar
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package libbun

import (
	"testing"

	. "github.com/ArnaudCalmettes/store"
	. "github.com/ArnaudCalmettes/store/test/helpers"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
)

func TestFilterBuilder(t *testing.T) {
	db := bun.NewDB(nil, sqlitedialect.New())

	type Person struct {
		bun.BaseModel `bun:"table:persons,alias:p"`

		ID       string `bun:",pk"`
		Name     string
		Age      int
		Referent *string
	}
	var model []*Person

	tableSpec, err := GetTableSpec[Person]()
	Require(t,
		NoError(err),
	)

	t.Run("nominal", func(t *testing.T) {
		builder, err := BuilderForFilter(
			Where("ID", "!=", ""),
			tableSpec,
		)
		Expect(t,
			NoError(err),
		)
		query := db.NewSelect().Model(&model).ApplyQueryBuilder(builder).String()
		Expect(t,
			Equal(
				`SELECT `+
					`"p"."id", "p"."name", "p"."age", "p"."referent" `+
					`FROM "persons" AS "p" WHERE ("id" != '')`,
				query,
			),
		)
	})
	t.Run("unknown field", func(t *testing.T) {
		_, err := BuilderForFilter(
			Where("BankAccount", ">", 100),
			tableSpec,
		)
		Expect(t,
			IsError(errNoSuchField, err),
		)
	})
	t.Run("coumpound", func(t *testing.T) {
		builder, err := BuilderForFilter(
			Any(
				All(Where("ID", "!=", ""), Where("Age", ">", 18)),
				Where("Name", "!=", "foo"),
			),
			tableSpec,
		)
		Expect(t,
			NoError(err),
		)
		query := db.NewSelect().Model(&model).ApplyQueryBuilder(builder).String()
		Expect(t,
			Equal(
				`SELECT `+
					`"p"."id", "p"."name", "p"."age", "p"."referent" `+
					`FROM "persons" AS "p" WHERE (("id" != '') AND ("age" > 18)) `+
					`OR (("name" != 'foo'))`,
				query,
			),
		)
	})
}
