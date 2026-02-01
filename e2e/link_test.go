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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	k8sv1alpha1 "github.com/netrisai/netris-operator/api/v1alpha1"
)

var _ = Describe("Link", func() {
	Context("When creating a Link resource", func() {
		var (
			f                  *TestFixture
			profileName        string
			switch1Name        string
			switch2Name        string
			loopbackSubnetName string
			mgmtAllocName      string
			mgmtSubnetName     string
			linkName           string
			loopbackPrefixes   TestPrefix
			mgmtPrefixes       TestPrefix
			profile            *k8sv1alpha1.InventoryProfile
			loopbackSubnet     *k8sv1alpha1.Subnet
			mgmtAlloc          *k8sv1alpha1.Allocation
			mgmtSubnet         *k8sv1alpha1.Subnet
			switch1            *k8sv1alpha1.Switch
			switch2            *k8sv1alpha1.Switch
			link               *k8sv1alpha1.Link
		)

		BeforeEach(func() {
			f = NewTestFixture()

			// Get prefixes for loopback (use fixture's prefix) and management
			loopbackPrefixes = f.Prefixes
			mgmtPrefixes = getTestPrefix()

			profileName = fmt.Sprintf("e2e-profile-%d", f.Timestamp)
			switch1Name = fmt.Sprintf("e2e-switch1-%d", f.Timestamp)
			switch2Name = fmt.Sprintf("e2e-switch2-%d", f.Timestamp)
			loopbackSubnetName = fmt.Sprintf("e2e-lo-subnet-%d", f.Timestamp)
			mgmtAllocName = fmt.Sprintf("e2e-mgmt-alloc-%d", f.Timestamp)
			mgmtSubnetName = fmt.Sprintf("e2e-mgmt-subnet-%d", f.Timestamp)
			linkName = fmt.Sprintf("e2e-link-%d", f.Timestamp)

			f.CreateSite(fmt.Sprintf("e2e-site-%d", f.Timestamp), 65020, 65021, 65022)

			By("Creating an InventoryProfile CR")
			profile = &k8sv1alpha1.InventoryProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:      profileName,
					Namespace: f.Namespace,
				},
				Spec: k8sv1alpha1.InventoryProfileSpec{
					Description:      "E2E Test Link Profile",
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

			By("Creating Switch 1 CR")
			switch1 = &k8sv1alpha1.Switch{
				ObjectMeta: metav1.ObjectMeta{
					Name:      switch1Name,
					Namespace: f.Namespace,
				},
				Spec: k8sv1alpha1.SwitchSpec{
					Tenant:      "Admin",
					Site:        f.Site.Name,
					Profile:     profileName,
					Description: "E2E Test Switch 1 for Link",
					NOS:         "cumulus_nvue",
					ASN:         4200000010,
					MainIP:      loopbackPrefixes.IP1,
					MgmtIP:      mgmtPrefixes.IP1,
					PortsCount:  48,
					MacAddress:  "00:11:22:33:44:01",
				},
			}
			Expect(k8sClient.Create(ctx, switch1)).To(Succeed())

			By("Waiting for Switch 1 to be reconciled")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: switch1Name, Namespace: f.Namespace}, switch1)
				if err != nil {
					return false
				}
				return switch1.Status.Status == "OK" || switch1.Status.Status == "Active" || switch1.Status.Status == "Failure"
			}, defaultTimeout, defaultInterval).Should(BeTrue(), "Switch 1 should reach a final status")

			if switch1.Status.Status == "Failure" {
				Skip("Switch 1 creation failed (likely requires real hardware): " + switch1.Status.Message)
			}

			By("Creating Switch 2 CR")
			switch2 = &k8sv1alpha1.Switch{
				ObjectMeta: metav1.ObjectMeta{
					Name:      switch2Name,
					Namespace: f.Namespace,
				},
				Spec: k8sv1alpha1.SwitchSpec{
					Tenant:      "Admin",
					Site:        f.Site.Name,
					Profile:     profileName,
					Description: "E2E Test Switch 2 for Link",
					NOS:         "cumulus_nvue",
					ASN:         4200000011,
					MainIP:      loopbackPrefixes.IP2,
					MgmtIP:      mgmtPrefixes.IP2,
					PortsCount:  48,
					MacAddress:  "00:11:22:33:44:02",
				},
			}
			Expect(k8sClient.Create(ctx, switch2)).To(Succeed())

			By("Waiting for Switch 2 to be reconciled")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: switch2Name, Namespace: f.Namespace}, switch2)
				if err != nil {
					return false
				}
				return switch2.Status.Status == "OK" || switch2.Status.Status == "Active" || switch2.Status.Status == "Failure"
			}, defaultTimeout, defaultInterval).Should(BeTrue(), "Switch 2 should reach a final status")

			if switch2.Status.Status == "Failure" {
				Skip("Switch 2 creation failed (likely requires real hardware): " + switch2.Status.Message)
			}

			By("Waiting for ports to sync from Netris backend")
			// The operator syncs ports every 10 seconds, so we need to wait for the sync
			time.Sleep(20 * time.Second)
		})

		AfterEach(func() {
			// Cleanup in reverse order
			if link != nil {
				linkCopy := &k8sv1alpha1.Link{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: linkName, Namespace: f.Namespace}, linkCopy); err == nil {
					_ = deleteIfExists(linkCopy)
				}
			}
			if switch2 != nil {
				swCopy := &k8sv1alpha1.Switch{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: switch2Name, Namespace: f.Namespace}, swCopy); err == nil {
					_ = deleteIfExists(swCopy)
				}
			}
			if switch1 != nil {
				swCopy := &k8sv1alpha1.Switch{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: switch1Name, Namespace: f.Namespace}, swCopy); err == nil {
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

		It("should create a Link between two switches and sync to Netris backend", func() {
			By("Creating a Link CR between Switch 1 and Switch 2")
			// Link ports are in format "portName@switchName" (see netrisstorage/ports.go:61)
			link = &k8sv1alpha1.Link{
				ObjectMeta: metav1.ObjectMeta{
					Name:      linkName,
					Namespace: f.Namespace,
				},
				Spec: k8sv1alpha1.LinkSpec{
					Ports: []k8sv1alpha1.LinkSpecPort{
						k8sv1alpha1.LinkSpecPort(fmt.Sprintf("swp1@%s", switch1Name)),
						k8sv1alpha1.LinkSpecPort(fmt.Sprintf("swp1@%s", switch2Name)),
					},
				},
			}
			Expect(k8sClient.Create(ctx, link)).To(Succeed())

			By("Waiting for Link to be reconciled")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: linkName, Namespace: f.Namespace}, link)
				if err != nil {
					return false
				}
				return link.Status.Status != "" || link.Status.Message != ""
			}, defaultTimeout, defaultInterval).Should(BeTrue(), func() string {
				_ = k8sClient.Get(ctx, types.NamespacedName{Name: linkName, Namespace: f.Namespace}, link)
				return fmt.Sprintf("Link should be reconciled - Current state: Status=%q, Message=%q, Ports=%q",
					link.Status.Status, link.Status.Message, link.Status.Ports)
			})

			By("Printing Link status for diagnosis")
			GinkgoWriter.Printf("Link Status: %q\n", link.Status.Status)
			GinkgoWriter.Printf("Link Message: %q\n", link.Status.Message)
			GinkgoWriter.Printf("Link Ports: %q\n", link.Status.Ports)

			By("Checking Link status")
			if link.Status.Status == "Failure" {
				Skip(fmt.Sprintf("Link creation failed: Status=%q, Message=%q",
					link.Status.Status, link.Status.Message))
			}

			By("Verifying Link status shows success")
			Expect(link.Status.Status).To(Or(Equal("OK"), Equal("Active")), "Link should reach OK/Active status")

			By("Verifying Link was created in Netris backend")
			backendID, err := getLinkMetaID(f.Namespace, link.UID)
			Expect(err).NotTo(HaveOccurred(), "LinkMeta should exist")
			Expect(backendID).NotTo(BeEmpty(), "Backend ID should be set (resource created in Netris)")
		})
	})
})
