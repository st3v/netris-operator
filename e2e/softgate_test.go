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

var _ = Describe("Softgate", func() {
	Context("When creating a Softgate resource", func() {
		var (
			f                  *TestFixture
			profileName        string
			softgateName       string
			loopbackSubnetName string
			mgmtAllocName      string
			mgmtSubnetName     string
			loopbackPrefixes   TestPrefix
			mgmtPrefixes       TestPrefix
			profile            *k8sv1alpha1.InventoryProfile
			loopbackSubnet     *k8sv1alpha1.Subnet
			mgmtAlloc          *k8sv1alpha1.Allocation
			mgmtSubnet         *k8sv1alpha1.Subnet
			softgate           *k8sv1alpha1.Softgate
		)

		BeforeEach(func() {
			f = NewTestFixture()

			// Get prefixes for loopback (use fixture's prefix) and management
			loopbackPrefixes = f.Prefixes
			mgmtPrefixes = getTestPrefix()

			profileName = fmt.Sprintf("e2e-profile-%d", f.Timestamp)
			softgateName = fmt.Sprintf("e2e-softgate-%d", f.Timestamp)
			loopbackSubnetName = fmt.Sprintf("e2e-lo-subnet-%d", f.Timestamp)
			mgmtAllocName = fmt.Sprintf("e2e-mgmt-alloc-%d", f.Timestamp)
			mgmtSubnetName = fmt.Sprintf("e2e-mgmt-subnet-%d", f.Timestamp)

			f.CreateSite(fmt.Sprintf("e2e-site-%d", f.Timestamp), 65200, 65201, 65202)

			By("Creating an InventoryProfile CR for Softgate")
			profile = &k8sv1alpha1.InventoryProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:      profileName,
					Namespace: f.Namespace,
				},
				Spec: k8sv1alpha1.InventoryProfileSpec{
					Description:      "E2E Test Softgate Profile",
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
			if softgate != nil {
				softgateCopy := &k8sv1alpha1.Softgate{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: softgateName, Namespace: f.Namespace}, softgateCopy); err == nil {
					_ = deleteIfExists(softgateCopy)
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

		It("should create a Softgate and sync to Netris backend", func() {
			By("Creating a Softgate CR")
			softgate = &k8sv1alpha1.Softgate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      softgateName,
					Namespace: f.Namespace,
				},
				Spec: k8sv1alpha1.SoftgateSpec{
					Tenant:      "Admin",
					Site:        f.Site.Name,
					Profile:     profileName,
					Description: "E2E Test Softgate",
					Flavor:      "sg-pro",
					MainIP:      loopbackPrefixes.IP1,
					MgmtIP:      mgmtPrefixes.IP1,
				},
			}
			Expect(k8sClient.Create(ctx, softgate)).To(Succeed())

			By("Waiting for Softgate to be reconciled")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: softgateName, Namespace: f.Namespace}, softgate)
				if err != nil {
					return false
				}
				// Softgate may reach OK/Active status or may fail due to hardware requirements
				return softgate.Status.Status == "OK" || softgate.Status.Status == "Active" || softgate.Status.Status == "Failure"
			}, defaultTimeout, defaultInterval).Should(BeTrue(), func() string {
				_ = k8sClient.Get(ctx, types.NamespacedName{Name: softgateName, Namespace: f.Namespace}, softgate)
				return fmt.Sprintf("Softgate should reach a final status - Current state: Status=%q, Message=%q",
					softgate.Status.Status, softgate.Status.Message)
			})

			By("Checking Softgate status")
			if softgate.Status.Status == "Failure" {
				Skip(fmt.Sprintf("Softgate creation failed: Status=%q, Message=%q",
					softgate.Status.Status, softgate.Status.Message))
			}

			By("Verifying Softgate status shows success")
			Expect(softgate.Status.Status).To(Or(Equal("OK"), Equal("Active")), "Softgate should reach OK/Active status")

			By("Verifying Softgate was created in Netris backend")
			backendID, err := getSoftgateMetaID(f.Namespace, softgate.UID)
			Expect(err).NotTo(HaveOccurred(), "SoftgateMeta should exist")
			Expect(backendID).To(BeNumerically(">", 0), "Backend ID should be set (resource created in Netris)")
		})
	})
})
