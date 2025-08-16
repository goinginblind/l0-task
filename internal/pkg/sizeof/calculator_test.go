package sizeof

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

// total shallow size is 80b
type fatherTest struct {
	Name   string // 16b
	Number int    // 8b
	Nested struct {
		AliasI  string //16b
		AliasII string // 16b
	}
	NestedBig []struct { // 24b
		NestedName string
		Tahiti     bool
	}
}

func TestSizeOf(t *testing.T) {
	s := fatherTest{
		// with deep size comments: total 80+153
		Name:   "John Marston", //12b
		Number: 1,
		Nested: struct {
			AliasI  string
			AliasII string
		}{
			"Rip van Winkle", //14
			"Jim Milton",     // 10
		},
		NestedBig: []struct { //72 its 3x24
			NestedName string
			Tahiti     bool
		}{
			{
				NestedName: "Arthur Morgan", //13
				Tahiti:     false,
			},
			{
				NestedName: "Dutch van der Linde", //19
				Tahiti:     true,
			},
			{
				NestedName: "Hosea Mathews", //13
				Tahiti:     false,
			},
		},
	}

	shallowSize := unsafe.Sizeof(s)
	deepSize := SizeOf(s)
	assert.Equal(t, deepSize, 233)
	fmt.Printf("Shallow size of struct: %d bytes\n", shallowSize)
	fmt.Printf("Deep size of struct: %d bytes\n", deepSize)
}
