package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getEnvName(t *testing.T) {
	type args struct {
		val string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "when value has pattern ${ENV_NAME} should return ENV_NAME",
			args: args{
				val: "${DB_PASS}",
			},
			want: "DB_PASS",
		},
		{
			name: "when value has pattern ${ENV_NAME} with space should return ENV_NAME",
			args: args{
				val: "${ DB_PASS } ",
			},
			want: "DB_PASS",
		},
		{
			name: "when value dont have pattern ${ENV_NAME} should return empty string",
			args: args{
				val: "DB_PASS",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getEnvName(tt.args.val); got != tt.want {
				t.Errorf("getEnvName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateConfig(t *testing.T) {
	type NestedConfig struct {
		Secret *string
	}

	type OptionalConfig struct {
		Detail string
	}

	type Config struct {
		Password string
		Nested   NestedConfig
		Optional *OptionalConfig
	}

	// Set environment variables for testing
	os.Setenv("DB_PASS", "password")
	os.Setenv("DB_SECRET", "supersecret")
	os.Setenv("DETAIL", "detailed_info")

	tests := []struct {
		name           string
		input          Config
		expectedOutput Config
	}{
		{
			name: "Simple substitution",
			input: Config{
				Password: "${DB_PASS}",
				Nested: NestedConfig{
					Secret: nil,
				},
				Optional: nil,
			},
			expectedOutput: Config{
				Password: "password",
				Nested: NestedConfig{
					Secret: nil,
				},
				Optional: nil,
			},
		},
		{
			name: "Nested substitution",
			input: Config{
				Password: "",
				Nested: NestedConfig{
					Secret: stringPointer("${DB_SECRET}"),
				},
				Optional: nil,
			},
			expectedOutput: Config{
				Password: "",
				Nested: NestedConfig{
					Secret: stringPointer("supersecret"),
				},
				Optional: nil,
			},
		},
		{
			name: "No substitution",
			input: Config{
				Password: "static_value",
				Nested: NestedConfig{
					Secret: stringPointer("static_secret"),
				},
				Optional: nil,
			},
			expectedOutput: Config{
				Password: "static_value",
				Nested: NestedConfig{
					Secret: stringPointer("static_secret"),
				},
				Optional: nil,
			},
		},
		{
			name: "Multiple substitutions",
			input: Config{
				Password: "${DB_PASS}",
				Nested: NestedConfig{
					Secret: stringPointer("${DB_SECRET}"),
				},
				Optional: &OptionalConfig{
					Detail: "${DETAIL}",
				},
			},
			expectedOutput: Config{
				Password: "password",
				Nested: NestedConfig{
					Secret: stringPointer("supersecret"),
				},
				Optional: &OptionalConfig{
					Detail: "detailed_info",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := UpdateConfig(&tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, tt.input)
		})
	}
}

// Helper function to create a string pointer
func stringPointer(s string) *string {
	return &s
}
