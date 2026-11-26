/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import "strings"

// LabelValue sanitizes an arbitrary string into a valid Kubernetes label value.
// Image tags may be digest-pinned references such as
// "v0.2.0@sha256:6ab9fa73...", which are not valid label values. The digest is
// dropped and the result is constrained to at most 63 characters of
// alphanumerics, '-', '_', '.' that start and end with an alphanumeric.
func LabelValue(s string) string {
	// drop any digest (everything from '@' onward)
	if i := strings.IndexByte(s, '@'); i >= 0 {
		s = s[:i]
	}

	// a bare digest or other reference may still carry a ':'; keep the first segment
	if i := strings.IndexByte(s, ':'); i >= 0 {
		s = s[:i]
	}

	// replace any remaining invalid characters
	s = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9',
			r == '-', r == '_', r == '.':
			return r
		default:
			return '-'
		}
	}, s)

	// label values are limited to 63 characters
	if len(s) > 63 {
		s = s[:63]
	}

	// must start and end with an alphanumeric
	return strings.TrimFunc(s, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9')
	})
}
