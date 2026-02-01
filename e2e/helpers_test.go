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

package e2e

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	k8sv1alpha1 "github.com/netrisai/netris-operator/api/v1alpha1"
)

// verifyNoNetrisResourcesRemaining checks that no Netris resources remain in the namespace.
// It uses netrisAPIResources discovered in BeforeSuite via the discovery client.
// Uses Eventually to wait for resources to be fully deleted (finalizers may delay deletion).
// This function should be called at the end of AfterEach to catch cleanup failures.
func verifyNoNetrisResourcesRemaining(namespace string) {
	// Track remaining resources for error message
	var lastRemainingResources []string

	Eventually(func() bool {
		var remainingResources []string

		for _, resource := range netrisAPIResources {
			// Create an unstructured list for this kind
			list := &unstructured.UnstructuredList{}
			list.SetGroupVersionKind(resource.GVK)
			list.SetKind(resource.GVK.Kind + "List")

			// Only apply namespace filter for namespaced resources
			var listOpts *client.ListOptions
			if resource.Namespaced {
				listOpts = &client.ListOptions{Namespace: namespace}
			} else {
				listOpts = &client.ListOptions{}
			}

			if err := k8sClient.List(ctx, list, listOpts); err == nil {
				for _, item := range list.Items {
					remainingResources = append(remainingResources, fmt.Sprintf("%s/%s", resource.GVK.Kind, item.GetName()))
				}
			}
		}

		lastRemainingResources = remainingResources
		return len(remainingResources) == 0
	}, defaultTimeout, defaultInterval).Should(BeTrue(), func() string {
		return fmt.Sprintf("Netris resources were not properly cleaned up: %v", lastRemainingResources)
	})
}

// waitForSiteStatus waits for a Site to reach a specific status and prints its state on timeout
func waitForSiteStatus(site *k8sv1alpha1.Site, key types.NamespacedName, checkFn func() bool, timeout, interval time.Duration, description string) {
	Eventually(func() bool {
		err := k8sClient.Get(ctx, key, site)
		if err != nil {
			return false
		}
		return checkFn()
	}, timeout, interval).Should(BeTrue(), func() string {
		_ = k8sClient.Get(ctx, key, site)
		return fmt.Sprintf("%s - Current state: Status=%s, Message=%s", description, site.Status.Status, site.Status.Message)
	})
}

// waitForAllocationStatus waits for an Allocation to reach a specific status
func waitForAllocationStatus(alloc *k8sv1alpha1.Allocation, key types.NamespacedName, checkFn func() bool, timeout, interval time.Duration, description string) {
	Eventually(func() bool {
		err := k8sClient.Get(ctx, key, alloc)
		if err != nil {
			return false
		}
		return checkFn()
	}, timeout, interval).Should(BeTrue(), func() string {
		_ = k8sClient.Get(ctx, key, alloc)
		return fmt.Sprintf("%s - Current state: Status=%s, Message=%s", description, alloc.Status.Status, alloc.Status.Message)
	})
}

// waitForSubnetStatus waits for a Subnet to reach a specific status
func waitForSubnetStatus(subnet *k8sv1alpha1.Subnet, key types.NamespacedName, checkFn func() bool, timeout, interval time.Duration, description string) {
	Eventually(func() bool {
		err := k8sClient.Get(ctx, key, subnet)
		if err != nil {
			return false
		}
		return checkFn()
	}, timeout, interval).Should(BeTrue(), func() string {
		_ = k8sClient.Get(ctx, key, subnet)
		return fmt.Sprintf("%s - Current state: Status=%s, Message=%s", description, subnet.Status.Status, subnet.Status.Message)
	})
}

// waitForInventoryProfileStatus waits for an InventoryProfile to reach a specific status
func waitForInventoryProfileStatus(profile *k8sv1alpha1.InventoryProfile, key types.NamespacedName, checkFn func() bool, timeout, interval time.Duration, description string) {
	Eventually(func() bool {
		err := k8sClient.Get(ctx, key, profile)
		if err != nil {
			return false
		}
		return checkFn()
	}, timeout, interval).Should(BeTrue(), func() string {
		_ = k8sClient.Get(ctx, key, profile)
		return fmt.Sprintf("%s - Current state: Status=%s, Message=%s", description, profile.Status.Status, profile.Status.Message)
	})
}

