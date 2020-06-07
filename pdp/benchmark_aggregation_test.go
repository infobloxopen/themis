// Performance measurement of client content tree
package pdp

import (
	"testing"

	"github.com/infobloxopen/go-trees/strtree"
)

func BenchmarkGetContentData(b *testing.B) {
	sm1 := strtree.NewTree()
	sm1.InplaceInsert("1-first", []string{"first-first"})
	sm1.InplaceInsert("2-second", []string{"second-first-1", "second-first-2"})
	sm1.InplaceInsert("3-third", []string{"third-first"})

	sm2 := strtree.NewTree()
	sm2.InplaceInsert("1-first", []string{"first-second-1"})
	sm2.InplaceInsert("2-second", []string{"second-second", "duplicate", "duplicate"})
	sm2.InplaceInsert("3-third", []string{"third-second"})

	sm3 := strtree.NewTree()
	sm3.InplaceInsert("1-first", []string{"first-third"})
	sm3.InplaceInsert("2-second", []string{"second-third", "second-third-2", "duplicate"})
	sm3.InplaceInsert("3-third", []string{"third-third-1"})

	ssm := strtree.NewTree()
	ssm.InplaceInsert("1-first", MakeContentStringMap(sm1))
	ssm.InplaceInsert("2-second", MakeContentStringMap(sm2))
	ssm.InplaceInsert("3-third", MakeContentStringMap(sm3))

	ci := MakeContentMappingItem(
		"str-str-los",
		TypeListOfStrings,
		MakeSignature(TypeString, TypeString),
		MakeContentStringMap(ssm),
	)

	ctx, _ := NewContext(nil, 0, nil)
	path := []Expression{MakeStringValue("1-first"), MakeStringValue("2-second")}

	b.Run("Get (original)-NoError", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			ci.Get(path, ctx)
		}
	})
	b.Run("GetAggregated-NoError", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			ci.GetAggregated(path, ctx, AggTypeDisable)
		}
	})

	pathA := []AttributeValue{MakeStringValue("1-first"), MakeStringValue("2-second")}
	b.Run("GetByValues-NoError", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			ci.GetByValues(pathA, AggTypeDisable)
		}
	})

	path = []Expression{MakeStringValue("unknown"), MakeStringValue("2-second")}
	b.Run("Get (original)-Missing1", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			ci.Get(path, ctx)
		}
	})
	b.Run("GetAggregated-Missing1", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			ci.GetAggregated(path, ctx, AggTypeDisable)
		}
	})

	path = []Expression{MakeStringValue("1-first"), MakeStringValue("unknown")}
	b.Run("Get (original)-Missing2", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			ci.Get(path, ctx)
		}
	})
	b.Run("GetAggregated-Missing2", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			ci.GetAggregated(path, ctx, AggTypeDisable)
		}
	})

	path = []Expression{MakeListOfStringsValue([]string{"1-first"}), MakeStringValue("2-second")}
	b.Run("GetAggregated-RF-V1-S", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			ci.GetAggregated(path, ctx, AggTypeReturnFirst)
		}
	})
	b.Run("GetAggregated-A-V1-S", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			ci.GetAggregated(path, ctx, AggTypeAppend)
		}
	})
	b.Run("GetAggregated-AU-V1-S", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			ci.GetAggregated(path, ctx, AggTypeAppendUnique)
		}
	})

	path = []Expression{MakeListOfStringsValue([]string{"1-first", "2-second"}), MakeStringValue("2-second")}
	b.Run("GetAggregated-RF-V2-S", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			ci.GetAggregated(path, ctx, AggTypeReturnFirst)
		}
	})
	b.Run("GetAggregated-A-V2-S", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			ci.GetAggregated(path, ctx, AggTypeAppend)
		}
	})
	b.Run("GetAggregated-AU-V2-S", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			ci.GetAggregated(path, ctx, AggTypeAppendUnique)
		}
	})

	path = []Expression{MakeListOfStringsValue([]string{"1-first", "2-second", "3-third"}), MakeStringValue("2-second")}
	b.Run("GetAggregated-RF-V3-S", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			ci.GetAggregated(path, ctx, AggTypeReturnFirst)
		}
	})
	b.Run("GetAggregated-A-V3-S", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			ci.GetAggregated(path, ctx, AggTypeAppend)
		}
	})
	b.Run("GetAggregated-AU-V3-S", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			ci.GetAggregated(path, ctx, AggTypeAppendUnique)
		}
	})

	path = []Expression{MakeListOfStringsValue([]string{"1-first", "2-second"}), MakeListOfStringsValue([]string{"2-second", "3-third"})}
	b.Run("GetAggregated-RF-V2-V2", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			ci.GetAggregated(path, ctx, AggTypeReturnFirst)
		}
	})
	b.Run("GetAggregated-A-V2-V2", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			ci.GetAggregated(path, ctx, AggTypeAppend)
		}
	})
	b.Run("GetAggregated-AU-V2-V2", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			ci.GetAggregated(path, ctx, AggTypeAppendUnique)
		}
	})
}
