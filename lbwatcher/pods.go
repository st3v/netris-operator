/*
Copyright 2021. Netris, Inc.

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
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getPodsByLabelSelector(k8sClient K8sClient, namespace, selectors string) (*v1.PodList, error) {
	ctx, cancel := context.WithTimeout(cntxt, contextTimeout)
	defer cancel()
	listOptions := metav1.ListOptions{
		LabelSelector: selectors,
	}
	pods, err := k8sClient.ListPods(ctx, namespace, listOptions)
	if err != nil {
		return pods, fmt.Errorf("{getPods} %s", err)
	}
	return pods, nil
}
