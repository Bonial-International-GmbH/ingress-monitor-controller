package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		options *Options
		valid   bool
	}{
		{
			name:    "default options are always valid",
			options: NewDefaultOptions(),
			valid:   true,
		},
		{
			name: "provider name must not be empty",
			options: func() *Options {
				o := NewDefaultOptions()
				o.ProviderName = ""
				return o
			}(),
			valid: false,
		},
		{
			name: "name template must not be empty",
			options: func() *Options {
				o := NewDefaultOptions()
				o.NameTemplate = ""
				return o
			}(),
			valid: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.options.Validate()
			if test.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
