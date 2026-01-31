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
	"errors"
	"testing"

	k8sv1alpha1 "github.com/netrisai/netris-operator/api/v1alpha1"
	"github.com/netrisai/netriswebapi/v2/types/ipam"
	"github.com/netrisai/netriswebapi/v2/types/l4lb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var errMock = errors.New("mock error")

func TestL4lbCompareFieldsForNewMeta(t *testing.T) {
	tests := []struct {
		name     string
		l4lb     *k8sv1alpha1.L4LB
		l4lbMeta *k8sv1alpha1.L4LBMeta
		expected bool
	}{
		{
			name: "no changes",
			l4lb: &k8sv1alpha1.L4LB{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "false",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			l4lbMeta: &k8sv1alpha1.L4LBMeta{
				Spec: k8sv1alpha1.L4LBMetaSpec{
					L4LBCRGeneration: 1,
					Imported:         false,
					Reclaim:          false,
				},
			},
			expected: false,
		},
		{
			name: "generation changed",
			l4lb: &k8sv1alpha1.L4LB{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 2,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "false",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			l4lbMeta: &k8sv1alpha1.L4LBMeta{
				Spec: k8sv1alpha1.L4LBMetaSpec{
					L4LBCRGeneration: 1,
					Imported:         false,
					Reclaim:          false,
				},
			},
			expected: true,
		},
		{
			name: "import annotation changed",
			l4lb: &k8sv1alpha1.L4LB{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "true",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			l4lbMeta: &k8sv1alpha1.L4LBMeta{
				Spec: k8sv1alpha1.L4LBMetaSpec{
					L4LBCRGeneration: 1,
					Imported:         false,
					Reclaim:          false,
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := l4lbCompareFieldsForNewMeta(tt.l4lb, tt.l4lbMeta)
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestL4lbMustUpdateAnnotations(t *testing.T) {
	tests := []struct {
		name     string
		l4lb     *k8sv1alpha1.L4LB
		expected bool
	}{
		{
			name: "no annotations",
			l4lb: &k8sv1alpha1.L4LB{
				ObjectMeta: metav1.ObjectMeta{},
			},
			expected: true,
		},
		{
			name: "valid annotations",
			l4lb: &k8sv1alpha1.L4LB{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "false",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			expected: false,
		},
		{
			name: "missing import annotation",
			l4lb: &k8sv1alpha1.L4LB{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := l4lbMustUpdateAnnotations(tt.l4lb)
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestL4lbUpdateDefaultAnnotations(t *testing.T) {
	tests := []struct {
		name                  string
		annotations           map[string]string
		expectedImport        string
		expectedReclaimPolicy string
	}{
		{
			name:                  "empty annotations get defaults",
			annotations:           map[string]string{},
			expectedImport:        "false",
			expectedReclaimPolicy: "delete",
		},
		{
			name: "preserves import true",
			annotations: map[string]string{
				"resource.k8s.netris.ai/import": "true",
			},
			expectedImport:        "true",
			expectedReclaimPolicy: "delete",
		},
		{
			name: "preserves reclaim retain",
			annotations: map[string]string{
				"resource.k8s.netris.ai/reclaimPolicy": "retain",
			},
			expectedImport:        "false",
			expectedReclaimPolicy: "retain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l4lb := &k8sv1alpha1.L4LB{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tt.annotations,
				},
			}

			l4lbUpdateDefaultAnnotations(l4lb)

			if l4lb.GetAnnotations()["resource.k8s.netris.ai/import"] != tt.expectedImport {
				t.Errorf("import: got %q, expected %q",
					l4lb.GetAnnotations()["resource.k8s.netris.ai/import"], tt.expectedImport)
			}
			if l4lb.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"] != tt.expectedReclaimPolicy {
				t.Errorf("reclaimPolicy: got %q, expected %q",
					l4lb.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"], tt.expectedReclaimPolicy)
			}
		})
	}
}

func TestCompareL4LBMetaAPIL4LBBackend(t *testing.T) {
	tests := []struct {
		name            string
		l4lbMetaBackends []k8sv1alpha1.L4LBMetaBackend
		apiL4LBBackends []l4lb.LBBackend
		expected        bool
	}{
		{
			name:            "both empty",
			l4lbMetaBackends: []k8sv1alpha1.L4LBMetaBackend{},
			apiL4LBBackends: []l4lb.LBBackend{},
			expected:        true,
		},
		{
			name: "matching backends",
			l4lbMetaBackends: []k8sv1alpha1.L4LBMetaBackend{
				{IP: "10.0.0.1", Port: 8080},
				{IP: "10.0.0.2", Port: 8080},
			},
			apiL4LBBackends: []l4lb.LBBackend{
				{IP: "10.0.0.1", Port: "8080"},
				{IP: "10.0.0.2", Port: "8080"},
			},
			expected: true,
		},
		{
			name: "different backends",
			l4lbMetaBackends: []k8sv1alpha1.L4LBMetaBackend{
				{IP: "10.0.0.1", Port: 8080},
			},
			apiL4LBBackends: []l4lb.LBBackend{
				{IP: "10.0.0.2", Port: "8080"},
			},
			expected: false,
		},
		{
			name: "different count",
			l4lbMetaBackends: []k8sv1alpha1.L4LBMetaBackend{
				{IP: "10.0.0.1", Port: 8080},
				{IP: "10.0.0.2", Port: 8080},
			},
			apiL4LBBackends: []l4lb.LBBackend{
				{IP: "10.0.0.1", Port: "8080"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareL4LBMetaAPIL4LBBackend(tt.l4lbMetaBackends, tt.apiL4LBBackends)
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestCompareL4LBMetaAPIL4LBHealthCheck(t *testing.T) {
	tests := []struct {
		name               string
		l4lbMetaHealthCheck k8sv1alpha1.L4LBMetaHealthCheck
		apiL4LBHealthCheck l4lb.LBHealthCheck
		expected           bool
	}{
		{
			name:               "both empty",
			l4lbMetaHealthCheck: k8sv1alpha1.L4LBMetaHealthCheck{},
			apiL4LBHealthCheck: l4lb.LBHealthCheck{},
			expected:           true,
		},
		{
			name: "matching TCP health check",
			l4lbMetaHealthCheck: k8sv1alpha1.L4LBMetaHealthCheck{
				TCP: &k8sv1alpha1.L4LBMetaHealthCheckTCP{
					Timeout: "2000",
				},
			},
			apiL4LBHealthCheck: l4lb.LBHealthCheck{
				TCP: l4lb.LBHealthCheckTCP{
					Timeout: "2000",
				},
			},
			expected: true,
		},
		{
			name: "matching HTTP health check",
			l4lbMetaHealthCheck: k8sv1alpha1.L4LBMetaHealthCheck{
				HTTP: &k8sv1alpha1.L4LBMetaHealthCheckHTTP{
					Timeout:     "2000",
					RequestPath: "/health",
				},
			},
			apiL4LBHealthCheck: l4lb.LBHealthCheck{
				HTTP: l4lb.LBHealthCheckHTTP{
					Timeout:     "2000",
					RequestPath: "/health",
				},
			},
			expected: true,
		},
		{
			name: "different timeout",
			l4lbMetaHealthCheck: k8sv1alpha1.L4LBMetaHealthCheck{
				TCP: &k8sv1alpha1.L4LBMetaHealthCheckTCP{
					Timeout: "2000",
				},
			},
			apiL4LBHealthCheck: l4lb.LBHealthCheck{
				TCP: l4lb.LBHealthCheckTCP{
					Timeout: "3000",
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareL4LBMetaAPIL4LBHealthCheck(tt.l4lbMetaHealthCheck, tt.apiL4LBHealthCheck)
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestL4LBMetaToNetris(t *testing.T) {
	tests := []struct {
		name           string
		l4lbMeta       *k8sv1alpha1.L4LBMeta
		expectedName   string
		expectedPort   int
		expectedProto  string
		expectedAuto   bool
	}{
		{
			name: "basic TCP load balancer",
			l4lbMeta: &k8sv1alpha1.L4LBMeta{
				Spec: k8sv1alpha1.L4LBMetaSpec{
					L4LBName:    "test-lb",
					Protocol:    "TCP",
					Port:        80,
					Tenant:      1,
					SiteID:      1,
					SiteName:    "site-1",
					Automatic:   true,
					Status:      "enabled",
					HealthCheck: &k8sv1alpha1.L4LBMetaHealthCheck{},
					Backend: []k8sv1alpha1.L4LBMetaBackend{
						{IP: "10.0.0.1", Port: 8080},
						{IP: "10.0.0.2", Port: 8080},
					},
				},
			},
			expectedName:  "test-lb",
			expectedPort:  80,
			expectedProto: "TCP",
			expectedAuto:  true,
		},
		{
			name: "UDP load balancer with explicit IP",
			l4lbMeta: &k8sv1alpha1.L4LBMeta{
				Spec: k8sv1alpha1.L4LBMetaSpec{
					L4LBName:    "udp-lb",
					Protocol:    "UDP",
					Port:        53,
					IP:          "203.0.113.1",
					Automatic:   false,
					Tenant:      2,
					SiteID:      1,
					HealthCheck: &k8sv1alpha1.L4LBMetaHealthCheck{},
					Backend: []k8sv1alpha1.L4LBMetaBackend{
						{IP: "10.0.0.10", Port: 53},
					},
				},
			},
			expectedName:  "udp-lb",
			expectedPort:  53,
			expectedProto: "UDP",
			expectedAuto:  false,
		},
		{
			name: "load balancer with HTTP health check",
			l4lbMeta: &k8sv1alpha1.L4LBMeta{
				Spec: k8sv1alpha1.L4LBMetaSpec{
					L4LBName: "http-health-lb",
					Protocol: "TCP",
					Port:     443,
					Tenant:   1,
					SiteID:   1,
					HealthCheck: &k8sv1alpha1.L4LBMetaHealthCheck{
						HTTP: &k8sv1alpha1.L4LBMetaHealthCheckHTTP{
							RequestPath: "/health",
							Timeout:     "5000",
						},
					},
				},
			},
			expectedName:  "http-health-lb",
			expectedPort:  443,
			expectedProto: "TCP",
			expectedAuto:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := L4LBMetaToNetris(tt.l4lbMeta)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Name != tt.expectedName {
				t.Errorf("Name: got %q, expected %q", result.Name, tt.expectedName)
			}
			if result.Port != tt.expectedPort {
				t.Errorf("Port: got %d, expected %d", result.Port, tt.expectedPort)
			}
			if result.Protocol != tt.expectedProto {
				t.Errorf("Protocol: got %q, expected %q", result.Protocol, tt.expectedProto)
			}
			if result.Automatic != tt.expectedAuto {
				t.Errorf("Automatic: got %v, expected %v", result.Automatic, tt.expectedAuto)
			}
		})
	}
}

func TestL4LBMetaToNetrisUpdate(t *testing.T) {
	tests := []struct {
		name          string
		l4lbMeta      *k8sv1alpha1.L4LBMeta
		expectedName  string
		expectedPort  int
		expectedProto string
	}{
		{
			name: "basic update",
			l4lbMeta: &k8sv1alpha1.L4LBMeta{
				Spec: k8sv1alpha1.L4LBMetaSpec{
					L4LBName:    "update-lb",
					Protocol:    "TCP",
					Port:        8080,
					Tenant:      1,
					SiteID:      1,
					Status:      "enabled",
					HealthCheck: &k8sv1alpha1.L4LBMetaHealthCheck{},
				},
			},
			expectedName:  "update-lb",
			expectedPort:  8080,
			expectedProto: "TCP",
		},
		{
			name: "update with TCP health check",
			l4lbMeta: &k8sv1alpha1.L4LBMeta{
				Spec: k8sv1alpha1.L4LBMetaSpec{
					L4LBName: "tcp-health-lb",
					Protocol: "TCP",
					Port:     443,
					HealthCheck: &k8sv1alpha1.L4LBMetaHealthCheck{
						TCP: &k8sv1alpha1.L4LBMetaHealthCheckTCP{
							RequestPath: "/",
							Timeout:     "3000",
						},
					},
				},
			},
			expectedName:  "tcp-health-lb",
			expectedPort:  443,
			expectedProto: "TCP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := L4LBMetaToNetrisUpdate(tt.l4lbMeta)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Name != tt.expectedName {
				t.Errorf("Name: got %q, expected %q", result.Name, tt.expectedName)
			}
			if result.Port != tt.expectedPort {
				t.Errorf("Port: got %d, expected %d", result.Port, tt.expectedPort)
			}
			if result.Protocol != tt.expectedProto {
				t.Errorf("Protocol: got %q, expected %q", result.Protocol, tt.expectedProto)
			}
		})
	}
}

func TestCompareL4LBMetaAPIL4LB(t *testing.T) {
	tests := []struct {
		name      string
		l4lbMeta  *k8sv1alpha1.L4LBMeta
		apiL4LB   *l4lb.LoadBalancer
		wantMatch bool
	}{
		{
			name: "all fields match",
			l4lbMeta: &k8sv1alpha1.L4LBMeta{
				Spec: k8sv1alpha1.L4LBMetaSpec{
					L4LBName:  "test-lb",
					IP:        "10.0.0.1",
					Automatic: false,
					Port:      80,
					Protocol:  "TCP",
					SiteID:    1,
					Tenant:    2,
					Status:    "enabled",
					VPCID:     3,
					HealthCheck: &k8sv1alpha1.L4LBMetaHealthCheck{
						TCP: &k8sv1alpha1.L4LBMetaHealthCheckTCP{
							RequestPath: "/health",
							Timeout:     "3000",
						},
					},
					Backend: []k8sv1alpha1.L4LBMetaBackend{
						{IP: "192.168.1.1", Port: 8080},
					},
				},
			},
			apiL4LB: &l4lb.LoadBalancer{
				Name:      "test-lb",
				IP:        "10.0.0.1",
				Automatic: false,
				Port:      80,
				Protocol:  "TCP",
				Site:      l4lb.IDName{ID: 1},
				Tenant:    l4lb.IDName{ID: 2},
				Status:    "enabled",
				Vpc:       l4lb.IDName{ID: 3},
				HealthCheck: l4lb.LBHealthCheck{
					TCP: l4lb.LBHealthCheckTCP{
						RequestPath: "/health",
						Timeout:     "3000",
					},
				},
				BackendIPs: []l4lb.LBBackend{
					{IP: "192.168.1.1", Port: "8080"},
				},
			},
			wantMatch: true,
		},
		{
			name: "name mismatch",
			l4lbMeta: &k8sv1alpha1.L4LBMeta{
				Spec: k8sv1alpha1.L4LBMetaSpec{
					L4LBName:    "lb-a",
					HealthCheck: &k8sv1alpha1.L4LBMetaHealthCheck{},
				},
			},
			apiL4LB: &l4lb.LoadBalancer{
				Name: "lb-b",
			},
			wantMatch: false,
		},
		{
			name: "IP mismatch",
			l4lbMeta: &k8sv1alpha1.L4LBMeta{
				Spec: k8sv1alpha1.L4LBMetaSpec{
					L4LBName:    "test-lb",
					IP:          "10.0.0.1",
					HealthCheck: &k8sv1alpha1.L4LBMetaHealthCheck{},
				},
			},
			apiL4LB: &l4lb.LoadBalancer{
				Name: "test-lb",
				IP:   "10.0.0.2",
			},
			wantMatch: false,
		},
		{
			name: "automatic mismatch",
			l4lbMeta: &k8sv1alpha1.L4LBMeta{
				Spec: k8sv1alpha1.L4LBMetaSpec{
					L4LBName:    "test-lb",
					Automatic:   true,
					HealthCheck: &k8sv1alpha1.L4LBMetaHealthCheck{},
				},
			},
			apiL4LB: &l4lb.LoadBalancer{
				Name:      "test-lb",
				Automatic: false,
			},
			wantMatch: false,
		},
		{
			name: "port mismatch",
			l4lbMeta: &k8sv1alpha1.L4LBMeta{
				Spec: k8sv1alpha1.L4LBMetaSpec{
					L4LBName:    "test-lb",
					Port:        80,
					HealthCheck: &k8sv1alpha1.L4LBMetaHealthCheck{},
				},
			},
			apiL4LB: &l4lb.LoadBalancer{
				Name: "test-lb",
				Port: 443,
			},
			wantMatch: false,
		},
		{
			name: "protocol mismatch",
			l4lbMeta: &k8sv1alpha1.L4LBMeta{
				Spec: k8sv1alpha1.L4LBMetaSpec{
					L4LBName:    "test-lb",
					Protocol:    "TCP",
					HealthCheck: &k8sv1alpha1.L4LBMetaHealthCheck{},
				},
			},
			apiL4LB: &l4lb.LoadBalancer{
				Name:     "test-lb",
				Protocol: "UDP",
			},
			wantMatch: false,
		},
		{
			name: "site mismatch",
			l4lbMeta: &k8sv1alpha1.L4LBMeta{
				Spec: k8sv1alpha1.L4LBMetaSpec{
					L4LBName:    "test-lb",
					SiteID:      1,
					HealthCheck: &k8sv1alpha1.L4LBMetaHealthCheck{},
				},
			},
			apiL4LB: &l4lb.LoadBalancer{
				Name: "test-lb",
				Site: l4lb.IDName{ID: 2},
			},
			wantMatch: false,
		},
		{
			name: "tenant mismatch",
			l4lbMeta: &k8sv1alpha1.L4LBMeta{
				Spec: k8sv1alpha1.L4LBMetaSpec{
					L4LBName:    "test-lb",
					Tenant:      1,
					HealthCheck: &k8sv1alpha1.L4LBMetaHealthCheck{},
				},
			},
			apiL4LB: &l4lb.LoadBalancer{
				Name:   "test-lb",
				Tenant: l4lb.IDName{ID: 2},
			},
			wantMatch: false,
		},
		{
			name: "status mismatch",
			l4lbMeta: &k8sv1alpha1.L4LBMeta{
				Spec: k8sv1alpha1.L4LBMetaSpec{
					L4LBName:    "test-lb",
					Status:      "enabled",
					HealthCheck: &k8sv1alpha1.L4LBMetaHealthCheck{},
				},
			},
			apiL4LB: &l4lb.LoadBalancer{
				Name:   "test-lb",
				Status: "disabled",
			},
			wantMatch: false,
		},
		{
			name: "VPC mismatch",
			l4lbMeta: &k8sv1alpha1.L4LBMeta{
				Spec: k8sv1alpha1.L4LBMetaSpec{
					L4LBName:    "test-lb",
					VPCID:       1,
					HealthCheck: &k8sv1alpha1.L4LBMetaHealthCheck{},
				},
			},
			apiL4LB: &l4lb.LoadBalancer{
				Name: "test-lb",
				Vpc:  l4lb.IDName{ID: 2},
			},
			wantMatch: false,
		},
		{
			name: "health check mismatch",
			l4lbMeta: &k8sv1alpha1.L4LBMeta{
				Spec: k8sv1alpha1.L4LBMetaSpec{
					L4LBName: "test-lb",
					HealthCheck: &k8sv1alpha1.L4LBMetaHealthCheck{
						TCP: &k8sv1alpha1.L4LBMetaHealthCheckTCP{
							RequestPath: "/health",
							Timeout:     "3000",
						},
					},
				},
			},
			apiL4LB: &l4lb.LoadBalancer{
				Name: "test-lb",
				HealthCheck: l4lb.LBHealthCheck{
					TCP: l4lb.LBHealthCheckTCP{
						RequestPath: "/different",
						Timeout:     "3000",
					},
				},
			},
			wantMatch: false,
		},
		{
			name: "backend mismatch",
			l4lbMeta: &k8sv1alpha1.L4LBMeta{
				Spec: k8sv1alpha1.L4LBMetaSpec{
					L4LBName:    "test-lb",
					HealthCheck: &k8sv1alpha1.L4LBMetaHealthCheck{},
					Backend: []k8sv1alpha1.L4LBMetaBackend{
						{IP: "192.168.1.1", Port: 8080},
					},
				},
			},
			apiL4LB: &l4lb.LoadBalancer{
				Name: "test-lb",
				BackendIPs: []l4lb.LBBackend{
					{IP: "192.168.1.2", Port: "8080"},
				},
			},
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareL4LBMetaAPIL4LB(tt.l4lbMeta, tt.apiL4LB)
			if got != tt.wantMatch {
				t.Errorf("got %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestFindSiteByIP(t *testing.T) {
	tests := []struct {
		name       string
		ip         string
		ipamData   []*ipam.IPAM
		ipamErr    error
		wantSiteID int
		wantErr    bool
	}{
		{
			name: "finds site for IP in subnet",
			ip:   "10.0.1.50",
			ipamData: []*ipam.IPAM{
				{
					Prefix: "10.0.0.0/16",
					Children: []*ipam.IPAM{
						{
							Prefix: "10.0.1.0/24",
							Sites:  []ipam.IDName{{ID: 42, Name: "site-1"}},
						},
					},
				},
			},
			wantSiteID: 42,
			wantErr:    false,
		},
		{
			name: "IP not in any subnet",
			ip:   "192.168.1.1",
			ipamData: []*ipam.IPAM{
				{
					Prefix: "10.0.0.0/16",
					Children: []*ipam.IPAM{
						{
							Prefix: "10.0.1.0/24",
							Sites:  []ipam.IDName{{ID: 42, Name: "site-1"}},
						},
					},
				},
			},
			wantSiteID: 0,
			wantErr:    true,
		},
		{
			name: "subnet has no sites",
			ip:   "10.0.1.50",
			ipamData: []*ipam.IPAM{
				{
					Prefix: "10.0.0.0/16",
					Children: []*ipam.IPAM{
						{
							Prefix: "10.0.1.0/24",
							Sites:  []ipam.IDName{},
						},
					},
				},
			},
			wantSiteID: 0,
			wantErr:    true,
		},
		{
			name: "multiple subnets finds correct one",
			ip:   "10.0.2.100",
			ipamData: []*ipam.IPAM{
				{
					Prefix: "10.0.0.0/16",
					Children: []*ipam.IPAM{
						{
							Prefix: "10.0.1.0/24",
							Sites:  []ipam.IDName{{ID: 1, Name: "site-1"}},
						},
						{
							Prefix: "10.0.2.0/24",
							Sites:  []ipam.IDName{{ID: 2, Name: "site-2"}},
						},
					},
				},
			},
			wantSiteID: 2,
			wantErr:    false,
		},
		{
			name:       "IPAM client error",
			ip:         "10.0.1.50",
			ipamData:   nil,
			ipamErr:    errMock,
			wantSiteID: 0,
			wantErr:    true,
		},
		{
			name: "invalid CIDR in subnet",
			ip:   "10.0.1.50",
			ipamData: []*ipam.IPAM{
				{
					Prefix: "10.0.0.0/16",
					Children: []*ipam.IPAM{
						{
							Prefix: "invalid-cidr",
							Sites:  []ipam.IDName{{ID: 42, Name: "site-1"}},
						},
					},
				},
			},
			wantSiteID: 0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &L4LBReconciler{
				IPAMClient: &MockIPAMClient{
					Data:   tt.ipamData,
					GetErr: tt.ipamErr,
				},
			}

			siteID, err := r.findSiteByIP(tt.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if siteID != tt.wantSiteID {
				t.Errorf("siteID = %d, want %d", siteID, tt.wantSiteID)
			}
		})
	}
}
