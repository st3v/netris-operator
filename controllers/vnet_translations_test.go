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
	"github.com/netrisai/netriswebapi/v2/types/vnet"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFindGatewayDuplicates(t *testing.T) {
	tests := []struct {
		name          string
		gateways      []k8sv1alpha1.VNetGateway
		wantDuplicate string
		wantFound     bool
	}{
		{
			name:          "empty list",
			gateways:      []k8sv1alpha1.VNetGateway{},
			wantDuplicate: "",
			wantFound:     false,
		},
		{
			name: "single gateway",
			gateways: []k8sv1alpha1.VNetGateway{
				{Prefix: "192.168.1.1/24"},
			},
			wantDuplicate: "",
			wantFound:     false,
		},
		{
			name: "no duplicates",
			gateways: []k8sv1alpha1.VNetGateway{
				{Prefix: "192.168.1.1/24"},
				{Prefix: "10.0.0.1/24"},
				{Prefix: "172.16.0.1/16"},
			},
			wantDuplicate: "",
			wantFound:     false,
		},
		{
			name: "duplicate found",
			gateways: []k8sv1alpha1.VNetGateway{
				{Prefix: "192.168.1.1/24"},
				{Prefix: "10.0.0.1/24"},
				{Prefix: "192.168.1.1/24"},
			},
			wantDuplicate: "192.168.1.1/24",
			wantFound:     true,
		},
		{
			name: "returns first duplicate",
			gateways: []k8sv1alpha1.VNetGateway{
				{Prefix: "192.168.1.1/24"},
				{Prefix: "192.168.1.1/24"},
				{Prefix: "10.0.0.1/24"},
				{Prefix: "10.0.0.1/24"},
			},
			wantDuplicate: "192.168.1.1/24",
			wantFound:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDup, gotFound := findGatewayDuplicates(tt.gateways)
			if gotDup != tt.wantDuplicate {
				t.Errorf("duplicate: got %q, want %q", gotDup, tt.wantDuplicate)
			}
			if gotFound != tt.wantFound {
				t.Errorf("found: got %v, want %v", gotFound, tt.wantFound)
			}
		})
	}
}

