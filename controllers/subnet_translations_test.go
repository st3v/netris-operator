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
	"github.com/netrisai/netriswebapi/v2/types/ipam"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSubnetCompareFieldsForNewMeta(t *testing.T) {
	tests := []struct {
		name       string
		subnet     *k8sv1alpha1.Subnet
		subnetMeta *k8sv1alpha1.SubnetMeta
		expected   bool
	}{
		{
			name: "no changes",
			subnet: &k8sv1alpha1.Subnet{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "false",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			subnetMeta: &k8sv1alpha1.SubnetMeta{
				Spec: k8sv1alpha1.SubnetMetaSpec{
					SubnetCRGeneration: 1,
					Imported:           false,
					Reclaim:            false,
				},
			},
			expected: false,
		},
		{
			name: "generation changed",
			subnet: &k8sv1alpha1.Subnet{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 2,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "false",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			subnetMeta: &k8sv1alpha1.SubnetMeta{
				Spec: k8sv1alpha1.SubnetMetaSpec{
					SubnetCRGeneration: 1,
					Imported:           false,
					Reclaim:            false,
				},
			},
			expected: true,
		},
		{
			name: "import annotation changed",
			subnet: &k8sv1alpha1.Subnet{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "true",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			subnetMeta: &k8sv1alpha1.SubnetMeta{
				Spec: k8sv1alpha1.SubnetMetaSpec{
					SubnetCRGeneration: 1,
					Imported:           false,
					Reclaim:            false,
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := subnetCompareFieldsForNewMeta(tt.subnet, tt.subnetMeta)
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestSubnetMustUpdateAnnotations(t *testing.T) {
	tests := []struct {
		name     string
		subnet   *k8sv1alpha1.Subnet
		expected bool
	}{
		{
			name: "no annotations",
			subnet: &k8sv1alpha1.Subnet{
				ObjectMeta: metav1.ObjectMeta{},
			},
			expected: true,
		},
		{
			name: "valid annotations",
			subnet: &k8sv1alpha1.Subnet{
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
			subnet: &k8sv1alpha1.Subnet{
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
			result := subnetMustUpdateAnnotations(tt.subnet)
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestSubnetUpdateDefaultAnnotations(t *testing.T) {
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
			subnet := &k8sv1alpha1.Subnet{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tt.annotations,
				},
			}

			subnetUpdateDefaultAnnotations(subnet)

			if subnet.GetAnnotations()["resource.k8s.netris.ai/import"] != tt.expectedImport {
				t.Errorf("import: got %q, expected %q",
					subnet.GetAnnotations()["resource.k8s.netris.ai/import"], tt.expectedImport)
			}
			if subnet.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"] != tt.expectedReclaimPolicy {
				t.Errorf("reclaimPolicy: got %q, expected %q",
					subnet.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"], tt.expectedReclaimPolicy)
			}
		})
	}
}

func TestCompareSubnetMetaSiteAPISubnetSite(t *testing.T) {
	tests := []struct {
		name           string
		subnetMetaSites []int
		apiSubnetSites []ipam.IDName
		expected       bool
	}{
		{
			name:           "both empty",
			subnetMetaSites: []int{},
			apiSubnetSites: []ipam.IDName{},
			expected:       true,
		},
		{
			name:           "matching sites",
			subnetMetaSites: []int{1, 2, 3},
			apiSubnetSites: []ipam.IDName{
				{ID: 1},
				{ID: 2},
				{ID: 3},
			},
			expected: true,
		},
		{
			name:           "different sites",
			subnetMetaSites: []int{1, 2},
			apiSubnetSites: []ipam.IDName{
				{ID: 1},
				{ID: 3},
			},
			expected: false,
		},
		{
			name:           "different count",
			subnetMetaSites: []int{1, 2},
			apiSubnetSites: []ipam.IDName{
				{ID: 1},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareSubnetMetaSiteAPISubnetSite(tt.subnetMetaSites, tt.apiSubnetSites)
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}
