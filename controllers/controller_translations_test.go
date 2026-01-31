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

func TestControllerCompareFieldsForNewMeta(t *testing.T) {
	tests := []struct {
		name           string
		controller     *k8sv1alpha1.Controller
		controllerMeta *k8sv1alpha1.ControllerMeta
		expected       bool
	}{
		{
			name: "no changes",
			controller: &k8sv1alpha1.Controller{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "false",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			controllerMeta: &k8sv1alpha1.ControllerMeta{
				Spec: k8sv1alpha1.ControllerMetaSpec{
					ControllerCRGeneration: 1,
					Imported:               false,
					Reclaim:                false,
				},
			},
			expected: false,
		},
		{
			name: "generation changed",
			controller: &k8sv1alpha1.Controller{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 2,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "false",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			controllerMeta: &k8sv1alpha1.ControllerMeta{
				Spec: k8sv1alpha1.ControllerMetaSpec{
					ControllerCRGeneration: 1,
					Imported:               false,
					Reclaim:                false,
				},
			},
			expected: true,
		},
		{
			name: "import changed",
			controller: &k8sv1alpha1.Controller{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "true",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			controllerMeta: &k8sv1alpha1.ControllerMeta{
				Spec: k8sv1alpha1.ControllerMetaSpec{
					ControllerCRGeneration: 1,
					Imported:               false,
					Reclaim:                false,
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := controllerCompareFieldsForNewMeta(tt.controller, tt.controllerMeta)
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestControllerMustUpdateAnnotations(t *testing.T) {
	tests := []struct {
		name       string
		controller *k8sv1alpha1.Controller
		expected   bool
	}{
		{
			name: "no annotations",
			controller: &k8sv1alpha1.Controller{
				ObjectMeta: metav1.ObjectMeta{},
			},
			expected: true,
		},
		{
			name: "valid annotations",
			controller: &k8sv1alpha1.Controller{
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
			controller: &k8sv1alpha1.Controller{
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
			result := controllerMustUpdateAnnotations(tt.controller)
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestControllerUpdateDefaultAnnotations(t *testing.T) {
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
			controller := &k8sv1alpha1.Controller{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tt.annotations,
				},
			}

			controllerUpdateDefaultAnnotations(controller)

			if controller.GetAnnotations()["resource.k8s.netris.ai/import"] != tt.expectedImport {
				t.Errorf("import: got %q, expected %q",
					controller.GetAnnotations()["resource.k8s.netris.ai/import"], tt.expectedImport)
			}
			if controller.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"] != tt.expectedReclaimPolicy {
				t.Errorf("reclaimPolicy: got %q, expected %q",
					controller.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"], tt.expectedReclaimPolicy)
			}
		})
	}
}
