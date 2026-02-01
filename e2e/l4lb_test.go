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

var _ = Describe("L4LB", func() {
	Context("When creating a L4LB resource", func() {
		var (
			f                     *TestFixture
			lbSubnetName          string
			backendSubnetName     string
			backendAllocationName string
			l4lbName              string
			lbPrefixes            TestPrefix
			backendPrefixes       TestPrefix
			backendAllocation     *k8sv1alpha1.Allocation
			lbSubnet              *k8sv1alpha1.Subnet
			backendSubnet         *k8sv1alpha1.Subnet
			l4lb                  *k8sv1alpha1.L4LB
		)

		BeforeEach(func() {
			f = NewTestFixture()

			// Get a second unique prefix for backend
			lbPrefixes = f.Prefixes
			backendPrefixes = getTestPrefix()

			lbSubnetName = fmt.Sprintf("e2e-lb-subnet-%d", f.Timestamp)
			backendAllocationName = fmt.Sprintf("e2e-backend-alloc-%d", f.Timestamp)
			backendSubnetName = fmt.Sprintf("e2e-backend-subnet-%d", f.Timestamp)
			l4lbName = fmt.Sprintf("e2e-l4lb-%d", f.Timestamp)

			f.CreateSite(fmt.Sprintf("e2e-site-%d", f.Timestamp), 65500, 65501, 65502)
			f.CreateAllocation(fmt.Sprintf("e2e-alloc-%d", f.Timestamp), lbPrefixes.Allocation)

			By("Creating a Load-Balancer Subnet CR")
			lbSubnet = &k8sv1alpha1.Subnet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      lbSubnetName,
					Namespace: f.Namespace,
				},
				Spec: k8sv1alpha1.SubnetSpec{
					Prefix:  lbPrefixes.Subnet,
					Tenant:  "Admin",
					Purpose: "load-balancer",
					Sites:   []string{f.Site.Name},
				},
			}
			Expect(k8sClient.Create(ctx, lbSubnet)).To(Succeed())

			By("Waiting for LB Subnet to be reconciled")
			waitForSubnetStatus(lbSubnet, types.NamespacedName{Name: lbSubnetName, Namespace: f.Namespace},
				func() bool { return isResourceReady(lbSubnet.Status.Status) },
				defaultTimeout, defaultInterval, "LB Subnet should reach OK/Active status")

			By("Creating a Backend Allocation CR for L4LB backend IPs")
			backendAllocation = &k8sv1alpha1.Allocation{
				ObjectMeta: metav1.ObjectMeta{
					Name:      backendAllocationName,
					Namespace: f.Namespace,
				},
				Spec: k8sv1alpha1.AllocationSpec{
					Prefix: backendPrefixes.Allocation,
					Tenant: "Admin",
				},
			}
			Expect(k8sClient.Create(ctx, backendAllocation)).To(Succeed())

			By("Waiting for Backend Allocation to be reconciled")
			waitForAllocationStatus(backendAllocation, types.NamespacedName{Name: backendAllocationName, Namespace: f.Namespace},
				func() bool { return isResourceReady(backendAllocation.Status.Status) },
				defaultTimeout, defaultInterval, "Backend Allocation should reach OK/Active status")

			By("Creating a Backend Subnet CR")
			backendSubnet = &k8sv1alpha1.Subnet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      backendSubnetName,
					Namespace: f.Namespace,
				},
				Spec: k8sv1alpha1.SubnetSpec{
					Prefix:  backendPrefixes.Subnet,
					Tenant:  "Admin",
					Purpose: "common",
					Sites:   []string{f.Site.Name},
				},
			}
			Expect(k8sClient.Create(ctx, backendSubnet)).To(Succeed())

			By("Waiting for Backend Subnet to be reconciled")
			waitForSubnetStatus(backendSubnet, types.NamespacedName{Name: backendSubnetName, Namespace: f.Namespace},
				func() bool { return isResourceReady(backendSubnet.Status.Status) },
				defaultTimeout, defaultInterval, "Backend Subnet should reach OK/Active status")
		})

		AfterEach(func() {
			// Cleanup in reverse order
			if l4lb != nil {
				l4lbCopy := &k8sv1alpha1.L4LB{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: l4lbName, Namespace: f.Namespace}, l4lbCopy); err == nil {
					_ = deleteIfExists(l4lbCopy)
				}
			}
			if backendSubnet != nil {
				subnetCopy := &k8sv1alpha1.Subnet{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: backendSubnetName, Namespace: f.Namespace}, subnetCopy); err == nil {
					_ = deleteIfExists(subnetCopy)
				}
			}
			if backendAllocation != nil {
				allocationCopy := &k8sv1alpha1.Allocation{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: backendAllocationName, Namespace: f.Namespace}, allocationCopy); err == nil {
					_ = deleteIfExists(allocationCopy)
				}
			}
			if lbSubnet != nil {
				subnetCopy := &k8sv1alpha1.Subnet{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: lbSubnetName, Namespace: f.Namespace}, subnetCopy); err == nil {
					_ = deleteIfExists(subnetCopy)
				}
			}
			f.Cleanup()
			f.VerifyCleanup()
		})

		It("should create a L4LB and sync to Netris backend", func() {
			By("Creating a L4LB CR")
			l4lb = &k8sv1alpha1.L4LB{
				ObjectMeta: metav1.ObjectMeta{
					Name:      l4lbName,
					Namespace: f.Namespace,
				},
				Spec: k8sv1alpha1.L4LBSpec{
					State:       "active",
					Site:        f.Site.Name,
					OwnerTenant: "Admin",
					Protocol:    "tcp",
					Frontend: k8sv1alpha1.L4LBFrontend{
						Port: 80,
						IP:   lbPrefixes.IP1,
					},
					Backend: []k8sv1alpha1.L4LBBackend{
						k8sv1alpha1.L4LBBackend(backendPrefixes.IP1WithPort),
						k8sv1alpha1.L4LBBackend(backendPrefixes.IP2WithPort),
					},
					Check: k8sv1alpha1.L4LBCheck{
						Type:    "tcp",
						Timeout: 3000,
					},
				},
			}
			Expect(k8sClient.Create(ctx, l4lb)).To(Succeed())

			By("Waiting for L4LB to be reconciled")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: l4lbName, Namespace: f.Namespace}, l4lb)
				if err != nil {
					return false
				}
				// L4LB controller sets Message to "Successfully reconciled" on success
				// Status field may be empty for newly created L4LBs
				return l4lb.Status.State == "active"
			}, defaultTimeout, defaultInterval).Should(BeTrue(), "L4LB should be successfully reconciled")

			By("Verifying L4LB status shows success")
			Expect(l4lb.Status.Message).To(Equal("Successfully reconciled"), "L4LB should have Successfully reconciled message")

			By("Verifying L4LB was created in Netris backend")
			backendID, err := getL4LBMetaID(f.Namespace, l4lb.UID)
			Expect(err).NotTo(HaveOccurred(), "L4LBMeta should exist")
			Expect(backendID).To(BeNumerically(">", 0), "Backend ID should be set (resource created in Netris)")
		})
	})
})
