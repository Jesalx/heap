// Package heap provides a generic priority queue built on top of container/heap.
package heap

import (
	"cmp"
	iheap "container/heap"
	"iter"
	"math/bits"
	"slices"
)

type lessFunc[T any] func(a, b T) bool

type innerHeap[T any] struct {
	items []T
	less  lessFunc[T]
}

func (h innerHeap[T]) Len() int           { return len(h.items) }
func (h innerHeap[T]) Less(i, j int) bool { return h.less(h.items[i], h.items[j]) }
func (h innerHeap[T]) Swap(i, j int)      { h.items[i], h.items[j] = h.items[j], h.items[i] }

func (h *innerHeap[T]) Push(x any) { h.items = append(h.items, x.(T)) }
func (h *innerHeap[T]) Pop() any {
	old := h.items
	n := len(old)
	x := old[n-1]
	var zero T
	old[n-1] = zero
	h.items = old[:n-1]
	return x
}

// Heap is a generic priority queue. Elements are ordered by the less function
// provided at construction time.
type Heap[T any] struct {
	inner innerHeap[T]
}

// New creates an empty Heap using less to determine element priority.
// The less function should return true when a has higher priority than b.
// It panics if less is nil.
func New[T any](less func(a, b T) bool) *Heap[T] {
	if less == nil {
		panic("heap: less function must not be nil")
	}
	return &Heap[T]{inner: innerHeap[T]{less: less}}
}

// NewFrom creates a Heap pre-populated with items. The input slice is cloned,
// so the caller's data is not modified. The heap is constructed in O(n) time
// via heap.Init rather than individual pushes. It panics if less is nil.
func NewFrom[T any](items []T, less func(a, b T) bool) *Heap[T] {
	if less == nil {
		panic("heap: less function must not be nil")
	}
	h := &Heap[T]{inner: innerHeap[T]{
		items: slices.Clone(items),
		less:  less,
	}}
	iheap.Init(&h.inner)
	return h
}

// NewMin creates an empty min-heap for any ordered type.
// Elements are popped in ascending order.
func NewMin[T cmp.Ordered]() *Heap[T] {
	return New(cmp.Less[T])
}

// NewMax creates an empty max-heap for any ordered type.
// Elements are popped in descending order.
func NewMax[T cmp.Ordered]() *Heap[T] {
	return New(func(a, b T) bool { return cmp.Less(b, a) })
}

// NewMinFrom creates a min-heap pre-populated with items.
// Elements are popped in ascending order. The input slice is not modified.
func NewMinFrom[T cmp.Ordered](items []T) *Heap[T] {
	return NewFrom(items, cmp.Less[T])
}

// NewMaxFrom creates a max-heap pre-populated with items.
// Elements are popped in descending order. The input slice is not modified.
func NewMaxFrom[T cmp.Ordered](items []T) *Heap[T] {
	return NewFrom(items, func(a, b T) bool { return cmp.Less(b, a) })
}

// Push adds val to the heap.
func (h *Heap[T]) Push(val T) { iheap.Push(&h.inner, val) }

// Pop removes and returns the highest-priority element.
// It returns the zero value and false if the heap is empty.
func (h *Heap[T]) Pop() (T, bool) {
	if h.inner.Len() == 0 {
		var zero T
		return zero, false
	}
	return iheap.Pop(&h.inner).(T), true
}

// PushPop pushes val onto the heap and then pops and returns the
// highest-priority element. It is more efficient than a separate Push
// followed by Pop because it requires at most one sift-down instead of
// a sift-up plus a sift-down.
func (h *Heap[T]) PushPop(val T) T {
	if h.inner.Len() == 0 || h.inner.less(val, h.inner.items[0]) {
		return val
	}
	old := h.inner.items[0]
	h.inner.items[0] = val
	iheap.Fix(&h.inner, 0)
	return old
}

// PopPush pops the highest-priority element and then pushes val.
// It is more efficient than a separate Pop followed by Push because it
// requires at most one sift-down instead of a sift-up plus a sift-down.
// It returns the zero value and false if the heap is empty.
func (h *Heap[T]) PopPush(val T) (T, bool) {
	if h.inner.Len() == 0 {
		var zero T
		return zero, false
	}
	old := h.inner.items[0]
	h.inner.items[0] = val
	iheap.Fix(&h.inner, 0)
	return old, true
}

// Peek returns the highest-priority element without removing it.
// It returns the zero value and false if the heap is empty.
func (h *Heap[T]) Peek() (T, bool) {
	if h.inner.Len() == 0 {
		var zero T
		return zero, false
	}
	return h.inner.items[0], true
}

