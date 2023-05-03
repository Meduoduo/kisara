package main

import (
	"fmt"

	"github.com/Yeuoly/kisara/src/helper"
)

func main() {
	testcase()
}

func testcase() {
	tree := helper.NewTree[string, string]()

	layer_1 := "layer-1"
	layer_2 := "layer-1-1"
	layer_3 := "layer-2"
	layer_4 := "layer-2-1"
	layer_5 := "layer-2-2"
	layer_6 := "layer-2-2-1"

	tree.AddNode("layer_1", &layer_1)
	tree.AddNode("layer_2", &layer_3)
	tree.AddToParent("layer_1", "layer_1_1", &layer_2)
	tree.AddToParent("layer_2", "layer_2_1", &layer_4)
	tree.AddToParent("layer_2", "layer_2_2", &layer_5)
	tree.AddToParent("layer_2_2", "layer_2_2_1", &layer_6)

	tree.WalkReverse(func(k string, v *string) {
		fmt.Printf("k: %s, v: %v\n", k, v)
	})
}
