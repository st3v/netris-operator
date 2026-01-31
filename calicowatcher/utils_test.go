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

package calicowatcher

import (
	"testing"

	"github.com/netrisai/netriswebapi/v2/types/ipam"
)

func TestFindIPAMByIP(t *testing.T) {
	tests := []struct {
		name        string
		ip          string
		subnets     []*ipam.IPAM
		wantPrefix  string
		wantErr     bool
		errContains string
	}{
		{
			name: "IP found in top-level subnet",
			ip:   "10.0.1.50",
			subnets: []*ipam.IPAM{
				{
					Prefix: "10.0.0.0/16",
				},
			},
			wantPrefix: "10.0.0.0/16",
			wantErr:    false,
		},
		{
			name: "IP found in child subnet",
			ip:   "10.0.1.50",
			subnets: []*ipam.IPAM{
				{
					Prefix: "10.0.0.0/16",
					Children: []*ipam.IPAM{
						{
							Prefix: "10.0.1.0/24",
						},
					},
				},
			},
			wantPrefix: "10.0.1.0/24",
			wantErr:    false,
		},
		{
			name: "IP found in deeply nested subnet",
			ip:   "10.0.1.50",
			subnets: []*ipam.IPAM{
				{
					Prefix: "10.0.0.0/8",
					Children: []*ipam.IPAM{
						{
							Prefix: "10.0.0.0/16",
							Children: []*ipam.IPAM{
								{
									Prefix: "10.0.1.0/24",
								},
							},
						},
					},
				},
			},
			wantPrefix: "10.0.1.0/24",
			wantErr:    false,
		},
		{
			name: "IP not found in any subnet",
			ip:   "192.168.1.1",
			subnets: []*ipam.IPAM{
				{
					Prefix: "10.0.0.0/16",
				},
			},
			wantErr:     true,
			errContains: "there are no subnet for specified IP address",
		},
		{
			name:        "empty subnets list",
			ip:          "10.0.1.50",
			subnets:     []*ipam.IPAM{},
			wantErr:     true,
			errContains: "there are no subnet for specified IP address",
		},
		{
			name: "multiple subnets, IP in second",
			ip:   "192.168.1.50",
			subnets: []*ipam.IPAM{
				{
					Prefix: "10.0.0.0/16",
				},
				{
					Prefix: "192.168.0.0/16",
					Children: []*ipam.IPAM{
						{
							Prefix: "192.168.1.0/24",
						},
					},
				},
			},
			wantPrefix: "192.168.1.0/24",
			wantErr:    false,
		},
		{
			name: "invalid subnet prefix",
			ip:   "10.0.1.50",
			subnets: []*ipam.IPAM{
				{
					Prefix: "invalid-prefix",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FindIPAMByIP(tt.ip, tt.subnets)

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
				if result == nil {
					t.Fatalf("expected result, got nil")
				}
				if result.Prefix != tt.wantPrefix {
					t.Errorf("prefix = %q, want %q", result.Prefix, tt.wantPrefix)
				}
			}
		})
	}
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
