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

func TestGetByKV(t *testing.T) {
	mc := newMemoryConfigCenter()

	mc.Save(Config{
		ID:      "rand0",
		Content: "{\"hello\": {\"world\": \"!\", \"me\": \"0\"}}",
	})
	mc.Save(Config{
		ID:      "rand1",
		Content: "{\"hello\": {\"world\": \"!\", \"me\": \"1\"}}",
	})

	c, _ := mc.GetByKV(map[string]string{"hello.world": "!", "hello.me": "1"}, "")

	fmt.Println(c)
}
