package libbun

import (
	"testing"

	. "github.com/ArnaudCalmettes/store/test/helpers"
	"github.com/uptrace/bun"
)

func TestGetTableSpec(t *testing.T) {
	t.Run("nominal", func(t *testing.T) {
		type Embedded struct {
			PhoneNumber string
			Addr        string `bun:"email_address"`
		}
		type Model struct {
			bun.BaseModel `bun:"table:users"`

			ID        string `bun:",pk"`
			FirstName string
			Ignored   int `bun:"-"`
			Embedded
		}
		spec, err := GetTableSpec[Model]()
		Expect(t,
			NoError(err),
			Equal(
				&TableSpec{
					TableName: "users",
					KeyField:  "ID",
					KeySQL:    "id",
					ColumnNames: map[string]string{
						"ID":          "id",
						"FirstName":   "first_name",
						"Addr":        "email_address",
						"PhoneNumber": "phone_number",
					},
				},
				spec,
			),
		)
	})
	t.Run("empty model", func(t *testing.T) {
		type Model struct {
			bun.BaseModel
		}
		spec, err := GetTableSpec[Model]()
		Expect(t,
			NoError(err),
			Equal(
				&TableSpec{
					ColumnNames: map[string]string{},
				},
				spec,
			),
		)
	})
	t.Run("not a struct", func(t *testing.T) {
		spec, err := GetTableSpec[int]()
		Expect(t,
			IsNilPointer(spec),
			IsError(errNotAStruct, err),
		)
	})
}

func TestTableSpecValidate(t *testing.T) {
	testCases := []struct {
		Name   string
		Input  TableSpec
		Expect []error
	}{
		{
			Name:   "empty",
			Input:  TableSpec{},
			Expect: []error{errMissingTableName, errNoPK, errNoColumns},
		},
		{
			Name: "no table name",
			Input: TableSpec{
				KeyField:    "ID",
				KeySQL:      "id",
				ColumnNames: map[string]string{"ID": "id"},
			},
			Expect: []error{errMissingTableName},
		},
		{
			Name: "no pk",
			Input: TableSpec{
				TableName:   "users",
				ColumnNames: map[string]string{"ID": "id"},
			},
			Expect: []error{errNoPK},
		},
		{
			Name: "no columns",
			Input: TableSpec{
				TableName: "users",
				KeyField:  "ID",
				KeySQL:    "id",
			},
			Expect: []error{errNoColumns},
		},
	}
	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.Input.Validate()
			if len(tc.Expect) == 0 {
				Expect(t, NoError(err))
			}
			for _, wantErr := range tc.Expect {
				Expect(t, IsError(wantErr, err))
			}
		})
	}
}
