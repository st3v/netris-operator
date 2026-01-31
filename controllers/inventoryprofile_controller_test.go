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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestInventoryProfileReconciler_InventoryProfileNotFound(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	r := &InventoryProfileReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "nonexistent-profile",
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

func TestInventoryProfileReconciler_SetsDefaultAnnotations(t *testing.T) {
	scheme := newTestScheme()

	profile := &k8sv1alpha1.InventoryProfile{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-profile",
			Namespace: "default",
			UID:       "profile-uid-123",
			Annotations: map[string]string{
				"test": "placeholder",
			},
		},
		Spec: k8sv1alpha1.InventoryProfileSpec{
			Description:      "Test profile",
			Timezone:         "UTC",
			AllowSSHFromIPv4: []string{"0.0.0.0/0"},
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, profile)

	r := &InventoryProfileReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-profile",
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

	updated := &k8sv1alpha1.InventoryProfile{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated InventoryProfile: %v", err)
	}

	if updated.GetAnnotations()["resource.k8s.netris.ai/import"] != "false" {
		t.Errorf("expected import annotation 'false', got %q", updated.GetAnnotations()["resource.k8s.netris.ai/import"])
	}
	if updated.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"] != "delete" {
		t.Errorf("expected reclaimPolicy 'delete', got %q", updated.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"])
	}
}

func TestInventoryProfileReconciler_SetsFinalizer(t *testing.T) {
	scheme := newTestScheme()

	profile := &k8sv1alpha1.InventoryProfile{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-profile",
			Namespace: "default",
			UID:       "profile-uid-456",
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
		},
		Spec: k8sv1alpha1.InventoryProfileSpec{
			Description:      "Test profile",
			Timezone:         "UTC",
			AllowSSHFromIPv4: []string{"0.0.0.0/0"},
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, profile)

	r := &InventoryProfileReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-profile",
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

	updated := &k8sv1alpha1.InventoryProfile{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated InventoryProfile: %v", err)
	}

	if len(updated.GetFinalizers()) == 0 || updated.GetFinalizers()[0] != "resource.k8s.netris.ai/delete" {
		t.Errorf("expected finalizer to be set, got %v", updated.GetFinalizers())
	}
}

func TestInventoryProfileReconciler_MetaFoundNoChanges(t *testing.T) {
	scheme := newTestScheme()

	profile := &k8sv1alpha1.InventoryProfile{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-profile",
			Namespace:  "default",
			UID:        "profile-uid-abc",
			Generation: 1,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.InventoryProfileSpec{
			Description:      "Test profile",
			Timezone:         "UTC",
			AllowSSHFromIPv4: []string{"0.0.0.0/0"},
		},
	}

	profileMeta := &k8sv1alpha1.InventoryProfileMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "profile-uid-abc",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.InventoryProfileMetaSpec{
			InventoryProfileCRGeneration: 1,
			Imported:                     false,
			Reclaim:                      false,
			InventoryProfileName:         "test-profile",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, profile, profileMeta)

	r := &InventoryProfileReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-profile",
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

func TestInventoryProfileReconciler_DeletionNoMeta_ClearsFinalizer(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	profile := &k8sv1alpha1.InventoryProfile{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-profile",
			Namespace:         "default",
			UID:               "profile-uid-nometa",
			DeletionTimestamp: &now,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.InventoryProfileSpec{
			Description:      "Test profile",
			Timezone:         "UTC",
			AllowSSHFromIPv4: []string{"0.0.0.0/0"},
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, profile)

	r := &InventoryProfileReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-profile",
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

	updated := &k8sv1alpha1.InventoryProfile{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated InventoryProfile: %v", err)
	}

	if len(updated.GetFinalizers()) != 0 {
		t.Errorf("expected finalizers to be cleared, got %v", updated.GetFinalizers())
	}
}

func TestInventoryProfileReconciler_DeletionWithReclaim_SkipsAPICall(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	profile := &k8sv1alpha1.InventoryProfile{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-profile",
			Namespace:         "default",
			UID:               "profile-uid-reclaim",
			DeletionTimestamp: &now,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "true",
				"resource.k8s.netris.ai/reclaimPolicy": "retain",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.InventoryProfileSpec{
			Description:      "Test profile",
			Timezone:         "UTC",
			AllowSSHFromIPv4: []string{"0.0.0.0/0"},
		},
	}

	profileMeta := &k8sv1alpha1.InventoryProfileMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "profile-uid-reclaim",
			Namespace:  "default",
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.InventoryProfileMetaSpec{
			ID:                   100,
			Reclaim:              true,
			InventoryProfileName: "test-profile",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, profile, profileMeta)

	r := &InventoryProfileReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-profile",
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

func TestInventoryProfileReconciler_CreatesMeta(t *testing.T) {
	scheme := newTestScheme()

	ip := &k8sv1alpha1.InventoryProfile{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-inventory-profile",
			Namespace:  "default",
			UID:        "inventoryprofile-uid-createsmeta",
			Generation: 1,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.InventoryProfileSpec{
			Description: "test profile",
			Timezone:    "UTC",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, ip)

	// Create test storage
	testStorage := newTestStorage(nil)

	r := &InventoryProfileReconciler{
		Client:                 fakeClient,
		Log:                    newTestLogger(),
		Scheme:                 scheme,
		InventoryProfileClient: &MockInventoryProfileClient{},
		NStorage:               testStorage,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-inventory-profile",
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

	// Verify InventoryProfileMeta was created
	ipMeta := &k8sv1alpha1.InventoryProfileMeta{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name:      "inventoryprofile-uid-createsmeta",
		Namespace: "default",
	}, ipMeta)
	if err != nil {
		t.Fatalf("expected InventoryProfileMeta to be created, got error: %v", err)
	}

	if ipMeta.Spec.InventoryProfileName != "test-inventory-profile" {
		t.Errorf("expected InventoryProfileMeta.Spec.InventoryProfileName to be 'test-inventory-profile', got %q", ipMeta.Spec.InventoryProfileName)
	}
}

func TestInventoryProfileReconciler_MetaFoundGenerationChanged(t *testing.T) {
	scheme := newTestScheme()

	ip := &k8sv1alpha1.InventoryProfile{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-inventory-profile",
			Namespace:  "default",
			UID:        "inventoryprofile-uid-genchange",
			Generation: 2, // Generation changed
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.InventoryProfileSpec{
			Description: "test profile",
			Timezone:    "UTC",
		},
	}

	ipMeta := &k8sv1alpha1.InventoryProfileMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "inventoryprofile-uid-genchange",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.InventoryProfileMetaSpec{
			InventoryProfileCRGeneration: 1, // Old generation
			Imported:                     false,
			Reclaim:                      false,
			InventoryProfileName:         "test-inventory-profile",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, ip, ipMeta)

	// Create test storage
	testStorage := newTestStorage(nil)

	r := &InventoryProfileReconciler{
		Client:                 fakeClient,
		Log:                    newTestLogger(),
		Scheme:                 scheme,
		InventoryProfileClient: &MockInventoryProfileClient{},
		NStorage:               testStorage,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-inventory-profile",
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

	// Verify InventoryProfileMeta was updated with new generation
	updatedMeta := &k8sv1alpha1.InventoryProfileMeta{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name:      "inventoryprofile-uid-genchange",
		Namespace: "default",
	}, updatedMeta)
	if err != nil {
		t.Fatalf("failed to get updated InventoryProfileMeta: %v", err)
	}

	if updatedMeta.Spec.InventoryProfileCRGeneration != 2 {
		t.Errorf("expected InventoryProfileMeta.Spec.InventoryProfileCRGeneration to be 2, got %d", updatedMeta.Spec.InventoryProfileCRGeneration)
	}
}
