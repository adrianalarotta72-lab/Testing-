package store

import (
	"fmt"
	"testing"
)

func BenchmarkSet(b *testing.B) {
	s := New()
	key := "benchkey"
	value := "benchvalue"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Set(key, value)
	}
}

func BenchmarkGet(b *testing.B) {
	s := New()
	key := "benchkey"
	value := "benchvalue"
	_ = s.Set(key, value)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.Get(key)
	}
}

func BenchmarkDelete(b *testing.B) {
	s := New()
	key := "benchkey"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		_ = s.Set(key, "value")
		b.StartTimer()

		_ = s.Delete(key)
	}
}

func BenchmarkExists(b *testing.B) {
	s := New()
	key := "benchkey"
	_ = s.Set(key, "value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Exists(key)
	}
}

func BenchmarkSetParallel(b *testing.B) {
	s := New()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key%d", i)
			value := fmt.Sprintf("value%d", i)
			_ = s.Set(key, value)
			i++
		}
	})
}

func BenchmarkGetParallel(b *testing.B) {
	s := New()
	// Prepopulate with some data
	for i := 0; i < 1000; i++ {
		_ = s.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key%d", i%1000)
			_, _ = s.Get(key)
			i++
		}
	})
}

func BenchmarkMixedOperations(b *testing.B) {
	s := New()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key%d", i%100)
			value := fmt.Sprintf("value%d", i)

			switch i % 3 {
			case 0:
				_ = s.Set(key, value)
			case 1:
				_, _ = s.Get(key)
			case 2:
				_ = s.Delete(key)
			}
			i++
		}
	})
}

func BenchmarkStoreWithDifferentSizes(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			s := New()

			// Prepopulate
			for i := 0; i < size; i++ {
				_ = s.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				key := fmt.Sprintf("key%d", i%size)
				_, _ = s.Get(key)
			}
		})
	}
}
