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
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestPatchL4LBStatus(t *testing.T) {
	scheme := newTestScheme()

	l4lb := &k8sv1alpha1.L4LB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-l4lb",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.L4LBSpec{
			State: "active",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, l4lb)

	u := &uniReconciler{
		Client:      fakeClient,
		Logger:      newTestLogger(),
		DebugLogger: newTestLogger(),
		NStorage:    newTestStorage(nil),
	}

	result, err := u.patchL4LBStatus(l4lb, "OK", "Success")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result.RequeueAfter == 0 {
		t.Errorf("expected RequeueAfter to be set")
	}

	// Verify status was set on the object
	if l4lb.Status.Status != "OK" {
		t.Errorf("Status: got %q, expected %q", l4lb.Status.Status, "OK")
	}
	if l4lb.Status.Message != "Success" {
		t.Errorf("Message: got %q, expected %q", l4lb.Status.Message, "Success")
	}
	if l4lb.Status.State != "active" {
		t.Errorf("State: got %q, expected %q", l4lb.Status.State, "active")
	}
}

func TestPatchL4LBStatus_DefaultState(t *testing.T) {
	scheme := newTestScheme()

	l4lb := &k8sv1alpha1.L4LB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-l4lb",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.L4LBSpec{
			// State is empty, should default to "active"
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, l4lb)

	u := &uniReconciler{
		Client:      fakeClient,
		Logger:      newTestLogger(),
		DebugLogger: newTestLogger(),
		NStorage:    newTestStorage(nil),
	}

	_, err := u.patchL4LBStatus(l4lb, "OK", "Success")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if l4lb.Status.State != "active" {
		t.Errorf("State: got %q, expected %q (default)", l4lb.Status.State, "active")
	}
}

func TestPatchL4LB(t *testing.T) {
	scheme := newTestScheme()

	l4lb := &k8sv1alpha1.L4LB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-l4lb",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.L4LBSpec{
			State: "active",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, l4lb)

	u := &uniReconciler{
		Client:      fakeClient,
		Logger:      newTestLogger(),
		DebugLogger: newTestLogger(),
		NStorage:    newTestStorage(nil),
	}

	// Modify the object
	l4lb.Spec.State = "disabled"

	result, err := u.patchL4LB(l4lb)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result.RequeueAfter == 0 {
		t.Errorf("expected RequeueAfter to be set")
	}

	// Verify the object was updated in the fake client
	updated := &k8sv1alpha1.L4LB{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{Name: "test-l4lb", Namespace: "default"}, updated)
	if err != nil {
		t.Fatalf("failed to get updated L4LB: %v", err)
	}

	if updated.Spec.State != "disabled" {
		t.Errorf("Spec.State: got %q, expected %q", updated.Spec.State, "disabled")
	}
}

func TestPatchSiteStatus(t *testing.T) {
	scheme := newTestScheme()

	site := &k8sv1alpha1.Site{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-site",
			Namespace: "default",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, site)

	u := &uniReconciler{
		Client:      fakeClient,
		Logger:      newTestLogger(),
		DebugLogger: newTestLogger(),
		NStorage:    newTestStorage(nil),
	}

	result, err := u.patchSiteStatus(site, "OK", "Site created")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result.RequeueAfter == 0 {
		t.Errorf("expected RequeueAfter to be set")
	}

	if site.Status.Status != "OK" {
		t.Errorf("Status: got %q, expected %q", site.Status.Status, "OK")
	}
	if site.Status.Message != "Site created" {
		t.Errorf("Message: got %q, expected %q", site.Status.Message, "Site created")
	}
}

