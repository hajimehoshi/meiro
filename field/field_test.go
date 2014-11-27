package field_test

import (
	"github.com/hajimehoshi/meiro/field"
	"math/rand"
	"testing"
	"time"
)

func BenchmarkCreate(b *testing.B) {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	field.Create(random, 100, 100, 10, 10)
}
