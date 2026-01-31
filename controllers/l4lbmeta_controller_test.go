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
	"testing"

	k8sv1alpha1 "github.com/netrisai/netris-operator/api/v1alpha1"
	"github.com/netrisai/netriswebapi/v2/types/vpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestL4LBMetaReconciler_L4LBMetaNotFound(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	r := &L4LBMetaReconciler{
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


func TestL4LBMetaReconciler_DeletionWithReclaim_SkipsAPICall(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	l4lbMeta := &k8sv1alpha1.L4LBMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "l4lb-meta-reclaim",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.L4LBMetaSpec{
			ID:       100,
			Reclaim:  true,
			L4LBName: "test-l4lb",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, l4lbMeta)

	r := &L4LBMetaReconciler{
		Client:     fakeClient,
		Log:        newTestLogger(),
		Scheme:     scheme,
		L4LBClient: &MockL4LBClient{},
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "l4lb-meta-reclaim",
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

	updated := &k8sv1alpha1.L4LBMeta{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated L4LBMeta: %v", err)
	}

	if len(updated.GetFinalizers()) != 0 {
		t.Errorf("expected finalizers to be cleared, got %v", updated.GetFinalizers())
	}
}

func TestL4LBMetaReconciler_DeletionWithZeroID_SkipsAPICall(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	l4lbMeta := &k8sv1alpha1.L4LBMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "l4lb-meta-zero",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.L4LBMetaSpec{
			ID:       0,
			Reclaim:  false,
			L4LBName: "test-l4lb",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, l4lbMeta)

	r := &L4LBMetaReconciler{
		Client:     fakeClient,
		Log:        newTestLogger(),
		Scheme:     scheme,
		L4LBClient: &MockL4LBClient{},
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "l4lb-meta-zero",
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



func TestL4LBMetaReconciler_DeletionWithID(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	l4lbMeta := &k8sv1alpha1.L4LBMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "l4lb-meta-delete",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.L4LBMetaSpec{
			ID:       100,
			Reclaim:  false, // Do not reclaim, actually delete
			L4LBName: "test-l4lb",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, l4lbMeta)

	r := &L4LBMetaReconciler{
		Client:     fakeClient,
		Log:        newTestLogger(),
		Scheme:     scheme,
		L4LBClient: &MockL4LBClient{},
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "l4lb-meta-delete",
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

	// Verify finalizer was cleared
	updated := &k8sv1alpha1.L4LBMeta{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated L4LBMeta: %v", err)
	}

	if len(updated.GetFinalizers()) != 0 {
		t.Errorf("expected finalizers to be cleared, got %v", updated.GetFinalizers())
	}
}

func TestPopulateMetaVPC(t *testing.T) {
	tests := []struct {
		name        string
		vpcID       int
		vpcs        []*vpc.VPC
		initialMeta *k8sv1alpha1.L4LBMeta
		l4lbCR      *k8sv1alpha1.L4LB
		wantVPCID   int
		wantVPCName string
		wantErr     bool
	}{
		{
			name:  "VPCID is 0 - clears VPC info",
			vpcID: 0,
			vpcs:  nil,
			initialMeta: &k8sv1alpha1.L4LBMeta{
				Spec: k8sv1alpha1.L4LBMetaSpec{
					VPCID:   10,
					VPCName: "old-vpc",
				},
			},
			l4lbCR:      &k8sv1alpha1.L4LB{},
			wantVPCID:   0,
			wantVPCName: "",
			wantErr:     false,
		},
		{
			name:  "VPC already set to same ID - no change",
			vpcID: 5,
			vpcs: []*vpc.VPC{
				{ID: 5, Name: "existing-vpc"},
			},
			initialMeta: &k8sv1alpha1.L4LBMeta{
				Spec: k8sv1alpha1.L4LBMetaSpec{
					VPCID:   5,
					VPCName: "existing-vpc",
				},
			},
			l4lbCR:      &k8sv1alpha1.L4LB{},
			wantVPCID:   5,
			wantVPCName: "existing-vpc",
			wantErr:     false,
		},
		{
			name:  "VPC found and set",
			vpcID: 10,
			vpcs: []*vpc.VPC{
				{ID: 10, Name: "my-vpc"},
			},
			initialMeta: &k8sv1alpha1.L4LBMeta{
				Spec: k8sv1alpha1.L4LBMetaSpec{
					VPCID:   0,
					VPCName: "",
				},
			},
			l4lbCR:      &k8sv1alpha1.L4LB{},
			wantVPCID:   10,
			wantVPCName: "my-vpc",
			wantErr:     false,
		},
		// Note: "VPC not found" case cannot be tested without mocking the download
		// function, as VPCStorage.FindByID attempts to download on cache miss.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &L4LBMetaReconciler{
				VPCID:    tt.vpcID,
				NStorage: newTestStorageWithVPCs(tt.vpcs),
			}

			err := r.populateMetaVPC(tt.initialMeta, tt.l4lbCR)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if tt.initialMeta.Spec.VPCID != tt.wantVPCID {
					t.Errorf("VPCID = %d, want %d", tt.initialMeta.Spec.VPCID, tt.wantVPCID)
				}
				if tt.initialMeta.Spec.VPCName != tt.wantVPCName {
					t.Errorf("VPCName = %q, want %q", tt.initialMeta.Spec.VPCName, tt.wantVPCName)
				}
			}
		})
	}
}
