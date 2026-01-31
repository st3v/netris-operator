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

func TestSwitchCompareFieldsForNewMeta(t *testing.T) {
	tests := []struct {
		name       string
		switchH    *k8sv1alpha1.Switch
		switchMeta *k8sv1alpha1.SwitchMeta
		expected   bool
	}{
		{
			name: "no changes",
			switchH: &k8sv1alpha1.Switch{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "false",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			switchMeta: &k8sv1alpha1.SwitchMeta{
				Spec: k8sv1alpha1.SwitchMetaSpec{
					SwitchCRGeneration: 1,
					Imported:           false,
					Reclaim:            false,
				},
			},
			expected: false,
		},
		{
			name: "generation changed",
			switchH: &k8sv1alpha1.Switch{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 2,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "false",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			switchMeta: &k8sv1alpha1.SwitchMeta{
				Spec: k8sv1alpha1.SwitchMetaSpec{
					SwitchCRGeneration: 1,
					Imported:           false,
					Reclaim:            false,
				},
			},
			expected: true,
		},
		{
			name: "import changed",
			switchH: &k8sv1alpha1.Switch{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "true",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			switchMeta: &k8sv1alpha1.SwitchMeta{
				Spec: k8sv1alpha1.SwitchMetaSpec{
					SwitchCRGeneration: 1,
					Imported:           false,
					Reclaim:            false,
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := switchCompareFieldsForNewMeta(tt.switchH, tt.switchMeta)
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestSwitchMustUpdateAnnotations(t *testing.T) {
	tests := []struct {
		name     string
		switchH  *k8sv1alpha1.Switch
		expected bool
	}{
		{
			name: "no annotations",
			switchH: &k8sv1alpha1.Switch{
				ObjectMeta: metav1.ObjectMeta{},
			},
			expected: true,
		},
		{
			name: "valid annotations",
			switchH: &k8sv1alpha1.Switch{
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
			name: "missing import",
			switchH: &k8sv1alpha1.Switch{
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
			result := switchMustUpdateAnnotations(tt.switchH)
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestSwitchUpdateDefaultAnnotations(t *testing.T) {
	tests := []struct {
		name                  string
		annotations           map[string]string
		expectedImport        string
		expectedReclaimPolicy string
	}{
		{
			name:                  "empty annotations",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switchH := &k8sv1alpha1.Switch{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tt.annotations,
				},
			}

			switchUpdateDefaultAnnotations(switchH)

			if switchH.GetAnnotations()["resource.k8s.netris.ai/import"] != tt.expectedImport {
				t.Errorf("import: got %q, expected %q",
					switchH.GetAnnotations()["resource.k8s.netris.ai/import"], tt.expectedImport)
			}
			if switchH.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"] != tt.expectedReclaimPolicy {
				t.Errorf("reclaimPolicy: got %q, expected %q",
					switchH.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"], tt.expectedReclaimPolicy)
			}
		})
	}
}
