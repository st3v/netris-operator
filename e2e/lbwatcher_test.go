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
	"net"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	k8sv1alpha1 "github.com/netrisai/netris-operator/api/v1alpha1"
)

// getNodeIPPrefix returns an allocation and subnet prefix that covers the first node's internal IP.
// This is needed because the lbwatcher maps pod hostIP -> subnet -> site.
func getNodeIPPrefix(namespace string) (allocationPrefix, subnetPrefix string, nodeIP string, err error) {
	nodeList := &corev1.NodeList{}
	if err = k8sClient.List(ctx, nodeList); err != nil {
		return "", "", "", fmt.Errorf("failed to list nodes: %w", err)
	}

	if len(nodeList.Items) == 0 {
		return "", "", "", fmt.Errorf("no nodes found in cluster")
	}

	// Find the first node's internal IP
	for _, addr := range nodeList.Items[0].Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			nodeIP = addr.Address
			break
		}
	}

	if nodeIP == "" {
		return "", "", "", fmt.Errorf("no internal IP found for node %s", nodeList.Items[0].Name)
	}

	// Parse the IP and create a /24 allocation and /28 subnet containing it
	ip := net.ParseIP(nodeIP)
	if ip == nil {
		return "", "", "", fmt.Errorf("failed to parse node IP: %s", nodeIP)
	}

	// For IPv4, create a /24 allocation from the first 3 octets
	ip4 := ip.To4()
	if ip4 == nil {
		return "", "", "", fmt.Errorf("node IP is not IPv4: %s", nodeIP)
	}

	// Create /24 allocation: x.y.z.0/24
	allocationPrefix = fmt.Sprintf("%d.%d.%d.0/24", ip4[0], ip4[1], ip4[2])
	// Create /28 subnet containing the node IP: x.y.z.(ip[3] & 0xF0)/28
	subnetBase := ip4[3] & 0xF0
	subnetPrefix = fmt.Sprintf("%d.%d.%d.%d/28", ip4[0], ip4[1], ip4[2], subnetBase)

	return allocationPrefix, subnetPrefix, nodeIP, nil
}

