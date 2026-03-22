package heap_test

import (
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jesalx/heap"
)

// fuzzReader treats a byte slice as a stream of random decisions.
type fuzzReader struct {
	data []byte
	pos  int
}

func (r *fuzzReader) next() (byte, bool) {
	if r.pos >= len(r.data) {
		return 0, false
	}
	b := r.data[r.pos]
	r.pos++
	return b, true
}

func (r *fuzzReader) nextInt() (int, bool) {
	b, ok := r.next()
	return int(b), ok
}

// FuzzHeapInvariant drives random operation sequences and checks the heap
// invariant after every operation. This generalises the fixed-seed
// TestHeapInvariantAfterOperations with corpus-guided exploration.
func FuzzHeapInvariant(f *testing.F) {
	f.Add([]byte{0, 42, 1, 2, 99, 3, 50, 4, 7, 5, 10, 20, 30, 6, 3, 15})
	f.Add([]byte{})
	f.Add([]byte{1, 1, 1, 1, 1, 1})
	f.Add([]byte{0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1})

	intLess := func(a, b int) bool { return a < b }

	f.Fuzz(func(t *testing.T, data []byte) {
		r := &fuzzReader{data: data}
		h := heap.New(intLess)

		for {
			op, ok := r.next()
			if !ok {
				break
			}

			switch op % 7 {
			case 0: // Push
				val, ok := r.nextInt()
				if !ok {
					continue
				}
				h.Push(val)

			case 1: // Pop
				v, ok := h.Pop()
				if !ok {
					if h.Len() != 0 {
						t.Fatal("Pop returned false on non-empty heap")
					}
					continue
				}
				_ = v

			case 2: // PushPop
				val, ok := r.nextInt()
				if !ok {
					continue
				}
				h.PushPop(val)

			case 3: // PopPush
				val, ok := r.nextInt()
				if !ok {
					continue
				}
				h.PopPush(val)

			case 4: // Remove
				val, ok := r.nextInt()
				if !ok {
					continue
				}
				h.Remove(func(v int) bool { return v == val })

			case 5: // Extend
				count, ok := r.nextInt()
				if !ok {
					continue
				}
				count = count%8 + 1
				items := make([]int, 0, count)
				for range count {
					v, ok := r.nextInt()
					if !ok {
						break
					}
					items = append(items, v)
				}
				h.Extend(items)

			case 6: // Merge
				count, ok := r.nextInt()
				if !ok {
					continue
				}
				count = count%5 + 1
				other := heap.New(intLess)
				for range count {
					v, ok := r.nextInt()
					if !ok {
						break
					}
					other.Push(v)
				}
				h.Merge(other)
				if other.Len() != 0 {
					t.Fatal("other heap not empty after Merge")
				}
			}

			assertHeapInvariant(t, h, intLess)
		}
	})
}

// FuzzPopOrdering pushes fuzzed values and verifies they pop in sorted order.
func FuzzPopOrdering(f *testing.F) {
	f.Add([]byte{5, 3, 7, 1, 4, 2, 6})
	f.Add([]byte{1, 1, 1, 1})
	f.Add([]byte{})
	f.Add([]byte{42})

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}

		h := heap.NewMin[byte]()
		for _, b := range data {
			h.Push(b)
		}

		if h.Len() != len(data) {
			t.Fatalf("len=%d, want %d", h.Len(), len(data))
		}

		got := drainAll(t, h)
		want := slices.Clone(data)
		slices.Sort(want)

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("drain order mismatch (-want +got):\n%s", diff)
		}
	})
}