func TestPatchAllocationStatus(t *testing.T) {
	scheme := newTestScheme()

	allocation := &k8sv1alpha1.Allocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-allocation",
			Namespace: "default",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, allocation)

	u := &uniReconciler{
		Client:      fakeClient,
		Logger:      newTestLogger(),
		DebugLogger: newTestLogger(),
		NStorage:    newTestStorage(nil),
	}

	result, err := u.patchAllocationStatus(allocation, "OK", "Allocation ready")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result.RequeueAfter == 0 {
		t.Errorf("expected RequeueAfter to be set")
	}

	if allocation.Status.Status != "OK" {
		t.Errorf("Status: got %q, expected %q", allocation.Status.Status, "OK")
	}
	if allocation.Status.Message != "Allocation ready" {
		t.Errorf("Message: got %q, expected %q", allocation.Status.Message, "Allocation ready")
	}
}

func TestPatchSubnetStatus(t *testing.T) {
	scheme := newTestScheme()

	subnet := &k8sv1alpha1.Subnet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-subnet",
			Namespace: "default",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, subnet)

	u := &uniReconciler{
		Client:      fakeClient,
		Logger:      newTestLogger(),
		DebugLogger: newTestLogger(),
		NStorage:    newTestStorage(nil),
	}

	result, err := u.patchSubnetStatus(subnet, "OK", "Subnet ready")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result.RequeueAfter == 0 {
		t.Errorf("expected RequeueAfter to be set")
	}

	if subnet.Status.Status != "OK" {
		t.Errorf("Status: got %q, expected %q", subnet.Status.Status, "OK")
	}
	if subnet.Status.Message != "Subnet ready" {
		t.Errorf("Message: got %q, expected %q", subnet.Status.Message, "Subnet ready")
	}
}

func TestPatchVNetStatus(t *testing.T) {
	scheme := newTestScheme()

	vnet := &k8sv1alpha1.VNet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-vnet",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.VNetSpec{
			State: "active",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, vnet)

	u := &uniReconciler{
		Client:      fakeClient,
		Logger:      newTestLogger(),
		DebugLogger: newTestLogger(),
		NStorage:    newTestStorage(nil),
	}

	result, err := u.patchVNetStatus(vnet, "OK", "VNet ready")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result.RequeueAfter == 0 {
		t.Errorf("expected RequeueAfter to be set")
	}

	if vnet.Status.Status != "OK" {
		t.Errorf("Status: got %q, expected %q", vnet.Status.Status, "OK")
	}
	if vnet.Status.Message != "VNet ready" {
		t.Errorf("Message: got %q, expected %q", vnet.Status.Message, "VNet ready")
	}
}

func TestPatchBGPStatus(t *testing.T) {
	scheme := newTestScheme()

	bgp := &k8sv1alpha1.BGP{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-bgp",
			Namespace: "default",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, bgp)

	u := &uniReconciler{
		Client:      fakeClient,
		Logger:      newTestLogger(),
		DebugLogger: newTestLogger(),
		NStorage:    newTestStorage(nil),
	}

	result, err := u.patchBGPStatus(bgp, "OK", "BGP ready")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result.RequeueAfter == 0 {
		t.Errorf("expected RequeueAfter to be set")
	}

	if bgp.Status.Status != "OK" {
		t.Errorf("Status: got %q, expected %q", bgp.Status.Status, "OK")
	}
	if bgp.Status.Message != "BGP ready" {
		t.Errorf("Message: got %q, expected %q", bgp.Status.Message, "BGP ready")
	}
}

func TestPatchNatStatus(t *testing.T) {
	scheme := newTestScheme()

	nat := &k8sv1alpha1.Nat{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-nat",
			Namespace: "default",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, nat)

	u := &uniReconciler{
		Client:      fakeClient,
		Logger:      newTestLogger(),
		DebugLogger: newTestLogger(),
		NStorage:    newTestStorage(nil),
	}

	result, err := u.patchNatStatus(nat, "OK", "NAT ready")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result.RequeueAfter == 0 {
		t.Errorf("expected RequeueAfter to be set")
	}

	if nat.Status.Status != "OK" {
		t.Errorf("Status: got %q, expected %q", nat.Status.Status, "OK")
	}
	if nat.Status.Message != "NAT ready" {
		t.Errorf("Message: got %q, expected %q", nat.Status.Message, "NAT ready")
	}
}

