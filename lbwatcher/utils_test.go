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

package lbwatcher

import (
	"testing"

	"github.com/netrisai/netris-operator/netrisstorage"
	"github.com/netrisai/netriswebapi/v2/types/ipam"
)

func TestFindSiteByIP(t *testing.T) {
	tests := []struct {
		name        string
		ip          string
		subnets     []*ipam.IPAM
		wantSite    string
		wantPrefix  string
		wantErr     bool
		errContains string
	}{
		{
			name: "IP found in subnet with site",
			ip:   "10.0.1.50",
			subnets: []*ipam.IPAM{
				{
					Prefix: "10.0.0.0/16",
					Children: []*ipam.IPAM{
						{
							Prefix: "10.0.1.0/24",
							Sites: []ipam.IDName{
								{ID: 1, Name: "dc1"},
							},
						},
					},
				},
			},
			wantSite:   "dc1",
			wantPrefix: "10.0.1.0/24",
			wantErr:    false,
		},
		{
			name: "IP not found in any subnet",
			ip:   "192.168.1.1",
			subnets: []*ipam.IPAM{
				{
					Prefix: "10.0.0.0/16",
					Children: []*ipam.IPAM{
						{
							Prefix: "10.0.1.0/24",
							Sites: []ipam.IDName{
								{ID: 1, Name: "dc1"},
							},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "There are no sites for specified IP address",
		},
		{
			name: "IP found in subnet without sites",
			ip:   "10.0.1.50",
			subnets: []*ipam.IPAM{
				{
					Prefix: "10.0.0.0/16",
					Children: []*ipam.IPAM{
						{
							Prefix: "10.0.1.0/24",
							Sites:  []ipam.IDName{}, // no sites
						},
					},
				},
			},
			wantErr:     true,
			errContains: "There are no sites for specified IP address",
		},
		{
			name:        "empty subnets list",
			ip:          "10.0.1.50",
			subnets:     []*ipam.IPAM{},
			wantErr:     true,
			errContains: "There are no sites for specified IP address",
		},
		{
			name: "multiple subnets, IP in second",
			ip:   "10.0.2.50",
			subnets: []*ipam.IPAM{
				{
					Prefix: "10.0.0.0/16",
					Children: []*ipam.IPAM{
						{
							Prefix: "10.0.1.0/24",
							Sites: []ipam.IDName{
								{ID: 1, Name: "dc1"},
							},
						},
						{
							Prefix: "10.0.2.0/24",
							Sites: []ipam.IDName{
								{ID: 2, Name: "dc2"},
							},
						},
					},
				},
			},
			wantSite:   "dc2",
			wantPrefix: "10.0.2.0/24",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subnetsStorage := netrisstorage.NewSubnetsStorage()
			subnetsStorage.Subnets = tt.subnets
			storage := &netrisstorage.Storage{
				SubnetsStorage: subnetsStorage,
			}

			watcher := &Watcher{
				NStorage: storage,
			}

			site, prefix, err := watcher.findSiteByIP(tt.ip)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if tt.errContains != "" && !containsSubstr(err.Error(), tt.errContains) {
					t.Errorf("error = %q, want containing %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if site.Name != tt.wantSite {
					t.Errorf("site.Name = %q, want %q", site.Name, tt.wantSite)
				}
				if prefix != tt.wantPrefix {
					t.Errorf("prefix = %q, want %q", prefix, tt.wantPrefix)
				}
			}
		})
	}
}

func containsSubstr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstrHelper(s, substr))
}

func containsSubstrHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
