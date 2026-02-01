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

var _ = Describe("VNet E2E", func() {
	Context("When creating a VNet resource", func() {
		var (
			f        *TestFixture
			vnetName string
			vnet     *k8sv1alpha1.VNet
		)

		BeforeEach(func() {
			f = NewTestFixture()
			vnetName = fmt.Sprintf("e2e-vnet-%d", f.Timestamp)

			f.CreateSite(fmt.Sprintf("e2e-site-%d", f.Timestamp), 65000, 65001, 65002)
		})

		AfterEach(func() {
			// Delete the vnet first (not managed by fixture since it's created in the test)
			if vnet != nil {
				vnetCopy := &k8sv1alpha1.VNet{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: vnetName, Namespace: f.Namespace}, vnetCopy); err == nil {
					_ = deleteIfExists(vnetCopy)
				}
			}
			f.Cleanup()
			f.VerifyCleanup()
		})

		It("should create a VNet and sync to Netris backend", func() {
			By("Creating a VNet CR")
			vnet = &k8sv1alpha1.VNet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      vnetName,
					Namespace: f.Namespace,
				},
				Spec: k8sv1alpha1.VNetSpec{
					Owner:        "Admin",
					GuestTenants: []string{},
					VlanID:       "100",
					Sites: []k8sv1alpha1.VNetSite{
						{
							Name:     f.Site.Name,
							Gateways: []k8sv1alpha1.VNetGateway{},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, vnet)).To(Succeed())

			By("Waiting for VNet to be reconciled")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: vnetName, Namespace: f.Namespace}, vnet)
				if err != nil {
					return false
				}
				// Check if status indicates successful sync (Provisioning means created in Netris)
				return vnet.Status.Status == "OK" || vnet.Status.Status == "Active" || vnet.Status.Status == "Provisioning"
			}, defaultTimeout, defaultInterval).Should(BeTrue(), "VNet should reach OK/Active/Provisioning status")

			By("Verifying VNet status shows success")
			Expect(vnet.Status.Message).To(Equal("Success"), "VNet should have Success message")

			By("Verifying VNet was created in Netris backend")
			backendID, err := getVNetMetaID(f.Namespace, vnet.UID)
			Expect(err).NotTo(HaveOccurred(), "VNetMeta should exist")
			Expect(backendID).To(BeNumerically(">", 0), "Backend ID should be set (resource created in Netris)")

			By("Verifying Site was created in Netris backend")
			siteBackendID, err := getSiteMetaID(f.Namespace, f.Site.UID)
			Expect(err).NotTo(HaveOccurred(), "SiteMeta should exist")
			Expect(siteBackendID).To(BeNumerically(">", 0), "Site backend ID should be set")
		})
	})
})
