# heap

[![CI](https://github.com/jesalx/heap/actions/workflows/ci.yml/badge.svg)](https://github.com/jesalx/heap/actions/workflows/ci.yml)

A generic priority queue for Go, built on top of `container/heap`.

## Install

```
go get github.com/jesalx/heap
```

## Usage

```go
import "github.com/jesalx/heap"

// Min-heap for ordered types
h := heap.NewMin[int]()
h.Push(5)
h.Push(3)
h.Push(7)

val, ok := h.Pop() // 3, true

// Max-heap
h := heap.NewMax[int]()

// Custom comparator
h := heap.New(func(a, b Task) bool {
    return a.Priority < b.Priority
})

// Bulk initialization (O(n) via heap.Init)
h := heap.NewFrom([]int{5, 3, 7, 1}, func(a, b int) bool {
    return a < b
})
```

### Iterating

`Values` returns an iterator that yields elements in priority order. The
iterator is destructive — each element is popped from the heap.

```go
h := heap.NewFrom([]int{5, 3, 7, 1}, func(a, b int) bool { return a < b })
for v := range h.Values() {
    fmt.Println(v) // 1, 3, 5, 7
}
```

`All` returns a non-destructive iterator over elements in heap-internal order
(not sorted). Use it when you need to inspect elements without modifying the
heap.

### Draining

`Drain` removes all elements and returns them as a sorted slice.

```go
h := heap.NewMin[int]()
h.Push(5)
h.Push(1)
h.Push(3)
sorted := h.Drain() // [1, 3, 5]
```

