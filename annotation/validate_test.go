package annotation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestInnerConfig2 struct {
	InnerField int `validate:"required,lt=100"`
}

type TestConfig2 struct {
	IntField       int               `validate:"required,gt=0"`
	UintField      uint              `validate:"required"`
	FloatField     float64           `validate:"required"`
	BoolField      bool              `validate:"required"`
	StringField    string            `validate:"required"`
	InnerConfig    TestInnerConfig2  `validate:"required"`
	InnerConfigPtr *TestInnerConfig2 `validate:"required"`
}

type ValidationStruct struct {
	RequiredField      int         `validate:"required"`
	LessThanField      int         `validate:"lt=10"`
	LessThanOrEqual    int         `validate:"lte=10"`
	GreaterThan        int         `validate:"gt=5"`
	GreaterThanOrEqual int         `validate:"gte=5"`
	LengthField        string      `validate:"len=5"`
	MinLengthField     string      `validate:"min_len=3"`
	MaxLengthField     string      `validate:"max_len=5"`
	NonBlankField      string      `validate:"non_blank"`
	PatternField       string      `validate:"pattern=^[a-zA-Z0-9]+$"`
	SliceField         []int       `validate:"min_len=2,max_len=5"`
	ArrayField         []int       `validate:"len=3"`
	NonBlankSliceField []string    `validate:"non_blank"`
	LessThanSliceField []int       `validate:"lt=10"`
	GreaterThanArray   []int       `validate:"gt=0"`
	ArrayField2        []int       `validate:"gt=0,min_len=1,max_len=2"`
	Config             TestConfig2 `validate:"required"`
}

