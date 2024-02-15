package inspect

import (
	"testing"
	"time"

	. "github.com/ArnaudCalmettes/store"
	. "github.com/ArnaudCalmettes/store/test/helpers"
)

type CmpTest struct {
	String  string
	Int     int
	Int8    int8
	Int16   int16
	Int32   int32
	Int64   int64
	UInt    uint
	UInt8   uint8
	UInt16  uint16
	UInt32  uint32
	UInt64  uint64
	Float32 float32
	Float64 float64
	Time    time.Time
	Ptr     *int
}

func TestNewCmp(t *testing.T) {
	t.Run("no such field", func(t *testing.T) {
		_, err := NewCmp[CmpTest](By("Field"))
		Expect(t,
			IsError(errNoSuchField, err),
		)
	})
	t.Run("nominal", func(t *testing.T) {
		cmp, err := NewCmp[CmpTest](By("String"))
		Expect(t,
			NoError(err),
			Equal(1, cmp(
				&CmpTest{String: "ZZZ"},
				&CmpTest{String: "AAA"},
			)),
			Equal(0, cmp(&CmpTest{}, &CmpTest{})),
			Equal(-1, cmp(
				&CmpTest{String: "AAA"},
				&CmpTest{String: "ZZZ"},
			)),
		)
	})
	t.Run("descending", func(t *testing.T) {
		cmp, err := NewCmp[CmpTest](By("Int").Desc())
		Expect(t,
			NoError(err),
			Equal(-1, cmp(&CmpTest{Int: 1337}, &CmpTest{Int: 42})),
			Equal(0, cmp(&CmpTest{Int: 1337}, &CmpTest{Int: 1337})),
			Equal(1, cmp(&CmpTest{Int: 42}, &CmpTest{Int: 1337})),
		)
	})
	t.Run("time", func(t *testing.T) {
		now := time.Now()
		yesterday := now.Add(-24 * time.Hour)
		tomorrow := now.Add(24 * time.Hour)

		cmp, err := NewCmp[CmpTest](By("Time"))
		Expect(t,
			NoError(err),
			Equal(1, cmp(&CmpTest{Time: now}, &CmpTest{Time: yesterday})),
			Equal(0, cmp(&CmpTest{Time: now}, &CmpTest{Time: now})),
			Equal(-1, cmp(&CmpTest{Time: now}, &CmpTest{Time: tomorrow})),
		)

		cmp, err = NewCmp[CmpTest](By("Time").Desc())
		Expect(t,
			NoError(err),
			Equal(-1, cmp(&CmpTest{Time: now}, &CmpTest{Time: yesterday})),
			Equal(0, cmp(&CmpTest{Time: now}, &CmpTest{Time: now})),
			Equal(1, cmp(&CmpTest{Time: now}, &CmpTest{Time: tomorrow})),
		)
	})
}

func TestNewCmpSupportedTypes(t *testing.T) {
	supported := []string{
		"String",
		"Int", "Int8", "Int16", "Int32", "Int64",
		"UInt", "UInt8", "UInt16", "UInt32", "UInt64",
		"Float32", "Float64", "Time",
	}
	t.Parallel()
	for _, field := range supported {
		t.Run(field, func(t *testing.T) {
			_, err := NewCmp[CmpTest](By(field))
			Expect(t,
				NoError(err),
			)
		})
	}
	t.Run("Ptr", func(t *testing.T) {
		_, err := NewCmp[CmpTest](By("Ptr"))
		Expect(t,
			IsError(errTypeNotSupported, err),
		)
	})
}
