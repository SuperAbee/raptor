package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRegexp(t *testing.T) {
	s0 := soRegexp.FindStringSubmatch("hello_plugin.so")

	assert.Equal(t, 2, len(s0))
	assert.Equal(t, "hello_plugin.so", s0[0])
	assert.Equal(t, "hello_plugin", s0[1])

	s1 := soRegexp.FindStringSubmatch("hello_plugin_demo.so")

	assert.Equal(t, 2, len(s1))
	assert.Equal(t, "hello_plugin_demo.so", s1[0])
	assert.Equal(t, "hello_plugin_demo", s1[1])

	s2 := soRegexp.FindStringSubmatch("hello-plugin.so")
	assert.Equal(t, 0, len(s2))
}
