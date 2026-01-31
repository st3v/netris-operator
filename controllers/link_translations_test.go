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

func TestLinkCompareFieldsForNewMeta(t *testing.T) {
	tests := []struct {
		name     string
		link     *k8sv1alpha1.Link
		linkMeta *k8sv1alpha1.LinkMeta
		expected bool
	}{
		{
			name: "no changes",
			link: &k8sv1alpha1.Link{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "false",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			linkMeta: &k8sv1alpha1.LinkMeta{
				Spec: k8sv1alpha1.LinkMetaSpec{
					LinkCRGeneration: 1,
					Imported:         false,
					Reclaim:          false,
				},
			},
			expected: false,
		},
		{
			name: "generation changed",
			link: &k8sv1alpha1.Link{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 2,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "false",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			linkMeta: &k8sv1alpha1.LinkMeta{
				Spec: k8sv1alpha1.LinkMetaSpec{
					LinkCRGeneration: 1,
					Imported:         false,
					Reclaim:          false,
				},
			},
			expected: true,
		},
		{
			name: "import changed",
			link: &k8sv1alpha1.Link{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "true",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			linkMeta: &k8sv1alpha1.LinkMeta{
				Spec: k8sv1alpha1.LinkMetaSpec{
					LinkCRGeneration: 1,
					Imported:         false,
					Reclaim:          false,
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := linkCompareFieldsForNewMeta(tt.link, tt.linkMeta)
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestLinkMustUpdateAnnotations(t *testing.T) {
	tests := []struct {
		name     string
		link     *k8sv1alpha1.Link
		expected bool
	}{
		{
			name: "no annotations",
			link: &k8sv1alpha1.Link{
				ObjectMeta: metav1.ObjectMeta{},
			},
			expected: true,
		},
		{
			name: "valid annotations",
			link: &k8sv1alpha1.Link{
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
			link: &k8sv1alpha1.Link{
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
			result := linkMustUpdateAnnotations(tt.link)
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestLinkUpdateDefaultAnnotations(t *testing.T) {
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
			link := &k8sv1alpha1.Link{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tt.annotations,
				},
			}

			linkUpdateDefaultAnnotations(link)

			if link.GetAnnotations()["resource.k8s.netris.ai/import"] != tt.expectedImport {
				t.Errorf("import: got %q, expected %q",
					link.GetAnnotations()["resource.k8s.netris.ai/import"], tt.expectedImport)
			}
			if link.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"] != tt.expectedReclaimPolicy {
				t.Errorf("reclaimPolicy: got %q, expected %q",
					link.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"], tt.expectedReclaimPolicy)
			}
		})
	}
}