// Len returns the number of elements in the heap.
func (h *Heap[T]) Len() int { return h.inner.Len() }

// Cap returns the current capacity of the heap's backing array.
func (h *Heap[T]) Cap() int { return cap(h.inner.items) }

// Merge absorbs all elements from other into h. After Merge, other is empty.
// It is a no-op if other is empty or h and other are the same heap.
// It adaptively chooses between individual pushes and heap.Init based on the
// relative sizes of h and other.
func (h *Heap[T]) Merge(other *Heap[T]) {
	if h == other || other.inner.Len() == 0 {
		return
	}
	if h.inner.Len() == 0 {
		// Steal the already-heapified slice directly.
		h.inner.items = other.inner.items
	} else {
		h.absorb(other.inner.items)
	}
	other.inner.items = nil
}

// Contains reports whether the heap has any element for which fn returns true.
// It performs a linear scan over the underlying storage.
func (h *Heap[T]) Contains(fn func(T) bool) bool {
	return slices.ContainsFunc(h.inner.items, fn)
}

// Remove removes and returns the first element for which fn returns true.
// It performs a linear scan over the underlying storage. If no element
// matches, it returns the zero value and false.
func (h *Heap[T]) Remove(fn func(T) bool) (T, bool) {
	i := slices.IndexFunc(h.inner.items, fn)
	if i == -1 {
		var zero T
		return zero, false
	}
	return iheap.Remove(&h.inner, i).(T), true
}

// Extend adds all items to the heap. It adaptively chooses between individual
// pushes and heap.Init based on the relative sizes of the heap and items.
func (h *Heap[T]) Extend(items []T) {
	if len(items) == 0 {
		return
	}
	h.absorb(items)
}

// absorb adds extra elements into h, choosing between O(m log n) individual
// pushes and O(n+m) bulk append + heap.Init based on relative sizes.
func (h *Heap[T]) absorb(extra []T) {
	n := h.inner.Len()
	m := len(extra)
	// Individual pushes cost O(m log(n+m)); bulk reinit costs O(n+m).
	// bits.Len approximates log₂, so 2*m*bits.Len ≈ 2·m·log₂(n+m).
	// When that is less than n+m the per-element pushes are cheaper;
	// otherwise we append everything and rebuild the heap in one pass.
	if n > 0 && 2*m*bits.Len(uint(n+m)) < n+m {
		h.inner.items = slices.Grow(h.inner.items, m)
		for _, v := range extra {
			iheap.Push(&h.inner, v)
		}
	} else {
		h.inner.items = append(h.inner.items, extra...)
		iheap.Init(&h.inner)
	}
}

// Clear removes all elements from the heap, retaining the underlying
// storage for future pushes.
func (h *Heap[T]) Clear() {
	clear(h.inner.items)
	h.inner.items = h.inner.items[:0]
}

// Clone returns a shallow copy of the heap. The cloned heap is independent
// of the original; modifications to one do not affect the other.
func (h *Heap[T]) Clone() *Heap[T] {
	return &Heap[T]{inner: innerHeap[T]{
		items: slices.Clone(h.inner.items),
		less:  h.inner.less,
	}}
}

// Grow increases the heap's backing-array capacity by at least n additional
// slots. It does not change the length. Grow panics if n is negative.
func (h *Heap[T]) Grow(n int) {
	h.inner.items = slices.Grow(h.inner.items, n)
}

// All returns an iterator that yields every element in the heap.
// The iteration order is arbitrary (backing-array order) and not
// guaranteed to be in priority order. The iterator is non-destructive:
// the heap is not modified.
func (h *Heap[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range h.inner.items {
			if !yield(v) {
				return
			}
		}
	}
}

// Values returns an iterator that yields elements in priority order.
// The iterator is destructive: each yielded element is popped from the heap.
func (h *Heap[T]) Values() iter.Seq[T] {
	return func(yield func(T) bool) {
		for h.inner.Len() > 0 {
			if !yield(iheap.Pop(&h.inner).(T)) {
				return
			}
		}
	}
}

// Drain removes and returns all elements in priority order.
// It returns nil if the heap is empty.
func (h *Heap[T]) Drain() []T {
	n := h.inner.Len()
	if n == 0 {
		return nil
	}
	result := make([]T, 0, n)
	for h.inner.Len() > 0 {
		result = append(result, iheap.Pop(&h.inner).(T))
	}
	return result
}
