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
	assert.Equal(t, "somestring", a.String("a"))
	assert.Equal(t, true, a.Bool("b"))
	assert.Equal(t, 42, a.Int("c"))
	assert.Equal(t, []string{"foo", "bar", "baz"}, a.StringSlice("d"))

	// default fallback
	assert.Equal(t, "thedefault", a.String("nonexistent", "thedefault"))
	assert.Equal(t, true, a.Bool("nonexistent", true))
	assert.Equal(t, 2, a.Int("nonexistent", 2))
	assert.Equal(t, []string{"thedefault"}, a.StringSlice("nonexistent", []string{"thedefault"}))

	// unparsable bool/int
	assert.Equal(t, false, a.Bool("invalidjson"))
	assert.Equal(t, 0, a.Int("invalidjson"))

	// zero value
	assert.Equal(t, "", a.String("nonexistent"))
	assert.Equal(t, false, a.Bool("nonexistent"))
	assert.Equal(t, 0, a.Int("nonexistent"))
	assert.Equal(t, []string(nil), a.StringSlice("nonexistent"))

	// json tests
	m := map[string]string{}
	require.NoError(t, a.JSON("e", &m))
	assert.Equal(t, map[string]string{"foo": "bar"}, m)

	m = map[string]string{}
	require.NoError(t, a.JSON("nonexistent", &m))
	assert.Equal(t, map[string]string{}, m)

	m = map[string]string{}
	require.Error(t, a.JSON("invalidjson", &m))
}
