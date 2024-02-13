package libbun

import (
	"errors"
	"reflect"
	"strings"

	"github.com/uptrace/bun"
)

type TableSpec struct {
	TableName   string
	KeyField    string
	KeySQL      string
	ColumnNames map[string]string
}

var (
	errMissingTableName = errors.New("missing table name")
	errNoPK             = errors.New("primary key not configured")
	errNoColumns        = errors.New("struct doesn't have any exported fields")
)

func (t *TableSpec) Validate() error {
	errs := make([]error, 0, 5)
	if t.TableName == "" {
		errs = append(errs, errMissingTableName)
	}
	if t.KeyField == "" || t.KeySQL == "" {
		errs = append(errs, errNoPK)
	}
	if len(t.ColumnNames) == 0 {
		errs = append(errs, errNoColumns)
	}
	return errors.Join(errs...)
}

var (
	baseModelType = reflect.TypeOf(bun.BaseModel{})
	errNotAStruct = errors.New("not a struct")
)

func GetTableSpec[T any]() (*TableSpec, error) {
	var zeroStruct T
	typ := reflect.TypeOf(zeroStruct)
	if typ.Kind() != reflect.Struct {
		return nil, errNotAStruct
	}
	spec := &TableSpec{
		ColumnNames: map[string]string{},
	}
	for _, field := range reflect.VisibleFields(typ) {
		tag := field.Tag.Get("bun")
		if field.Anonymous {
			if field.Type == baseModelType {
				spec.TableName = parseTableName(tag)
			}
			continue
		}
		if tag == "-" {
			continue
		}
		name := field.Name
		column := getColumnNameFromTag(tag)
		if column == "" {
			column = toColumnName(name)
		}
		spec.ColumnNames[name] = column
		if isPK(tag) {
			spec.KeyField = name
			spec.KeySQL = column
		}
	}
	return spec, nil
}

func parseTableName(tag string) string {
	options := strings.Split(tag, ",")
	for _, option := range options {
		if name, ok := strings.CutPrefix(option, "table:"); ok {
			return name
		}
	}
	return ""
}

func getColumnNameFromTag(tag string) string {
	if tag == "" {
		return ""
	}
	name, _, _ := strings.Cut(tag, ",")
	return name
}

func isPK(tag string) bool {
	return strings.Contains(tag, ",pk")
}