var _ = Describe("LBWatcher", func() {
	// LBWatcher needs a longer timeout since it polls every 10 seconds
	const lbwatcherTimeout = time.Second * 180

	Context("When creating a LoadBalancer service with backend pods", func() {
		var (
			f                     *TestFixture
			lbSubnetName          string
			backendAllocationName string
			backendSubnetName     string
			serviceName           string
			podName               string
			nodeIP                string
			lbSubnet              *k8sv1alpha1.Subnet
			backendAllocation     *k8sv1alpha1.Allocation
			backendSubnet         *k8sv1alpha1.Subnet
			service               *corev1.Service
			pod                   *corev1.Pod
		)

		BeforeEach(func() {
			f = NewTestFixture()

			lbSubnetName = fmt.Sprintf("e2e-lbw-lb-subnet-%d", f.Timestamp)
			backendAllocationName = fmt.Sprintf("e2e-lbw-be-alloc-%d", f.Timestamp)
			backendSubnetName = fmt.Sprintf("e2e-lbw-be-subnet-%d", f.Timestamp)
			serviceName = fmt.Sprintf("e2e-lbw-svc-%d", f.Timestamp)
			podName = fmt.Sprintf("e2e-lbw-pod-%d", f.Timestamp)

			f.CreateSite(fmt.Sprintf("e2e-lbw-site-%d", f.Timestamp), 65000, 65001, 65002)
			f.CreateAllocation(fmt.Sprintf("e2e-lbw-alloc-%d", f.Timestamp), f.Prefixes.Allocation)

			By("Creating a load-balancer Subnet CR")
			lbSubnet = &k8sv1alpha1.Subnet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      lbSubnetName,
					Namespace: f.Namespace,
				},
				Spec: k8sv1alpha1.SubnetSpec{
					Prefix:  f.Prefixes.Subnet,
					Tenant:  "Admin",
					Purpose: "load-balancer",
					Sites:   []string{f.Site.Name},
				},
			}
			Expect(k8sClient.Create(ctx, lbSubnet)).To(Succeed())

			By("Waiting for LB Subnet to be reconciled")
			waitForSubnetStatus(lbSubnet, types.NamespacedName{Name: lbSubnetName, Namespace: f.Namespace},
				func() bool { return isResourceReady(lbSubnet.Status.Status) },
				lbwatcherTimeout, defaultInterval, "LB Subnet should reach OK/Active status")

			By("Getting node IP to create matching backend subnet")
			backendAllocPrefix, backendSubPrefix, foundNodeIP, err := getNodeIPPrefix(f.Namespace)
			Expect(err).NotTo(HaveOccurred(), "Should find a node with internal IP")
			nodeIP = foundNodeIP

			By(fmt.Sprintf("Creating a backend Allocation CR covering node IP %s", nodeIP))
			backendAllocation = &k8sv1alpha1.Allocation{
				ObjectMeta: metav1.ObjectMeta{
					Name:      backendAllocationName,
					Namespace: f.Namespace,
				},
				Spec: k8sv1alpha1.AllocationSpec{
					Prefix: backendAllocPrefix,
					Tenant: "Admin",
				},
			}
			Expect(k8sClient.Create(ctx, backendAllocation)).To(Succeed())

			By("Waiting for Backend Allocation to be reconciled")
			waitForAllocationStatus(backendAllocation, types.NamespacedName{Name: backendAllocationName, Namespace: f.Namespace},
				func() bool { return isResourceReady(backendAllocation.Status.Status) },
				lbwatcherTimeout, defaultInterval, "Backend Allocation should reach OK/Active status")

			By(fmt.Sprintf("Creating a backend Subnet CR with prefix %s", backendSubPrefix))
			backendSubnet = &k8sv1alpha1.Subnet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      backendSubnetName,
					Namespace: f.Namespace,
				},
				Spec: k8sv1alpha1.SubnetSpec{
					Prefix:  backendSubPrefix,
					Tenant:  "Admin",
					Purpose: "common",
					Sites:   []string{f.Site.Name},
				},
			}
			Expect(k8sClient.Create(ctx, backendSubnet)).To(Succeed())

			By("Waiting for Backend Subnet to be reconciled")
			waitForSubnetStatus(backendSubnet, types.NamespacedName{Name: backendSubnetName, Namespace: f.Namespace},
				func() bool { return isResourceReady(backendSubnet.Status.Status) },
				lbwatcherTimeout, defaultInterval, "Backend Subnet should reach OK/Active status")
		})

		AfterEach(func() {
			// Cleanup in reverse order
			if pod != nil {
				podCopy := &corev1.Pod{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: podName, Namespace: f.Namespace}, podCopy); err == nil {
					_ = k8sClient.Delete(ctx, podCopy)
				}
			}

			if service != nil {
				svcCopy := &corev1.Service{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: serviceName, Namespace: f.Namespace}, svcCopy); err == nil {
					_ = deleteIfExists(svcCopy)
				}
			}

			// Clean up any auto-created L4LBs for this service
			l4lbList := &k8sv1alpha1.L4LBList{}
			if err := k8sClient.List(ctx, l4lbList, &client.ListOptions{Namespace: f.Namespace}); err == nil {
				for _, l4lb := range l4lbList.Items {
					if l4lb.GetServiceName() == serviceName {
						_ = k8sClient.Delete(ctx, &l4lb)
					}
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

		It("should automatically create an L4LB when a LoadBalancer service is created", func() {
			By("Creating a pod that will be a backend for the LoadBalancer service")
			pod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName,
					Namespace: f.Namespace,
					Labels: map[string]string{
						"app": serviceName,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:alpine",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, pod)).To(Succeed())

			By("Waiting for pod to be running with hostIP")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: podName, Namespace: f.Namespace}, pod)
				if err != nil {
					return false
				}
				return pod.Status.Phase == corev1.PodRunning && pod.Status.HostIP != ""
			}, lbwatcherTimeout, defaultInterval).Should(BeTrue(), "Pod should be running with hostIP")

			By("Creating a LoadBalancer service")
			service = &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      serviceName,
					Namespace: f.Namespace,
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
					Ports: []corev1.ServicePort{
						{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromInt(80),
							Protocol:   corev1.ProtocolTCP,
						},
					},
					Selector: map[string]string{
						"app": serviceName,
					},
				},
			}
			Expect(k8sClient.Create(ctx, service)).To(Succeed())

			By("Waiting for the LoadBalancer service to get a UID")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: serviceName, Namespace: f.Namespace}, service)
				if err != nil {
					return false
				}
				return service.UID != ""
			}, lbwatcherTimeout, defaultInterval).Should(BeTrue(), "Service should have a UID")

			By("Waiting for L4LB to be auto-created by LBWatcher")
			// LBWatcher runs every 10 seconds and creates L4LBs for LoadBalancer services
			// The L4LB name format: {svc-name}-{svc-namespace}-{svc-uid}-{protocol}-{port}
			var createdL4LB *k8sv1alpha1.L4LB
			Eventually(func() bool {
				l4lbList := &k8sv1alpha1.L4LBList{}
				err := k8sClient.List(ctx, l4lbList, &client.ListOptions{Namespace: f.Namespace})
				if err != nil {
					return false
				}

				for i := range l4lbList.Items {
					l4lb := &l4lbList.Items[i]
					// Check if this L4LB was created for our service
					if l4lb.GetServiceName() == serviceName && l4lb.GetServiceNamespace() == f.Namespace {
						createdL4LB = l4lb
						return true
					}
				}
				return false
			}, lbwatcherTimeout, defaultInterval).Should(BeTrue(), "L4LB should be auto-created for LoadBalancer service")

			By("Verifying L4LB has correct service annotations")
			Expect(createdL4LB).NotTo(BeNil())
			Expect(createdL4LB.GetServiceName()).To(Equal(serviceName))
			Expect(createdL4LB.GetServiceNamespace()).To(Equal(f.Namespace))
			Expect(createdL4LB.GetServiceUID()).To(Equal(string(service.UID)))

			By("Verifying L4LB spec is correctly configured")
			Expect(createdL4LB.Spec.Protocol).To(Equal("tcp"))
			Expect(createdL4LB.Spec.Frontend.Port).To(Equal(80))
			Expect(createdL4LB.Spec.State).To(Equal("active"))
			Expect(createdL4LB.Spec.Site).To(Equal(f.Site.Name))

			By("Verifying L4LB has backends configured from pod hostIP")
			Expect(len(createdL4LB.Spec.Backend)).To(BeNumerically(">=", 1), "L4LB should have at least one backend")

			By("Verifying L4LB gets reconciled successfully")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      createdL4LB.Name,
					Namespace: createdL4LB.Namespace,
				}, createdL4LB)
				if err != nil {
					return false
				}
				return createdL4LB.Status.State == "active"
			}, lbwatcherTimeout, defaultInterval).Should(BeTrue(), "L4LB should be successfully reconciled")
		})
	})
})