// FuzzPushPopEquivalence verifies PushPop(v) behaves identically to Push(v)
// followed by Pop().
func FuzzPushPopEquivalence(f *testing.F) {
	f.Add([]byte{5, 3, 7}, byte(4))
	f.Add([]byte{5, 3, 7}, byte(1))
	f.Add([]byte{5, 3, 7}, byte(10))
	f.Add([]byte{}, byte(5))
	f.Add([]byte{1}, byte(1))

	f.Fuzz(func(t *testing.T, items []byte, val byte) {
		// PushPop path
		pp := heap.NewMinFrom(items)
		gotPP := pp.PushPop(val)

		// Separate Push then Pop path
		sep := heap.NewMinFrom(items)
		sep.Push(val)
		gotSep, ok := sep.Pop()
		if !ok {
			t.Fatal("Pop returned false after Push on non-empty heap")
		}

		if gotPP != gotSep {
			t.Errorf("PushPop(%d)=%d, Push+Pop=%d", val, gotPP, gotSep)
		}

		remainPP := drainAll(t, pp)
		remainSep := drainAll(t, sep)
		if diff := cmp.Diff(remainSep, remainPP); diff != "" {
			t.Errorf("remaining heap mismatch (-want +got):\n%s", diff)
		}
	})
}

// FuzzPopPushEquivalence verifies PopPush(v) behaves identically to Pop()
// followed by Push(v).
func FuzzPopPushEquivalence(f *testing.F) {
	f.Add([]byte{5, 3, 7}, byte(4))
	f.Add([]byte{5, 3, 7}, byte(1))
	f.Add([]byte{5, 3, 7}, byte(10))
	f.Add([]byte{1}, byte(1))

	f.Fuzz(func(t *testing.T, items []byte, val byte) {
		if len(items) == 0 {
			return
		}

		// PopPush path
		pp := heap.NewMinFrom(items)
		gotPP, okPP := pp.PopPush(val)
		if !okPP {
			t.Fatal("PopPush returned false on non-empty heap")
		}

		// Separate Pop then Push path
		sep := heap.NewMinFrom(items)
		gotSep, okSep := sep.Pop()
		if !okSep {
			t.Fatal("Pop returned false on non-empty heap")
		}
		sep.Push(val)

		if gotPP != gotSep {
			t.Errorf("PopPush(%d)=%d, Pop+Push=%d", val, gotPP, gotSep)
		}

		remainPP := drainAll(t, pp)
		remainSep := drainAll(t, sep)
		if diff := cmp.Diff(remainSep, remainPP); diff != "" {
			t.Errorf("remaining heap mismatch (-want +got):\n%s", diff)
		}
	})
}

// FuzzNewFromEquivalence verifies NewFrom produces the same drain order as
// pushing elements individually.
func FuzzNewFromEquivalence(f *testing.F) {
	f.Add([]byte{5, 3, 7, 1, 4, 2, 6})
	f.Add([]byte{})
	f.Add([]byte{42})
	f.Add([]byte{1, 1, 1, 1})

	f.Fuzz(func(t *testing.T, items []byte) {
		fromHeap := heap.NewMinFrom(items)

		pushHeap := heap.NewMin[byte]()
		for _, v := range items {
			pushHeap.Push(v)
		}

		gotFrom := drainAll(t, fromHeap)
		gotPush := drainAll(t, pushHeap)
		if diff := cmp.Diff(gotPush, gotFrom); diff != "" {
			t.Errorf("NewFrom vs Push-one-by-one mismatch (-want +got):\n%s", diff)
		}
	})
}

// FuzzDrainEquivalence verifies Drain() returns the same elements in the same
// order as collecting Values().
func FuzzDrainEquivalence(f *testing.F) {
	f.Add([]byte{5, 3, 7, 1, 4, 2, 6})
	f.Add([]byte{})
	f.Add([]byte{42})

	f.Fuzz(func(t *testing.T, items []byte) {
		dHeap := heap.NewMinFrom(items)
		vHeap := heap.NewMinFrom(items)

		gotDrain := dHeap.Drain()
		gotValues := slices.Collect(vHeap.Values())

		if len(items) == 0 {
			if gotDrain != nil {
				t.Errorf("Drain on empty heap returned %v, want nil", gotDrain)
			}
			if len(gotValues) != 0 {
				t.Errorf("Values on empty heap returned %v, want empty", gotValues)
			}
			return
		}

		if diff := cmp.Diff(gotValues, gotDrain); diff != "" {
			t.Errorf("Drain vs Values mismatch (-want +got):\n%s", diff)
		}
	})
}
