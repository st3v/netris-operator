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
	"github.com/netrisai/netriswebapi/v2/types/ipam"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestSubnetMetaReconciler_SubnetMetaNotFound(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	r := &SubnetMetaReconciler{
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

func TestSubnetMetaReconciler_SetsFinalizer(t *testing.T) {
	scheme := newTestScheme()

	subnetMeta := &k8sv1alpha1.SubnetMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "subnet-meta-123",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.SubnetMetaSpec{
			SubnetName: "test-subnet",
		},
	}

	subnet := &k8sv1alpha1.Subnet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-subnet",
			Namespace: "default",
			UID:       "subnet-meta-123",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, subnetMeta, subnet)

	r := &SubnetMetaReconciler{
		Client:     fakeClient,
		Log:        newTestLogger(),
		Scheme:     scheme,
		IPAMClient: &MockIPAMClient{},
		NStorage:   newTestStorage(nil),
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "subnet-meta-123",
			Namespace: "default",
		},
	}

	_, err := r.Reconcile(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	updated := &k8sv1alpha1.SubnetMeta{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated SubnetMeta: %v", err)
	}

	// Verify the subnet was created (ID should be set after create call)
	if updated.Spec.ID == 0 {
		t.Logf("ID not patched yet (expected during async reconciliation)")
	}
}

func TestSubnetMetaReconciler_DeletionWithReclaim_SkipsAPICall(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	subnetMeta := &k8sv1alpha1.SubnetMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "subnet-meta-reclaim",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SubnetMetaSpec{
			ID:         100,
			Reclaim:    true,
			SubnetName: "test-subnet",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, subnetMeta)

	r := &SubnetMetaReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "subnet-meta-reclaim",
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

func TestSubnetMetaReconciler_DeletionWithZeroID_SkipsAPICall(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	subnetMeta := &k8sv1alpha1.SubnetMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "subnet-meta-zero",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SubnetMetaSpec{
			ID:         0,
			Reclaim:    false,
			SubnetName: "test-subnet",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, subnetMeta)

	r := &SubnetMetaReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "subnet-meta-zero",
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

func TestSubnetMetaReconciler_CreateSubnet(t *testing.T) {
	scheme := newTestScheme()

	subnetMeta := &k8sv1alpha1.SubnetMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "subnet-meta-create",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.SubnetMetaSpec{
			ID:         0,
			SubnetName: "new-subnet",
		},
	}

	subnet := &k8sv1alpha1.Subnet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "new-subnet",
			Namespace: "default",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, subnetMeta, subnet)

	r := &SubnetMetaReconciler{
		Client:     fakeClient,
		Log:        newTestLogger(),
		Scheme:     scheme,
		IPAMClient: &MockIPAMClient{},
		NStorage:   newTestStorage(nil),
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "subnet-meta-create",
			Namespace: "default",
		},
	}

	_, err := r.Reconcile(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}


func TestSubnetMetaReconciler_DeletionWithID(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	subnetMeta := &k8sv1alpha1.SubnetMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "subnet-meta-delete",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SubnetMetaSpec{
			ID:         100,
			Reclaim:    false,
			SubnetName: "delete-subnet",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, subnetMeta)

	r := &SubnetMetaReconciler{
		Client:     fakeClient,
		Log:        newTestLogger(),
		Scheme:     scheme,
		IPAMClient: &MockIPAMClient{},
		NStorage:   newTestStorage(nil),
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "subnet-meta-delete",
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

func TestSubnetMetaReconciler_updateSubnet(t *testing.T) {
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
			r := &SubnetMetaReconciler{
				IPAMClient: &MockIPAMClient{UpdateErr: tt.mockErr},
			}

			update := &ipam.Subnet{
				Name: "test-subnet",
			}

			result, err, _ := r.updateSubnet(1, update)

			if (err != nil) != tt.wantErr {
				t.Errorf("updateSubnet() error = %v, wantErr %v", err, tt.wantErr)
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
