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

var _ = Describe("Allocation", func() {
	Context("When creating an Allocation resource", func() {
		var (
			f              *TestFixture
			allocationName string
			allocation     *k8sv1alpha1.Allocation
		)

		BeforeEach(func() {
			f = NewTestFixture()
			allocationName = fmt.Sprintf("e2e-alloc-%d", f.Timestamp)
		})

		AfterEach(func() {
			// Cleanup Allocation (not managed by fixture since it's created in the test)
			if allocation != nil {
				allocationCopy := &k8sv1alpha1.Allocation{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: allocationName, Namespace: f.Namespace}, allocationCopy); err == nil {
					_ = deleteIfExists(allocationCopy)
				}
			}
			f.Cleanup()
			f.VerifyCleanup()
		})

		It("should create an Allocation and sync to Netris backend", func() {
			By("Creating an Allocation CR")
			allocation = &k8sv1alpha1.Allocation{
				ObjectMeta: metav1.ObjectMeta{
					Name:      allocationName,
					Namespace: f.Namespace,
				},
				Spec: k8sv1alpha1.AllocationSpec{
					Prefix: f.Prefixes.Allocation,
					Tenant: "Admin",
				},
			}
			Expect(k8sClient.Create(ctx, allocation)).To(Succeed())

			By("Waiting for Allocation to be reconciled")
			waitForAllocationStatus(allocation, types.NamespacedName{Name: allocationName, Namespace: f.Namespace},
				func() bool { return isResourceReady(allocation.Status.Status) },
				defaultTimeout, defaultInterval, "Allocation should reach OK/Active status")

			By("Verifying Allocation status shows success")
			Expect(allocation.Status.Message).To(Equal("Success"), "Allocation should have Success message")

			By("Verifying Allocation was created in Netris backend")
			backendID, err := getAllocationMetaID(f.Namespace, allocation.UID)
			Expect(err).NotTo(HaveOccurred(), "AllocationMeta should exist")
			Expect(backendID).To(BeNumerically(">", 0), "Backend ID should be set (resource created in Netris)")
		})
	})
})