func TestValidateStruct(t *testing.T) {
	tests := []struct {
		name      string
		input     ValidationStruct
		expectErr bool
	}{
		{
			name: "All valid",
			input: ValidationStruct{
				RequiredField:      1,
				LessThanField:      9,
				LessThanOrEqual:    10,
				GreaterThan:        6,
				GreaterThanOrEqual: 5,
				LengthField:        "abcde",
				MinLengthField:     "abcd",
				MaxLengthField:     "abc",
				NonBlankField:      "non_blank",
				PatternField:       "abc123",
				SliceField:         []int{1, 2},
				ArrayField:         []int{1, 2, 3},
				NonBlankSliceField: []string{"a", "b"},
				LessThanSliceField: []int{1, 2, 3},
				GreaterThanArray:   []int{1, 2, 3},
				ArrayField2:        []int{1},
				Config: TestConfig2{
					IntField:       1,
					UintField:      2,
					FloatField:     3.4,
					BoolField:      true,
					StringField:    "hello",
					InnerConfig:    TestInnerConfig2{InnerField: 99},
					InnerConfigPtr: &TestInnerConfig2{InnerField: 99},
				},
			},
			expectErr: false,
		},
		{
			name: "ArrayField2 invalid",
			input: ValidationStruct{
				RequiredField:      1,
				LessThanField:      9,
				LessThanOrEqual:    10,
				GreaterThan:        6,
				GreaterThanOrEqual: 5,
				LengthField:        "abcde",
				MinLengthField:     "abcd",
				MaxLengthField:     "abc",
				NonBlankField:      "non_blank",
				PatternField:       "abc123",
				SliceField:         []int{1, 2},
				ArrayField:         []int{1, 2, 3},
				NonBlankSliceField: []string{"a", "b"},
				LessThanSliceField: []int{1, 2, 3},
				GreaterThanArray:   []int{1, 2, 3},
				ArrayField2:        []int{-1, 0, 7},
				Config: TestConfig2{
					IntField:       1,
					UintField:      2,
					FloatField:     3.4,
					BoolField:      true,
					StringField:    "hello",
					InnerConfig:    TestInnerConfig2{InnerField: 99},
					InnerConfigPtr: &TestInnerConfig2{InnerField: 99},
				},
			},
			expectErr: true,
		},
		{
			name: "Required field missing",
			input: ValidationStruct{
				LessThanField:      9,
				LessThanOrEqual:    10,
				GreaterThan:        6,
				GreaterThanOrEqual: 5,
				LengthField:        "abcde",
				MinLengthField:     "abcd",
				MaxLengthField:     "abc",
				NonBlankField:      "non_blank",
				PatternField:       "abc123",
				SliceField:         []int{1, 2},
				ArrayField:         []int{1, 2, 3},
				NonBlankSliceField: []string{"a", "b"},
				LessThanSliceField: []int{1, 2, 3},
				GreaterThanArray:   []int{1, 2, 3},
				Config: TestConfig2{
					IntField:       1,
					UintField:      2,
					FloatField:     3.4,
					BoolField:      true,
					StringField:    "hello",
					InnerConfig:    TestInnerConfig2{InnerField: 99},
					InnerConfigPtr: &TestInnerConfig2{InnerField: 99},
				},
			},
			expectErr: true,
		},
		{
			name: "Less than field invalid",
			input: ValidationStruct{
				RequiredField:      1,
				LessThanField:      10,
				LessThanOrEqual:    10,
				GreaterThan:        6,
				GreaterThanOrEqual: 5,
				LengthField:        "abcde",
				MinLengthField:     "abcd",
				MaxLengthField:     "abc",
				NonBlankField:      "non_blank",
				PatternField:       "abc123",
				SliceField:         []int{1, 2},
				ArrayField:         []int{1, 2, 3},
				NonBlankSliceField: []string{"a", "b"},
				LessThanSliceField: []int{1, 2, 3},
				GreaterThanArray:   []int{1, 2, 3},
				Config: TestConfig2{
					IntField:       1,
					UintField:      2,
					FloatField:     3.4,
					BoolField:      true,
					StringField:    "hello",
					InnerConfig:    TestInnerConfig2{InnerField: 99},
					InnerConfigPtr: &TestInnerConfig2{InnerField: 99},
				},
			},
			expectErr: true,
		},
		{
			name: "Pattern field invalid",
			input: ValidationStruct{
				RequiredField:      1,
				LessThanField:      9,
				LessThanOrEqual:    10,
				GreaterThan:        6,
				GreaterThanOrEqual: 5,
				LengthField:        "abcde",
				MinLengthField:     "abcd",
				MaxLengthField:     "abc",
				NonBlankField:      "non_blank",
				PatternField:       "abc 123",
				SliceField:         []int{1, 2},
				ArrayField:         []int{1, 2, 3},
				NonBlankSliceField: []string{"a", "b"},
				LessThanSliceField: []int{1, 2, 3},
				GreaterThanArray:   []int{1, 2, 3},
				Config: TestConfig2{
					IntField:       1,
					UintField:      2,
					FloatField:     3.4,
					BoolField:      true,
					StringField:    "hello",
					InnerConfig:    TestInnerConfig2{InnerField: 99},
					InnerConfigPtr: &TestInnerConfig2{InnerField: 99},
				},
			},
			expectErr: true,
		},
		{
			name: "Slice field length invalid",
			input: ValidationStruct{
				RequiredField:      1,
				LessThanField:      9,
				LessThanOrEqual:    10,
				GreaterThan:        6,
				GreaterThanOrEqual: 5,
				LengthField:        "abcde",
				MinLengthField:     "abcd",
				MaxLengthField:     "abc",
				NonBlankField:      "non_blank",
				PatternField:       "abc123",
				SliceField:         []int{1},
				ArrayField:         []int{1, 2, 3},
				NonBlankSliceField: []string{"a", "b"},
				LessThanSliceField: []int{1, 2, 3},
				GreaterThanArray:   []int{1, 2, 3},
				Config: TestConfig2{
					IntField:       1,
					UintField:      2,
					FloatField:     3.4,
					BoolField:      true,
					StringField:    "hello",
					InnerConfig:    TestInnerConfig2{InnerField: 99},
					InnerConfigPtr: &TestInnerConfig2{InnerField: 99},
				},
			},
			expectErr: true,
		},
		{
			name: "Array field length invalid",
			input: ValidationStruct{
				RequiredField:      1,
				LessThanField:      9,
				LessThanOrEqual:    10,
				GreaterThan:        6,
				GreaterThanOrEqual: 5,
				LengthField:        "abcde",
				MinLengthField:     "abcd",
				MaxLengthField:     "abc",
				NonBlankField:      "non_blank",
				PatternField:       "abc123",
				SliceField:         []int{1, 2},
				ArrayField:         []int{1, 2},
				NonBlankSliceField: []string{"a", "b"},
				LessThanSliceField: []int{1, 2, 3},
				GreaterThanArray:   []int{1, 2, 3},
				Config: TestConfig2{
					IntField:       1,
					UintField:      2,
					FloatField:     3.4,
					BoolField:      true,
					StringField:    "hello",
					InnerConfig:    TestInnerConfig2{InnerField: 99},
					InnerConfigPtr: &TestInnerConfig2{InnerField: 99},
				},
			},
			expectErr: true,
		},
		{
			name: "Non-blank slice field invalid",
			input: ValidationStruct{
				RequiredField:      1,
				LessThanField:      9,
				LessThanOrEqual:    10,
				GreaterThan:        6,
				GreaterThanOrEqual: 5,
				LengthField:        "abcde",
				MinLengthField:     "abcd",
				MaxLengthField:     "abc",
				NonBlankField:      "non_blank",
				PatternField:       "abc123",
				SliceField:         []int{1, 2},
				ArrayField:         []int{1, 2, 3},
				NonBlankSliceField: []string{"a", ""},
				LessThanSliceField: []int{1, 2, 3},
				GreaterThanArray:   []int{1, 2, 3},
				Config: TestConfig2{
					IntField:       1,
					UintField:      2,
					FloatField:     3.4,
					BoolField:      true,
					StringField:    "hello",
					InnerConfig:    TestInnerConfig2{InnerField: 99},
					InnerConfigPtr: &TestInnerConfig2{InnerField: 99},
				},
			},
			expectErr: true,
		},
		{
			name: "Less than slice field invalid",
			input: ValidationStruct{
				RequiredField:      1,
				LessThanField:      9,
				LessThanOrEqual:    10,
				GreaterThan:        6,
				GreaterThanOrEqual: 5,
				LengthField:        "abcde",
				MinLengthField:     "abcd",
				MaxLengthField:     "abc",
				NonBlankField:      "non_blank",
				PatternField:       "abc123",
				SliceField:         []int{1, 2},
				ArrayField:         []int{1, 2, 3},
				NonBlankSliceField: []string{"a", "b"},
				LessThanSliceField: []int{1, 2, 11},
				GreaterThanArray:   []int{1, 2, 3},
				Config: TestConfig2{
					IntField:       1,
					UintField:      2,
					FloatField:     3.4,
					BoolField:      true,
					StringField:    "hello",
					InnerConfig:    TestInnerConfig2{InnerField: 99},
					InnerConfigPtr: &TestInnerConfig2{InnerField: 99},
				},
			},
			expectErr: true,
		},
		{
			name: "Greater than array field invalid",
			input: ValidationStruct{
				RequiredField:      1,
				LessThanField:      9,
				LessThanOrEqual:    10,
				GreaterThan:        6,
				GreaterThanOrEqual: 5,
				LengthField:        "abcde",
				MinLengthField:     "abcd",
				MaxLengthField:     "abc",
				NonBlankField:      "non_blank",
				PatternField:       "abc123",
				SliceField:         []int{1, 2},
				ArrayField:         []int{1, 2, 3},
				NonBlankSliceField: []string{"a", "b"},
				LessThanSliceField: []int{1, 2, 3},
				GreaterThanArray:   []int{0, 2, 3},
				Config: TestConfig2{
					IntField:       1,
					UintField:      2,
					FloatField:     3.4,
					BoolField:      true,
					StringField:    "hello",
					InnerConfig:    TestInnerConfig2{InnerField: 99},
					InnerConfigPtr: &TestInnerConfig2{InnerField: 99},
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(&tt.input)
			t.Logf("err: %v", err)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_matchString(t *testing.T) {
	tests := []struct {
		pattern string
		s       string
		want    bool
	}{
		// Test digit strings
		{"^[0-9]+$", "123456", true},
		{"^[0-9]+$", "123abc", false},
		{Digits, "123456", true},

		// Test character strings
		{"^[a-zA-Z]+$", "abcXYZ", true},
		{"^[a-zA-Z]+$", "abc123", false},
		{Alphabets, "abc123", false},

		// Test hexadecimal strings
		{"^[a-fA-F0-9]+$", "1a2B3C", true},
		{"^[a-fA-F0-9]+$", "1a2B3CZ", false},
		{Hex, "1a2B3CZ", false},

		// Test email strings
		{`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, "test.email@example.com", true},
		{`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, "test.email@example", false},
		{Email, "test.email@example", false},

		// Test alphanumeric strings
		{"^[a-zA-Z0-9]+$", "abc123XYZ", true},
		{"^[a-zA-Z0-9]+$", "abc 123", false},
		{Alphanumeric, "abc 123", false},

		// Test pattern with special characters
		{`^\w+@\w+\.\w+$`, "user@domain.com", true},
		{`^\w+@\w+\.\w+$`, "user@domain", false},

		// Test empty pattern and string
		{"", "", false},
		{"^$", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			if got := matchString(tt.pattern, tt.s); got != tt.want {
				t.Errorf("matchString(%q, %q) = %v, want %v", tt.pattern, tt.s, got, tt.want)
			}
		})
	}
}
