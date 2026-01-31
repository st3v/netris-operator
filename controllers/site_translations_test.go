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
	"github.com/netrisai/netriswebapi/v2/types/site"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSiteCompareFieldsForNewMeta(t *testing.T) {
	tests := []struct {
		name     string
		site     *k8sv1alpha1.Site
		siteMeta *k8sv1alpha1.SiteMeta
		expected bool
	}{
		{
			name: "no changes",
			site: &k8sv1alpha1.Site{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "false",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			siteMeta: &k8sv1alpha1.SiteMeta{
				Spec: k8sv1alpha1.SiteMetaSpec{
					SiteCRGeneration: 1,
					Imported:         false,
					Reclaim:          false,
				},
			},
			expected: false,
		},
		{
			name: "generation changed",
			site: &k8sv1alpha1.Site{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 2,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "false",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			siteMeta: &k8sv1alpha1.SiteMeta{
				Spec: k8sv1alpha1.SiteMetaSpec{
					SiteCRGeneration: 1,
					Imported:         false,
					Reclaim:          false,
				},
			},
			expected: true,
		},
		{
			name: "import annotation changed",
			site: &k8sv1alpha1.Site{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "true",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			siteMeta: &k8sv1alpha1.SiteMeta{
				Spec: k8sv1alpha1.SiteMetaSpec{
					SiteCRGeneration: 1,
					Imported:         false,
					Reclaim:          false,
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := siteCompareFieldsForNewMeta(tt.site, tt.siteMeta)
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestSiteMustUpdateAnnotations(t *testing.T) {
	tests := []struct {
		name     string
		site     *k8sv1alpha1.Site
		expected bool
	}{
		{
			name: "no annotations",
			site: &k8sv1alpha1.Site{
				ObjectMeta: metav1.ObjectMeta{},
			},
			expected: true,
		},
		{
			name: "valid annotations",
			site: &k8sv1alpha1.Site{
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
			site: &k8sv1alpha1.Site{
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
			result := siteMustUpdateAnnotations(tt.site)
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestSiteUpdateDefaultAnnotations(t *testing.T) {
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
			site := &k8sv1alpha1.Site{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tt.annotations,
				},
			}

			siteUpdateDefaultAnnotations(site)

			if site.GetAnnotations()["resource.k8s.netris.ai/import"] != tt.expectedImport {
				t.Errorf("import: got %q, expected %q",
					site.GetAnnotations()["resource.k8s.netris.ai/import"], tt.expectedImport)
			}
			if site.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"] != tt.expectedReclaimPolicy {
				t.Errorf("reclaimPolicy: got %q, expected %q",
					site.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"], tt.expectedReclaimPolicy)
			}
		})
	}
}

func TestSiteMetaToNetris(t *testing.T) {
	tests := []struct {
		name     string
		siteMeta *k8sv1alpha1.SiteMeta
		expected *site.Site
	}{
		{
			name: "converts all fields correctly",
			siteMeta: &k8sv1alpha1.SiteMeta{
				Spec: k8sv1alpha1.SiteMetaSpec{
					SiteName:            "test-site",
					PublicASN:           65000,
					RohASN:              65001,
					VMASN:               65002,
					RohRoutingProfileID: 1,
					SiteMesh:            "hub",
					ACLDefaultPolicy:    "permit",
				},
			},
			expected: &site.Site{
				Name:      "test-site",
				PublicAsn: 65000,
				RohAsn:    65001,
				VMAsn:     65002,
				SiteMesh:  site.IDName{Value: "hub"},
				AclPolicy: "permit",
			},
		},
		{
			name: "handles spoke site mesh",
			siteMeta: &k8sv1alpha1.SiteMeta{
				Spec: k8sv1alpha1.SiteMetaSpec{
					SiteName:            "spoke-site",
					PublicASN:           65100,
					SiteMesh:            "spoke",
					ACLDefaultPolicy:    "deny",
					RohRoutingProfileID: 2,
				},
			},
			expected: &site.Site{
				Name:      "spoke-site",
				PublicAsn: 65100,
				SiteMesh:  site.IDName{Value: "spoke"},
				AclPolicy: "deny",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SiteMetaToNetris(tt.siteMeta)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Name != tt.expected.Name {
				t.Errorf("Name: got %q, expected %q", result.Name, tt.expected.Name)
			}
			if result.PublicAsn != tt.expected.PublicAsn {
				t.Errorf("PublicAsn: got %d, expected %d", result.PublicAsn, tt.expected.PublicAsn)
			}
			if result.SiteMesh.Value != tt.expected.SiteMesh.Value {
				t.Errorf("SiteMesh: got %q, expected %q", result.SiteMesh.Value, tt.expected.SiteMesh.Value)
			}
			if result.AclPolicy != tt.expected.AclPolicy {
				t.Errorf("AclPolicy: got %q, expected %q", result.AclPolicy, tt.expected.AclPolicy)
			}
		})
	}
}

func TestSiteMetaToNetrisUpdate(t *testing.T) {
	siteMeta := &k8sv1alpha1.SiteMeta{
		Spec: k8sv1alpha1.SiteMetaSpec{
			ID:                  42,
			SiteName:            "updated-site",
			PublicASN:           65200,
			RohASN:              65201,
			VMASN:               65202,
			RohRoutingProfileID: 3,
			SiteMesh:            "hub",
			ACLDefaultPolicy:    "permit",
		},
	}

	result, err := SiteMetaToNetrisUpdate(siteMeta)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != 42 {
		t.Errorf("ID: got %d, expected 42", result.ID)
	}
	if result.Name != "updated-site" {
		t.Errorf("Name: got %q, expected 'updated-site'", result.Name)
	}
	if result.PublicAsn != 65200 {
		t.Errorf("PublicAsn: got %d, expected 65200", result.PublicAsn)
	}
	if result.RohProfile.ID != 3 {
		t.Errorf("RohProfile.ID: got %d, expected 3", result.RohProfile.ID)
	}
}

func TestCompareSiteMetaAPIESite(t *testing.T) {
	tests := []struct {
		name     string
		siteMeta *k8sv1alpha1.SiteMeta
		apiSite  *site.Site
		expected bool
	}{
		{
			name: "all fields match",
			siteMeta: &k8sv1alpha1.SiteMeta{
				Spec: k8sv1alpha1.SiteMetaSpec{
					SiteName:            "test-site",
					PublicASN:           65000,
					RohASN:              65001,
					VMASN:               65002,
					RohRoutingProfileID: 1,
					SiteMesh:            "hub",
					ACLDefaultPolicy:    "permit",
				},
			},
			apiSite: &site.Site{
				Name:       "test-site",
				PublicAsn:  65000,
				RohAsn:     65001,
				VMAsn:      65002,
				RohProfile: &site.RohProfile{ID: 1},
				SiteMesh:   site.IDName{Value: "hub"},
				AclPolicy:  "permit",
			},
			expected: true,
		},
		{
			name: "name mismatch",
			siteMeta: &k8sv1alpha1.SiteMeta{
				Spec: k8sv1alpha1.SiteMetaSpec{
					SiteName: "site-a",
				},
			},
			apiSite: &site.Site{
				Name:       "site-b",
				RohProfile: &site.RohProfile{},
			},
			expected: false,
		},
		{
			name: "public ASN mismatch",
			siteMeta: &k8sv1alpha1.SiteMeta{
				Spec: k8sv1alpha1.SiteMetaSpec{
					SiteName:  "test-site",
					PublicASN: 65000,
				},
			},
			apiSite: &site.Site{
				Name:       "test-site",
				PublicAsn:  65999,
				RohProfile: &site.RohProfile{},
			},
			expected: false,
		},
		{
			name: "ACL policy mismatch",
			siteMeta: &k8sv1alpha1.SiteMeta{
				Spec: k8sv1alpha1.SiteMetaSpec{
					SiteName:         "test-site",
					ACLDefaultPolicy: "permit",
				},
			},
			apiSite: &site.Site{
				Name:       "test-site",
				AclPolicy:  "deny",
				RohProfile: &site.RohProfile{},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareSiteMetaAPIESite(tt.siteMeta, tt.apiSite, newTestLogger())
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}
