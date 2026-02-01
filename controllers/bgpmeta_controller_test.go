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
	"github.com/netrisai/netriswebapi/v2/types/bgp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestBGPMetaReconciler_BGPMetaNotFound(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	r := &BGPMetaReconciler{
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

func TestBGPMetaReconciler_DeletionWithReclaim_SkipsAPICall(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	bgpMeta := &k8sv1alpha1.BGPMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "bgp-meta-reclaim",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.BGPMetaSpec{
			ID:      100,
			Reclaim: true,
			BGPName: "test-bgp",
		},
	}

	bgpCR := &k8sv1alpha1.BGP{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-bgp",
			Namespace: "default",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, bgpMeta, bgpCR)

	r := &BGPMetaReconciler{
		Client:    fakeClient,
		Log:       newTestLogger(),
		Scheme:    scheme,
		NStorage:  newTestStorage(nil),
		BGPClient: &MockBGPClient{},
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "bgp-meta-reclaim",
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

func TestBGPMetaReconciler_DeletionWithZeroID_SkipsAPICall(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	bgpMeta := &k8sv1alpha1.BGPMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "bgp-meta-zero",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.BGPMetaSpec{
			ID:      0,
			Reclaim: false,
			BGPName: "test-bgp",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, bgpMeta)

	r := &BGPMetaReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "bgp-meta-zero",
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

func TestBGPMetaReconciler_CreateBGP(t *testing.T) {
	scheme := newTestScheme()

	bgpMeta := &k8sv1alpha1.BGPMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "bgp-meta-create",
			Namespace:  "default",
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.BGPMetaSpec{
			ID:      0, // No ID means create
			BGPName: "test-bgp",
		},
	}

	bgpCR := &k8sv1alpha1.BGP{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-bgp",
			Namespace: "default",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, bgpMeta, bgpCR)

	r := &BGPMetaReconciler{
		Client:    fakeClient,
		Log:       newTestLogger(),
		Scheme:    scheme,
		NStorage:  newTestStorage(nil),
		BGPClient: &MockBGPClient{},
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "bgp-meta-create",
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
	updated := &k8sv1alpha1.BGPMeta{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated BGPMeta: %v", err)
	}

	if updated.Spec.ID == 0 {
		t.Errorf("expected ID to be set after creation, got 0")
	}
}

func TestBGPMetaReconciler_DeletionWithID(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	bgpMeta := &k8sv1alpha1.BGPMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "bgp-meta-delete",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.BGPMetaSpec{
			ID:      100,
			Reclaim: false, // Do not reclaim, actually delete
			BGPName: "test-bgp",
		},
	}

	bgpCR := &k8sv1alpha1.BGP{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-bgp",
			Namespace: "default",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, bgpMeta, bgpCR)

	mockClient := &MockBGPClient{}
	r := &BGPMetaReconciler{
		Client:    fakeClient,
		Log:       newTestLogger(),
		Scheme:    scheme,
		NStorage:  newTestStorage(nil),
		BGPClient: mockClient,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "bgp-meta-delete",
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

	// Note: BGPMeta controller returns early when DeletionTimestamp is set.
	// Actual deletion is handled by the parent BGP controller.
	// The BGPMeta reconciler should NOT call the API delete.
	if mockClient.DeleteCalled {
		t.Error("BGPMeta reconciler should not call API Delete - that's handled by BGP controller")
	}
}

func TestUpdateBGP(t *testing.T) {
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
			mockClient := &MockBGPClient{UpdateErr: tt.mockErr}

			update := &bgp.EBGPUpdate{
				Name: "test-bgp",
			}

			result, err, _ := updateBGP(1, update, mockClient)

			if (err != nil) != tt.wantErr {
				t.Errorf("updateBGP() error = %v, wantErr %v", err, tt.wantErr)
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

			// Verify mock was called
			if !mockClient.UpdateCalled {
				t.Error("expected Update to be called")
			}
			if mockClient.LastUpdateID != 1 {
				t.Errorf("expected Update called with ID 1, got %d", mockClient.LastUpdateID)
			}
		})
	}
}
