package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnnotations(t *testing.T) {
	a := Annotations{
		"a":           "somestring",
		"b":           "true",
		"c":           "42",
		"d":           "foo,bar,baz",
		"e":           `{"foo":"bar"}`,
		"invalidjson": `{invalidjson`,
	}

	// existent value
	assert.Equal(t, "somestring", a.StringValue("a"))
	assert.Equal(t, true, a.BoolValue("b"))
	assert.Equal(t, 42, a.IntValue("c"))
	assert.Equal(t, []string{"foo", "bar", "baz"}, a.StringSliceValue("d"))

	// default fallback
	assert.Equal(t, "thedefault", a.StringValue("nonexistent", "thedefault"))
	assert.Equal(t, true, a.BoolValue("nonexistent", true))
	assert.Equal(t, 2, a.IntValue("nonexistent", 2))
	assert.Equal(t, []string{"thedefault"}, a.StringSliceValue("nonexistent", []string{"thedefault"}))

	// unparsable bool/int
	assert.Equal(t, false, a.BoolValue("invalidjson"))
	assert.Equal(t, 0, a.IntValue("invalidjson"))

	// zero value
	assert.Equal(t, "", a.StringValue("nonexistent"))
	assert.Equal(t, false, a.BoolValue("nonexistent"))
	assert.Equal(t, 0, a.IntValue("nonexistent"))
	assert.Equal(t, []string(nil), a.StringSliceValue("nonexistent"))

	// json tests
	m := map[string]string{}
	require.NoError(t, a.ParseJSON("e", &m))
	assert.Equal(t, map[string]string{"foo": "bar"}, m)

	m = map[string]string{}
	require.NoError(t, a.ParseJSON("nonexistent", &m))
	assert.Equal(t, map[string]string{}, m)

	m = map[string]string{}
	require.Error(t, a.ParseJSON("invalidjson", &m))
}