func TestPatchSoftgateStatus(t *testing.T) {
	scheme := newTestScheme()

	softgate := &k8sv1alpha1.Softgate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-softgate",
			Namespace: "default",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, softgate)

	u := &uniReconciler{
		Client:      fakeClient,
		Logger:      newTestLogger(),
		DebugLogger: newTestLogger(),
		NStorage:    newTestStorage(nil),
	}

	result, err := u.patchSoftgateStatus(softgate, "OK", "Softgate ready")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result.RequeueAfter == 0 {
		t.Errorf("expected RequeueAfter to be set")
	}

	if softgate.Status.Status != "OK" {
		t.Errorf("Status: got %q, expected %q", softgate.Status.Status, "OK")
	}
	if softgate.Status.Message != "Softgate ready" {
		t.Errorf("Message: got %q, expected %q", softgate.Status.Message, "Softgate ready")
	}
}

func TestPatchSwitchStatus(t *testing.T) {
	scheme := newTestScheme()

	sw := &k8sv1alpha1.Switch{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-switch",
			Namespace: "default",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, sw)

	u := &uniReconciler{
		Client:      fakeClient,
		Logger:      newTestLogger(),
		DebugLogger: newTestLogger(),
		NStorage:    newTestStorage(nil),
	}

	result, err := u.patchSwitchStatus(sw, "OK", "Switch ready")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result.RequeueAfter == 0 {
		t.Errorf("expected RequeueAfter to be set")
	}

	if sw.Status.Status != "OK" {
		t.Errorf("Status: got %q, expected %q", sw.Status.Status, "OK")
	}
	if sw.Status.Message != "Switch ready" {
		t.Errorf("Message: got %q, expected %q", sw.Status.Message, "Switch ready")
	}
}

func TestPatchControllerStatus(t *testing.T) {
	scheme := newTestScheme()

	controller := &k8sv1alpha1.Controller{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-controller",
			Namespace: "default",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, controller)

	u := &uniReconciler{
		Client:      fakeClient,
		Logger:      newTestLogger(),
		DebugLogger: newTestLogger(),
		NStorage:    newTestStorage(nil),
	}

	result, err := u.patchControllerStatus(controller, "OK", "Controller ready")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result.RequeueAfter == 0 {
		t.Errorf("expected RequeueAfter to be set")
	}

	if controller.Status.Status != "OK" {
		t.Errorf("Status: got %q, expected %q", controller.Status.Status, "OK")
	}
	if controller.Status.Message != "Controller ready" {
		t.Errorf("Message: got %q, expected %q", controller.Status.Message, "Controller ready")
	}
}

func TestPatchInventoryProfileStatus(t *testing.T) {
	scheme := newTestScheme()

	profile := &k8sv1alpha1.InventoryProfile{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-profile",
			Namespace: "default",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, profile)

	u := &uniReconciler{
		Client:      fakeClient,
		Logger:      newTestLogger(),
		DebugLogger: newTestLogger(),
		NStorage:    newTestStorage(nil),
	}

	result, err := u.patchInventoryProfileStatus(profile, "OK", "Profile ready")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result.RequeueAfter == 0 {
		t.Errorf("expected RequeueAfter to be set")
	}

	if profile.Status.Status != "OK" {
		t.Errorf("Status: got %q, expected %q", profile.Status.Status, "OK")
	}
	if profile.Status.Message != "Profile ready" {
		t.Errorf("Message: got %q, expected %q", profile.Status.Message, "Profile ready")
	}
}

