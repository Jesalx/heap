package heap_test

import (
	"math/rand/v2"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jesalx/heap"
)

func drainAll[T any](t *testing.T, h *heap.Heap[T]) []T {
	t.Helper()
	var result []T
	for h.Len() > 0 {
		v, ok := h.Pop()
		if !ok {
			t.Fatal("Pop returned ok=false on non-empty heap")
		}
		result = append(result, v)
	}
	return result
}

func assertHeapInvariant[T any](t *testing.T, h *heap.Heap[T], less func(a, b T) bool) {
	t.Helper()
	items := slices.Collect(h.All())
	for i := 1; i < len(items); i++ {
		parent := (i - 1) / 2
		if less(items[i], items[parent]) {
			t.Fatalf("heap invariant violated: items[%d] has higher priority than parent items[%d]", i, parent)
		}
	}
}

func TestNewNilLess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		fn   func()
	}{
		{
			name: "New",
			fn:   func() { heap.New[int](nil) },
		},
		{
			name: "NewFrom",
			fn:   func() { heap.NewFrom([]int{1, 2, 3}, nil) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer func() {
				r := recover()
				if r == nil {
					t.Fatal("expected panic, got none")
				}
				msg, ok := r.(string)
				if !ok {
					t.Fatalf("expected string panic, got %T: %v", r, r)
				}
				if msg != "heap: less function must not be nil" {
					t.Errorf("unexpected panic message: %q", msg)
				}
			}()
			tt.fn()
		})
	}
}

func TestIntOrdering(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		less func(a, b int) bool
		push []int
		want []int
	}{
		{
			name: "min heap",
			less: func(a, b int) bool { return a < b },
			push: []int{5, 3, 7, 1},
			want: []int{1, 3, 5, 7},
		},
		{
			name: "max heap",
			less: func(a, b int) bool { return a > b },
			push: []int{5, 3, 7, 1},
			want: []int{7, 5, 3, 1},
		},
		{
			name: "duplicates",
			less: func(a, b int) bool { return a < b },
			push: []int{3, 1, 3, 1, 2},
			want: []int{1, 1, 2, 3, 3},
		},
		{
			name: "single element",
			less: func(a, b int) bool { return a < b },
			push: []int{42},
			want: []int{42},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := heap.New(tt.less)
			for _, v := range tt.push {
				h.Push(v)
			}
			got := drainAll(t, h)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("drain mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFloat64Ordering(t *testing.T) {
	t.Parallel()

	h := heap.New(func(a, b float64) bool { return a < b })
	for _, v := range []float64{3.14, 1.41, 2.72, 0.58} {
		h.Push(v)
	}

	want := []float64{0.58, 1.41, 2.72, 3.14}
	got := drainAll(t, h)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("drain mismatch (-want +got):\n%s", diff)
	}
}

func TestStringOrdering(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		less func(a, b string) bool
		push []string
		want []string
	}{
		{
			name: "lexicographic",
			less: func(a, b string) bool { return a < b },
			push: []string{"cherry", "apple", "banana", "date"},
			want: []string{"apple", "banana", "cherry", "date"},
		},
		{
			name: "case insensitive",
			less: func(a, b string) bool { return strings.ToLower(a) < strings.ToLower(b) },
			push: []string{"Banana", "apple", "Cherry"},
			want: []string{"apple", "Banana", "Cherry"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := heap.New(tt.less)
			for _, v := range tt.push {
				h.Push(v)
			}
			got := drainAll(t, h)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("drain mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

type task struct {
	name     string
	priority int
}

func TestStructHeap(t *testing.T) {
	t.Parallel()

	h := heap.New(func(a, b task) bool { return a.priority < b.priority })

	h.Push(task{"low", 10})
	h.Push(task{"critical", 1})
	h.Push(task{"medium", 5})
	h.Push(task{"high", 3})

	want := []task{
		{"critical", 1},
		{"high", 3},
		{"medium", 5},
		{"low", 10},
	}
	got := drainAll(t, h)
	if diff := cmp.Diff(want, got, cmp.AllowUnexported(task{})); diff != "" {
		t.Errorf("drain mismatch (-want +got):\n%s", diff)
	}
}

func TestPointerHeap(t *testing.T) {
	t.Parallel()

	h := heap.New(func(a, b *task) bool { return a.priority < b.priority })

	h.Push(&task{"low", 10})
	h.Push(&task{"critical", 1})
	h.Push(&task{"high", 3})

	want := []*task{
		{"critical", 1},
		{"high", 3},
		{"low", 10},
	}
	got := drainAll(t, h)
	if diff := cmp.Diff(want, got, cmp.AllowUnexported(task{})); diff != "" {
		t.Errorf("drain mismatch (-want +got):\n%s", diff)
	}
}

func TestTimeHeap(t *testing.T) {
	t.Parallel()

	h := heap.New(func(a, b time.Time) bool { return a.Before(b) })

	now := time.Now()
	t1 := now.Add(3 * time.Hour)
	t2 := now.Add(1 * time.Hour)
	t3 := now.Add(2 * time.Hour)

	h.Push(t1)
	h.Push(t2)
	h.Push(t3)

	want := []time.Time{t2, t3, t1}
	got := drainAll(t, h)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("drain mismatch (-want +got):\n%s", diff)
	}
}

func TestDurationHeap(t *testing.T) {
	t.Parallel()

	h := heap.New(func(a, b time.Duration) bool { return a < b })

	h.Push(5 * time.Second)
	h.Push(1 * time.Millisecond)
	h.Push(3 * time.Minute)

	want := []time.Duration{1 * time.Millisecond, 5 * time.Second, 3 * time.Minute}
	got := drainAll(t, h)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("drain mismatch (-want +got):\n%s", diff)
	}
}

func TestBoolHeap(t *testing.T) {
	t.Parallel()

	h := heap.New(func(a, b bool) bool {
		if a == b {
			return false
		}
		return !a
	})

	h.Push(true)
	h.Push(false)
	h.Push(true)

	want := []bool{false, true, true}
	got := drainAll(t, h)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("drain mismatch (-want +got):\n%s", diff)
	}
}

func TestByteSliceHeap(t *testing.T) {
	t.Parallel()

	h := heap.New(func(a, b []byte) bool { return string(a) < string(b) })

	h.Push([]byte("banana"))
	h.Push([]byte("apple"))
	h.Push([]byte("cherry"))

	want := [][]byte{[]byte("apple"), []byte("banana"), []byte("cherry")}
	got := drainAll(t, h)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("drain mismatch (-want +got):\n%s", diff)
	}
}

func TestEmptyHeap(t *testing.T) {
	t.Parallel()

	t.Run("pop int", func(t *testing.T) {
		t.Parallel()
		h := heap.New(func(a, b int) bool { return a < b })
		val, ok := h.Pop()
		if ok {
			t.Error("expected ok=false")
		}
		if val != 0 {
			t.Errorf("expected zero value, got %d", val)
		}
	})

	t.Run("peek int", func(t *testing.T) {
		t.Parallel()
		h := heap.New(func(a, b int) bool { return a < b })
		val, ok := h.Peek()
		if ok {
			t.Error("expected ok=false")
		}
		if val != 0 {
			t.Errorf("expected zero value, got %d", val)
		}
	})

	t.Run("pop string", func(t *testing.T) {
		t.Parallel()
		h := heap.New(func(a, b string) bool { return a < b })
		val, ok := h.Pop()
		if ok {
			t.Error("expected ok=false")
		}
		if val != "" {
			t.Errorf("expected zero value, got %q", val)
		}
	})

	t.Run("pop pointer", func(t *testing.T) {
		t.Parallel()
		h := heap.New(func(a, b *task) bool { return a.priority < b.priority })
		val, ok := h.Pop()
		if ok {
			t.Error("expected ok=false")
		}
		if val != nil {
			t.Errorf("expected nil, got %+v", val)
		}
	})
}

func TestPeek(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		push     []int
		popFirst int
		wantPeek int
		wantLen  int
	}{
		{
			name:     "returns min without removing",
			push:     []int{5, 3, 7},
			wantPeek: 3,
			wantLen:  3,
		},
		{
			name:     "updates after pop",
			push:     []int{2, 1, 3},
			popFirst: 1,
			wantPeek: 2,
			wantLen:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := heap.New(func(a, b int) bool { return a < b })
			for _, v := range tt.push {
				h.Push(v)
			}
			for range tt.popFirst {
				h.Pop()
			}

			val, ok := h.Peek()
			if !ok {
				t.Fatal("expected ok=true")
			}
			if val != tt.wantPeek {
				t.Errorf("got %d, want %d", val, tt.wantPeek)
			}
			if h.Len() != tt.wantLen {
				t.Errorf("len=%d, want %d", h.Len(), tt.wantLen)
			}
		})
	}
}

