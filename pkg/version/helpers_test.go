// Copyright 2020 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package version

import (
	"testing"
)

func TestCompareIAMAwareVersionStrings(t *testing.T) {
	tests := []*struct {
		v1, v2          string
		expectedGreater bool
	}{
		{"v1", "v2", false},
		{"v2", "v1", true},
		{"v10", "v2", true},
		{"v1", "v2alpha1", true},
		{"v1", "v2beta1", true},
		{"v1alpha2", "v1alpha1", true},
		{"v1beta1", "v2alpha3", true},
		{"v1alpha10", "v1alpha2", true},
		{"v1beta10", "v1beta2", true},
		{"foo", "v1beta2", false},
		{"bar", "foo", false},
		{"version1", "version2", false},  // Non iam-like versions are sorted alphabetically
		{"version1", "version10", false}, // Non iam-like versions are sorted alphabetically
	}

	for _, tc := range tests {
		if e, a := tc.expectedGreater, CompareIAMAwareVersionStrings(tc.v1, tc.v2) > 0; e != a {
			if e {
				t.Errorf("expected %s to be greater than %s", tc.v1, tc.v2)
			} else {
				t.Errorf("expected %s to be less than than %s", tc.v1, tc.v2)
			}
		}
	}
}
