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
	"github.com/netrisai/netriswebapi/v2/types/inventory"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSoftgateCompareFieldsForNewMeta(t *testing.T) {
	tests := []struct {
		name         string
		softgate     *k8sv1alpha1.Softgate
		softgateMeta *k8sv1alpha1.SoftgateMeta
		expected     bool
	}{
		{
			name: "no changes",
			softgate: &k8sv1alpha1.Softgate{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "false",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			softgateMeta: &k8sv1alpha1.SoftgateMeta{
				Spec: k8sv1alpha1.SoftgateMetaSpec{
					SoftgateCRGeneration: 1,
					Imported:             false,
					Reclaim:              false,
				},
			},
			expected: false,
		},
		{
			name: "generation changed",
			softgate: &k8sv1alpha1.Softgate{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 2,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "false",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			softgateMeta: &k8sv1alpha1.SoftgateMeta{
				Spec: k8sv1alpha1.SoftgateMetaSpec{
					SoftgateCRGeneration: 1,
					Imported:             false,
					Reclaim:              false,
				},
			},
			expected: true,
		},
		{
			name: "import annotation changed",
			softgate: &k8sv1alpha1.Softgate{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "true",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			softgateMeta: &k8sv1alpha1.SoftgateMeta{
				Spec: k8sv1alpha1.SoftgateMetaSpec{
					SoftgateCRGeneration: 1,
					Imported:             false,
					Reclaim:              false,
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := softgateCompareFieldsForNewMeta(tt.softgate, tt.softgateMeta)
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestSoftgateMustUpdateAnnotations(t *testing.T) {
	tests := []struct {
		name     string
		softgate *k8sv1alpha1.Softgate
		expected bool
	}{
		{
			name: "no annotations",
			softgate: &k8sv1alpha1.Softgate{
				ObjectMeta: metav1.ObjectMeta{},
			},
			expected: true,
		},
		{
			name: "valid annotations",
			softgate: &k8sv1alpha1.Softgate{
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
			softgate: &k8sv1alpha1.Softgate{
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
			result := softgateMustUpdateAnnotations(tt.softgate)
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestSoftgateUpdateDefaultAnnotations(t *testing.T) {
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
			softgate := &k8sv1alpha1.Softgate{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tt.annotations,
				},
			}

			softgateUpdateDefaultAnnotations(softgate)

			if softgate.GetAnnotations()["resource.k8s.netris.ai/import"] != tt.expectedImport {
				t.Errorf("import: got %q, expected %q",
					softgate.GetAnnotations()["resource.k8s.netris.ai/import"], tt.expectedImport)
			}
			if softgate.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"] != tt.expectedReclaimPolicy {
				t.Errorf("reclaimPolicy: got %q, expected %q",
					softgate.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"], tt.expectedReclaimPolicy)
			}
		})
	}
}

func TestCompareSoftgateMetaAPIESoftgate(t *testing.T) {
	tests := []struct {
		name         string
		softgateMeta *k8sv1alpha1.SoftgateMeta
		apiSoftgate  *inventory.HW
		expected     bool
	}{
		{
			name: "all fields match",
			softgateMeta: &k8sv1alpha1.SoftgateMeta{
				Spec: k8sv1alpha1.SoftgateMetaSpec{
					SoftgateName: "softgate-1",
					Description:  "Test softgate",
					TenantID:     1,
					SiteID:       2,
					ProfileID:    3,
					MainIP:       "10.0.0.1",
					MgmtIP:       "192.168.1.1",
				},
			},
			apiSoftgate: &inventory.HW{
				Name:        "softgate-1",
				Description: "Test softgate",
				Tenant:      inventory.IDName{ID: 1},
				Site:        inventory.IDName{ID: 2},
				Profile:     inventory.IDName{ID: 3},
				MainIP:      inventory.HWMainIP{Address: "10.0.0.1"},
				MgmtIP:      inventory.HWMgmtIP{Address: "192.168.1.1"},
			},
			expected: true,
		},
		{
			name: "name mismatch",
			softgateMeta: &k8sv1alpha1.SoftgateMeta{
				Spec: k8sv1alpha1.SoftgateMetaSpec{
					SoftgateName: "softgate-a",
				},
			},
			apiSoftgate: &inventory.HW{
				Name: "softgate-b",
			},
			expected: false,
		},
		{
			name: "description mismatch",
			softgateMeta: &k8sv1alpha1.SoftgateMeta{
				Spec: k8sv1alpha1.SoftgateMetaSpec{
					SoftgateName: "softgate-1",
					Description:  "Description A",
				},
			},
			apiSoftgate: &inventory.HW{
				Name:        "softgate-1",
				Description: "Description B",
			},
			expected: false,
		},
		{
			name: "tenant mismatch",
			softgateMeta: &k8sv1alpha1.SoftgateMeta{
				Spec: k8sv1alpha1.SoftgateMetaSpec{
					SoftgateName: "softgate-1",
					TenantID:     1,
				},
			},
			apiSoftgate: &inventory.HW{
				Name:   "softgate-1",
				Tenant: inventory.IDName{ID: 2},
			},
			expected: false,
		},
		{
			name: "site mismatch",
			softgateMeta: &k8sv1alpha1.SoftgateMeta{
				Spec: k8sv1alpha1.SoftgateMetaSpec{
					SoftgateName: "softgate-1",
					SiteID:       1,
				},
			},
			apiSoftgate: &inventory.HW{
				Name: "softgate-1",
				Site: inventory.IDName{ID: 2},
			},
			expected: false,
		},
		{
			name: "profile mismatch",
			softgateMeta: &k8sv1alpha1.SoftgateMeta{
				Spec: k8sv1alpha1.SoftgateMetaSpec{
					SoftgateName: "softgate-1",
					ProfileID:    1,
				},
			},
			apiSoftgate: &inventory.HW{
				Name:    "softgate-1",
				Profile: inventory.IDName{ID: 2},
			},
			expected: false,
		},
		{
			name: "main IP mismatch",
			softgateMeta: &k8sv1alpha1.SoftgateMeta{
				Spec: k8sv1alpha1.SoftgateMetaSpec{
					SoftgateName: "softgate-1",
					MainIP:       "10.0.0.1",
				},
			},
			apiSoftgate: &inventory.HW{
				Name:   "softgate-1",
				MainIP: inventory.HWMainIP{Address: "10.0.0.2"},
			},
			expected: false,
		},
		{
			name: "mgmt IP mismatch",
			softgateMeta: &k8sv1alpha1.SoftgateMeta{
				Spec: k8sv1alpha1.SoftgateMetaSpec{
					SoftgateName: "softgate-1",
					MgmtIP:       "192.168.1.1",
				},
			},
			apiSoftgate: &inventory.HW{
				Name:   "softgate-1",
				MgmtIP: inventory.HWMgmtIP{Address: "192.168.1.2"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareSoftgateMetaAPIESoftgate(tt.softgateMeta, tt.apiSoftgate, newTestLogger())
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}