func TestLen(t *testing.T) {
	t.Parallel()

	h := heap.New(func(a, b int) bool { return a < b })

	if h.Len() != 0 {
		t.Fatalf("new heap len=%d, want 0", h.Len())
	}

	h.Push(1)
	h.Push(2)
	if h.Len() != 2 {
		t.Fatalf("after two pushes len=%d, want 2", h.Len())
	}

	h.Pop()
	if h.Len() != 1 {
		t.Fatalf("after pop len=%d, want 1", h.Len())
	}

	h.Pop()
	if h.Len() != 0 {
		t.Fatalf("after final pop len=%d, want 0", h.Len())
	}
}

func TestMixedPushPop(t *testing.T) {
	t.Parallel()

	h := heap.New(func(a, b int) bool { return a < b })

	h.Push(5)
	h.Push(3)

	got, _ := h.Pop()
	if got != 3 {
		t.Fatalf("got %d, want 3", got)
	}

	h.Push(1)
	h.Push(4)

	want := []int{1, 4, 5}
	got2 := drainAll(t, h)
	if diff := cmp.Diff(want, got2); diff != "" {
		t.Errorf("drain mismatch (-want +got):\n%s", diff)
	}
}

func TestLargeHeap(t *testing.T) {
	t.Parallel()

	h := heap.New(func(a, b int) bool { return a < b })

	n := 10_000
	for i := n; i > 0; i-- {
		h.Push(i)
	}

	if h.Len() != n {
		t.Fatalf("len=%d, want %d", h.Len(), n)
	}

	prev := 0
	for i := range n {
		got, ok := h.Pop()
		if !ok {
			t.Fatalf("pop %d: expected ok=true", i)
		}
		if got <= prev {
			t.Fatalf("pop %d: got %d, not greater than previous %d", i, got, prev)
		}
		prev = got
	}
}