func TestPatchSoftgate(t *testing.T) {
	scheme := newTestScheme()

	softgate := &k8sv1alpha1.Softgate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-softgate",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.SoftgateSpec{
			MainIP: "10.0.0.1",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, softgate)

	u := &uniReconciler{
		Client:      fakeClient,
		Logger:      newTestLogger(),
		DebugLogger: newTestLogger(),
		NStorage:    newTestStorage(nil),
	}

	// Modify the object
	softgate.Spec.MainIP = "10.0.0.2"

	result, err := u.patchSoftgate(softgate)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result.RequeueAfter == 0 {
		t.Errorf("expected RequeueAfter to be set")
	}

	// Verify the object was updated
	updated := &k8sv1alpha1.Softgate{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{Name: "test-softgate", Namespace: "default"}, updated)
	if err != nil {
		t.Fatalf("failed to get updated Softgate: %v", err)
	}

	if updated.Spec.MainIP != "10.0.0.2" {
		t.Errorf("MainIP: got %q, expected %q", updated.Spec.MainIP, "10.0.0.2")
	}
}

func TestPatchSwitch(t *testing.T) {
	scheme := newTestScheme()

	sw := &k8sv1alpha1.Switch{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-switch",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.SwitchSpec{
			MainIP: "10.0.0.1",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, sw)

	u := &uniReconciler{
		Client:      fakeClient,
		Logger:      newTestLogger(),
		DebugLogger: newTestLogger(),
		NStorage:    newTestStorage(nil),
	}

	// Modify the object
	sw.Spec.MainIP = "10.0.0.2"

	result, err := u.patchSwitch(sw)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result.RequeueAfter == 0 {
		t.Errorf("expected RequeueAfter to be set")
	}

	// Verify the object was updated
	updated := &k8sv1alpha1.Switch{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{Name: "test-switch", Namespace: "default"}, updated)
	if err != nil {
		t.Fatalf("failed to get updated Switch: %v", err)
	}

	if updated.Spec.MainIP != "10.0.0.2" {
		t.Errorf("MainIP: got %q, expected %q", updated.Spec.MainIP, "10.0.0.2")
	}
}

func TestPatchController(t *testing.T) {
	scheme := newTestScheme()

	controller := &k8sv1alpha1.Controller{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-controller",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.ControllerSpec{
			MainIP: "10.0.0.1",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, controller)

	u := &uniReconciler{
		Client:      fakeClient,
		Logger:      newTestLogger(),
		DebugLogger: newTestLogger(),
		NStorage:    newTestStorage(nil),
	}

	// Modify the object
	controller.Spec.MainIP = "10.0.0.2"

	result, err := u.patchController(controller)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result.RequeueAfter == 0 {
		t.Errorf("expected RequeueAfter to be set")
	}

	// Verify the object was updated
	updated := &k8sv1alpha1.Controller{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{Name: "test-controller", Namespace: "default"}, updated)
	if err != nil {
		t.Fatalf("failed to get updated Controller: %v", err)
	}

	if updated.Spec.MainIP != "10.0.0.2" {
		t.Errorf("MainIP: got %q, expected %q", updated.Spec.MainIP, "10.0.0.2")
	}
}

func TestPatchLinkStatus(t *testing.T) {
	scheme := newTestScheme()

	link := &k8sv1alpha1.Link{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-link",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.LinkSpec{
			Ports: []k8sv1alpha1.LinkSpecPort{"port1", "port2"},
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, link)

	u := &uniReconciler{
		Client:      fakeClient,
		Logger:      newTestLogger(),
		DebugLogger: newTestLogger(),
		NStorage:    newTestStorage(nil),
	}

	result, err := u.patchLinkStatus(link, "OK", "Link ready")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result.RequeueAfter == 0 {
		t.Errorf("expected RequeueAfter to be set")
	}

	if link.Status.Status != "OK" {
		t.Errorf("Status: got %q, expected %q", link.Status.Status, "OK")
	}
	if link.Status.Message != "Link ready" {
		t.Errorf("Message: got %q, expected %q", link.Status.Message, "Link ready")
	}
}
