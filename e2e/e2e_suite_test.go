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
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	k8sv1alpha1 "github.com/netrisai/netris-operator/api/v1alpha1"
)

const (
	// defaultTimeout is the standard timeout for waiting on resource status
	defaultTimeout = time.Second * 120
	// defaultInterval is the standard polling interval for status checks
	defaultInterval = time.Second * 2
)

// NetrisAPIResource holds information about a discovered Netris API resource
type NetrisAPIResource struct {
	GVK        schema.GroupVersionKind
	Namespaced bool
}

var (
	cfg       *rest.Config
	k8sClient client.Client
	ctx       context.Context
	cancel    context.CancelFunc

	// testNamespace is the namespace created for e2e tests
	testNamespace string

	// netrisAPIResources holds the discovered Netris API resources (discovered once in BeforeSuite)
	netrisAPIResources []NetrisAPIResource

	// prefixCounter provides unique IP prefixes for each test
	prefixCounter atomic.Int32
)

// TestPrefix holds prefix information for a test
type TestPrefix struct {
	Allocation  string // e.g., "100.64.1.0/24"
	Subnet      string // e.g., "100.64.1.0/28"
	IP1         string // e.g., "100.64.1.1"
	IP2         string // e.g., "100.64.1.2"
	IP1WithPort string // e.g., "100.64.1.1:8080"
	IP2WithPort string // e.g., "100.64.1.2:8080"
}

// getTestPrefix returns a unique prefix set for this test
// Each call returns a new non-overlapping /24 from the 100.64.0.0/10 range
func getTestPrefix() TestPrefix {
	n := int(prefixCounter.Add(1))
	// Use 100.64.0.0 - 100.127.255.0 range (64 * 256 = 16384 unique /24s)
	octet2 := 64 + (n / 256)
	octet3 := n % 256
	base := fmt.Sprintf("100.%d.%d", octet2, octet3)
	return TestPrefix{
		Allocation:  fmt.Sprintf("%s.0/24", base),
		Subnet:      fmt.Sprintf("%s.0/28", base),
		IP1:         fmt.Sprintf("%s.1", base),
		IP2:         fmt.Sprintf("%s.2", base),
		IP1WithPort: fmt.Sprintf("%s.1:8080", base),
		IP2WithPort: fmt.Sprintf("%s.2:8080", base),
	}
}

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	suiteConfig, reporterConfig := GinkgoConfiguration()
	RunSpecs(t, "E2E Suite", suiteConfig, reporterConfig)
}

var _ = BeforeSuite(func() {
	ctx, cancel = context.WithCancel(context.Background())

	By("Setting up Kubernetes client")
	var err error
	cfg, err = config.GetConfig()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = k8sv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	By("Creating test namespace")
	testNamespace = fmt.Sprintf("netris-e2e-%d", time.Now().UnixNano())
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
		},
	}
	Expect(k8sClient.Create(ctx, ns)).To(Succeed())

	By("Discovering Netris API resources")
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())

	// Get all API resources for the Netris group
	resourceList, err := discoveryClient.ServerResourcesForGroupVersion(k8sv1alpha1.GroupVersion.String())
	Expect(err).NotTo(HaveOccurred())

	// Build list of resources for cleanup verification
	for _, resource := range resourceList.APIResources {
		// Skip subresources (they contain "/")
		if !strings.Contains(resource.Name, "/") {
			netrisAPIResources = append(netrisAPIResources, NetrisAPIResource{
				GVK: schema.GroupVersionKind{
					Group:   k8sv1alpha1.GroupVersion.Group,
					Version: k8sv1alpha1.GroupVersion.Version,
					Kind:    resource.Kind,
				},
				Namespaced: resource.Namespaced,
			})
		}
	}
})

var _ = AfterSuite(func() {
	By("Deleting test namespace")
	if k8sClient != nil && testNamespace != "" {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: testNamespace,
			},
		}
		_ = k8sClient.Delete(ctx, ns)
	}

	cancel()
	By("Tearing down test environment")
})
