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

package calico

import (
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name               string
		options            Options
		wantContextTimeout bool
	}{
		{
			name:               "default options",
			options:            Options{},
			wantContextTimeout: false,
		},
		{
			name:               "custom context timeout",
			options:            Options{ContextTimeout: 30},
			wantContextTimeout: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.options)
			if c == nil {
				t.Errorf("New() returned nil")
			}
			if c != nil && c.options.ContextTimeout != tt.options.ContextTimeout {
				t.Errorf("options.ContextTimeout = %d, want %d", c.options.ContextTimeout, tt.options.ContextTimeout)
			}
		})
	}
}

func TestGenerateBGPPeer(t *testing.T) {
	tests := []struct {
		name          string
		peerName      string
		namespace     string
		ip            string
		asn           int
		wantNamespace string
	}{
		{
			name:          "with namespace",
			peerName:      "test-peer",
			namespace:     "kube-system",
			ip:            "10.0.0.1",
			asn:           65000,
			wantNamespace: "kube-system",
		},
		{
			name:          "empty namespace defaults to default",
			peerName:      "test-peer",
			namespace:     "",
			ip:            "10.0.0.1",
			asn:           65001,
			wantNamespace: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(Options{})
			peer := c.GenerateBGPPeer(tt.peerName, tt.namespace, tt.ip, tt.asn)

			if peer == nil {
				t.Fatalf("GenerateBGPPeer() returned nil")
			}
			if peer.Metadata.Name != tt.peerName {
				t.Errorf("peer.Metadata.Name = %q, want %q", peer.Metadata.Name, tt.peerName)
			}
			if peer.Metadata.Namespace != tt.wantNamespace {
				t.Errorf("peer.Metadata.Namespace = %q, want %q", peer.Metadata.Namespace, tt.wantNamespace)
			}
			if peer.Spec.PeerIP != tt.ip {
				t.Errorf("peer.Spec.PeerIP = %q, want %q", peer.Spec.PeerIP, tt.ip)
			}
			if peer.Spec.ASNumber != tt.asn {
				t.Errorf("peer.Spec.ASNumber = %d, want %d", peer.Spec.ASNumber, tt.asn)
			}
			if peer.TypeMeta.Kind != "BGPPeer" {
				t.Errorf("peer.TypeMeta.Kind = %q, want %q", peer.TypeMeta.Kind, "BGPPeer")
			}
			if peer.TypeMeta.APIVersion != "crd.projectcalico.org/v1" {
				t.Errorf("peer.TypeMeta.APIVersion = %q, want %q", peer.TypeMeta.APIVersion, "crd.projectcalico.org/v1")
			}
		})
	}
}

func TestIsMissingResource(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "missing resource error",
			err:  errors.New("the server could not find the requested resource"),
			want: true,
		},
		{
			name: "different error",
			err:  errors.New("connection refused"),
			want: false,
		},
		{
			name: "partial match error",
			err:  errors.New("the server could not find"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsMissingResource(tt.err)
			if got != tt.want {
				t.Errorf("IsMissingResource() = %v, want %v", got, tt.want)
			}
		})
	}
}
