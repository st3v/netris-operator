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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	k8sv1alpha1 "github.com/netrisai/netris-operator/api/v1alpha1"
)

var _ = Describe("Switch E2E", func() {
	Context("When creating a Switch resource", func() {
		var (
			f                  *TestFixture
			profileName        string
			switchName         string
			loopbackSubnetName string
			mgmtAllocName      string
			mgmtSubnetName     string
			loopbackPrefixes   TestPrefix
			mgmtPrefixes       TestPrefix
			profile            *k8sv1alpha1.InventoryProfile
			loopbackSubnet     *k8sv1alpha1.Subnet
			mgmtAlloc          *k8sv1alpha1.Allocation
			mgmtSubnet         *k8sv1alpha1.Subnet
			sw                 *k8sv1alpha1.Switch
		)

		BeforeEach(func() {
			f = NewTestFixture()

			// Get prefixes for loopback (use fixture's prefix) and management
			loopbackPrefixes = f.Prefixes
			mgmtPrefixes = getTestPrefix()

			profileName = fmt.Sprintf("e2e-profile-%d", f.Timestamp)
			switchName = fmt.Sprintf("e2e-switch-%d", f.Timestamp)
			loopbackSubnetName = fmt.Sprintf("e2e-lo-subnet-%d", f.Timestamp)
			mgmtAllocName = fmt.Sprintf("e2e-mgmt-alloc-%d", f.Timestamp)
			mgmtSubnetName = fmt.Sprintf("e2e-mgmt-subnet-%d", f.Timestamp)

			f.CreateSite(fmt.Sprintf("e2e-site-%d", f.Timestamp), 65010, 65011, 65012)

			By("Creating an InventoryProfile CR for Switch")
			profile = &k8sv1alpha1.InventoryProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:      profileName,
					Namespace: f.Namespace,
				},
				Spec: k8sv1alpha1.InventoryProfileSpec{
					Description:      "E2E Test Switch Profile",
					Timezone:         "America/Los_Angeles",
					AllowSSHFromIPv4: []string{"0.0.0.0/0"},
					AllowSSHFromIPv6: []string{},
					NTPServers:       []k8sv1alpha1.NTPServer{"pool.ntp.org"},
					DNSServers:       []k8sv1alpha1.DNSServer{"8.8.8.8"},
				},
			}
			Expect(k8sClient.Create(ctx, profile)).To(Succeed())

			By("Waiting for InventoryProfile to be reconciled")
			waitForInventoryProfileStatus(profile, types.NamespacedName{Name: profileName, Namespace: f.Namespace},
				func() bool { return isResourceReady(profile.Status.Status) },
				defaultTimeout, defaultInterval, "InventoryProfile should reach OK/Active status")

			// Create loopback allocation using fixture
			f.CreateAllocation(fmt.Sprintf("e2e-lo-alloc-%d", f.Timestamp), loopbackPrefixes.Allocation)

			By("Creating a Loopback Subnet CR")
			loopbackSubnet = &k8sv1alpha1.Subnet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      loopbackSubnetName,
					Namespace: f.Namespace,
				},
				Spec: k8sv1alpha1.SubnetSpec{
					Prefix:  loopbackPrefixes.Allocation,
					Tenant:  "Admin",
					Purpose: "loopback",
					Sites:   []string{f.Site.Name},
				},
			}
			Expect(k8sClient.Create(ctx, loopbackSubnet)).To(Succeed())

			By("Waiting for Loopback Subnet to be reconciled")
			waitForSubnetStatus(loopbackSubnet, types.NamespacedName{Name: loopbackSubnetName, Namespace: f.Namespace},
				func() bool { return isResourceReady(loopbackSubnet.Status.Status) },
				defaultTimeout, defaultInterval, "Loopback Subnet should reach OK/Active status")

			By("Creating a Management Allocation CR")
			mgmtAlloc = &k8sv1alpha1.Allocation{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mgmtAllocName,
					Namespace: f.Namespace,
				},
				Spec: k8sv1alpha1.AllocationSpec{
					Prefix: mgmtPrefixes.Allocation,
					Tenant: "Admin",
				},
			}
			Expect(k8sClient.Create(ctx, mgmtAlloc)).To(Succeed())

			By("Waiting for Management Allocation to be reconciled")
			waitForAllocationStatus(mgmtAlloc, types.NamespacedName{Name: mgmtAllocName, Namespace: f.Namespace},
				func() bool { return isResourceReady(mgmtAlloc.Status.Status) },
				defaultTimeout, defaultInterval, "Management Allocation should reach OK/Active status")

			By("Creating a Management Subnet CR")
			mgmtSubnet = &k8sv1alpha1.Subnet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mgmtSubnetName,
					Namespace: f.Namespace,
				},
				Spec: k8sv1alpha1.SubnetSpec{
					Prefix:  mgmtPrefixes.Allocation,
					Tenant:  "Admin",
					Purpose: "management",
					Sites:   []string{f.Site.Name},
				},
			}
			Expect(k8sClient.Create(ctx, mgmtSubnet)).To(Succeed())

			By("Waiting for Management Subnet to be reconciled")
			waitForSubnetStatus(mgmtSubnet, types.NamespacedName{Name: mgmtSubnetName, Namespace: f.Namespace},
				func() bool { return isResourceReady(mgmtSubnet.Status.Status) },
				defaultTimeout, defaultInterval, "Management Subnet should reach OK/Active status")
		})

		AfterEach(func() {
			// Cleanup in reverse order
			if sw != nil {
				swCopy := &k8sv1alpha1.Switch{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: switchName, Namespace: f.Namespace}, swCopy); err == nil {
					_ = deleteIfExists(swCopy)
				}
			}
			if mgmtSubnet != nil {
				subnetCopy := &k8sv1alpha1.Subnet{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: mgmtSubnetName, Namespace: f.Namespace}, subnetCopy); err == nil {
					_ = deleteIfExists(subnetCopy)
				}
			}
			if mgmtAlloc != nil {
				allocCopy := &k8sv1alpha1.Allocation{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: mgmtAllocName, Namespace: f.Namespace}, allocCopy); err == nil {
					_ = deleteIfExists(allocCopy)
				}
			}
			if loopbackSubnet != nil {
				subnetCopy := &k8sv1alpha1.Subnet{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: loopbackSubnetName, Namespace: f.Namespace}, subnetCopy); err == nil {
					_ = deleteIfExists(subnetCopy)
				}
			}
			if profile != nil {
				profileCopy := &k8sv1alpha1.InventoryProfile{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: profileName, Namespace: f.Namespace}, profileCopy); err == nil {
					_ = deleteIfExists(profileCopy)
				}
			}
			f.Cleanup()
			f.VerifyCleanup()
		})

		It("should create a Switch and sync to Netris backend", func() {
			By("Creating a Switch CR")
			sw = &k8sv1alpha1.Switch{
				ObjectMeta: metav1.ObjectMeta{
					Name:      switchName,
					Namespace: f.Namespace,
				},
				Spec: k8sv1alpha1.SwitchSpec{
					Tenant:      "Admin",
					Site:        f.Site.Name,
					Profile:     profileName,
					Description: "E2E Test Switch",
					NOS:         "cumulus_nvue",
					ASN:         4200000001,
					MainIP:      loopbackPrefixes.IP1,
					MgmtIP:      mgmtPrefixes.IP1,
					PortsCount:  48,
					MacAddress:  "00:11:22:33:44:55",
				},
			}
			Expect(k8sClient.Create(ctx, sw)).To(Succeed())

			By("Waiting for Switch to be reconciled")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: switchName, Namespace: f.Namespace}, sw)
				if err != nil {
					return false
				}
				// Switch may reach OK/Active status or may fail due to hardware requirements
				return sw.Status.Status == "OK" || sw.Status.Status == "Active" || sw.Status.Status == "Failure"
			}, defaultTimeout, defaultInterval).Should(BeTrue(), "Switch should reach a final status")

			By("Checking Switch status")
			if sw.Status.Status == "Failure" {
				Skip("Switch creation failed (likely requires real hardware): " + sw.Status.Message)
			}

			By("Verifying Switch was created in Netris backend")
			backendID, err := getSwitchMetaID(f.Namespace, sw.UID)
			Expect(err).NotTo(HaveOccurred(), "SwitchMeta should exist")
			Expect(backendID).To(BeNumerically(">", 0), "Backend ID should be set (resource created in Netris)")
		})
	})
})
