package main

import "fmt"
import "golang.org/x/tour/tree"

// Walk walks the tree t sending all values
// from the tree to the channel ch.
func Walk(t *tree.Tree, ch chan int) {
	innerWalk(t, ch)
	close(ch)
}

func innerWalk(t *tree.Tree, ch chan int) {
	if t.Left != nil {
		innerWalk(t.Left, ch)
	}
	ch <- t.Value
	if t.Right != nil {
		innerWalk(t.Right, ch)
	}
}

// Same determines whether the trees
// t1 and t2 contain the same values.
func Same(t1, t2 *tree.Tree) bool {
	ch1 := make(chan int)
	ch2 := make(chan int)

	go Walk(t1, ch1)
	go Walk(t2, ch2)

	for elem1 := range ch1 {
		elem2, ok := <-ch2
		if !ok {
			return false
		}
		if elem1 != elem2 {
			return false
		}
	}

	// make sure ch1 is closed
	_, ok := <-ch1
	return !ok
}

func main() {
	fmt.Printf("true: %v\n", Same(tree.New(1), tree.New(1)))
	fmt.Printf("false: %v\n", Same(tree.New(1), tree.New(2)))
}
