package future_test

import (
	"testing"

	"github.com/dancantos/future"
)

type bigdata struct {
	s1, s2, s3, s4, s5, s6, s7, s8, s9, s0 string
	i1, i2, i3, i4, i5, i6, i7, i8, i9, i0 int
	f1, f2, f3, f4, f5, f6, f7, f8, f9, f0 float64
	n1, n2, n3, n4, n5, n6, n7, n8, n9, n0 struct {
		a string
		b int
	}
}

func BenchmarkFuture(b *testing.B) {
	for i := 0; i < b.N; i++ {
		futureVal := future.Go(func() bigdata {
			return bigdata{}
		})
		for j := 0; j < 1; j++ {
			futureVal.Get()
		}
	}
}
