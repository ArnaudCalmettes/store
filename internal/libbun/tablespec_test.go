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
