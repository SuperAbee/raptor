package configcenter

import (
	"fmt"
	"testing"
)

func TestMemory(t *testing.T) {
	mc := newMemoryConfigCenter()
	mc.Save(Config{
		ID:      "1",
		Content: "11",
	})
	mc.Get("1")
	mc.OnChange("1", func(config Config) {
		fmt.Printf("OnChange: %v\n", config)
	})
	mc.Save(Config{
		ID:      "1",
		Content: "12",
	})
}
