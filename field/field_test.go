package field_test

import (
	"github.com/hajimehoshi/meiro/field"
	"testing"
)

func BenchmarkCreate(b *testing.B) {
	field.Create(300, 300)
}
