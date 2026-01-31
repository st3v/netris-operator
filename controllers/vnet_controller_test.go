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
	"context"
	"errors"
	"testing"

	k8sv1alpha1 "github.com/netrisai/netris-operator/api/v1alpha1"
	"github.com/netrisai/netriswebapi/v2/types/dhcp"
	"github.com/netrisai/netriswebapi/v2/types/site"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestVNetReconciler_VNetNotFound(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	r := &VNetReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "nonexistent-vnet",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result.Requeue || result.RequeueAfter != 0 {
		t.Errorf("expected no requeue, got %v", result)
	}
}

func TestVNetReconciler_SetsDefaultAnnotations(t *testing.T) {
	scheme := newTestScheme()

	vnet := &k8sv1alpha1.VNet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-vnet",
			Namespace: "default",
			UID:       "vnet-uid-123",
			Annotations: map[string]string{
				"test": "placeholder",
			},
		},
		Spec: k8sv1alpha1.VNetSpec{
			Owner: "admin",
			Sites: []k8sv1alpha1.VNetSite{
				{Name: "site1"},
			},
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, vnet)

	r := &VNetReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-vnet",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result.RequeueAfter != 0 {
		t.Errorf("expected no requeue delay, got %v", result.RequeueAfter)
	}

	// Verify annotations were set
	updated := &k8sv1alpha1.VNet{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated VNet: %v", err)
	}

	if updated.GetAnnotations()["resource.k8s.netris.ai/import"] != "false" {
		t.Errorf("expected import annotation 'false', got %q", updated.GetAnnotations()["resource.k8s.netris.ai/import"])
	}
	if updated.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"] != "delete" {
		t.Errorf("expected reclaimPolicy 'delete', got %q", updated.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"])
	}
}

func TestVNetReconciler_SetsFinalizer(t *testing.T) {
	scheme := newTestScheme()

	vnet := &k8sv1alpha1.VNet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-vnet",
			Namespace: "default",
			UID:       "vnet-uid-456",
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
		},
		Spec: k8sv1alpha1.VNetSpec{
			Owner: "admin",
			Sites: []k8sv1alpha1.VNetSite{
				{Name: "site1"},
			},
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, vnet)

	r := &VNetReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-vnet",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result.RequeueAfter != 0 {
		t.Errorf("expected no requeue delay, got %v", result.RequeueAfter)
	}

	// Verify finalizer was set
	updated := &k8sv1alpha1.VNet{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated VNet: %v", err)
	}

	if len(updated.GetFinalizers()) == 0 || updated.GetFinalizers()[0] != "resource.k8s.netris.ai/delete" {
		t.Errorf("expected finalizer to be set, got %v", updated.GetFinalizers())
	}
}

func TestVNetReconciler_DuplicateGatewayError(t *testing.T) {
	scheme := newTestScheme()

	vnet := &k8sv1alpha1.VNet{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-vnet",
			Namespace:  "default",
			UID:        "vnet-uid-dup",
			Generation: 1,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.VNetSpec{
			Owner: "admin",
			Sites: []k8sv1alpha1.VNetSite{
				{
					Name: "site1",
					Gateways: []k8sv1alpha1.VNetGateway{
						{Prefix: "10.0.0.1/24"},
						{Prefix: "10.0.0.1/24"}, // duplicate
					},
				},
			},
		},
	}

	// Need meta to exist so we get past meta lookup
	vnetMeta := &k8sv1alpha1.VNetMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vnet-uid-dup",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.VNetMetaSpec{
			VnetCRGeneration: 1,
			Imported:         false,
			Reclaim:          false,
			VnetName:         "test-vnet",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, vnet, vnetMeta)

	r := &VNetReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-vnet",
			Namespace: "default",
		},
	}

	// This should return an error due to duplicate gateways
	_, err := r.Reconcile(req)
	if err != nil {
		t.Errorf("expected no error (status patched), got %v", err)
	}

	// Verify status was updated with failure
	updated := &k8sv1alpha1.VNet{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated VNet: %v", err)
	}

	if updated.Status.Status != "Failure" {
		t.Errorf("expected status 'Failure', got %q", updated.Status.Status)
	}
}

func TestVNetReconciler_MetaFoundNoChanges(t *testing.T) {
	scheme := newTestScheme()

	vnet := &k8sv1alpha1.VNet{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-vnet",
			Namespace:  "default",
			UID:        "vnet-uid-abc",
			Generation: 1,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.VNetSpec{
			Owner: "admin",
			Sites: []k8sv1alpha1.VNetSite{
				{Name: "site1"},
			},
		},
	}

	vnetMeta := &k8sv1alpha1.VNetMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vnet-uid-abc",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.VNetMetaSpec{
			VnetCRGeneration: 1, // matches
			Imported:         false,
			Reclaim:          false,
			VnetName:         "test-vnet",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, vnet, vnetMeta)

	r := &VNetReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-vnet",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result.RequeueAfter == 0 {
		t.Error("expected requeue with interval")
	}
}

