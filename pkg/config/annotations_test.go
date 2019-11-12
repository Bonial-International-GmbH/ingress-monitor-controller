package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnnotations(t *testing.T) {
	annotations := Annotations{
		"a":           "somestring",
		"b":           "true",
		"c":           "42",
		"d":           "foo,bar,baz",
		"e":           `{"foo":"bar"}`,
		"invalidjson": `{invalidjson`,
	}

	// existent value
	assert.Equal(t, "somestring", annotations.StringValue("a"))
	assert.Equal(t, true, annotations.BoolValue("b"))
	assert.Equal(t, 42, annotations.IntValue("c"))
	assert.Equal(t, []string{"foo", "bar", "baz"}, annotations.StringSliceValue("d"))

	// default fallback
	assert.Equal(t, "thedefault", annotations.StringValue("nonexistent", "thedefault"))
	assert.Equal(t, true, annotations.BoolValue("nonexistent", true))
	assert.Equal(t, 2, annotations.IntValue("nonexistent", 2))
	assert.Equal(t, []string{"thedefault"}, annotations.StringSliceValue("nonexistent", []string{"thedefault"}))

	// unparsable bool/int
	assert.Equal(t, false, annotations.BoolValue("invalidjson"))
	assert.Equal(t, 0, annotations.IntValue("invalidjson"))

	// zero value
	assert.Equal(t, "", annotations.StringValue("nonexistent"))
	assert.Equal(t, false, annotations.BoolValue("nonexistent"))
	assert.Equal(t, 0, annotations.IntValue("nonexistent"))
	assert.Equal(t, []string(nil), annotations.StringSliceValue("nonexistent"))

	// json tests
	dest := map[string]string{}
	require.NoError(t, annotations.ParseJSON("e", &dest))
	assert.Equal(t, map[string]string{"foo": "bar"}, dest)

	dest = map[string]string{}
	require.NoError(t, annotations.ParseJSON("nonexistent", &dest))
	assert.Equal(t, map[string]string{}, dest)

	dest = map[string]string{}
	require.Error(t, annotations.ParseJSON("invalidjson", &dest))
}
