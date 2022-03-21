package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"plugin"
	"raptor/executor"
	"raptor/filter"
	"regexp"
	"strings"
)

var soRegexp = regexp.MustCompile(`^([\w_]+).so$`)

func init() {
	pwd, err := os.Getwd()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	err = filepath.Walk(pwd + "/plugins", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			ss := soRegexp.FindStringSubmatch(info.Name())
			// if info.Name() == "hello_plugin.so"
			// then ss == ["hello_plugin.so", "hello_plugin"]
			// we need plugin name ss[1]
			if len(ss) != 2 {
				return fmt.Errorf("no match string for plugin name with %s", soRegexp)
			}
			load(path, ss[1])
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}

}

func load(path, pluginName string) {
	plug, err := plugin.Open(path)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	log.Printf("%s: plugin opened\n", path)

	v, err := plug.Lookup(strings.ToUpper(pluginName))
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	f, ok := v.(filter.Filter)
	if ok {
		filter.Filters[pluginName] = f
	}
	e, ok := v.(executor.Executor)
	if ok {
		executor.Executors[pluginName] = e
	}
	if !ok {
		log.Printf("%s: unexpected type from module symbol\n", path)
	}
}
