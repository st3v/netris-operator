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
			name:                  "nil annotations get defaults",
			annotations:           nil,
			expectedImport:        "false",
			expectedReclaimPolicy: "delete",
		},
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

func TestCompareSwitchMetaAPIESwitch(t *testing.T) {
	tests := []struct {
		name       string
		switchMeta *k8sv1alpha1.SwitchMeta
		apiSwitch  *inventory.HW
		expected   bool
	}{
		{
			name: "all fields match",
			switchMeta: &k8sv1alpha1.SwitchMeta{
				Spec: k8sv1alpha1.SwitchMetaSpec{
					SwitchName:  "switch-1",
					Description: "Test switch",
					TenantID:    1,
					SiteID:      2,
					NOS:         inventory.NOS{Tag: "cumulus_linux"},
					ASN:         65000,
					PortsCount:  48,
					MacAddress:  "00:11:22:33:44:55",
					ProfileID:   3,
					MainIP:      "10.0.0.1",
					MgmtIP:      "192.168.1.1",
				},
			},
			apiSwitch: &inventory.HW{
				Name:        "switch-1",
				Description: "Test switch",
				Tenant:      inventory.IDName{ID: 1},
				Site:        inventory.IDName{ID: 2},
				Nos:         inventory.HWNOS{Tag: "cumulus_linux"},
				Asn:         65000,
				PortCount:   48,
				MacAddress:  "00:11:22:33:44:55",
				Profile:     inventory.IDName{ID: 3},
				MainIP:      inventory.HWMainIP{Address: "10.0.0.1"},
				MgmtIP:      inventory.HWMgmtIP{Address: "192.168.1.1"},
			},
			expected: true,
		},
		{
			name: "name mismatch",
			switchMeta: &k8sv1alpha1.SwitchMeta{
				Spec: k8sv1alpha1.SwitchMetaSpec{
					SwitchName: "switch-a",
				},
			},
			apiSwitch: &inventory.HW{
				Name: "switch-b",
			},
			expected: false,
		},
		{
			name: "description mismatch",
			switchMeta: &k8sv1alpha1.SwitchMeta{
				Spec: k8sv1alpha1.SwitchMetaSpec{
					SwitchName:  "switch-1",
					Description: "Description A",
				},
			},
			apiSwitch: &inventory.HW{
				Name:        "switch-1",
				Description: "Description B",
			},
			expected: false,
		},
		{
			name: "tenant mismatch",
			switchMeta: &k8sv1alpha1.SwitchMeta{
				Spec: k8sv1alpha1.SwitchMetaSpec{
					SwitchName: "switch-1",
					TenantID:   1,
				},
			},
			apiSwitch: &inventory.HW{
				Name:   "switch-1",
				Tenant: inventory.IDName{ID: 2},
			},
			expected: false,
		},
		{
			name: "site mismatch",
			switchMeta: &k8sv1alpha1.SwitchMeta{
				Spec: k8sv1alpha1.SwitchMetaSpec{
					SwitchName: "switch-1",
					SiteID:     1,
				},
			},
			apiSwitch: &inventory.HW{
				Name: "switch-1",
				Site: inventory.IDName{ID: 2},
			},
			expected: false,
		},
		{
			name: "NOS mismatch",
			switchMeta: &k8sv1alpha1.SwitchMeta{
				Spec: k8sv1alpha1.SwitchMetaSpec{
					SwitchName: "switch-1",
					NOS:        inventory.NOS{Tag: "cumulus_linux"},
				},
			},
			apiSwitch: &inventory.HW{
				Name: "switch-1",
				Nos:  inventory.HWNOS{Tag: "sonic"},
			},
			expected: false,
		},
		{
			name: "ASN mismatch",
			switchMeta: &k8sv1alpha1.SwitchMeta{
				Spec: k8sv1alpha1.SwitchMetaSpec{
					SwitchName: "switch-1",
					ASN:        65000,
				},
			},
			apiSwitch: &inventory.HW{
				Name: "switch-1",
				Asn:  65001,
			},
			expected: false,
		},
		{
			name: "port count mismatch",
			switchMeta: &k8sv1alpha1.SwitchMeta{
				Spec: k8sv1alpha1.SwitchMetaSpec{
					SwitchName: "switch-1",
					PortsCount: 48,
				},
			},
			apiSwitch: &inventory.HW{
				Name:      "switch-1",
				PortCount: 32,
			},
			expected: false,
		},
		{
			name: "MAC address mismatch",
			switchMeta: &k8sv1alpha1.SwitchMeta{
				Spec: k8sv1alpha1.SwitchMetaSpec{
					SwitchName: "switch-1",
					MacAddress: "00:11:22:33:44:55",
				},
			},
			apiSwitch: &inventory.HW{
				Name:       "switch-1",
				MacAddress: "00:11:22:33:44:66",
			},
			expected: false,
		},
		{
			name: "profile mismatch",
			switchMeta: &k8sv1alpha1.SwitchMeta{
				Spec: k8sv1alpha1.SwitchMetaSpec{
					SwitchName: "switch-1",
					ProfileID:  1,
				},
			},
			apiSwitch: &inventory.HW{
				Name:    "switch-1",
				Profile: inventory.IDName{ID: 2},
			},
			expected: false,
		},
		{
			name: "main IP mismatch",
			switchMeta: &k8sv1alpha1.SwitchMeta{
				Spec: k8sv1alpha1.SwitchMetaSpec{
					SwitchName: "switch-1",
					MainIP:     "10.0.0.1",
				},
			},
			apiSwitch: &inventory.HW{
				Name:   "switch-1",
				MainIP: inventory.HWMainIP{Address: "10.0.0.2"},
			},
			expected: false,
		},
		{
			name: "mgmt IP mismatch",
			switchMeta: &k8sv1alpha1.SwitchMeta{
				Spec: k8sv1alpha1.SwitchMetaSpec{
					SwitchName: "switch-1",
					MgmtIP:     "192.168.1.1",
				},
			},
			apiSwitch: &inventory.HW{
				Name:   "switch-1",
				MgmtIP: inventory.HWMgmtIP{Address: "192.168.1.2"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareSwitchMetaAPIESwitch(tt.switchMeta, tt.apiSwitch, newTestLogger())
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}
