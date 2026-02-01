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

var _ = Describe("Subnet E2E", func() {
	Context("When creating a Subnet resource", func() {
		var (
			f          *TestFixture
			subnetName string
			subnet     *k8sv1alpha1.Subnet
		)

		BeforeEach(func() {
			f = NewTestFixture()
			subnetName = fmt.Sprintf("e2e-subnet-%d", f.Timestamp)

			f.CreateSite(fmt.Sprintf("e2e-site-%d", f.Timestamp), 65100, 65101, 65102)
			f.CreateAllocation(fmt.Sprintf("e2e-alloc-%d", f.Timestamp), f.Prefixes.Allocation)
		})

		AfterEach(func() {
			// Delete the subnet first (not managed by fixture since it's created in the test)
			if subnet != nil {
				subnetCopy := &k8sv1alpha1.Subnet{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: subnetName, Namespace: f.Namespace}, subnetCopy); err == nil {
					_ = deleteIfExists(subnetCopy)
				}
			}
			f.Cleanup()
			f.VerifyCleanup()
		})

		It("should create a Subnet and sync to Netris backend", func() {
			By("Creating a Subnet CR")
			subnet = &k8sv1alpha1.Subnet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      subnetName,
					Namespace: f.Namespace,
				},
				Spec: k8sv1alpha1.SubnetSpec{
					Prefix:  f.Prefixes.Subnet,
					Tenant:  "Admin",
					Purpose: "common",
					Sites:   []string{f.Site.Name},
				},
			}
			Expect(k8sClient.Create(ctx, subnet)).To(Succeed())

			By("Waiting for Subnet to be reconciled")
			waitForSubnetStatus(subnet, types.NamespacedName{Name: subnetName, Namespace: f.Namespace},
				func() bool { return isResourceReady(subnet.Status.Status) },
				defaultTimeout, defaultInterval, "Subnet should reach OK/Active status")

			By("Verifying Subnet status shows success")
			Expect(subnet.Status.Message).To(Equal("Success"), "Subnet should have Success message")

			By("Verifying Subnet was created in Netris backend")
			subnetBackendID, err := getSubnetMetaID(f.Namespace, subnet.UID)
			Expect(err).NotTo(HaveOccurred(), "SubnetMeta should exist")
			Expect(subnetBackendID).To(BeNumerically(">", 0), "Subnet backend ID should be set (resource created in Netris)")

			By("Verifying Allocation was created in Netris backend")
			allocationBackendID, err := getAllocationMetaID(f.Namespace, f.Allocation.UID)
			Expect(err).NotTo(HaveOccurred(), "AllocationMeta should exist")
			Expect(allocationBackendID).To(BeNumerically(">", 0), "Allocation backend ID should be set")

			By("Verifying Site was created in Netris backend")
			siteBackendID, err := getSiteMetaID(f.Namespace, f.Site.UID)
			Expect(err).NotTo(HaveOccurred(), "SiteMeta should exist")
			Expect(siteBackendID).To(BeNumerically(">", 0), "Site backend ID should be set")
		})
	})
})