func TestNewFrom(t *testing.T) {
	t.Parallel()

	intLess := func(a, b int) bool { return a < b }

	t.Run("matches push-one-by-one", func(t *testing.T) {
		t.Parallel()

		items := []int{5, 3, 7, 1, 4, 2, 6}
		fromHeap := heap.NewFrom(items, intLess)

		pushHeap := heap.New(intLess)
		for _, v := range items {
			pushHeap.Push(v)
		}

		got := drainAll(t, fromHeap)
		want := drainAll(t, pushHeap)
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("does not mutate input slice", func(t *testing.T) {
		t.Parallel()

		original := []int{5, 3, 7, 1}
		snapshot := slices.Clone(original)

		h := heap.NewFrom(original, intLess)
		drainAll(t, h)

		if diff := cmp.Diff(snapshot, original); diff != "" {
			t.Errorf("input slice was mutated (-want +got):\n%s", diff)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()

		h := heap.NewFrom([]int{}, intLess)
		if h.Len() != 0 {
			t.Fatalf("len=%d, want 0", h.Len())
		}
		_, ok := h.Pop()
		if ok {
			t.Error("expected ok=false on empty heap")
		}
	})

	t.Run("single element", func(t *testing.T) {
		t.Parallel()

		h := heap.NewFrom([]int{42}, intLess)
		if h.Len() != 1 {
			t.Fatalf("len=%d, want 1", h.Len())
		}
		val, ok := h.Pop()
		if !ok {
			t.Fatal("expected ok=true")
		}
		if val != 42 {
			t.Errorf("got %d, want 42", val)
		}
	})
}

func TestValues(t *testing.T) {
	t.Parallel()

	intLess := func(a, b int) bool { return a < b }

	t.Run("drain via collect", func(t *testing.T) {
		t.Parallel()

		h := heap.NewFrom([]int{5, 3, 7, 1, 4}, intLess)
		got := slices.Collect(h.Values())
		want := []int{1, 3, 4, 5, 7}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
		if h.Len() != 0 {
			t.Errorf("heap should be empty after full iteration, len=%d", h.Len())
		}
	})

	t.Run("break early", func(t *testing.T) {
		t.Parallel()

		h := heap.NewFrom([]int{5, 3, 7, 1, 4}, intLess)
		var got []int
		for v := range h.Values() {
			got = append(got, v)
			if len(got) == 2 {
				break
			}
		}

		want := []int{1, 3}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("early break mismatch (-want +got):\n%s", diff)
		}
		if h.Len() != 3 {
			t.Errorf("remaining len=%d, want 3", h.Len())
		}
	})

	t.Run("empty heap", func(t *testing.T) {
		t.Parallel()

		h := heap.New(intLess)
		got := slices.Collect(h.Values())
		if len(got) != 0 {
			t.Errorf("expected no values, got %v", got)
		}
	})
}

func TestPushAfterDrain(t *testing.T) {
	t.Parallel()

	h := heap.New(func(a, b int) bool { return a < b })
	h.Push(3)
	h.Push(1)
	drainAll(t, h)

	h.Push(5)
	h.Push(2)
	h.Push(4)

	got := drainAll(t, h)
	want := []int{2, 4, 5}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("drain mismatch (-want +got):\n%s", diff)
	}
}

func TestAllIdenticalElements(t *testing.T) {
	t.Parallel()

	h := heap.New(func(a, b int) bool { return a < b })
	for range 5 {
		h.Push(7)
	}

	got := drainAll(t, h)
	want := []int{7, 7, 7, 7, 7}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("drain mismatch (-want +got):\n%s", diff)
	}
}