// waitForSwitchStatus waits for a Switch to reach a specific status
func waitForSwitchStatus(sw *k8sv1alpha1.Switch, key types.NamespacedName, checkFn func() bool, timeout, interval time.Duration, description string) {
	Eventually(func() bool {
		err := k8sClient.Get(ctx, key, sw)
		if err != nil {
			return false
		}
		return checkFn()
	}, timeout, interval).Should(BeTrue(), func() string {
		_ = k8sClient.Get(ctx, key, sw)
		return fmt.Sprintf("%s - Current state: Status=%s, Message=%s", description, sw.Status.Status, sw.Status.Message)
	})
}

// getVNetMetaID retrieves the backend ID from the VNetMeta resource
// The Meta resource is named after the parent CR's UID
func getVNetMetaID(namespace string, vnetUID types.UID) (int, error) {
	meta := &k8sv1alpha1.VNetMeta{}
	err := k8sClient.Get(ctx, types.NamespacedName{
		Name:      string(vnetUID),
		Namespace: namespace,
	}, meta)
	if err != nil {
		return 0, err
	}
	return meta.Spec.ID, nil
}

// getSiteMetaID retrieves the backend ID from the SiteMeta resource
func getSiteMetaID(namespace string, siteUID types.UID) (int, error) {
	meta := &k8sv1alpha1.SiteMeta{}
	err := k8sClient.Get(ctx, types.NamespacedName{
		Name:      string(siteUID),
		Namespace: namespace,
	}, meta)
	if err != nil {
		return 0, err
	}
	return meta.Spec.ID, nil
}

// getAllocationMetaID retrieves the backend ID from the AllocationMeta resource
func getAllocationMetaID(namespace string, allocationUID types.UID) (int, error) {
	meta := &k8sv1alpha1.AllocationMeta{}
	err := k8sClient.Get(ctx, types.NamespacedName{
		Name:      string(allocationUID),
		Namespace: namespace,
	}, meta)
	if err != nil {
		return 0, err
	}
	return meta.Spec.ID, nil
}

// getSubnetMetaID retrieves the backend ID from the SubnetMeta resource
func getSubnetMetaID(namespace string, subnetUID types.UID) (int, error) {
	meta := &k8sv1alpha1.SubnetMeta{}
	err := k8sClient.Get(ctx, types.NamespacedName{
		Name:      string(subnetUID),
		Namespace: namespace,
	}, meta)
	if err != nil {
		return 0, err
	}
	return meta.Spec.ID, nil
}

// getInventoryProfileMetaID retrieves the backend ID from the InventoryProfileMeta resource
func getInventoryProfileMetaID(namespace string, profileUID types.UID) (int, error) {
	meta := &k8sv1alpha1.InventoryProfileMeta{}
	err := k8sClient.Get(ctx, types.NamespacedName{
		Name:      string(profileUID),
		Namespace: namespace,
	}, meta)
	if err != nil {
		return 0, err
	}
	return meta.Spec.ID, nil
}

// getNatMetaID retrieves the backend ID from the NatMeta resource
func getNatMetaID(namespace string, natUID types.UID) (int, error) {
	meta := &k8sv1alpha1.NatMeta{}
	err := k8sClient.Get(ctx, types.NamespacedName{
		Name:      string(natUID),
		Namespace: namespace,
	}, meta)
	if err != nil {
		return 0, err
	}
	return meta.Spec.ID, nil
}

// getBGPMetaID retrieves the backend ID from the BGPMeta resource
func getBGPMetaID(namespace string, bgpUID types.UID) (int, error) {
	meta := &k8sv1alpha1.BGPMeta{}
	err := k8sClient.Get(ctx, types.NamespacedName{
		Name:      string(bgpUID),
		Namespace: namespace,
	}, meta)
	if err != nil {
		return 0, err
	}
	return meta.Spec.ID, nil
}

// getL4LBMetaID retrieves the backend ID from the L4LBMeta resource
func getL4LBMetaID(namespace string, l4lbUID types.UID) (int, error) {
	meta := &k8sv1alpha1.L4LBMeta{}
	err := k8sClient.Get(ctx, types.NamespacedName{
		Name:      string(l4lbUID),
		Namespace: namespace,
	}, meta)
	if err != nil {
		return 0, err
	}
	return meta.Spec.ID, nil
}

// getControllerMetaID retrieves the backend ID from the ControllerMeta resource
func getControllerMetaID(namespace string, controllerUID types.UID) (int, error) {
	meta := &k8sv1alpha1.ControllerMeta{}
	err := k8sClient.Get(ctx, types.NamespacedName{
		Name:      string(controllerUID),
		Namespace: namespace,
	}, meta)
	if err != nil {
		return 0, err
	}
	return meta.Spec.ID, nil
}

