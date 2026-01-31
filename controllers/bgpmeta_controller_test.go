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

	r := &BGPMetaReconciler{
		Client:    fakeClient,
		Log:       newTestLogger(),
		Scheme:    scheme,
		NStorage:  newTestStorage(nil),
		BGPClient: &MockBGPClient{},
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
}