func TestVNetReconciler_DeletionNoMeta_ClearsFinalizer(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	vnet := &k8sv1alpha1.VNet{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-vnet",
			Namespace:         "default",
			UID:               "vnet-uid-nometa",
			DeletionTimestamp: &now,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.VNetSpec{
			Owner: "admin",
			Sites: []k8sv1alpha1.VNetSite{
				{Name: "site1"},
			},
		},
	}

	// No VNetMeta exists
	fakeClient := fake.NewFakeClientWithScheme(scheme, vnet)

	r := &VNetReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-vnet",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result.RequeueAfter != 0 || result.Requeue {
		t.Errorf("expected no requeue after deletion, got %v", result)
	}

	// Verify finalizer was cleared
	updated := &k8sv1alpha1.VNet{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated VNet: %v", err)
	}

	if len(updated.GetFinalizers()) != 0 {
		t.Errorf("expected finalizers to be cleared, got %v", updated.GetFinalizers())
	}
}

func TestVNetReconciler_DeletionWithReclaim_SkipsAPICall(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	vnet := &k8sv1alpha1.VNet{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-vnet",
			Namespace:         "default",
			UID:               "vnet-uid-reclaim",
			DeletionTimestamp: &now,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "true",
				"resource.k8s.netris.ai/reclaimPolicy": "retain",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.VNetSpec{
			Owner: "admin",
			Sites: []k8sv1alpha1.VNetSite{
				{Name: "site1"},
			},
		},
	}

	vnetMeta := &k8sv1alpha1.VNetMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "vnet-uid-reclaim",
			Namespace:  "default",
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.VNetMetaSpec{
			ID:       100,
			Reclaim:  true, // Skip API call
			VnetName: "test-vnet",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, vnet, vnetMeta)

	r := &VNetReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
		// Cred is nil - if API call were attempted, it would panic
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-vnet",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result.Requeue {
		t.Errorf("expected no requeue, got Requeue=true")
	}
}

// TestVnetToVnetMeta_BasicConversion tests the VnetToVnetMeta conversion
// with a basic VNet that has no gateways.
func TestVnetToVnetMeta_BasicConversion(t *testing.T) {
	scheme := newTestScheme()

	vnet := &k8sv1alpha1.VNet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-vnet",
			Namespace: "default",
			UID:       "vnet-uid-basic",
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
		},
		Spec: k8sv1alpha1.VNetSpec{
			Owner: "admin",
			State: "active",
			Sites: []k8sv1alpha1.VNetSite{
				{Name: "site1"},
			},
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, vnet)
	testStorage := newTestStorage([]*site.Site{
		{ID: 1, Name: "site1"},
	})

	r := &VNetReconciler{
		Client:     fakeClient,
		Log:        newTestLogger(),
		Scheme:     scheme,
		VNetClient: &MockVNetClient{},
		DHCPClient: &MockDHCPClient{Data: []*dhcp.DHCPOptionSet{{ID: 1, Name: "default"}}},
		NStorage:   testStorage,
	}

	vnetMeta, err := r.VnetToVnetMeta(vnet)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if vnetMeta.Spec.VnetName != "test-vnet" {
		t.Errorf("expected VnetName 'test-vnet', got %q", vnetMeta.Spec.VnetName)
	}
	if vnetMeta.Spec.Owner != "admin" {
		t.Errorf("expected Owner 'admin', got %q", vnetMeta.Spec.Owner)
	}
	if vnetMeta.Spec.State != "active" {
		t.Errorf("expected State 'active', got %q", vnetMeta.Spec.State)
	}
	if len(vnetMeta.Spec.Sites) != 1 || vnetMeta.Spec.Sites[0].Name != "site1" {
		t.Errorf("expected one site 'site1', got %v", vnetMeta.Spec.Sites)
	}
	if vnetMeta.Spec.Imported != false {
		t.Errorf("expected Imported false, got %v", vnetMeta.Spec.Imported)
	}
	if vnetMeta.Spec.Reclaim != false {
		t.Errorf("expected Reclaim false, got %v", vnetMeta.Spec.Reclaim)
	}
}