// getSoftgateMetaID retrieves the backend ID from the SoftgateMeta resource
func getSoftgateMetaID(namespace string, softgateUID types.UID) (int, error) {
	meta := &k8sv1alpha1.SoftgateMeta{}
	err := k8sClient.Get(ctx, types.NamespacedName{
		Name:      string(softgateUID),
		Namespace: namespace,
	}, meta)
	if err != nil {
		return 0, err
	}
	return meta.Spec.ID, nil
}

// getSwitchMetaID retrieves the backend ID from the SwitchMeta resource
func getSwitchMetaID(namespace string, switchUID types.UID) (int, error) {
	meta := &k8sv1alpha1.SwitchMeta{}
	err := k8sClient.Get(ctx, types.NamespacedName{
		Name:      string(switchUID),
		Namespace: namespace,
	}, meta)
	if err != nil {
		return 0, err
	}
	return meta.Spec.ID, nil
}

// getLinkMetaID retrieves the backend ID from the LinkMeta resource
// Note: LinkMeta.Spec.ID is a string (format: "local-remote")
func getLinkMetaID(namespace string, linkUID types.UID) (string, error) {
	meta := &k8sv1alpha1.LinkMeta{}
	err := k8sClient.Get(ctx, types.NamespacedName{
		Name:      string(linkUID),
		Namespace: namespace,
	}, meta)
	if err != nil {
		return "", err
	}
	return meta.Spec.ID, nil
}

// deleteIfExists deletes a resource if it exists, ignoring NotFound errors
func deleteIfExists(obj runtime.Object) error {
	err := k8sClient.Delete(ctx, obj)
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}

// isResourceReady checks if a resource has reached OK or Active status
func isResourceReady(status string) bool {
	return status == "OK" || status == "Active"
}

// createAndWaitForSite creates a Site and waits for it to be ready
func createAndWaitForSite(name, namespace string, publicASN, rohASN, vmASN int) *k8sv1alpha1.Site {
	site := &k8sv1alpha1.Site{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: k8sv1alpha1.SiteSpec{
			PublicASN:         publicASN,
			RohASN:            rohASN,
			VMASN:             vmASN,
			RohRoutingProfile: "default",
			SiteMesh:          "disabled",
			ACLDefaultPolicy:  "permit",
		},
	}

	By(fmt.Sprintf("Creating Site %s", name))
	Expect(k8sClient.Create(ctx, site)).To(Succeed())

	By("Waiting for Site to be reconciled")
	waitForSiteStatus(site, types.NamespacedName{Name: name, Namespace: namespace},
		func() bool { return isResourceReady(site.Status.Status) },
		defaultTimeout, defaultInterval, "Site should reach OK/Active status")

	return site
}

// createAndWaitForAllocation creates an Allocation and waits for it to be ready
func createAndWaitForAllocation(name, namespace, prefix, tenant string) *k8sv1alpha1.Allocation {
	allocation := &k8sv1alpha1.Allocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: k8sv1alpha1.AllocationSpec{
			Prefix: prefix,
			Tenant: tenant,
		},
	}

	By(fmt.Sprintf("Creating Allocation %s", name))
	Expect(k8sClient.Create(ctx, allocation)).To(Succeed())

	By("Waiting for Allocation to be reconciled")
	waitForAllocationStatus(allocation, types.NamespacedName{Name: name, Namespace: namespace},
		func() bool { return isResourceReady(allocation.Status.Status) },
		defaultTimeout, defaultInterval, "Allocation should reach OK/Active status")

	return allocation
}

// createAndWaitForSubnet creates a Subnet and waits for it to be ready
func createAndWaitForSubnet(name, namespace, prefix, tenant, purpose string, sites []string) *k8sv1alpha1.Subnet {
	subnet := &k8sv1alpha1.Subnet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: k8sv1alpha1.SubnetSpec{
			Prefix:  prefix,
			Tenant:  tenant,
			Purpose: purpose,
			Sites:   sites,
		},
	}

	By(fmt.Sprintf("Creating Subnet %s", name))
	Expect(k8sClient.Create(ctx, subnet)).To(Succeed())

	By("Waiting for Subnet to be reconciled")
	waitForSubnetStatus(subnet, types.NamespacedName{Name: name, Namespace: namespace},
		func() bool { return isResourceReady(subnet.Status.Status) },
		defaultTimeout, defaultInterval, "Subnet should reach OK/Active status")

	return subnet
}

