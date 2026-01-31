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

	k8sv1alpha1 "github.com/netrisai/netris-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBgpCompareFieldsForNewMeta(t *testing.T) {
	tests := []struct {
		name     string
		bgp      *k8sv1alpha1.BGP
		bgpMeta  *k8sv1alpha1.BGPMeta
		expected bool
	}{
		{
			name: "no changes",
			bgp: &k8sv1alpha1.BGP{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "false",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			bgpMeta: &k8sv1alpha1.BGPMeta{
				Spec: k8sv1alpha1.BGPMetaSpec{
					BGPCRGeneration: 1,
					Imported:        false,
					Reclaim:         false,
				},
			},
			expected: false,
		},
		{
			name: "generation changed",
			bgp: &k8sv1alpha1.BGP{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 2,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "false",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			bgpMeta: &k8sv1alpha1.BGPMeta{
				Spec: k8sv1alpha1.BGPMetaSpec{
					BGPCRGeneration: 1,
					Imported:        false,
					Reclaim:         false,
				},
			},
			expected: true,
		},
		{
			name: "import annotation changed",
			bgp: &k8sv1alpha1.BGP{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "true",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			bgpMeta: &k8sv1alpha1.BGPMeta{
				Spec: k8sv1alpha1.BGPMetaSpec{
					BGPCRGeneration: 1,
					Imported:        false,
					Reclaim:         false,
				},
			},
			expected: true,
		},
		{
			name: "reclaim annotation changed",
			bgp: &k8sv1alpha1.BGP{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "false",
						"resource.k8s.netris.ai/reclaimPolicy": "retain",
					},
				},
			},
			bgpMeta: &k8sv1alpha1.BGPMeta{
				Spec: k8sv1alpha1.BGPMetaSpec{
					BGPCRGeneration: 1,
					Imported:        false,
					Reclaim:         false,
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bgpCompareFieldsForNewMeta(tt.bgp, tt.bgpMeta)
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestBgpMustUpdateAnnotations(t *testing.T) {
	tests := []struct {
		name     string
		bgp      *k8sv1alpha1.BGP
		expected bool
	}{
		{
			name: "no annotations",
			bgp: &k8sv1alpha1.BGP{
				ObjectMeta: metav1.ObjectMeta{},
			},
			expected: true,
		},
		{
			name: "valid annotations",
			bgp: &k8sv1alpha1.BGP{
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
			name: "import true reclaim retain",
			bgp: &k8sv1alpha1.BGP{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "true",
						"resource.k8s.netris.ai/reclaimPolicy": "retain",
					},
				},
			},
			expected: false,
		},
		{
			name: "missing import annotation",
			bgp: &k8sv1alpha1.BGP{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			expected: true,
		},
		{
			name: "missing reclaimPolicy annotation",
			bgp: &k8sv1alpha1.BGP{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import": "false",
					},
				},
			},
			expected: true,
		},
		{
			name: "invalid import value",
			bgp: &k8sv1alpha1.BGP{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "invalid",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			expected: true,
		},
		{
			name: "invalid reclaimPolicy value",
			bgp: &k8sv1alpha1.BGP{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "false",
						"resource.k8s.netris.ai/reclaimPolicy": "invalid",
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bgpMustUpdateAnnotations(tt.bgp)
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestBgpUpdateDefaultAnnotations(t *testing.T) {
	tests := []struct {
		name                   string
		annotations            map[string]string
		expectedImport         string
		expectedReclaimPolicy  string
	}{
		{
			name:                   "empty annotations get defaults",
			annotations:            map[string]string{},
			expectedImport:         "false",
			expectedReclaimPolicy:  "delete",
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
		{
			name: "invalid values get overwritten with defaults",
			annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "invalid",
				"resource.k8s.netris.ai/reclaimPolicy": "invalid",
			},
			expectedImport:        "false",
			expectedReclaimPolicy: "delete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bgp := &k8sv1alpha1.BGP{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tt.annotations,
				},
			}

			bgpUpdateDefaultAnnotations(bgp)

			if bgp.GetAnnotations()["resource.k8s.netris.ai/import"] != tt.expectedImport {
				t.Errorf("import: got %q, expected %q",
					bgp.GetAnnotations()["resource.k8s.netris.ai/import"], tt.expectedImport)
			}
			if bgp.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"] != tt.expectedReclaimPolicy {
				t.Errorf("reclaimPolicy: got %q, expected %q",
					bgp.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"], tt.expectedReclaimPolicy)
			}
		})
	}
}

func TestOptionalRouteMapID(t *testing.T) {
	tests := []struct {
		name        string
		value       int
		expectNil   bool
	}{
		{
			name:      "zero returns nil",
			value:     0,
			expectNil: true,
		},
		{
			name:      "negative returns nil",
			value:     -1,
			expectNil: true,
		},
		{
			name:      "positive value returns pointer",
			value:     42,
			expectNil: false,
		},
		{
			name:      "one returns pointer",
			value:     1,
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := optionalRouteMapID(tt.value)
			if tt.expectNil {
				if result != nil {
					t.Errorf("expected nil for value %d, got %v", tt.value, *result)
				}
			} else {
				if result == nil {
					t.Errorf("expected non-nil for value %d", tt.value)
				} else if *result != tt.value {
					t.Errorf("got %d, expected %d", *result, tt.value)
				}
			}
		})
	}
}
