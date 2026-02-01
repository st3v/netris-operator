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

var _ = Describe("NAT E2E", func() {
	Context("When creating a NAT resource", func() {
		var (
			f             *TestFixture
			natSubnetName string
			natName       string
			natSubnet     *k8sv1alpha1.Subnet
			nat           *k8sv1alpha1.Nat
		)

		BeforeEach(func() {
			f = NewTestFixture()
			natSubnetName = fmt.Sprintf("e2e-nat-subnet-%d", f.Timestamp)
			natName = fmt.Sprintf("e2e-nat-%d", f.Timestamp)

			f.CreateSite(fmt.Sprintf("e2e-site-%d", f.Timestamp), 65300, 65301, 65302)
			f.CreateAllocation(fmt.Sprintf("e2e-alloc-%d", f.Timestamp), f.Prefixes.Allocation)

			By("Creating a NAT Subnet CR")
			natSubnet = &k8sv1alpha1.Subnet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      natSubnetName,
					Namespace: f.Namespace,
				},
				Spec: k8sv1alpha1.SubnetSpec{
					Prefix:  f.Prefixes.Subnet,
					Tenant:  "Admin",
					Purpose: "nat",
					Sites:   []string{f.Site.Name},
				},
			}
			Expect(k8sClient.Create(ctx, natSubnet)).To(Succeed())

			By("Waiting for NAT Subnet to be reconciled")
			waitForSubnetStatus(natSubnet, types.NamespacedName{Name: natSubnetName, Namespace: f.Namespace},
				func() bool { return isResourceReady(natSubnet.Status.Status) },
				defaultTimeout, defaultInterval, "NAT Subnet should reach OK/Active status")
		})

		AfterEach(func() {
			// Cleanup in reverse order
			if nat != nil {
				natCopy := &k8sv1alpha1.Nat{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: natName, Namespace: f.Namespace}, natCopy); err == nil {
					_ = deleteIfExists(natCopy)
				}
			}
			if natSubnet != nil {
				subnetCopy := &k8sv1alpha1.Subnet{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: natSubnetName, Namespace: f.Namespace}, subnetCopy); err == nil {
					_ = deleteIfExists(subnetCopy)
				}
			}
			f.Cleanup()
			f.VerifyCleanup()
		})

		It("should create a NAT rule and sync to Netris backend", func() {
			By("Creating a NAT CR")
			nat = &k8sv1alpha1.Nat{
				ObjectMeta: metav1.ObjectMeta{
					Name:      natName,
					Namespace: f.Namespace,
				},
				Spec: k8sv1alpha1.NatSpec{
					Comment:    "E2E Test NAT Rule",
					State:      "enabled",
					Site:       f.Site.Name,
					Action:     "snat",
					Protocol:   "all",
					SrcAddress: "10.0.0.0/8",
					DstAddress: f.Prefixes.Subnet,
					SnatToIP:   f.Prefixes.IP1,
				},
			}
			Expect(k8sClient.Create(ctx, nat)).To(Succeed())

			By("Waiting for NAT to be reconciled")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: natName, Namespace: f.Namespace}, nat)
				if err != nil {
					return false
				}
				return isResourceReady(nat.Status.Status)
			}, defaultTimeout, defaultInterval).Should(BeTrue(), "NAT should reach OK/Active status")

			By("Verifying NAT status shows success")
			Expect(nat.Status.Message).To(Equal("Success"), "NAT should have Success message")

			By("Verifying NAT was created in Netris backend")
			backendID, err := getNatMetaID(f.Namespace, nat.UID)
			Expect(err).NotTo(HaveOccurred(), "NatMeta should exist")
			Expect(backendID).To(BeNumerically(">", 0), "Backend ID should be set (resource created in Netris)")
		})
	})
})
