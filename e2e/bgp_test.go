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

var _ = Describe("BGP", func() {
	Context("When creating a BGP resource", func() {
		var (
			f          *TestFixture
			subnetName string
			vnetName   string
			bgpName    string
			subnet     *k8sv1alpha1.Subnet
			vnet       *k8sv1alpha1.VNet
			bgp        *k8sv1alpha1.BGP
		)

		BeforeEach(func() {
			f = NewTestFixture()
			subnetName = fmt.Sprintf("e2e-subnet-%d", f.Timestamp)
			vnetName = fmt.Sprintf("e2e-vnet-%d", f.Timestamp)
			bgpName = fmt.Sprintf("e2e-bgp-%d", f.Timestamp)

			f.CreateSite(fmt.Sprintf("e2e-site-%d", f.Timestamp), 65400, 65401, 65402)
			f.CreateAllocation(fmt.Sprintf("e2e-alloc-%d", f.Timestamp), f.Prefixes.Allocation)

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

			By("Creating a VNet CR for BGP transport")
			gatewayPrefix := fmt.Sprintf("%s/28", f.Prefixes.IP1)
			vnet = &k8sv1alpha1.VNet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      vnetName,
					Namespace: f.Namespace,
				},
				Spec: k8sv1alpha1.VNetSpec{
					Owner:        "Admin",
					GuestTenants: []string{},
					VlanID:       "200",
					Sites: []k8sv1alpha1.VNetSite{
						{
							Name: f.Site.Name,
							Gateways: []k8sv1alpha1.VNetGateway{
								{
									Prefix: gatewayPrefix,
								},
							},
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
				return vnet.Status.Status == "OK" || vnet.Status.Status == "Active" || vnet.Status.Status == "Provisioning"
			}, defaultTimeout, defaultInterval).Should(BeTrue(), "VNet should reach OK/Active/Provisioning status")
		})

		AfterEach(func() {
			// Cleanup in reverse order
			if bgp != nil {
				bgpCopy := &k8sv1alpha1.BGP{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: bgpName, Namespace: f.Namespace}, bgpCopy); err == nil {
					_ = deleteIfExists(bgpCopy)
				}
			}
			if vnet != nil {
				vnetCopy := &k8sv1alpha1.VNet{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: vnetName, Namespace: f.Namespace}, vnetCopy); err == nil {
					_ = deleteIfExists(vnetCopy)
				}
			}
			if subnet != nil {
				subnetCopy := &k8sv1alpha1.Subnet{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: subnetName, Namespace: f.Namespace}, subnetCopy); err == nil {
					_ = deleteIfExists(subnetCopy)
				}
			}
			f.Cleanup()
			f.VerifyCleanup()
		})

		It("should create a BGP peering and sync to Netris backend", func() {
			By("Creating a BGP CR")
			localIP := fmt.Sprintf("%s/28", f.Prefixes.IP1)
			remoteIP := fmt.Sprintf("%s/28", f.Prefixes.IP2)
			bgp = &k8sv1alpha1.BGP{
				ObjectMeta: metav1.ObjectMeta{
					Name:      bgpName,
					Namespace: f.Namespace,
				},
				Spec: k8sv1alpha1.BGPSpec{
					Site:       f.Site.Name,
					NeighborAS: 65500,
					Transport: k8sv1alpha1.BGPTransport{
						Type: "vnet",
						Name: vnetName,
					},
					LocalIP:     localIP,
					RemoteIP:    remoteIP,
					Description: "E2E Test BGP Peering",
					State:       "enabled",
					Timers: k8sv1alpha1.BGPTimers{
						Hello:   3,
						Hold:    10,
						Connect: 10,
					},
				},
			}
			Expect(k8sClient.Create(ctx, bgp)).To(Succeed())

			By("Waiting for BGP to be reconciled")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: bgpName, Namespace: f.Namespace}, bgp)
				if err != nil {
					return false
				}
				// Check if status indicates successful sync (Provisioning means created in Netris)
				return bgp.Status.Status == "OK" || bgp.Status.Status == "Active" || bgp.Status.Status == "Provisioning"
			}, defaultTimeout, defaultInterval).Should(BeTrue(), "BGP should reach OK/Active/Provisioning status")

			By("Verifying BGP status shows success")
			Expect(bgp.Status.Message).To(Equal("Success"), "BGP should have Success message")

			By("Verifying BGP was created in Netris backend")
			backendID, err := getBGPMetaID(f.Namespace, bgp.UID)
			Expect(err).NotTo(HaveOccurred(), "BGPMeta should exist")
			Expect(backendID).To(BeNumerically(">", 0), "Backend ID should be set (resource created in Netris)")
		})
	})
})
