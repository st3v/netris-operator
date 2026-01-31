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

package lbwatcher

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// K8sClient abstracts Kubernetes client operations for testing
type K8sClient interface {
	ListServices(ctx context.Context, namespace string, opts metav1.ListOptions) (*v1.ServiceList, error)
	GetService(ctx context.Context, namespace, name string, opts metav1.GetOptions) (*v1.Service, error)
	UpdateServiceStatus(ctx context.Context, namespace string, service *v1.Service, opts metav1.UpdateOptions) (*v1.Service, error)
	ListPods(ctx context.Context, namespace string, opts metav1.ListOptions) (*v1.PodList, error)
}

// k8sClientWrapper wraps a real Kubernetes clientset
type k8sClientWrapper struct {
	clientset *kubernetes.Clientset
}

// NewK8sClient creates a new K8sClient from a Kubernetes clientset
func NewK8sClient(clientset *kubernetes.Clientset) K8sClient {
	return &k8sClientWrapper{clientset: clientset}
}

func (c *k8sClientWrapper) ListServices(ctx context.Context, namespace string, opts metav1.ListOptions) (*v1.ServiceList, error) {
	return c.clientset.CoreV1().Services(namespace).List(ctx, opts)
}

func (c *k8sClientWrapper) GetService(ctx context.Context, namespace, name string, opts metav1.GetOptions) (*v1.Service, error) {
	return c.clientset.CoreV1().Services(namespace).Get(ctx, name, opts)
}

func (c *k8sClientWrapper) UpdateServiceStatus(ctx context.Context, namespace string, service *v1.Service, opts metav1.UpdateOptions) (*v1.Service, error) {
	return c.clientset.CoreV1().Services(namespace).UpdateStatus(ctx, service, opts)
}

func (c *k8sClientWrapper) ListPods(ctx context.Context, namespace string, opts metav1.ListOptions) (*v1.PodList, error) {
	return c.clientset.CoreV1().Pods(namespace).List(ctx, opts)
}
