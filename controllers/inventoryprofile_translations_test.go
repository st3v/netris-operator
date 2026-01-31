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
	"github.com/netrisai/netriswebapi/v1/types/inventoryprofile"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestInventoryProfileCompareFieldsForNewMeta(t *testing.T) {
	tests := []struct {
		name                 string
		inventoryProfile     *k8sv1alpha1.InventoryProfile
		inventoryProfileMeta *k8sv1alpha1.InventoryProfileMeta
		expected             bool
	}{
		{
			name: "no changes",
			inventoryProfile: &k8sv1alpha1.InventoryProfile{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "false",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			inventoryProfileMeta: &k8sv1alpha1.InventoryProfileMeta{
				Spec: k8sv1alpha1.InventoryProfileMetaSpec{
					InventoryProfileCRGeneration: 1,
					Imported:                     false,
					Reclaim:                      false,
				},
			},
			expected: false,
		},
		{
			name: "generation changed",
			inventoryProfile: &k8sv1alpha1.InventoryProfile{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 2,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "false",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			inventoryProfileMeta: &k8sv1alpha1.InventoryProfileMeta{
				Spec: k8sv1alpha1.InventoryProfileMetaSpec{
					InventoryProfileCRGeneration: 1,
					Imported:                     false,
					Reclaim:                      false,
				},
			},
			expected: true,
		},
		{
			name: "import annotation changed",
			inventoryProfile: &k8sv1alpha1.InventoryProfile{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "true",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			inventoryProfileMeta: &k8sv1alpha1.InventoryProfileMeta{
				Spec: k8sv1alpha1.InventoryProfileMetaSpec{
					InventoryProfileCRGeneration: 1,
					Imported:                     false,
					Reclaim:                      false,
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inventoryProfileCompareFieldsForNewMeta(tt.inventoryProfile, tt.inventoryProfileMeta)
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestInventoryProfileMustUpdateAnnotations(t *testing.T) {
	tests := []struct {
		name             string
		inventoryProfile *k8sv1alpha1.InventoryProfile
		expected         bool
	}{
		{
			name: "no annotations",
			inventoryProfile: &k8sv1alpha1.InventoryProfile{
				ObjectMeta: metav1.ObjectMeta{},
			},
			expected: true,
		},
		{
			name: "valid annotations",
			inventoryProfile: &k8sv1alpha1.InventoryProfile{
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
			inventoryProfile: &k8sv1alpha1.InventoryProfile{
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
			result := inventoryProfileMustUpdateAnnotations(tt.inventoryProfile)
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestInventoryProfileUpdateDefaultAnnotations(t *testing.T) {
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
			inventoryProfile := &k8sv1alpha1.InventoryProfile{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tt.annotations,
				},
			}

			inventoryProfileUpdateDefaultAnnotations(inventoryProfile)

			if inventoryProfile.GetAnnotations()["resource.k8s.netris.ai/import"] != tt.expectedImport {
				t.Errorf("import: got %q, expected %q",
					inventoryProfile.GetAnnotations()["resource.k8s.netris.ai/import"], tt.expectedImport)
			}
			if inventoryProfile.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"] != tt.expectedReclaimPolicy {
				t.Errorf("reclaimPolicy: got %q, expected %q",
					inventoryProfile.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"], tt.expectedReclaimPolicy)
			}
		})
	}
}

func TestCompareInventoryProfileAPIInventoryProfileCustomRules(t *testing.T) {
	tests := []struct {
		name              string
		inventoryRules    []k8sv1alpha1.InventoryProfileCustomRule
		apiProfileRules   []inventoryprofile.CustomRule
		expected          bool
	}{
		{
			name:            "both empty",
			inventoryRules:  []k8sv1alpha1.InventoryProfileCustomRule{},
			apiProfileRules: []inventoryprofile.CustomRule{},
			expected:        true,
		},
		{
			name: "matching rules",
			inventoryRules: []k8sv1alpha1.InventoryProfileCustomRule{
				{SrcSubnet: "10.0.0.0/24", SrcPort: "8080", DstPort: "80", Protocol: "tcp"},
			},
			apiProfileRules: []inventoryprofile.CustomRule{
				{SrcSubnet: "10.0.0.0/24", SrcPort: "8080", DstPort: "80", Protocol: "tcp"},
			},
			expected: true,
		},
		{
			name: "different rules",
			inventoryRules: []k8sv1alpha1.InventoryProfileCustomRule{
				{SrcSubnet: "10.0.0.0/24", SrcPort: "8080", DstPort: "80", Protocol: "tcp"},
			},
			apiProfileRules: []inventoryprofile.CustomRule{
				{SrcSubnet: "192.168.0.0/24", SrcPort: "8080", DstPort: "80", Protocol: "tcp"},
			},
			expected: false,
		},
		{
			name: "different count",
			inventoryRules: []k8sv1alpha1.InventoryProfileCustomRule{
				{SrcSubnet: "10.0.0.0/24", SrcPort: "8080", DstPort: "80", Protocol: "tcp"},
				{SrcSubnet: "192.168.0.0/24", SrcPort: "443", DstPort: "443", Protocol: "tcp"},
			},
			apiProfileRules: []inventoryprofile.CustomRule{
				{SrcSubnet: "10.0.0.0/24", SrcPort: "8080", DstPort: "80", Protocol: "tcp"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareInventoryProfileAPIInventoryProfileCustomRules(tt.inventoryRules, tt.apiProfileRules)
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestUnmarshalTimezone(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedTzCode string
	}{
		{
			name:           "empty string",
			input:          "",
			expectedTzCode: "",
		},
		{
			name:           "valid timezone json",
			input:          `{"label":"America/New_York","tzCode":"America/New_York"}`,
			expectedTzCode: "America/New_York",
		},
		{
			name:           "invalid json",
			input:          "not json",
			expectedTzCode: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := unmarshalTimezone(tt.input)
			if result.TzCode != tt.expectedTzCode {
				t.Errorf("got TzCode %q, expected %q", result.TzCode, tt.expectedTzCode)
			}
		})
	}
}
