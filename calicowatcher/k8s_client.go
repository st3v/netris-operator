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

package calicowatcher

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// K8sClient abstracts Kubernetes client operations for testing
type K8sClient interface {
	ListNodes(ctx context.Context, opts metav1.ListOptions) (*v1.NodeList, error)
	PatchNode(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions) (*v1.Node, error)
}

// k8sClientWrapper wraps a real Kubernetes clientset
type k8sClientWrapper struct {
	clientset *kubernetes.Clientset
}

// NewK8sClient creates a new K8sClient from a Kubernetes clientset
func NewK8sClient(clientset *kubernetes.Clientset) K8sClient {
	return &k8sClientWrapper{clientset: clientset}
}

func (c *k8sClientWrapper) ListNodes(ctx context.Context, opts metav1.ListOptions) (*v1.NodeList, error) {
	return c.clientset.CoreV1().Nodes().List(ctx, opts)
}

func (c *k8sClientWrapper) PatchNode(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions) (*v1.Node, error) {
	return c.clientset.CoreV1().Nodes().Patch(ctx, name, pt, data, opts)
}
