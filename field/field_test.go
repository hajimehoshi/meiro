package field_test

import (
	"github.com/hajimehoshi/meiro/field"
	"math/rand"
	"testing"
)

func BenchmarkCreate(b *testing.B) {
	random := rand.New(rand.NewSource(0))
	field.Create(random, 1000, 1000)
}
