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

var _ = Describe("InventoryProfile", func() {
	Context("When creating an InventoryProfile resource", func() {
		var (
			f           *TestFixture
			profileName string
			profile     *k8sv1alpha1.InventoryProfile
		)

		BeforeEach(func() {
			f = NewTestFixture()
			profileName = fmt.Sprintf("e2e-profile-%d", f.Timestamp)
		})

		AfterEach(func() {
			// Cleanup InventoryProfile (not managed by fixture since it's created in the test)
			if profile != nil {
				profileCopy := &k8sv1alpha1.InventoryProfile{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: profileName, Namespace: f.Namespace}, profileCopy); err == nil {
					_ = deleteIfExists(profileCopy)
				}
			}
			f.Cleanup()
			f.VerifyCleanup()
		})

		It("should create an InventoryProfile and sync to Netris backend", func() {
			By("Creating an InventoryProfile CR")
			profile = &k8sv1alpha1.InventoryProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:      profileName,
					Namespace: f.Namespace,
				},
				Spec: k8sv1alpha1.InventoryProfileSpec{
					Description:      "E2E Test Profile",
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

			By("Verifying InventoryProfile status shows success")
			Expect(profile.Status.Message).To(Equal("Success"), "InventoryProfile should have Success message")

			By("Verifying InventoryProfile was created in Netris backend")
			backendID, err := getInventoryProfileMetaID(f.Namespace, profile.UID)
			Expect(err).NotTo(HaveOccurred(), "InventoryProfileMeta should exist")
			Expect(backendID).To(BeNumerically(">", 0), "Backend ID should be set (resource created in Netris)")
		})
	})
})
