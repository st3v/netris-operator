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
	"github.com/netrisai/netriswebapi/v2/types/vnet"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestVNetMetaReconciler_VNetMetaNotFound(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	r := &VNetMetaReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "nonexistent-meta",
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

func TestVNetMetaReconciler_DeletionWithReclaim_SkipsAPICall(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	vnetMeta := &k8sv1alpha1.VNetMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "vnet-meta-reclaim",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.VNetMetaSpec{
			ID:       100,
			Reclaim:  true,
			VnetName: "test-vnet",
		},
	}

	vnetCR := &k8sv1alpha1.VNet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-vnet",
			Namespace: "default",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, vnetMeta, vnetCR)

	r := &VNetMetaReconciler{
		Client:     fakeClient,
		Log:        newTestLogger(),
		Scheme:     scheme,
		NStorage:   newTestStorage(nil),
		VNetClient: &MockVNetClient{},
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "vnet-meta-reclaim",
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

func TestVNetMetaReconciler_DeletionWithZeroID_SkipsAPICall(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	vnetMeta := &k8sv1alpha1.VNetMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "vnet-meta-zero",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.VNetMetaSpec{
			ID:       0,
			Reclaim:  false,
			VnetName: "test-vnet",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, vnetMeta)

	r := &VNetMetaReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "vnet-meta-zero",
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

func TestVNetMetaReconciler_CreateVNet(t *testing.T) {
	scheme := newTestScheme()

	vnetMeta := &k8sv1alpha1.VNetMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "vnet-meta-create",
			Namespace:  "default",
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.VNetMetaSpec{
			ID:       0, // No ID means create
			VnetName: "test-vnet",
		},
	}

	vnetCR := &k8sv1alpha1.VNet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-vnet",
			Namespace: "default",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, vnetMeta, vnetCR)

	r := &VNetMetaReconciler{
		Client:     fakeClient,
		Log:        newTestLogger(),
		Scheme:     scheme,
		NStorage:   newTestStorage(nil),
		VNetClient: &MockVNetClient{},
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "vnet-meta-create",
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

	// Verify ID was set after creation
	updated := &k8sv1alpha1.VNetMeta{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated VNetMeta: %v", err)
	}

	if updated.Spec.ID == 0 {
		t.Errorf("expected ID to be set after creation, got 0")
	}
}

func TestVNetMetaReconciler_UpdateVNet(t *testing.T) {
	scheme := newTestScheme()

	vnetMeta := &k8sv1alpha1.VNetMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "vnet-meta-update",
			Namespace:  "default",
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.VNetMetaSpec{
			ID:       100, // Existing ID means update
			VnetName: "test-vnet",
		},
	}

	vnetCR := &k8sv1alpha1.VNet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-vnet",
			Namespace: "default",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, vnetMeta, vnetCR)

	r := &VNetMetaReconciler{
		Client:     fakeClient,
		Log:        newTestLogger(),
		Scheme:     scheme,
		NStorage:   newTestStorage(nil),
		VNetClient: &MockVNetClient{DetailedData: map[int]*vnet.VNetDetailed{100: {ID: 100, Name: "test-vnet"}}},
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "vnet-meta-update",
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

func TestVNetMetaReconciler_updateVNet(t *testing.T) {
	tests := []struct {
		name        string
		mockErr     error
		wantErr     bool
		errSubstr   string
		wantRequeue bool
	}{
		{
			name:        "success",
			mockErr:     nil,
			wantErr:     false,
			wantRequeue: false,
		},
		{
			name:        "client error",
			mockErr:     errors.New("connection refused"),
			wantErr:     true,
			errSubstr:   "connection refused",
			wantRequeue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &VNetMetaReconciler{
				VNetClient: &MockVNetClient{UpdateErr: tt.mockErr},
			}

			update := &vnet.VNetUpdate{
				Name: "test-vnet",
			}

			result, err, _ := r.updateVNet(1, update)

			if (err != nil) != tt.wantErr {
				t.Errorf("updateVNet() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && tt.errSubstr != "" {
				if !containsSubstr(err.Error(), tt.errSubstr) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errSubstr)
				}
			}
			if !tt.wantErr {
				if result.Requeue != tt.wantRequeue {
					t.Errorf("result.Requeue = %v, want %v", result.Requeue, tt.wantRequeue)
				}
				if result.RequeueAfter != 0 {
					t.Errorf("result.RequeueAfter = %v, want 0", result.RequeueAfter)
				}
			}
		})
	}
}

func TestVNetMetaReconciler_DeletionWithID(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	vnetMeta := &k8sv1alpha1.VNetMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "vnet-meta-delete",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.VNetMetaSpec{
			ID:       100,
			Reclaim:  false, // Do not reclaim, actually delete
			VnetName: "test-vnet",
		},
	}

	vnetCR := &k8sv1alpha1.VNet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-vnet",
			Namespace: "default",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, vnetMeta, vnetCR)

	r := &VNetMetaReconciler{
		Client:     fakeClient,
		Log:        newTestLogger(),
		Scheme:     scheme,
		NStorage:   newTestStorage(nil),
		VNetClient: &MockVNetClient{},
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "vnet-meta-delete",
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
