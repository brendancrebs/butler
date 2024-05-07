// Copyright (c) 2023 - 2024 Schweitzer Engineering Laboratories, Inc.
// SEL Confidential

package butler_test

import (
	"testing"

	"selinc.com/butler/code/butler"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_difference(t *testing.T) {
	tests := []struct {
		desc     string
		a, b     []string
		expected []string
	}{
		{
			desc:     "both empty",
			a:        []string{},
			b:        []string{},
			expected: []string(nil),
		},
		{
			desc:     "a is empty",
			a:        []string{},
			b:        []string{"str1", "str2"},
			expected: []string(nil),
		},
		{
			desc:     "b is empty",
			a:        []string{"str1", "str2"},
			b:        []string{},
			expected: []string{"str1", "str2"},
		},
		{
			desc:     "no common elements",
			a:        []string{"str1", "str2"},
			b:        []string{"str3", "str4"},
			expected: []string{"str1", "str2"},
		},
		{
			desc:     "all elements in a are in b",
			a:        []string{"str1", "str2"},
			b:        []string{"str1", "str2", "str3"},
			expected: []string(nil),
		},
		{
			desc:     "some elements in a are in b",
			a:        []string{"str1", "str2", "str3"},
			b:        []string{"str2", "str3"},
			expected: []string{"str1"},
		},
		{
			desc:     "duplicates in a and b",
			a:        []string{"str1", "str1", "str2"},
			b:        []string{"str2", "str2"},
			expected: []string{"str1", "str1"},
		},
	}

	for _, test := range tests {
		Convey(test.desc, t, func() {
			So(butler.Difference(test.a, test.b), ShouldEqual, test.expected)
		})
	}
}
