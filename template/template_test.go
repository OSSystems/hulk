package template

import (
	"errors"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpandVariable(t *testing.T) {
	testCases := []struct {
		name           string
		content        string
		env            map[string]string
		expectedResult []string
		expectedError  error
	}{
		{
			"RequiredVariableValueFound",
			"{VARIABLE}",
			map[string]string{"VARIABLE": "value"},
			[]string{"value"},
			nil,
		},

		{
			"RequiredVariableValueNotFound",
			"{VARIABLE}",
			map[string]string{},
			nil,
			errors.New("No value for required variable"),
		},

		{
			"OptionalVariableValueFound",
			"{VARIABLE}?",
			map[string]string{"VARIABLE": "value"},
			[]string{"value"},
			nil,
		},

		{
			"OptionalVariableValueNotFound",
			"{VARIABLE}?",
			map[string]string{},
			nil,
			errors.New("No value for optional variable"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			list, err := Expand(tc.content, tc.env)

			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expectedResult, list)
		})
	}
}

func TestExpandArray(t *testing.T) {
	testCases := []struct {
		name           string
		content        string
		env            map[string]string
		expectedResult []string
		expectedError  error
	}{
		{
			"RequiredArrayValueFound",
			"{ARRAY[]}",
			map[string]string{"ARRAY": "value1 value2"},
			[]string{"value1", "value2"},
			nil,
		},

		{
			"RequiredArrayValueNotFound",
			"{ARRAY[]}",
			map[string]string{},
			nil,
			errors.New("No value for required variable"),
		},

		{
			"OptionalArrayValueFound",
			"{ARRAY[]}?",
			map[string]string{"ARRAY": "value1 value2"},
			[]string{"value1", "value2"},
			nil,
		},

		{
			"OptionalArrayValueNotFound",
			"{ARRAY[]}?",
			map[string]string{},
			nil,
			errors.New("No value for optional variable"),
		},

		{
			"CustomArraySeparator",
			"{ARRAY[,]}",
			map[string]string{"ARRAY": "value1,value2"},
			[]string{"value1", "value2"},
			nil,
		},

		{
			"DifferentArraySize",
			"{ARRAY1[]} {ARRAY2[]}",
			map[string]string{"ARRAY1": "value1 value2", "ARRAY2": "value3"},
			nil,
			errors.New("Array size differs: ARRAY2 (1 should be 2)"),
		},

		{
			"WithoutVariable",
			"VARIABLE",
			map[string]string{},
			[]string{"VARIABLE"},
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			list, err := Expand(tc.content, tc.env)

			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			sort.Strings(list)

			assert.Equal(t, tc.expectedResult, list)
		})
	}
}
