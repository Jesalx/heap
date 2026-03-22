package heap_test

import (
	"cmp"
	"fmt"
	"sort"

	"github.com/jesalx/heap"
)

func ExampleNewMin() {
	h := heap.NewMin[int]()
	h.Push(5)
	h.Push(3)
	h.Push(7)
	h.Push(1)

	for v := range h.Values() {
		fmt.Println(v)
	}
	// Output:
	// 1
	// 3
	// 5
	// 7
}

func ExampleNewMax() {
	h := heap.NewMax[int]()
	h.Push(5)
	h.Push(3)
	h.Push(7)
	h.Push(1)

	for v := range h.Values() {
		fmt.Println(v)
	}
	// Output:
	// 7
	// 5
	// 3
	// 1
}

func ExampleNew() {
	type Job struct {
		Name     string
		Priority int
	}

	h := heap.New(func(a, b Job) bool {
		return a.Priority < b.Priority
	})

	h.Push(Job{"email", 5})
	h.Push(Job{"backup", 10})
	h.Push(Job{"alert", 1})

	for v := range h.Values() {
		fmt.Println(v.Name)
	}
	// Output:
	// alert
	// email
	// backup
}

func ExampleNewFrom() {
	h := heap.NewFrom([]int{5, 3, 7, 1, 4}, func(a, b int) bool { return a < b })

	fmt.Println(h.Drain())
	// Output:
	// [1 3 4 5 7]
}

func ExampleNewMinFrom() {
	h := heap.NewMinFrom([]int{5, 3, 7, 1})

	for v := range h.Values() {
		fmt.Println(v)
	}
	// Output:
	// 1
	// 3
	// 5
	// 7
}

func ExampleNewMaxFrom() {
	h := heap.NewMaxFrom([]int{5, 3, 7, 1})

	for v := range h.Values() {
		fmt.Println(v)
	}
	// Output:
	// 7
	// 5
	// 3
	// 1
}

func ExampleHeap_PushPop() {
	h := heap.NewMinFrom([]int{3, 5, 7})

	// PushPop(1): 1 has higher priority than root (3), returned immediately.
	fmt.Println(h.PushPop(1))

	// PushPop(4): 4 replaces root (3), which is returned.
	fmt.Println(h.PushPop(4))

	fmt.Println(h.Drain())
	// Output:
	// 1
	// 3
	// [4 5 7]
}

func ExampleHeap_Contains() {
	h := heap.NewMinFrom([]int{2, 4, 6, 8})

	fmt.Println(h.Contains(func(v int) bool { return v == 4 }))
	fmt.Println(h.Contains(func(v int) bool { return v == 5 }))

	// Works with any predicate — find the first even number > 5.
	fmt.Println(h.Contains(func(v int) bool { return v > 5 && v%2 == 0 }))
	// Output:
	// true
	// false
	// true
}

func ExampleHeap_Values() {
	h := heap.NewMinFrom([]string{"cherry", "apple", "banana"})

	for v := range h.Values() {
		fmt.Println(v)
	}
	// Output:
	// apple
	// banana
	// cherry
}

func ExampleHeap_Drain() {
	h := heap.NewFrom([]int{5, 1, 3}, cmp.Less[int])

	sorted := h.Drain()
	fmt.Println(sorted)
	fmt.Println("empty:", h.Len() == 0)
	// Output:
	// [1 3 5]
	// empty: true
}

func ExampleHeap_Clone() {
	h := heap.NewMinFrom([]int{3, 1, 5})
	c := h.Clone()

	h.Push(0)

	fmt.Println("original:", h.Drain())
	fmt.Println("clone:", c.Drain())
	// Output:
	// original: [0 1 3 5]
	// clone: [1 3 5]
}

func ExampleHeap_Grow() {
	h := heap.NewMin[int]()
	h.Grow(100)

	h.Push(3)
	h.Push(1)
	h.Push(2)

	fmt.Println(h.Drain())
	// Output:
	// [1 2 3]
}

func ExampleHeap_All() {
	h := heap.NewMinFrom([]int{3, 1, 4, 1, 5})

	// Collect all elements (arbitrary order), then sort for display.
	var all []int
	for v := range h.All() {
		all = append(all, v)
	}
	sort.Ints(all)
	fmt.Println("all:", all)

	// The heap is unchanged — we can still pop from it.
	fmt.Println("len:", h.Len())
	// Output:
	// all: [1 1 3 4 5]
	// len: 5
}

func ExampleHeap_PopPush() {
	h := heap.NewMinFrom([]int{3, 5, 7})

	old, _ := h.PopPush(4)
	fmt.Println("popped:", old)
	fmt.Println("heap:", h.Drain())
	// Output:
	// popped: 3
	// heap: [4 5 7]
}

func ExampleHeap_Remove() {
	h := heap.NewMinFrom([]int{3, 1, 4, 1, 5})
	val, ok := h.Remove(func(v int) bool { return v == 4 })
	fmt.Println(val, ok)
	fmt.Println(h.Drain())
	// Output:
	// 4 true
	// [1 1 3 5]
}

func ExampleHeap_Extend() {
	h := heap.NewMinFrom([]int{3, 1})
	h.Extend([]int{5, 2, 4})
	fmt.Println(h.Drain())
	// Output:
	// [1 2 3 4 5]
}

func ExampleHeap_Merge() {
	h1 := heap.NewMinFrom([]int{1, 3, 5})
	h2 := heap.NewMinFrom([]int{2, 4, 6})

	h1.Merge(h2)
	fmt.Println("merged:", h1.Drain())
	fmt.Println("other empty:", h2.Len() == 0)
	// Output:
	// merged: [1 2 3 4 5 6]
	// other empty: true
}
