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

func TestNatCompareFieldsForNewMeta(t *testing.T) {
	tests := []struct {
		name     string
		nat      *k8sv1alpha1.Nat
		natMeta  *k8sv1alpha1.NatMeta
		expected bool
	}{
		{
			name: "no changes",
			nat: &k8sv1alpha1.Nat{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "false",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			natMeta: &k8sv1alpha1.NatMeta{
				Spec: k8sv1alpha1.NatMetaSpec{
					NatCRGeneration: 1,
					Imported:        false,
					Reclaim:         false,
				},
			},
			expected: false,
		},
		{
			name: "generation changed",
			nat: &k8sv1alpha1.Nat{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 2,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "false",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			natMeta: &k8sv1alpha1.NatMeta{
				Spec: k8sv1alpha1.NatMetaSpec{
					NatCRGeneration: 1,
					Imported:        false,
					Reclaim:         false,
				},
			},
			expected: true,
		},
		{
			name: "import annotation changed",
			nat: &k8sv1alpha1.Nat{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "true",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			natMeta: &k8sv1alpha1.NatMeta{
				Spec: k8sv1alpha1.NatMetaSpec{
					NatCRGeneration: 1,
					Imported:        false,
					Reclaim:         false,
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := natCompareFieldsForNewMeta(tt.nat, tt.natMeta)
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestNatMustUpdateAnnotations(t *testing.T) {
	tests := []struct {
		name     string
		nat      *k8sv1alpha1.Nat
		expected bool
	}{
		{
			name: "no annotations",
			nat: &k8sv1alpha1.Nat{
				ObjectMeta: metav1.ObjectMeta{},
			},
			expected: true,
		},
		{
			name: "valid annotations",
			nat: &k8sv1alpha1.Nat{
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
			nat: &k8sv1alpha1.Nat{
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
			result := natMustUpdateAnnotations(tt.nat)
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestNatUpdateDefaultAnnotations(t *testing.T) {
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
			nat := &k8sv1alpha1.Nat{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tt.annotations,
				},
			}

			natUpdateDefaultAnnotations(nat)

			if nat.GetAnnotations()["resource.k8s.netris.ai/import"] != tt.expectedImport {
				t.Errorf("import: got %q, expected %q",
					nat.GetAnnotations()["resource.k8s.netris.ai/import"], tt.expectedImport)
			}
			if nat.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"] != tt.expectedReclaimPolicy {
				t.Errorf("reclaimPolicy: got %q, expected %q",
					nat.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"], tt.expectedReclaimPolicy)
			}
		})
	}
}