func TestCompareVNetMetaAPIVnetSites(t *testing.T) {
	tests := []struct {
		name      string
		k8sSites  []k8sv1alpha1.VNetMetaSite
		apiSites  []vnet.VNetDetailedSite
		wantMatch bool
	}{
		{
			name:      "both empty",
			k8sSites:  []k8sv1alpha1.VNetMetaSite{},
			apiSites:  []vnet.VNetDetailedSite{},
			wantMatch: true,
		},
		{
			name: "matching sites",
			k8sSites: []k8sv1alpha1.VNetMetaSite{
				{Name: "site-a"},
				{Name: "site-b"},
			},
			apiSites: []vnet.VNetDetailedSite{
				{ID: 1, Name: "site-a"},
				{ID: 2, Name: "site-b"},
			},
			wantMatch: true,
		},
		{
			name: "api has site not in k8s",
			k8sSites: []k8sv1alpha1.VNetMetaSite{
				{Name: "site-a"},
			},
			apiSites: []vnet.VNetDetailedSite{
				{ID: 1, Name: "site-a"},
				{ID: 2, Name: "site-b"},
			},
			wantMatch: false,
		},
		{
			name: "k8s has more sites than api",
			k8sSites: []k8sv1alpha1.VNetMetaSite{
				{Name: "site-a"},
				{Name: "site-b"},
			},
			apiSites: []vnet.VNetDetailedSite{
				{ID: 1, Name: "site-a"},
			},
			wantMatch: true, // function only checks if API sites exist in k8s
		},
		{
			name:     "k8s empty api has sites",
			k8sSites: []k8sv1alpha1.VNetMetaSite{},
			apiSites: []vnet.VNetDetailedSite{
				{ID: 1, Name: "site-a"},
			},
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareVNetMetaAPIVnetSites(tt.k8sSites, tt.apiSites)
			if got != tt.wantMatch {
				t.Errorf("got %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestCompareVNetMetaAPIVnetTenants(t *testing.T) {
	tests := []struct {
		name       string
		k8sTenants []string
		apiTenants []vnet.VNetDetailedGuestTenant
		wantMatch  bool
	}{
		{
			name:       "both empty",
			k8sTenants: []string{},
			apiTenants: []vnet.VNetDetailedGuestTenant{},
			wantMatch:  true,
		},
		{
			name:       "matching tenants",
			k8sTenants: []string{"tenant-a", "tenant-b"},
			apiTenants: []vnet.VNetDetailedGuestTenant{
				{ID: 1, Name: "tenant-a"},
				{ID: 2, Name: "tenant-b"},
			},
			wantMatch: true,
		},
		{
			name:       "different tenants",
			k8sTenants: []string{"tenant-a"},
			apiTenants: []vnet.VNetDetailedGuestTenant{
				{ID: 1, Name: "tenant-b"},
			},
			wantMatch: false,
		},
		{
			name:       "different count",
			k8sTenants: []string{"tenant-a", "tenant-b"},
			apiTenants: []vnet.VNetDetailedGuestTenant{
				{ID: 1, Name: "tenant-a"},
			},
			wantMatch: false,
		},
		{
			name:       "same tenants different order still matches",
			k8sTenants: []string{"tenant-b", "tenant-a"},
			apiTenants: []vnet.VNetDetailedGuestTenant{
				{ID: 1, Name: "tenant-a"},
				{ID: 2, Name: "tenant-b"},
			},
			wantMatch: true, // diff library treats same elements as matching
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareVNetMetaAPIVnetTenants(tt.k8sTenants, tt.apiTenants)
			if got != tt.wantMatch {
				t.Errorf("got %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestVnetMustUpdateAnnotations(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]string
		wantUpdate  bool
	}{
		{
			name:        "no annotations",
			annotations: map[string]string{},
			wantUpdate:  true,
		},
		{
			name: "valid annotations",
			annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			wantUpdate: false,
		},
		{
			name: "import true reclaim retain",
			annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "true",
				"resource.k8s.netris.ai/reclaimPolicy": "retain",
			},
			wantUpdate: false,
		},
		{
			name: "missing import annotation",
			annotations: map[string]string{
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			wantUpdate: true,
		},
		{
			name: "missing reclaimPolicy annotation",
			annotations: map[string]string{
				"resource.k8s.netris.ai/import": "false",
			},
			wantUpdate: true,
		},
		{
			name: "invalid import value",
			annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "invalid",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			wantUpdate: true,
		},
		{
			name: "invalid reclaimPolicy value",
			annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "invalid",
			},
			wantUpdate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vnet := &k8sv1alpha1.VNet{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tt.annotations,
				},
			}
			got := vnetMustUpdateAnnotations(vnet)
			if got != tt.wantUpdate {
				t.Errorf("got %v, want %v", got, tt.wantUpdate)
			}
		})
	}
}

func TestVnetUpdateDefaultAnnotations(t *testing.T) {
	tests := []struct {
		name              string
		inputAnnotations  map[string]string
		wantImport        string
		wantReclaimPolicy string
	}{
		{
			name:              "empty annotations get defaults",
			inputAnnotations:  map[string]string{},
			wantImport:        "false",
			wantReclaimPolicy: "delete",
		},
		{
			name: "preserves import true",
			inputAnnotations: map[string]string{
				"resource.k8s.netris.ai/import": "true",
			},
			wantImport:        "true",
			wantReclaimPolicy: "delete",
		},
		{
			name: "preserves reclaim retain",
			inputAnnotations: map[string]string{
				"resource.k8s.netris.ai/reclaimPolicy": "retain",
			},
			wantImport:        "false",
			wantReclaimPolicy: "retain",
		},
		{
			name: "invalid values get overwritten with defaults",
			inputAnnotations: map[string]string{
				"resource.k8s.netris.ai/import":        "invalid",
				"resource.k8s.netris.ai/reclaimPolicy": "invalid",
			},
			wantImport:        "false",
			wantReclaimPolicy: "delete",
		},
		{
			name: "preserves other annotations",
			inputAnnotations: map[string]string{
				"other-annotation": "value",
			},
			wantImport:        "false",
			wantReclaimPolicy: "delete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vnet := &k8sv1alpha1.VNet{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tt.inputAnnotations,
				},
			}
			vnetUpdateDefaultAnnotations(vnet)

			gotImport := vnet.GetAnnotations()["resource.k8s.netris.ai/import"]
			gotReclaim := vnet.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"]

			if gotImport != tt.wantImport {
				t.Errorf("import: got %q, want %q", gotImport, tt.wantImport)
			}
			if gotReclaim != tt.wantReclaimPolicy {
				t.Errorf("reclaimPolicy: got %q, want %q", gotReclaim, tt.wantReclaimPolicy)
			}
		})
	}
}

func TestVnetCompareFieldsForNewMeta(t *testing.T) {
	tests := []struct {
		name            string
		vnetGen         int64
		vnetAnnotations map[string]string
		metaGen         int64
		metaImported    bool
		metaReclaim     bool
		wantChanged     bool
	}{
		{
			name:    "no changes",
			vnetGen: 1,
			vnetAnnotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			metaGen:      1,
			metaImported: false,
			metaReclaim:  false,
			wantChanged:  false,
		},
		{
			name:    "generation changed",
			vnetGen: 2,
			vnetAnnotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			metaGen:      1,
			metaImported: false,
			metaReclaim:  false,
			wantChanged:  true,
		},
		{
			name:    "import annotation changed",
			vnetGen: 1,
			vnetAnnotations: map[string]string{
				"resource.k8s.netris.ai/import":        "true",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			metaGen:      1,
			metaImported: false,
			metaReclaim:  false,
			wantChanged:  true,
		},
		{
			name:    "reclaim annotation changed",
			vnetGen: 1,
			vnetAnnotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "retain",
			},
			metaGen:      1,
			metaImported: false,
			metaReclaim:  false,
			wantChanged:  true,
		},
		{
			name:            "missing annotations treated as false",
			vnetGen:         1,
			vnetAnnotations: map[string]string{},
			metaGen:         1,
			metaImported:    false,
			metaReclaim:     false,
			wantChanged:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vnet := &k8sv1alpha1.VNet{
				ObjectMeta: metav1.ObjectMeta{
					Generation:  tt.vnetGen,
					Annotations: tt.vnetAnnotations,
				},
			}
			vnetMeta := &k8sv1alpha1.VNetMeta{
				Spec: k8sv1alpha1.VNetMetaSpec{
					VnetCRGeneration: tt.metaGen,
					Imported:         tt.metaImported,
					Reclaim:          tt.metaReclaim,
				},
			}

			got := vnetCompareFieldsForNewMeta(vnet, vnetMeta)
			if got != tt.wantChanged {
				t.Errorf("got %v, want %v", got, tt.wantChanged)
			}
		})
	}
}
