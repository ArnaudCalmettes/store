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