// TestVnetToVnetMeta_WithGateways tests the VnetToVnetMeta conversion
// with a VNet that has gateways including DHCP configuration.
func TestVnetToVnetMeta_WithGateways(t *testing.T) {
	scheme := newTestScheme()

	vnet := &k8sv1alpha1.VNet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-vnet-gw",
			Namespace: "default",
			UID:       "vnet-uid-gateway",
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
		},
		Spec: k8sv1alpha1.VNetSpec{
			Owner: "admin",
			State: "active",
			Sites: []k8sv1alpha1.VNetSite{
				{
					Name: "site1",
					Gateways: []k8sv1alpha1.VNetGateway{
						{
							Prefix:        "10.0.0.1/24",
							DHCP:          "enabled",
							DHCPStartIP:   "10.0.0.100",
							DHCPEndIP:     "10.0.0.200",
							DHCPOptionSet: "default",
						},
					},
				},
			},
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, vnet)
	testStorage := newTestStorage([]*site.Site{
		{ID: 1, Name: "site1"},
	})

	r := &VNetReconciler{
		Client:     fakeClient,
		Log:        newTestLogger(),
		Scheme:     scheme,
		VNetClient: &MockVNetClient{},
		DHCPClient: &MockDHCPClient{Data: []*dhcp.DHCPOptionSet{{ID: 42, Name: "default"}}},
		NStorage:   testStorage,
	}

	vnetMeta, err := r.VnetToVnetMeta(vnet)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if vnetMeta.Spec.VnetName != "test-vnet-gw" {
		t.Errorf("expected VnetName 'test-vnet-gw', got %q", vnetMeta.Spec.VnetName)
	}
	if len(vnetMeta.Spec.Gateways) != 1 {
		t.Fatalf("expected 1 gateway, got %d", len(vnetMeta.Spec.Gateways))
	}

	gw := vnetMeta.Spec.Gateways[0]
	if gw.Gateway != "10.0.0.1" {
		t.Errorf("expected Gateway '10.0.0.1', got %q", gw.Gateway)
	}
	if gw.GwLength != 24 {
		t.Errorf("expected GwLength 24, got %d", gw.GwLength)
	}
	if gw.Version != "ipv4" {
		t.Errorf("expected Version 'ipv4', got %q", gw.Version)
	}
	if gw.DHCP != true {
		t.Errorf("expected DHCP true, got %v", gw.DHCP)
	}
	if gw.DHCPStartIP != "10.0.0.100" {
		t.Errorf("expected DHCPStartIP '10.0.0.100', got %q", gw.DHCPStartIP)
	}
	if gw.DHCPEndIP != "10.0.0.200" {
		t.Errorf("expected DHCPEndIP '10.0.0.200', got %q", gw.DHCPEndIP)
	}
	if gw.DHCPOptionSetID != 42 {
		t.Errorf("expected DHCPOptionSetID 42, got %d", gw.DHCPOptionSetID)
	}
}

func TestVNetReconciler_deleteVNet(t *testing.T) {
	tests := []struct {
		name      string
		vnetMeta  *k8sv1alpha1.VNetMeta
		deleteErr error
		wantErr   bool
		errSubstr string
	}{
		{
			name: "success - deletes via API and CRs",
			vnetMeta: &k8sv1alpha1.VNetMeta{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vnet-meta-1",
					Namespace: "default",
				},
				Spec: k8sv1alpha1.VNetMetaSpec{
					ID:      100,
					Reclaim: false,
				},
			},
			deleteErr: nil,
			wantErr:   false,
		},
		{
			name:      "nil meta - only deletes CR",
			vnetMeta:  nil,
			deleteErr: nil,
			wantErr:   false,
		},
		{
			name: "reclaim true - skips API call",
			vnetMeta: &k8sv1alpha1.VNetMeta{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vnet-meta-reclaim",
					Namespace: "default",
				},
				Spec: k8sv1alpha1.VNetMetaSpec{
					ID:      100,
					Reclaim: true,
				},
			},
			deleteErr: nil,
			wantErr:   false,
		},
		{
			name: "zero ID - skips API call",
			vnetMeta: &k8sv1alpha1.VNetMeta{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vnet-meta-zero",
					Namespace: "default",
				},
				Spec: k8sv1alpha1.VNetMetaSpec{
					ID:      0,
					Reclaim: false,
				},
			},
			deleteErr: nil,
			wantErr:   false,
		},
		{
			name: "API error",
			vnetMeta: &k8sv1alpha1.VNetMeta{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vnet-meta-err",
					Namespace: "default",
				},
				Spec: k8sv1alpha1.VNetMetaSpec{
					ID:      100,
					Reclaim: false,
				},
			},
			deleteErr: errors.New("connection refused"),
			wantErr:   true,
			errSubstr: "connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme := newTestScheme()

			vnet := &k8sv1alpha1.VNet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-vnet",
					Namespace: "default",
				},
			}

			objects := []runtime.Object{vnet}
			if tt.vnetMeta != nil {
				objects = append(objects, tt.vnetMeta)
			}

			fakeClient := fake.NewFakeClientWithScheme(scheme, objects...)

			r := &VNetReconciler{
				Client:     fakeClient,
				Log:        newTestLogger(),
				Scheme:     scheme,
				VNetClient: &MockVNetClient{DeleteErr: tt.deleteErr},
			}

			_, err := r.deleteVNet(vnet, tt.vnetMeta)

			if (err != nil) != tt.wantErr {
				t.Errorf("deleteVNet() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && tt.errSubstr != "" {
				if !containsSubstr(err.Error(), tt.errSubstr) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errSubstr)
				}
			}
		})
	}
}
