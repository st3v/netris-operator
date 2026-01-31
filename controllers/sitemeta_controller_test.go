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

func TestSiteMetaReconciler_SiteMetaNotFound(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	r := &SiteMetaReconciler{
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


func TestSiteMetaReconciler_DeletionWithReclaim_SkipsAPICall(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	siteMeta := &k8sv1alpha1.SiteMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "site-meta-reclaim",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SiteMetaSpec{
			ID:       100,
			Reclaim:  true, // Skip API call
			SiteName: "test-site",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, siteMeta)

	r := &SiteMetaReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
		// Cred is nil - if API call were attempted, it would panic
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "site-meta-reclaim",
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
	updated := &k8sv1alpha1.SiteMeta{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated SiteMeta: %v", err)
	}

	if len(updated.GetFinalizers()) != 0 {
		t.Errorf("expected finalizers to be cleared, got %v", updated.GetFinalizers())
	}
}

func TestSiteMetaReconciler_DeletionWithZeroID_SkipsAPICall(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	siteMeta := &k8sv1alpha1.SiteMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "site-meta-zero",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SiteMetaSpec{
			ID:       0, // Zero ID skips API call
			Reclaim:  false,
			SiteName: "test-site",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, siteMeta)

	r := &SiteMetaReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
		// Cred is nil - if API call were attempted, it would panic
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "site-meta-zero",
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

func TestSiteMetaReconciler_CreateSite(t *testing.T) {
	scheme := newTestScheme()

	siteMeta := &k8sv1alpha1.SiteMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "site-meta-create",
			Namespace:  "default",
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SiteMetaSpec{
			ID:       0, // No ID means create
			SiteName: "test-site",
		},
	}

	site := &k8sv1alpha1.Site{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-site",
			Namespace: "default",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, siteMeta, site)

	r := &SiteMetaReconciler{
		Client:     fakeClient,
		Log:        newTestLogger(),
		Scheme:     scheme,
		NStorage:   newTestStorage(nil),
		SiteClient: &MockSiteClient{},
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "site-meta-create",
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
	updated := &k8sv1alpha1.SiteMeta{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated SiteMeta: %v", err)
	}

	if updated.Spec.ID == 0 {
		t.Errorf("expected ID to be set after creation, got 0")
	}
}


func TestSiteMetaReconciler_DeletionWithID(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	siteMeta := &k8sv1alpha1.SiteMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "site-meta-delete",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SiteMetaSpec{
			ID:       100,
			Reclaim:  false, // Do not reclaim, actually delete
			SiteName: "test-site",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, siteMeta)

	r := &SiteMetaReconciler{
		Client:     fakeClient,
		Log:        newTestLogger(),
		Scheme:     scheme,
		SiteClient: &MockSiteClient{},
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "site-meta-delete",
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
	updated := &k8sv1alpha1.SiteMeta{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated SiteMeta: %v", err)
	}

	if len(updated.GetFinalizers()) != 0 {
		t.Errorf("expected finalizers to be cleared, got %v", updated.GetFinalizers())
	}
}
