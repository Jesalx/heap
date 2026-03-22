package heap_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/jesalx/heap"
)

func randInts(n int) []int {
	r := rand.New(rand.NewPCG(0, 0))
	s := make([]int, n)
	for i := range s {
		s[i] = r.IntN(n)
	}
	return s
}

func BenchmarkPush(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		items := randInts(n)
		b.Run(sizeName(n), func(b *testing.B) {
			for b.Loop() {
				h := heap.NewMin[int]()
				for _, v := range items {
					h.Push(v)
				}
			}
		})
	}
}

func BenchmarkPop(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		items := randInts(n)
		less := func(a, b int) bool { return a < b }
		b.Run(sizeName(n), func(b *testing.B) {
			for b.Loop() {
				h := heap.NewFrom(items, less)
				for h.Len() > 0 {
					h.Pop()
				}
			}
		})
	}
}

func BenchmarkPeek(b *testing.B) {
	h := heap.NewFrom(randInts(1_000), func(a, b int) bool { return a < b })
	b.ResetTimer()
	for b.Loop() {
		h.Peek()
	}
}

func BenchmarkNewFrom(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		items := randInts(n)
		less := func(a, b int) bool { return a < b }
		b.Run(sizeName(n), func(b *testing.B) {
			for b.Loop() {
				heap.NewFrom(items, less)
			}
		})
	}
}

func BenchmarkNewFromVsPush(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		items := randInts(n)
		less := func(a, b int) bool { return a < b }

		b.Run(sizeName(n)+"/NewFrom", func(b *testing.B) {
			for b.Loop() {
				heap.NewFrom(items, less)
			}
		})

		b.Run(sizeName(n)+"/Push", func(b *testing.B) {
			for b.Loop() {
				h := heap.New(less)
				for _, v := range items {
					h.Push(v)
				}
			}
		})
	}
}

func BenchmarkPushPop(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		items := randInts(n)
		less := func(a, b int) bool { return a < b }
		b.Run(sizeName(n), func(b *testing.B) {
			for b.Loop() {
				h := heap.NewFrom(items, less)
				for _, v := range items {
					h.PushPop(v)
				}
			}
		})
	}
}

func BenchmarkPushPopVsPushAndPop(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		items := randInts(n)
		less := func(a, b int) bool { return a < b }

		b.Run(sizeName(n)+"/PushPop", func(b *testing.B) {
			for b.Loop() {
				h := heap.NewFrom(items, less)
				for _, v := range items {
					h.PushPop(v)
				}
			}
		})

		b.Run(sizeName(n)+"/PushAndPop", func(b *testing.B) {
			for b.Loop() {
				h := heap.NewFrom(items, less)
				for _, v := range items {
					h.Push(v)
					h.Pop()
				}
			}
		})
	}
}

func BenchmarkContains(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		items := randInts(n)
		less := func(a, b int) bool { return a < b }

		b.Run(sizeName(n)+"/hit", func(b *testing.B) {
			h := heap.NewFrom(items, less)
			target := items[n-1]
			b.ResetTimer()
			for b.Loop() {
				h.Contains(func(v int) bool { return v == target })
			}
		})

		b.Run(sizeName(n)+"/miss", func(b *testing.B) {
			h := heap.NewFrom(items, less)
			target := -1
			b.ResetTimer()
			for b.Loop() {
				h.Contains(func(v int) bool { return v == target })
			}
		})
	}
}

func BenchmarkDrain(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		items := randInts(n)
		less := func(a, b int) bool { return a < b }
		b.Run(sizeName(n), func(b *testing.B) {
			for b.Loop() {
				h := heap.NewFrom(items, less)
				h.Drain()
			}
		})
	}
}

func BenchmarkClear(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		items := randInts(n)
		less := func(a, b int) bool { return a < b }
		b.Run(sizeName(n), func(b *testing.B) {
			heaps := make([]*heap.Heap[int], b.N)
			for i := range heaps {
				heaps[i] = heap.NewFrom(items, less)
			}
			b.ResetTimer()
			for i := range b.N {
				heaps[i].Clear()
			}
		})
	}
}

func BenchmarkValues(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		items := randInts(n)
		less := func(a, b int) bool { return a < b }
		b.Run(sizeName(n), func(b *testing.B) {
			for b.Loop() {
				h := heap.NewFrom(items, less)
				for range h.Values() {
				}
			}
		})
	}
}

