package field_test

import (
	"github.com/hajimehoshi/meiro/field"
	"math/rand"
	"testing"
)

func BenchmarkCreate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		random := rand.New(rand.NewSource(0))
		field.Create(random, 100, 100, 10, 10)
	}
}
