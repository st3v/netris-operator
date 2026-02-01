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

var _ = Describe("Site", func() {
	Context("When creating a Site resource", func() {
		var (
			f        *TestFixture
			siteName string
			site     *k8sv1alpha1.Site
		)

		BeforeEach(func() {
			f = NewTestFixture()
			siteName = fmt.Sprintf("e2e-site-%d", f.Timestamp)
		})

		AfterEach(func() {
			// Cleanup Site (not managed by fixture since it's created in the test)
			if site != nil {
				siteCopy := &k8sv1alpha1.Site{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: siteName, Namespace: f.Namespace}, siteCopy); err == nil {
					_ = deleteIfExists(siteCopy)
				}
			}
			f.Cleanup()
			f.VerifyCleanup()
		})

		It("should create a Site and sync to Netris backend", func() {
			By("Creating a Site CR")
			site = &k8sv1alpha1.Site{
				ObjectMeta: metav1.ObjectMeta{
					Name:      siteName,
					Namespace: f.Namespace,
				},
				Spec: k8sv1alpha1.SiteSpec{
					PublicASN:         65000,
					RohASN:            65001,
					VMASN:             65002,
					RohRoutingProfile: "default",
					SiteMesh:          "disabled",
					ACLDefaultPolicy:  "permit",
				},
			}
			Expect(k8sClient.Create(ctx, site)).To(Succeed())

			By("Waiting for Site to be reconciled")
			waitForSiteStatus(site, types.NamespacedName{Name: siteName, Namespace: f.Namespace},
				func() bool { return isResourceReady(site.Status.Status) },
				defaultTimeout, defaultInterval, "Site should reach OK/Active status")

			By("Verifying Site status shows success")
			Expect(site.Status.Message).To(Equal("Success"), "Site should have Success message")

			By("Verifying Site was created in Netris backend")
			backendID, err := getSiteMetaID(f.Namespace, site.UID)
			Expect(err).NotTo(HaveOccurred(), "SiteMeta should exist")
			Expect(backendID).To(BeNumerically(">", 0), "Backend ID should be set (resource created in Netris)")
		})
	})
})