// createAndWaitForInventoryProfile creates an InventoryProfile and waits for it to be ready
func createAndWaitForInventoryProfile(name, namespace, description string) *k8sv1alpha1.InventoryProfile {
	profile := &k8sv1alpha1.InventoryProfile{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: k8sv1alpha1.InventoryProfileSpec{
			Description: description,
		},
	}

	By(fmt.Sprintf("Creating InventoryProfile %s", name))
	Expect(k8sClient.Create(ctx, profile)).To(Succeed())

	By("Waiting for InventoryProfile to be reconciled")
	waitForInventoryProfileStatus(profile, types.NamespacedName{Name: name, Namespace: namespace},
		func() bool { return isResourceReady(profile.Status.Status) },
		defaultTimeout, defaultInterval, "InventoryProfile should reach OK/Active status")

	return profile
}

// TestFixture manages common test resources with automatic cleanup
type TestFixture struct {
	Namespace string
	Timestamp int64
	Prefixes  TestPrefix

	// Common resources
	Site       *k8sv1alpha1.Site
	Allocation *k8sv1alpha1.Allocation
	Subnet     *k8sv1alpha1.Subnet

	// Cleanup stack (LIFO)
	cleanupFns []func()
}

// NewTestFixture creates a new test fixture using the suite's test namespace
func NewTestFixture() *TestFixture {
	return &TestFixture{
		Namespace:  testNamespace,
		Timestamp:  time.Now().UnixNano(),
		Prefixes:   getTestPrefix(),
		cleanupFns: []func(){},
	}
}

// RegisterCleanup adds a cleanup function to be called in LIFO order
func (f *TestFixture) RegisterCleanup(fn func()) {
	f.cleanupFns = append(f.cleanupFns, fn)
}

// Cleanup runs all registered cleanup functions in reverse order
func (f *TestFixture) Cleanup() {
	// Run cleanup functions in reverse order (LIFO)
	for i := len(f.cleanupFns) - 1; i >= 0; i-- {
		f.cleanupFns[i]()
	}
}

// VerifyCleanup checks that no Netris resources remain in the namespace
func (f *TestFixture) VerifyCleanup() {
	verifyNoNetrisResourcesRemaining(f.Namespace)
}

// CreateSite creates a Site resource and registers it for cleanup
func (f *TestFixture) CreateSite(name string, publicASN, rohASN, vmASN int) *k8sv1alpha1.Site {
	f.Site = createAndWaitForSite(name, f.Namespace, publicASN, rohASN, vmASN)
	f.RegisterCleanup(func() {
		siteCopy := &k8sv1alpha1.Site{}
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: f.Namespace}, siteCopy); err == nil {
			_ = deleteIfExists(siteCopy)
		}
	})
	return f.Site
}

// CreateAllocation creates an Allocation resource and registers it for cleanup
func (f *TestFixture) CreateAllocation(name, prefix string) *k8sv1alpha1.Allocation {
	f.Allocation = createAndWaitForAllocation(name, f.Namespace, prefix, "Admin")
	f.RegisterCleanup(func() {
		allocationCopy := &k8sv1alpha1.Allocation{}
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: f.Namespace}, allocationCopy); err == nil {
			_ = deleteIfExists(allocationCopy)
		}
	})
	return f.Allocation
}

// CreateSubnet creates a Subnet resource and registers it for cleanup
func (f *TestFixture) CreateSubnet(name, prefix, purpose string, sites []string) *k8sv1alpha1.Subnet {
	f.Subnet = createAndWaitForSubnet(name, f.Namespace, prefix, "Admin", purpose, sites)
	f.RegisterCleanup(func() {
		subnetCopy := &k8sv1alpha1.Subnet{}
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: f.Namespace}, subnetCopy); err == nil {
			_ = deleteIfExists(subnetCopy)
		}
	})
	return f.Subnet
}

// CreateInventoryProfile creates an InventoryProfile resource and registers it for cleanup
func (f *TestFixture) CreateInventoryProfile(name, description string) *k8sv1alpha1.InventoryProfile {
	profile := createAndWaitForInventoryProfile(name, f.Namespace, description)
	f.RegisterCleanup(func() {
		profileCopy := &k8sv1alpha1.InventoryProfile{}
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: f.Namespace}, profileCopy); err == nil {
			_ = deleteIfExists(profileCopy)
		}
	})
	return profile
}
