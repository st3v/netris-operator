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
	"github.com/netrisai/netriswebapi/v1/types/tenant"
	"github.com/netrisai/netriswebapi/v2/types/site"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestSubnetReconciler_SubnetNotFound(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	r := &SubnetReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "nonexistent-subnet",
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

func TestSubnetReconciler_SetsDefaultAnnotations(t *testing.T) {
	scheme := newTestScheme()

	subnet := &k8sv1alpha1.Subnet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-subnet",
			Namespace: "default",
			UID:       "subnet-uid-123",
			Annotations: map[string]string{
				"test": "placeholder",
			},
		},
		Spec: k8sv1alpha1.SubnetSpec{
			Prefix:  "10.0.0.0/24",
			Tenant:  "admin",
			Purpose: "common",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, subnet)

	r := &SubnetReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-subnet",
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

	updated := &k8sv1alpha1.Subnet{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated Subnet: %v", err)
	}

	if updated.GetAnnotations()["resource.k8s.netris.ai/import"] != "false" {
		t.Errorf("expected import annotation 'false', got %q", updated.GetAnnotations()["resource.k8s.netris.ai/import"])
	}
	if updated.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"] != "delete" {
		t.Errorf("expected reclaimPolicy 'delete', got %q", updated.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"])
	}
}

func TestSubnetReconciler_SetsFinalizer(t *testing.T) {
	scheme := newTestScheme()

	subnet := &k8sv1alpha1.Subnet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-subnet",
			Namespace: "default",
			UID:       "subnet-uid-456",
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
		},
		Spec: k8sv1alpha1.SubnetSpec{
			Prefix:  "10.0.0.0/24",
			Tenant:  "admin",
			Purpose: "common",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, subnet)

	r := &SubnetReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-subnet",
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

	updated := &k8sv1alpha1.Subnet{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated Subnet: %v", err)
	}

	if len(updated.GetFinalizers()) == 0 || updated.GetFinalizers()[0] != "resource.k8s.netris.ai/delete" {
		t.Errorf("expected finalizer to be set, got %v", updated.GetFinalizers())
	}
}

