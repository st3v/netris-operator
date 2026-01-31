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
	"github.com/netrisai/netriswebapi/v1/types/inventoryprofile"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestInventoryProfileMetaReconciler_InventoryProfileMetaNotFound(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	r := &InventoryProfileMetaReconciler{
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

func TestInventoryProfileMetaReconciler_SetsFinalizer(t *testing.T) {
	scheme := newTestScheme()

	inventoryProfileMeta := &k8sv1alpha1.InventoryProfileMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "inventoryprofile-meta-123",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.InventoryProfileMetaSpec{
			InventoryProfileName: "test-profile",
		},
	}

	inventoryProfileCR := &k8sv1alpha1.InventoryProfile{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-profile",
			Namespace: "default",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, inventoryProfileMeta, inventoryProfileCR)

	r := &InventoryProfileMetaReconciler{
		Client:                 fakeClient,
		Log:                    newTestLogger(),
		Scheme:                 scheme,
		NStorage:               newTestStorage(nil),
		InventoryProfileClient: &MockInventoryProfileClient{},
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "inventoryprofile-meta-123",
			Namespace: "default",
		},
	}

	_, err := r.Reconcile(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestInventoryProfileMetaReconciler_DeletionWithReclaim_SkipsAPICall(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	inventoryProfileMeta := &k8sv1alpha1.InventoryProfileMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "inventoryprofile-meta-reclaim",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.InventoryProfileMetaSpec{
			ID:                   100,
			Reclaim:              true,
			InventoryProfileName: "test-profile",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, inventoryProfileMeta)

	r := &InventoryProfileMetaReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "inventoryprofile-meta-reclaim",
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

func TestInventoryProfileMetaReconciler_DeletionWithZeroID_SkipsAPICall(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	inventoryProfileMeta := &k8sv1alpha1.InventoryProfileMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "inventoryprofile-meta-zero",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.InventoryProfileMetaSpec{
			ID:                   0,
			Reclaim:              false,
			InventoryProfileName: "test-profile",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, inventoryProfileMeta)

	r := &InventoryProfileMetaReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "inventoryprofile-meta-zero",
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

func TestInventoryProfileMetaReconciler_CreateInventoryProfile(t *testing.T) {
	scheme := newTestScheme()

	inventoryProfileMeta := &k8sv1alpha1.InventoryProfileMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "inventoryprofile-meta-create",
			Namespace:  "default",
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.InventoryProfileMetaSpec{
			ID:                   0, // No ID means create
			InventoryProfileName: "test-profile",
		},
	}

	inventoryProfileCR := &k8sv1alpha1.InventoryProfile{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-profile",
			Namespace: "default",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, inventoryProfileMeta, inventoryProfileCR)

	r := &InventoryProfileMetaReconciler{
		Client:                 fakeClient,
		Log:                    newTestLogger(),
		Scheme:                 scheme,
		NStorage:               newTestStorage(nil),
		InventoryProfileClient: &MockInventoryProfileClient{},
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "inventoryprofile-meta-create",
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
	updated := &k8sv1alpha1.InventoryProfileMeta{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated InventoryProfileMeta: %v", err)
	}

	if updated.Spec.ID == 0 {
		t.Errorf("expected ID to be set after creation, got 0")
	}
}

func TestInventoryProfileMetaReconciler_UpdateInventoryProfile(t *testing.T) {
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
			r := &InventoryProfileMetaReconciler{
				InventoryProfileClient: &MockInventoryProfileClient{UpdateErr: tt.mockErr},
			}

			profile := &inventoryprofile.ProfileW{
				ID:   1,
				Name: "test-profile",
			}

			result, err, _ := r.updateInventoryProfile(1, profile)

			if (err != nil) != tt.wantErr {
				t.Errorf("updateInventoryProfile() error = %v, wantErr %v", err, tt.wantErr)
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
