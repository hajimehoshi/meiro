package field_test

import (
	"github.com/hajimehoshi/meiro/field"
	"math/rand"
	"testing"
)

type NullWriter struct{}

func (w *NullWriter) Write(bytes []byte) (int, error) {
	return len(bytes), nil
}

func BenchmarkCreate(b *testing.B) {
	random := rand.New(rand.NewSource(0))
	f := field.Create(random, 100, 100, 10, 10)
	f.WriteSVG(&NullWriter{})
}