func BenchmarkClone(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		items := randInts(n)
		less := func(a, b int) bool { return a < b }
		b.Run(sizeName(n), func(b *testing.B) {
			h := heap.NewFrom(items, less)
			b.ResetTimer()
			for b.Loop() {
				h.Clone()
			}
		})
	}
}

func BenchmarkGrow(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		less := func(a, b int) bool { return a < b }
		b.Run(sizeName(n), func(b *testing.B) {
			for b.Loop() {
				h := heap.New(less)
				h.Grow(n)
			}
		})
	}
}

func BenchmarkAll(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		items := randInts(n)
		less := func(a, b int) bool { return a < b }
		b.Run(sizeName(n), func(b *testing.B) {
			h := heap.NewFrom(items, less)
			b.ResetTimer()
			for b.Loop() {
				for range h.All() {
				}
			}
		})
	}
}

func BenchmarkPopPush(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		items := randInts(n)
		less := func(a, b int) bool { return a < b }
		b.Run(sizeName(n), func(b *testing.B) {
			for b.Loop() {
				h := heap.NewFrom(items, less)
				for _, v := range items {
					h.PopPush(v)
				}
			}
		})
	}
}

func BenchmarkPopPushVsPopAndPush(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		items := randInts(n)
		less := func(a, b int) bool { return a < b }

		b.Run(sizeName(n)+"/PopPush", func(b *testing.B) {
			for b.Loop() {
				h := heap.NewFrom(items, less)
				for _, v := range items {
					h.PopPush(v)
				}
			}
		})

		b.Run(sizeName(n)+"/PopAndPush", func(b *testing.B) {
			for b.Loop() {
				h := heap.NewFrom(items, less)
				for _, v := range items {
					h.Pop()
					h.Push(v)
				}
			}
		})
	}
}

func BenchmarkMerge(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		items1 := randInts(n)
		items2 := randInts(n)
		less := func(a, b int) bool { return a < b }
		b.Run(sizeName(n), func(b *testing.B) {
			for b.Loop() {
				h1 := heap.NewFrom(items1, less)
				h2 := heap.NewFrom(items2, less)
				h1.Merge(h2)
			}
		})
	}
}

func BenchmarkMergeVsPushLoop(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		items1 := randInts(n)
		items2 := randInts(n)
		less := func(a, b int) bool { return a < b }

		b.Run(sizeName(n)+"/Merge", func(b *testing.B) {
			for b.Loop() {
				h1 := heap.NewFrom(items1, less)
				h2 := heap.NewFrom(items2, less)
				h1.Merge(h2)
			}
		})

		b.Run(sizeName(n)+"/PushLoop", func(b *testing.B) {
			for b.Loop() {
				h := heap.NewFrom(items1, less)
				for _, v := range items2 {
					h.Push(v)
				}
			}
		})
	}
}

func BenchmarkExtendVsPush(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		items := randInts(n)
		less := func(a, b int) bool { return a < b }

		b.Run(sizeName(n)+"/Extend", func(b *testing.B) {
			for b.Loop() {
				h := heap.New(less)
				h.Extend(items)
			}
		})

		b.Run(sizeName(n)+"/Push", func(b *testing.B) {
			for b.Loop() {
				h := heap.New(less)
				for _, v := range items {
					h.Push(v)
				}
			}
		})
	}
}

func BenchmarkExtendAsymmetric(b *testing.B) {
	less := func(a, b int) bool { return a < b }
	for _, n := range []int{1_000, 10_000, 100_000} {
		base := randInts(n)
		for _, m := range []int{1, 10, 100, 1_000} {
			if m > n {
				continue
			}
			extra := randInts(m)
			b.Run(fmt.Sprintf("n=%s/m=%s", sizeName(n), sizeName(m)), func(b *testing.B) {
				for b.Loop() {
					h := heap.NewFrom(base, less)
					h.Extend(extra)
				}
			})
		}
	}
}

func BenchmarkMergeAsymmetric(b *testing.B) {
	less := func(a, b int) bool { return a < b }
	for _, n := range []int{1_000, 10_000, 100_000} {
		base := randInts(n)
		for _, m := range []int{1, 10, 100, 1_000} {
			if m > n {
				continue
			}
			extra := randInts(m)
			b.Run(fmt.Sprintf("n=%s/m=%s", sizeName(n), sizeName(m)), func(b *testing.B) {
				for b.Loop() {
					h1 := heap.NewFrom(base, less)
					h2 := heap.NewFrom(extra, less)
					h1.Merge(h2)
				}
			})
		}
	}
}

func sizeName(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%dM", n/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%dK", n/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}