func TestSubnetReconciler_MetaFoundNoChanges(t *testing.T) {
	scheme := newTestScheme()

	subnet := &k8sv1alpha1.Subnet{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-subnet",
			Namespace:  "default",
			UID:        "subnet-uid-abc",
			Generation: 1,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SubnetSpec{
			Prefix:  "10.0.0.0/24",
			Tenant:  "admin",
			Purpose: "common",
		},
	}

	subnetMeta := &k8sv1alpha1.SubnetMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "subnet-uid-abc",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.SubnetMetaSpec{
			SubnetCRGeneration: 1,
			Imported:           false,
			Reclaim:            false,
			SubnetName:         "test-subnet",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, subnet, subnetMeta)

	r := &SubnetReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-subnet",
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

func TestSubnetReconciler_DeletionNoMeta_ClearsFinalizer(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	subnet := &k8sv1alpha1.Subnet{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-subnet",
			Namespace:         "default",
			UID:               "subnet-uid-nometa",
			DeletionTimestamp: &now,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SubnetSpec{
			Prefix:  "10.0.0.0/24",
			Tenant:  "admin",
			Purpose: "common",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, subnet)

	r := &SubnetReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-subnet",
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

	updated := &k8sv1alpha1.Subnet{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated Subnet: %v", err)
	}

	if len(updated.GetFinalizers()) != 0 {
		t.Errorf("expected finalizers to be cleared, got %v", updated.GetFinalizers())
	}
}

func TestSubnetReconciler_DeletionWithReclaim_SkipsAPICall(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	subnet := &k8sv1alpha1.Subnet{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-subnet",
			Namespace:         "default",
			UID:               "subnet-uid-reclaim",
			DeletionTimestamp: &now,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "true",
				"resource.k8s.netris.ai/reclaimPolicy": "retain",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SubnetSpec{
			Prefix:  "10.0.0.0/24",
			Tenant:  "admin",
			Purpose: "common",
		},
	}

	subnetMeta := &k8sv1alpha1.SubnetMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "subnet-uid-reclaim",
			Namespace:  "default",
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SubnetMetaSpec{
			ID:         100,
			Reclaim:    true,
			SubnetName: "test-subnet",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, subnet, subnetMeta)

	r := &SubnetReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-subnet",
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

func TestSubnetReconciler_CreatesMeta(t *testing.T) {
	scheme := newTestScheme()

	subnet := &k8sv1alpha1.Subnet{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-subnet",
			Namespace:  "default",
			UID:        "subnet-uid-createsmeta",
			Generation: 1,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SubnetSpec{
			Tenant: "admin",
			Sites:  []string{"dc1"},
			Prefix: "10.0.0.0/24",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, subnet)

	// Create test storage with sites and tenants
	testStorage := newTestStorage([]*site.Site{
		{ID: 1, Name: "dc1"},
	})
	testStorage.TenantsStorage.Tenants = []*tenant.Tenant{
		{ID: 1, Name: "admin"},
	}

	r := &SubnetReconciler{
		Client:     fakeClient,
		Log:        newTestLogger(),
		Scheme:     scheme,
		IPAMClient: &MockIPAMClient{},
		NStorage:   testStorage,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-subnet",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result.RequeueAfter == 0 {
		t.Error("expected requeue with interval after creating meta")
	}

	// Verify SubnetMeta was created
	subnetMeta := &k8sv1alpha1.SubnetMeta{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name:      "subnet-uid-createsmeta",
		Namespace: "default",
	}, subnetMeta)
	if err != nil {
		t.Fatalf("expected SubnetMeta to be created, got error: %v", err)
	}

	if subnetMeta.Spec.SubnetName != "test-subnet" {
		t.Errorf("expected SubnetMeta.Spec.SubnetName to be 'test-subnet', got %q", subnetMeta.Spec.SubnetName)
	}
}

func TestSubnetReconciler_MetaFoundGenerationChanged(t *testing.T) {
	scheme := newTestScheme()

	subnet := &k8sv1alpha1.Subnet{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-subnet",
			Namespace:  "default",
			UID:        "subnet-uid-genchange",
			Generation: 2, // Generation changed
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SubnetSpec{
			Tenant: "admin",
			Sites:  []string{"dc1"},
			Prefix: "10.0.0.0/24",
		},
	}

	subnetMeta := &k8sv1alpha1.SubnetMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "subnet-uid-genchange",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.SubnetMetaSpec{
			SubnetCRGeneration: 1, // Old generation
			Imported:           false,
			Reclaim:            false,
			SubnetName:         "test-subnet",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, subnet, subnetMeta)

	// Create test storage with sites and tenants
	testStorage := newTestStorage([]*site.Site{
		{ID: 1, Name: "dc1"},
	})
	testStorage.TenantsStorage.Tenants = []*tenant.Tenant{
		{ID: 1, Name: "admin"},
	}

	r := &SubnetReconciler{
		Client:     fakeClient,
		Log:        newTestLogger(),
		Scheme:     scheme,
		IPAMClient: &MockIPAMClient{},
		NStorage:   testStorage,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-subnet",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result.RequeueAfter == 0 {
		t.Error("expected requeue with interval after updating meta")
	}

	// Verify SubnetMeta was updated with new generation
	updatedMeta := &k8sv1alpha1.SubnetMeta{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name:      "subnet-uid-genchange",
		Namespace: "default",
	}, updatedMeta)
	if err != nil {
		t.Fatalf("failed to get updated SubnetMeta: %v", err)
	}

	if updatedMeta.Spec.SubnetCRGeneration != 2 {
		t.Errorf("expected SubnetMeta.Spec.SubnetCRGeneration to be 2, got %d", updatedMeta.Spec.SubnetCRGeneration)
	}
}
