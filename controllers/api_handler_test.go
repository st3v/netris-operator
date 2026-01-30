/*
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

package controllers

import (
	"testing"

	"github.com/netrisai/netris-operator/netrisstorage"
	"github.com/netrisai/netriswebapi/v2/types/site"
)

// newTestStorage creates a Storage with pre-populated sites for testing.
func newTestStorage(sites []*site.Site) *netrisstorage.Storage {
	sitesStorage := netrisstorage.NewSitesStorage()
	sitesStorage.Sites = sites
	return &netrisstorage.Storage{
		SitesStorage: sitesStorage,
	}
}

func TestGetSites(t *testing.T) {
	tests := []struct {
		name     string
		sites    []*site.Site
		names    []string
		expected map[string]int
	}{
		{
			name: "finds all sites",
			sites: []*site.Site{
				{ID: 1, Name: "site-a"},
				{ID: 2, Name: "site-b"},
				{ID: 3, Name: "site-c"},
			},
			names: []string{"site-a", "site-b"},
			expected: map[string]int{
				"site-a": 1,
				"site-b": 2,
			},
		},
		{
			name: "missing sites return zero",
			sites: []*site.Site{
				{ID: 1, Name: "site-a"},
			},
			names: []string{"site-a", "site-missing"},
			expected: map[string]int{
				"site-a":       1,
				"site-missing": 0,
			},
		},
		{
			name:     "empty input returns empty map",
			sites:    []*site.Site{{ID: 1, Name: "site-a"}},
			names:    []string{},
			expected: map[string]int{},
		},
		{
			name:  "all sites missing returns zeros",
			sites: []*site.Site{},
			names: []string{"missing-a", "missing-b"},
			expected: map[string]int{
				"missing-a": 0,
				"missing-b": 0,
			},
		},
		{
			name: "duplicate names in input are deduplicated",
			sites: []*site.Site{
				{ID: 5, Name: "site-x"},
			},
			names: []string{"site-x", "site-x", "site-x"},
			expected: map[string]int{
				"site-x": 5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := newTestStorage(tt.sites)
			result := getSites(tt.names, storage)

			if len(result) != len(tt.expected) {
				t.Errorf("got %d entries, expected %d", len(result), len(tt.expected))
			}

			for name, expectedID := range tt.expected {
				if gotID, ok := result[name]; !ok {
					t.Errorf("missing key %q in result", name)
				} else if gotID != expectedID {
					t.Errorf("for site %q: got ID %d, expected %d", name, gotID, expectedID)
				}
			}
		})
	}
}