func TestNewMin(t *testing.T) {
	t.Parallel()

	t.Run("int", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMin[int]()
		for _, v := range []int{5, 3, 7, 1} {
			h.Push(v)
		}
		got := drainAll(t, h)
		want := []int{1, 3, 5, 7}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("float64", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMin[float64]()
		for _, v := range []float64{3.14, 1.41, 2.72} {
			h.Push(v)
		}
		got := drainAll(t, h)
		want := []float64{1.41, 2.72, 3.14}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMin[string]()
		for _, v := range []string{"cherry", "apple", "banana"} {
			h.Push(v)
		}
		got := drainAll(t, h)
		want := []string{"apple", "banana", "cherry"}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMin[int]()
		if h.Len() != 0 {
			t.Fatalf("len=%d, want 0", h.Len())
		}
		_, ok := h.Pop()
		if ok {
			t.Error("expected ok=false on empty heap")
		}
	})

	t.Run("single element", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMin[int]()
		h.Push(42)
		val, ok := h.Pop()
		if !ok {
			t.Fatal("expected ok=true")
		}
		if val != 42 {
			t.Errorf("got %d, want 42", val)
		}
	})

	t.Run("duplicates", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMin[int]()
		for _, v := range []int{3, 1, 3, 1, 2} {
			h.Push(v)
		}
		got := drainAll(t, h)
		want := []int{1, 1, 2, 3, 3}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("equivalent to New with less-than", func(t *testing.T) {
		t.Parallel()
		items := []int{5, 3, 7, 1, 4, 2, 6}

		minHeap := heap.NewMin[int]()
		manualHeap := heap.New(func(a, b int) bool { return a < b })
		for _, v := range items {
			minHeap.Push(v)
			manualHeap.Push(v)
		}

		got := drainAll(t, minHeap)
		want := drainAll(t, manualHeap)
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestNewMax(t *testing.T) {
	t.Parallel()

	t.Run("int", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMax[int]()
		for _, v := range []int{5, 3, 7, 1} {
			h.Push(v)
		}
		got := drainAll(t, h)
		want := []int{7, 5, 3, 1}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("float64", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMax[float64]()
		for _, v := range []float64{3.14, 1.41, 2.72} {
			h.Push(v)
		}
		got := drainAll(t, h)
		want := []float64{3.14, 2.72, 1.41}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMax[string]()
		for _, v := range []string{"cherry", "apple", "banana"} {
			h.Push(v)
		}
		got := drainAll(t, h)
		want := []string{"cherry", "banana", "apple"}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMax[int]()
		if h.Len() != 0 {
			t.Fatalf("len=%d, want 0", h.Len())
		}
		_, ok := h.Pop()
		if ok {
			t.Error("expected ok=false on empty heap")
		}
	})

	t.Run("single element", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMax[int]()
		h.Push(42)
		val, ok := h.Pop()
		if !ok {
			t.Fatal("expected ok=true")
		}
		if val != 42 {
			t.Errorf("got %d, want 42", val)
		}
	})

	t.Run("duplicates", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMax[int]()
		for _, v := range []int{3, 1, 3, 1, 2} {
			h.Push(v)
		}
		got := drainAll(t, h)
		want := []int{3, 3, 2, 1, 1}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("equivalent to New with greater-than", func(t *testing.T) {
		t.Parallel()
		items := []int{5, 3, 7, 1, 4, 2, 6}

		maxHeap := heap.NewMax[int]()
		manualHeap := heap.New(func(a, b int) bool { return a > b })
		for _, v := range items {
			maxHeap.Push(v)
			manualHeap.Push(v)
		}

		got := drainAll(t, maxHeap)
		want := drainAll(t, manualHeap)
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestNewMinFrom(t *testing.T) {
	t.Parallel()

	t.Run("correct ordering", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{5, 3, 7, 1, 4})
		got := drainAll(t, h)
		want := []int{1, 3, 4, 5, 7}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("input slice not mutated", func(t *testing.T) {
		t.Parallel()
		original := []int{5, 3, 7, 1}
		snapshot := slices.Clone(original)
		h := heap.NewMinFrom(original)
		drainAll(t, h)
		if diff := cmp.Diff(snapshot, original); diff != "" {
			t.Errorf("input slice was mutated (-want +got):\n%s", diff)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{})
		if h.Len() != 0 {
			t.Fatalf("len=%d, want 0", h.Len())
		}
	})

	t.Run("single element", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{42})
		val, ok := h.Pop()
		if !ok {
			t.Fatal("expected ok=true")
		}
		if val != 42 {
			t.Errorf("got %d, want 42", val)
		}
	})

	t.Run("equivalent to NewFrom with cmp.Less", func(t *testing.T) {
		t.Parallel()
		items := []int{5, 3, 7, 1, 4, 2, 6}
		minFrom := heap.NewMinFrom(items)
		explicit := heap.NewFrom(items, func(a, b int) bool { return a < b })
		got := drainAll(t, minFrom)
		want := drainAll(t, explicit)
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestNewMaxFrom(t *testing.T) {
	t.Parallel()

	t.Run("correct ordering", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMaxFrom([]int{5, 3, 7, 1, 4})
		got := drainAll(t, h)
		want := []int{7, 5, 4, 3, 1}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("input slice not mutated", func(t *testing.T) {
		t.Parallel()
		original := []int{5, 3, 7, 1}
		snapshot := slices.Clone(original)
		h := heap.NewMaxFrom(original)
		drainAll(t, h)
		if diff := cmp.Diff(snapshot, original); diff != "" {
			t.Errorf("input slice was mutated (-want +got):\n%s", diff)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMaxFrom([]int{})
		if h.Len() != 0 {
			t.Fatalf("len=%d, want 0", h.Len())
		}
	})

	t.Run("single element", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMaxFrom([]int{42})
		val, ok := h.Pop()
		if !ok {
			t.Fatal("expected ok=true")
		}
		if val != 42 {
			t.Errorf("got %d, want 42", val)
		}
	})

	t.Run("equivalent to NewFrom with reverse cmp.Less", func(t *testing.T) {
		t.Parallel()
		items := []int{5, 3, 7, 1, 4, 2, 6}
		maxFrom := heap.NewMaxFrom(items)
		explicit := heap.NewFrom(items, func(a, b int) bool { return a > b })
		got := drainAll(t, maxFrom)
		want := drainAll(t, explicit)
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestClear(t *testing.T) {
	t.Parallel()

	t.Run("clears populated heap", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMin[int]()
		for _, v := range []int{5, 3, 7, 1} {
			h.Push(v)
		}
		h.Clear()
		if h.Len() != 0 {
			t.Fatalf("len=%d, want 0", h.Len())
		}
		_, ok := h.Pop()
		if ok {
			t.Error("expected ok=false after clear")
		}
	})

	t.Run("clear empty", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMin[int]()
		h.Clear()
		if h.Len() != 0 {
			t.Fatalf("len=%d, want 0", h.Len())
		}
	})

	t.Run("push after clear", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMin[int]()
		for _, v := range []int{5, 3, 7} {
			h.Push(v)
		}
		h.Clear()
		h.Push(10)
		h.Push(2)
		got := drainAll(t, h)
		want := []int{2, 10}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("double clear", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMin[int]()
		h.Push(1)
		h.Clear()
		h.Clear()
		if h.Len() != 0 {
			t.Fatalf("len=%d, want 0", h.Len())
		}
	})

	t.Run("pointer type zeroed", func(t *testing.T) {
		t.Parallel()
		h := heap.New(func(a, b *task) bool { return a.priority < b.priority })
		h.Push(&task{"a", 1})
		h.Push(&task{"b", 2})
		h.Clear()
		if h.Len() != 0 {
			t.Fatalf("len=%d, want 0", h.Len())
		}
		val, ok := h.Pop()
		if ok {
			t.Error("expected ok=false after clear")
		}
		if val != nil {
			t.Errorf("expected nil, got %+v", val)
		}
	})
}

func TestDrain(t *testing.T) {
	t.Parallel()

	t.Run("min ordering", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMin[int]()
		for _, v := range []int{5, 3, 7, 1} {
			h.Push(v)
		}
		got := h.Drain()
		want := []int{1, 3, 5, 7}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("max ordering", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMax[int]()
		for _, v := range []int{5, 3, 7, 1} {
			h.Push(v)
		}
		got := h.Drain()
		want := []int{7, 5, 3, 1}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("empty returns nil", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMin[int]()
		got := h.Drain()
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("single element", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMin[int]()
		h.Push(42)
		got := h.Drain()
		want := []int{42}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("heap empty after drain", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMin[int]()
		for _, v := range []int{3, 1, 2} {
			h.Push(v)
		}
		h.Drain()
		if h.Len() != 0 {
			t.Fatalf("len=%d, want 0", h.Len())
		}
	})

	t.Run("push after drain", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMin[int]()
		h.Push(3)
		h.Push(1)
		h.Drain()

		h.Push(5)
		h.Push(2)
		got := h.Drain()
		want := []int{2, 5}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("equivalent to slices.Collect(Values)", func(t *testing.T) {
		t.Parallel()
		items := []int{5, 3, 7, 1, 4, 2, 6}
		intLess := func(a, b int) bool { return a < b }

		drainHeap := heap.NewFrom(items, intLess)
		valuesHeap := heap.NewFrom(items, intLess)

		got := drainHeap.Drain()
		want := slices.Collect(valuesHeap.Values())
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("duplicates", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMin[int]()
		for _, v := range []int{3, 1, 3, 1, 2} {
			h.Push(v)
		}
		got := h.Drain()
		want := []int{1, 1, 2, 3, 3}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestPushPop(t *testing.T) {
	t.Parallel()

	t.Run("empty heap", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMin[int]()
		got := h.PushPop(5)
		if got != 5 {
			t.Errorf("got %d, want 5", got)
		}
		if h.Len() != 0 {
			t.Errorf("len=%d, want 0", h.Len())
		}
	})

	t.Run("val higher priority than root", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{3, 5, 7})
		got := h.PushPop(1)
		if got != 1 {
			t.Errorf("got %d, want 1", got)
		}
		want := []int{3, 5, 7}
		if diff := cmp.Diff(want, drainAll(t, h)); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("val lower priority than root", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{3, 5, 7})
		got := h.PushPop(4)
		if got != 3 {
			t.Errorf("got %d, want 3", got)
		}
		want := []int{4, 5, 7}
		if diff := cmp.Diff(want, drainAll(t, h)); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("equal priority", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{3, 5, 7})
		got := h.PushPop(3)
		if got != 3 {
			t.Errorf("got %d, want 3", got)
		}
		want := []int{3, 5, 7}
		if diff := cmp.Diff(want, drainAll(t, h)); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("single element heap", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{10})
		got := h.PushPop(5)
		if got != 5 {
			t.Errorf("got %d, want 5", got)
		}
		val, ok := h.Pop()
		if !ok {
			t.Fatal("expected ok=true")
		}
		if val != 10 {
			t.Errorf("got %d, want 10", val)
		}
	})

	t.Run("equivalence to Push then Pop", func(t *testing.T) {
		t.Parallel()
		items := []int{5, 3, 7, 1, 4, 2, 6}
		val := 4

		pushPop := heap.NewMinFrom(items)
		gotPP := pushPop.PushPop(val)

		separate := heap.NewMinFrom(items)
		separate.Push(val)
		gotSep, _ := separate.Pop()

		if gotPP != gotSep {
			t.Errorf("PushPop=%d, Push+Pop=%d", gotPP, gotSep)
		}

		wantPP := drainAll(t, pushPop)
		wantSep := drainAll(t, separate)
		if diff := cmp.Diff(wantSep, wantPP); diff != "" {
			t.Errorf("remaining heap mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("consecutive PushPops", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{3, 5, 7})
		var results []int
		for _, v := range []int{1, 4, 6, 2} {
			results = append(results, h.PushPop(v))
		}
		want := []int{1, 3, 4, 2}
		if diff := cmp.Diff(want, results); diff != "" {
			t.Errorf("results mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("max heap", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMaxFrom([]int{7, 5, 3})
		got := h.PushPop(10)
		if got != 10 {
			t.Errorf("got %d, want 10", got)
		}

		got = h.PushPop(4)
		if got != 7 {
			t.Errorf("got %d, want 7", got)
		}
	})
}

func TestContains(t *testing.T) {
	t.Parallel()

	t.Run("element present", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{5, 3, 7, 1})
		if !h.Contains(func(v int) bool { return v == 3 }) {
			t.Error("expected Contains to return true for 3")
		}
	})

	t.Run("element absent", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{5, 3, 7, 1})
		if h.Contains(func(v int) bool { return v == 99 }) {
			t.Error("expected Contains to return false for 99")
		}
	})

	t.Run("empty heap", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMin[int]()
		if h.Contains(func(v int) bool { return v == 1 }) {
			t.Error("expected Contains to return false on empty heap")
		}
	})

	t.Run("struct field predicate", func(t *testing.T) {
		t.Parallel()
		h := heap.New(func(a, b task) bool { return a.priority < b.priority })
		h.Push(task{"low", 10})
		h.Push(task{"critical", 1})
		h.Push(task{"medium", 5})

		if !h.Contains(func(t task) bool { return t.name == "critical" }) {
			t.Error("expected to find task 'critical'")
		}
		if h.Contains(func(t task) bool { return t.name == "high" }) {
			t.Error("expected not to find task 'high'")
		}
	})

	t.Run("after pop and push", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{1, 2, 3})
		h.Pop()
		if h.Contains(func(v int) bool { return v == 1 }) {
			t.Error("expected 1 to be gone after pop")
		}
		h.Push(10)
		if !h.Contains(func(v int) bool { return v == 10 }) {
			t.Error("expected 10 to be present after push")
		}
	})

	t.Run("multiple matches", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{2, 4, 6, 8})
		if !h.Contains(func(v int) bool { return v%2 == 0 }) {
			t.Error("expected to find an even number")
		}
	})
}

func TestCap(t *testing.T) {
	t.Parallel()

	t.Run("empty heap", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMin[int]()
		if h.Cap() < 0 {
			t.Errorf("Cap()=%d, want >= 0", h.Cap())
		}
	})

	t.Run("after Grow", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMin[int]()
		h.Grow(100)
		if h.Cap() < 100 {
			t.Errorf("Cap()=%d after Grow(100), want >= 100", h.Cap())
		}
	})

	t.Run("after pushes", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMin[int]()
		for i := range 50 {
			h.Push(i)
		}
		if h.Cap() < h.Len() {
			t.Errorf("Cap()=%d < Len()=%d", h.Cap(), h.Len())
		}
	})
}

func TestRemove(t *testing.T) {
	t.Parallel()

	intLess := func(a, b int) bool { return a < b }

	t.Run("element present", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{5, 3, 7, 1, 4})
		val, ok := h.Remove(func(v int) bool { return v == 3 })
		if !ok {
			t.Fatal("expected ok=true")
		}
		if val != 3 {
			t.Errorf("got %d, want 3", val)
		}
		got := drainAll(t, h)
		want := []int{1, 4, 5, 7}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("element absent", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{5, 3, 7, 1})
		val, ok := h.Remove(func(v int) bool { return v == 99 })
		if ok {
			t.Error("expected ok=false for absent element")
		}
		if val != 0 {
			t.Errorf("got %d, want zero value 0", val)
		}
		if h.Len() != 4 {
			t.Errorf("len=%d, want 4 (heap should be unchanged)", h.Len())
		}
	})

	t.Run("empty heap", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMin[int]()
		val, ok := h.Remove(func(v int) bool { return v == 1 })
		if ok {
			t.Error("expected ok=false on empty heap")
		}
		if val != 0 {
			t.Errorf("got %d, want zero value 0", val)
		}
	})

	t.Run("removes first match", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{3, 1, 3, 1, 2})
		val, ok := h.Remove(func(v int) bool { return v == 3 })
		if !ok {
			t.Fatal("expected ok=true")
		}
		if val != 3 {
			t.Errorf("got %d, want 3", val)
		}
		// Only one 3 should be removed.
		got := drainAll(t, h)
		want := []int{1, 1, 2, 3}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("struct field predicate", func(t *testing.T) {
		t.Parallel()
		h := heap.New(func(a, b task) bool { return a.priority < b.priority })
		h.Push(task{"low", 10})
		h.Push(task{"critical", 1})
		h.Push(task{"medium", 5})

		val, ok := h.Remove(func(t task) bool { return t.name == "medium" })
		if !ok {
			t.Fatal("expected ok=true")
		}
		if val.name != "medium" || val.priority != 5 {
			t.Errorf("got %+v, want {medium 5}", val)
		}
		got := drainAll(t, h)
		want := []task{{"critical", 1}, {"low", 10}}
		if diff := cmp.Diff(want, got, cmp.AllowUnexported(task{})); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("heap invariant preserved", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{10, 20, 30, 40, 50, 60, 70, 80})
		h.Remove(func(v int) bool { return v == 40 })
		assertHeapInvariant(t, h, intLess)
	})

	t.Run("push after remove", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{5, 3, 7})
		h.Remove(func(v int) bool { return v == 5 })
		h.Push(1)
		got := drainAll(t, h)
		want := []int{1, 3, 7}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestExtend(t *testing.T) {
	t.Parallel()

	intLess := func(a, b int) bool { return a < b }

	t.Run("basic extend", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{3, 1})
		h.Extend([]int{5, 2, 4})
		got := drainAll(t, h)
		want := []int{1, 2, 3, 4, 5}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("extend empty heap", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMin[int]()
		h.Extend([]int{3, 1, 2})
		got := drainAll(t, h)
		want := []int{1, 2, 3}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("extend with empty slice", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{1, 2, 3})
		h.Extend([]int{})
		got := drainAll(t, h)
		want := []int{1, 2, 3}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("extend with nil slice", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{1, 2, 3})
		h.Extend(nil)
		got := drainAll(t, h)
		want := []int{1, 2, 3}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("push after extend", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{3, 5})
		h.Extend([]int{2, 4})
		h.Push(1)
		got := drainAll(t, h)
		want := []int{1, 2, 3, 4, 5}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("max heap", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMaxFrom([]int{7, 5})
		h.Extend([]int{6, 4, 2})
		got := drainAll(t, h)
		want := []int{7, 6, 5, 4, 2}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("heap invariant preserved", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{10, 20, 30})
		h.Extend([]int{5, 15, 25, 35})
		assertHeapInvariant(t, h, intLess)
	})

	t.Run("small extend into large heap", func(t *testing.T) {
		t.Parallel()
		base := make([]int, 1000)
		for i := range base {
			base[i] = i + 10
		}
		h := heap.NewMinFrom(base)
		h.Extend([]int{3, 1, 7})
		assertHeapInvariant(t, h, intLess)
		if h.Len() != 1003 {
			t.Fatalf("len=%d, want 1003", h.Len())
		}
		if v, _ := h.Peek(); v != 1 {
			t.Errorf("peek=%d, want 1", v)
		}
	})

	t.Run("large extend into small heap", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{1, 5, 3})
		extra := make([]int, 500)
		for i := range extra {
			extra[i] = i + 10
		}
		h.Extend(extra)
		assertHeapInvariant(t, h, intLess)
		if v, _ := h.Peek(); v != 1 {
			t.Errorf("peek=%d, want 1", v)
		}
		if h.Len() != 503 {
			t.Fatalf("len=%d, want 503", h.Len())
		}
	})
}

func TestPeekStability(t *testing.T) {
	t.Parallel()

	h := heap.NewFrom([]int{5, 3, 7, 1}, func(a, b int) bool { return a < b })

	for range 3 {
		val, ok := h.Peek()
		if !ok {
			t.Fatal("expected ok=true")
		}
		if val != 1 {
			t.Errorf("got %d, want 1", val)
		}
	}

	if h.Len() != 4 {
		t.Errorf("len=%d, want 4 (Peek should not modify heap)", h.Len())
	}
}

func TestClone(t *testing.T) {
	t.Parallel()

	intLess := func(a, b int) bool { return a < b }

	t.Run("preserves ordering", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{5, 3, 7, 1, 4})
		c := h.Clone()
		got := drainAll(t, c)
		want := []int{1, 3, 4, 5, 7}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("independent after clone", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{3, 5, 7})
		c := h.Clone()

		h.Push(1)
		c.Push(10)

		gotH := drainAll(t, h)
		wantH := []int{1, 3, 5, 7}
		if diff := cmp.Diff(wantH, gotH); diff != "" {
			t.Errorf("original mismatch (-want +got):\n%s", diff)
		}

		gotC := drainAll(t, c)
		wantC := []int{3, 5, 7, 10}
		if diff := cmp.Diff(wantC, gotC); diff != "" {
			t.Errorf("clone mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("empty heap", func(t *testing.T) {
		t.Parallel()
		h := heap.New(intLess)
		c := h.Clone()
		if c.Len() != 0 {
			t.Fatalf("len=%d, want 0", c.Len())
		}
		c.Push(1)
		if h.Len() != 0 {
			t.Error("push to clone affected original")
		}
	})

	t.Run("mutations to clone do not affect original", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{1, 2, 3, 4, 5})
		c := h.Clone()

		c.Pop()
		c.Pop()
		c.Push(100)

		gotH := drainAll(t, h)
		wantH := []int{1, 2, 3, 4, 5}
		if diff := cmp.Diff(wantH, gotH); diff != "" {
			t.Errorf("original mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestGrow(t *testing.T) {
	t.Parallel()

	t.Run("increases capacity", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMin[int]()
		h.Grow(100)
		for i := range 100 {
			h.Push(i)
		}
		got := drainAll(t, h)
		want := make([]int, 100)
		for i := range want {
			want[i] = i
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("preserves existing elements", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{3, 1, 2})
		h.Grow(50)
		got := drainAll(t, h)
		want := []int{1, 2, 3}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("negative panics", func(t *testing.T) {
		t.Parallel()
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic for negative Grow")
			}
		}()
		h := heap.NewMin[int]()
		h.Grow(-1)
	})
}

func TestAll(t *testing.T) {
	t.Parallel()

	t.Run("yields all elements", func(t *testing.T) {
		t.Parallel()
		input := []int{5, 3, 7, 1, 4}
		h := heap.NewMinFrom(input)
		got := slices.Collect(h.All())
		sort.Ints(got)
		want := slices.Clone(input)
		sort.Ints(want)
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("non-destructive", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{3, 1, 2})
		_ = slices.Collect(h.All())
		if h.Len() != 3 {
			t.Errorf("len=%d, want 3", h.Len())
		}
		got := drainAll(t, h)
		want := []int{1, 2, 3}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("empty heap", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMin[int]()
		got := slices.Collect(h.All())
		if len(got) != 0 {
			t.Errorf("expected no values, got %v", got)
		}
	})

	t.Run("early break", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{5, 3, 7, 1, 4})
		count := 0
		for range h.All() {
			count++
			if count == 2 {
				break
			}
		}
		if count != 2 {
			t.Errorf("count=%d, want 2", count)
		}
		if h.Len() != 5 {
			t.Errorf("len=%d, want 5 (All should not modify heap)", h.Len())
		}
	})
}

func TestPopPush(t *testing.T) {
	t.Parallel()

	t.Run("empty heap", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMin[int]()
		val, ok := h.PopPush(5)
		if ok {
			t.Error("expected ok=false on empty heap")
		}
		if val != 0 {
			t.Errorf("got %d, want 0", val)
		}
		if h.Len() != 0 {
			t.Errorf("len=%d, want 0", h.Len())
		}
	})

	t.Run("val lower priority than root", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{3, 5, 7})
		old, ok := h.PopPush(10)
		if !ok {
			t.Fatal("expected ok=true")
		}
		if old != 3 {
			t.Errorf("got %d, want 3", old)
		}
		want := []int{5, 7, 10}
		if diff := cmp.Diff(want, drainAll(t, h)); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("val higher priority than root", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{3, 5, 7})
		old, ok := h.PopPush(1)
		if !ok {
			t.Fatal("expected ok=true")
		}
		if old != 3 {
			t.Errorf("got %d, want 3", old)
		}
		want := []int{1, 5, 7}
		if diff := cmp.Diff(want, drainAll(t, h)); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("equal priority", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{3, 5, 7})
		old, ok := h.PopPush(3)
		if !ok {
			t.Fatal("expected ok=true")
		}
		if old != 3 {
			t.Errorf("got %d, want 3", old)
		}
		want := []int{3, 5, 7}
		if diff := cmp.Diff(want, drainAll(t, h)); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("single element", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{10})
		old, ok := h.PopPush(5)
		if !ok {
			t.Fatal("expected ok=true")
		}
		if old != 10 {
			t.Errorf("got %d, want 10", old)
		}
		val, ok := h.Pop()
		if !ok {
			t.Fatal("expected ok=true")
		}
		if val != 5 {
			t.Errorf("got %d, want 5", val)
		}
	})

	t.Run("equivalence to Pop then Push", func(t *testing.T) {
		t.Parallel()
		items := []int{5, 3, 7, 1, 4, 2, 6}
		val := 4

		ppHeap := heap.NewMinFrom(items)
		gotPP, _ := ppHeap.PopPush(val)

		sepHeap := heap.NewMinFrom(items)
		gotSep, _ := sepHeap.Pop()
		sepHeap.Push(val)

		if gotPP != gotSep {
			t.Errorf("PopPush=%d, Pop+Push=%d", gotPP, gotSep)
		}

		wantPP := drainAll(t, ppHeap)
		wantSep := drainAll(t, sepHeap)
		if diff := cmp.Diff(wantSep, wantPP); diff != "" {
			t.Errorf("remaining heap mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("consecutive operations", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{3, 5, 7})
		var results []int
		for _, v := range []int{1, 10, 4, 2} {
			old, ok := h.PopPush(v)
			if !ok {
				t.Fatal("expected ok=true")
			}
			results = append(results, old)
		}
		want := []int{3, 1, 5, 4}
		if diff := cmp.Diff(want, results); diff != "" {
			t.Errorf("results mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("max heap", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMaxFrom([]int{7, 5, 3})

		old, ok := h.PopPush(10)
		if !ok {
			t.Fatal("expected ok=true")
		}
		if old != 7 {
			t.Errorf("got %d, want 7", old)
		}

		old, ok = h.PopPush(1)
		if !ok {
			t.Fatal("expected ok=true")
		}
		if old != 10 {
			t.Errorf("got %d, want 10", old)
		}
	})
}

func TestMerge(t *testing.T) {
	t.Parallel()

	t.Run("basic merge", func(t *testing.T) {
		t.Parallel()
		h1 := heap.NewMinFrom([]int{1, 3, 5})
		h2 := heap.NewMinFrom([]int{2, 4, 6})
		h1.Merge(h2)
		got := drainAll(t, h1)
		want := []int{1, 2, 3, 4, 5, 6}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("other emptied", func(t *testing.T) {
		t.Parallel()
		h1 := heap.NewMinFrom([]int{1, 3})
		h2 := heap.NewMinFrom([]int{2, 4})
		h1.Merge(h2)
		if h2.Len() != 0 {
			t.Errorf("other len=%d, want 0", h2.Len())
		}
	})

	t.Run("merge into empty", func(t *testing.T) {
		t.Parallel()
		h1 := heap.NewMin[int]()
		h2 := heap.NewMinFrom([]int{3, 1, 2})
		h1.Merge(h2)
		got := drainAll(t, h1)
		want := []int{1, 2, 3}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("merge empty", func(t *testing.T) {
		t.Parallel()
		h1 := heap.NewMinFrom([]int{1, 2, 3})
		h2 := heap.NewMin[int]()
		h1.Merge(h2)
		got := drainAll(t, h1)
		want := []int{1, 2, 3}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("self merge", func(t *testing.T) {
		t.Parallel()
		h := heap.NewMinFrom([]int{1, 2, 3})
		h.Merge(h)
		got := drainAll(t, h)
		want := []int{1, 2, 3}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("push after merge", func(t *testing.T) {
		t.Parallel()
		h1 := heap.NewMinFrom([]int{3, 5})
		h2 := heap.NewMinFrom([]int{2, 4})
		h1.Merge(h2)
		h1.Push(1)
		got := drainAll(t, h1)
		want := []int{1, 2, 3, 4, 5}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("max heap", func(t *testing.T) {
		t.Parallel()
		h1 := heap.NewMaxFrom([]int{7, 5, 3})
		h2 := heap.NewMaxFrom([]int{6, 4, 2})
		h1.Merge(h2)
		got := drainAll(t, h1)
		want := []int{7, 6, 5, 4, 3, 2}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("small merge into large heap", func(t *testing.T) {
		t.Parallel()
		less := func(a, b int) bool { return a < b }
		base := make([]int, 1000)
		for i := range base {
			base[i] = i + 10
		}
		h1 := heap.NewMinFrom(base)
		h2 := heap.NewMinFrom([]int{3, 1, 7})
		h1.Merge(h2)
		assertHeapInvariant(t, h1, less)
		if h1.Len() != 1003 {
			t.Fatalf("len=%d, want 1003", h1.Len())
		}
		if h2.Len() != 0 {
			t.Errorf("other len=%d, want 0", h2.Len())
		}
		if v, _ := h1.Peek(); v != 1 {
			t.Errorf("peek=%d, want 1", v)
		}
	})

	t.Run("large merge into small heap", func(t *testing.T) {
		t.Parallel()
		less := func(a, b int) bool { return a < b }
		h1 := heap.NewMinFrom([]int{1, 5, 3})
		extra := make([]int, 500)
		for i := range extra {
			extra[i] = i + 10
		}
		h2 := heap.NewMinFrom(extra)
		h1.Merge(h2)
		assertHeapInvariant(t, h1, less)
		if h1.Len() != 503 {
			t.Fatalf("len=%d, want 503", h1.Len())
		}
		if h2.Len() != 0 {
			t.Errorf("other len=%d, want 0", h2.Len())
		}
	})
}

func TestHeapInvariantAfterOperations(t *testing.T) {
	t.Parallel()

	intLess := func(a, b int) bool { return a < b }
	rng := rand.New(rand.NewPCG(42, 0))
	h := heap.New(intLess)

	for op := range 10_000 {
		switch rng.IntN(7) {
		case 0:
			h.Push(rng.IntN(1000))
		case 1:
			h.Pop()
		case 2:
			h.PushPop(rng.IntN(1000))
		case 3:
			h.PopPush(rng.IntN(1000))
		case 4:
			target := rng.IntN(1000)
			h.Remove(func(v int) bool { return v == target })
		case 5:
			items := make([]int, rng.IntN(10))
			for i := range items {
				items[i] = rng.IntN(1000)
			}
			h.Extend(items)
		case 6:
			other := heap.New(intLess)
			for range rng.IntN(5) {
				other.Push(rng.IntN(1000))
			}
			h.Merge(other)
		}

		if (op+1)%1_000 == 0 {
			assertHeapInvariant(t, h, intLess)
		}
	}

	assertHeapInvariant(t, h, intLess)
}

func TestLessPanicRecovery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		op   func(h *heap.Heap[int])
	}{
		{
			name: "Push",
			op:   func(h *heap.Heap[int]) { h.Push(0) },
		},
		{
			name: "Pop",
			op:   func(h *heap.Heap[int]) { h.Pop() },
		},
		{
			name: "PushPop",
			op:   func(h *heap.Heap[int]) { h.PushPop(0) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var comparisons int
			armed := false
			h := heap.New(func(a, b int) bool {
				if armed {
					comparisons++
					if comparisons >= 3 {
						panic("boom")
					}
				}
				return a < b
			})
			for i := range 10 {
				h.Push(i)
			}

			// Arm the panic and trigger it during the operation.
			armed = true
			func() {
				defer func() { _ = recover() }()
				tt.op(h)
			}()

			// Heap should still be usable: Len should return without
			// panicking and the count should be reasonable.
			n := h.Len()
			if n < 1 {
				t.Fatalf("Len()=%d after recovered panic, want >= 1", n)
			}

			// Subsequent operations with a well-behaved less should
			// not panic.
			safe := heap.NewMin[int]()
			for v := range h.All() {
				safe.Push(v)
			}
			if safe.Len() != n {
				t.Fatalf("safe heap len=%d, want %d", safe.Len(), n)
			}
		})
	}
}
