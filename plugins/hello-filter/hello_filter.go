package main

import (
	"fmt"
	"raptor/proto"
)

type HelloFilter struct {

}

func (h *HelloFilter) filter(config proto.Config) error {
	fmt.Println("hello filter")
	return nil
}
