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
	"fmt"
	"testing"

	"github.com/netrisai/netris-operator/api/v1alpha1"
	"github.com/netrisai/netris-operator/calicowatcher/calico"
	"github.com/netrisai/netris-operator/netrisstorage"
	"github.com/netrisai/netriswebapi/v2/types/ipam"
	"github.com/netrisai/netriswebapi/v2/types/site"
	"github.com/netrisai/netriswebapi/v2/types/vnet"
	"go.uber.org/zap/zapcore"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func init() {
	// Initialize logger and debugLogger for tests
	logger = zap.New(zap.UseDevMode(true))
	debugLogger = logger.V(int(zapcore.WarnLevel))
}

func newTestScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)
	return scheme
}

func TestNewWatcher(t *testing.T) {
	tests := []struct {
		name    string
		storage *netrisstorage.Storage
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil storage returns error",
			storage: nil,
			wantErr: true,
			errMsg:  "please provide NStorage",
		},
		{
			name:    "valid storage returns watcher",
			storage: &netrisstorage.Storage{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			watcher, err := NewWatcher(tt.storage, nil, Options{})

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if err.Error() != tt.errMsg {
					t.Errorf("error = %q, want %q", err.Error(), tt.errMsg)
				}
				if watcher != nil {
					t.Errorf("expected nil watcher when error, got %v", watcher)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if watcher == nil {
					t.Errorf("expected watcher, got nil")
				}
				if watcher != nil && watcher.NStorage != tt.storage {
					t.Errorf("watcher.NStorage = %v, want %v", watcher.NStorage, tt.storage)
				}
			}
		})
	}
}

func TestValidateASNRange(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantA   int
		wantB   int
		wantErr bool
	}{
		{
			name:    "valid range",
			input:   "4230000000-4239999999",
			wantA:   4230000000,
			wantB:   4239999999,
			wantErr: false,
		},
		{
			name:    "small valid range",
			input:   "65000-65100",
			wantA:   65000,
			wantB:   65100,
			wantErr: false,
		},
		{
			name:    "invalid format - no dash",
			input:   "4230000000",
			wantErr: true,
		},
		{
			name:    "invalid format - too many parts",
			input:   "100-200-300",
			wantErr: true,
		},
		{
			name:    "invalid - first number not numeric",
			input:   "abc-4239999999",
			wantErr: true,
		},
		{
			name:    "invalid - second number not numeric",
			input:   "4230000000-xyz",
			wantErr: true,
		},
		{
			name:    "invalid - first number zero",
			input:   "0-100",
			wantErr: true,
		},
		{
			name:    "invalid - second number exceeds max",
			input:   "100-4294967295",
			wantErr: true,
		},
		{
			name:    "invalid - first >= second",
			input:   "200-100",
			wantErr: true,
		},
		{
			name:    "invalid - equal values",
			input:   "100-100",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Watcher{}
			a, b, err := w.validateASNRange(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if a != tt.wantA {
					t.Errorf("a = %d, want %d", a, tt.wantA)
				}
				if b != tt.wantB {
					t.Errorf("b = %d, want %d", b, tt.wantB)
				}
			}
		})
	}
}

func TestCheckBGPConfigurations(t *testing.T) {
	tests := []struct {
		name     string
		bgpConfs []*calico.BGPConfiguration
		want     bool
	}{
		{
			name:     "empty list",
			bgpConfs: []*calico.BGPConfiguration{},
			want:     false,
		},
		{
			name: "no calico annotation",
			bgpConfs: []*calico.BGPConfiguration{
				{
					Metadata: metav1.ObjectMeta{
						Annotations: map[string]string{
							"other-annotation": "value",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "calico annotation false",
			bgpConfs: []*calico.BGPConfiguration{
				{
					Metadata: metav1.ObjectMeta{
						Annotations: map[string]string{
							"manage.k8s.netris.ai/calico": "false",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "calico annotation true",
			bgpConfs: []*calico.BGPConfiguration{
				{
					Metadata: metav1.ObjectMeta{
						Annotations: map[string]string{
							"manage.k8s.netris.ai/calico": "true",
						},
					},
				},
			},
			want: true,
		},
		{
			name: "multiple configs, second has annotation",
			bgpConfs: []*calico.BGPConfiguration{
				{
					Metadata: metav1.ObjectMeta{
						Annotations: map[string]string{},
					},
				},
				{
					Metadata: metav1.ObjectMeta{
						Annotations: map[string]string{
							"manage.k8s.netris.ai/calico": "true",
						},
					},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Watcher{
				data: data{
					bgpConfs: tt.bgpConfs,
				},
			}

			got := w.checkBGPConfigurations()
			if got != tt.want {
				t.Errorf("checkBGPConfigurations() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareBGPs(t *testing.T) {
	tests := []struct {
		name            string
		generatedBGPs   []*v1alpha1.BGP
		bgpList         []*v1alpha1.BGP
		wantCreateCount int
		wantDeleteCount int
		wantUpdateCount int
	}{
		{
			name:            "empty lists",
			generatedBGPs:   []*v1alpha1.BGP{},
			bgpList:         []*v1alpha1.BGP{},
			wantCreateCount: 0,
			wantDeleteCount: 0,
			wantUpdateCount: 0,
		},
		{
			name: "new BGP to create",
			generatedBGPs: []*v1alpha1.BGP{
				{ObjectMeta: metav1.ObjectMeta{Name: "new-bgp"}},
			},
			bgpList:         []*v1alpha1.BGP{},
			wantCreateCount: 1,
			wantDeleteCount: 0,
			wantUpdateCount: 0,
		},
		{
			name:          "existing BGP to delete",
			generatedBGPs: []*v1alpha1.BGP{},
			bgpList: []*v1alpha1.BGP{
				{ObjectMeta: metav1.ObjectMeta{Name: "old-bgp"}},
			},
			wantCreateCount: 0,
			wantDeleteCount: 1,
			wantUpdateCount: 0,
		},
		{
			name: "matching BGP no update",
			generatedBGPs: []*v1alpha1.BGP{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "same-bgp"},
					Spec:       v1alpha1.BGPSpec{Site: "site1", NeighborAS: 65000},
				},
			},
			bgpList: []*v1alpha1.BGP{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "same-bgp"},
					Spec:       v1alpha1.BGPSpec{Site: "site1", NeighborAS: 65000},
				},
			},
			wantCreateCount: 0,
			wantDeleteCount: 0,
			wantUpdateCount: 0,
		},
		{
			name: "matching BGP needs update",
			generatedBGPs: []*v1alpha1.BGP{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "update-bgp"},
					Spec:       v1alpha1.BGPSpec{Site: "site1", NeighborAS: 65001},
				},
			},
			bgpList: []*v1alpha1.BGP{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "update-bgp"},
					Spec:       v1alpha1.BGPSpec{Site: "site1", NeighborAS: 65000},
				},
			},
			wantCreateCount: 0,
			wantDeleteCount: 0,
			wantUpdateCount: 1,
		},
		{
			name: "mixed operations",
			generatedBGPs: []*v1alpha1.BGP{
				{ObjectMeta: metav1.ObjectMeta{Name: "new-bgp"}},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "update-bgp"},
					Spec:       v1alpha1.BGPSpec{NeighborAS: 65001},
				},
			},
			bgpList: []*v1alpha1.BGP{
				{ObjectMeta: metav1.ObjectMeta{Name: "delete-bgp"}},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "update-bgp"},
					Spec:       v1alpha1.BGPSpec{NeighborAS: 65000},
				},
			},
			wantCreateCount: 1,
			wantDeleteCount: 1,
			wantUpdateCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Watcher{
				data: data{
					generatedBGPs: tt.generatedBGPs,
					bgpList:       tt.bgpList,
				},
			}

			toCreate, toDelete, toUpdate := w.compareBGPs()

			if len(toCreate) != tt.wantCreateCount {
				t.Errorf("toCreate count = %d, want %d", len(toCreate), tt.wantCreateCount)
			}
			if len(toDelete) != tt.wantDeleteCount {
				t.Errorf("toDelete count = %d, want %d", len(toDelete), tt.wantDeleteCount)
			}
			if len(toUpdate) != tt.wantUpdateCount {
				t.Errorf("toUpdate count = %d, want %d", len(toUpdate), tt.wantUpdateCount)
			}
		})
	}
}

func TestGetBGPs(t *testing.T) {
	scheme := newTestScheme()

	tests := []struct {
		name        string
		existingBGPs []v1alpha1.BGP
		wantCount   int
		wantErr     bool
	}{
		{
			name:        "no BGPs exist",
			existingBGPs: []v1alpha1.BGP{},
			wantCount:   0,
			wantErr:     false,
		},
		{
			name: "multiple BGPs exist",
			existingBGPs: []v1alpha1.BGP{
				{ObjectMeta: metav1.ObjectMeta{Name: "bgp1", Namespace: "default"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "bgp2", Namespace: "default"}},
			},
			wantCount: 2,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objects := make([]runtime.Object, len(tt.existingBGPs))
			for i := range tt.existingBGPs {
				objects[i] = &tt.existingBGPs[i]
			}

			fakeClient := fake.NewFakeClientWithScheme(scheme, objects...)

			w := &Watcher{
				client: fakeClient,
			}

			result, err := w.getBGPs()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Fatalf("expected result, got nil")
				}
				if len(result.Items) != tt.wantCount {
					t.Errorf("got %d items, want %d", len(result.Items), tt.wantCount)
				}
			}
		})
	}
}

func TestCreateBGP(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	w := &Watcher{
		client: fakeClient,
	}

	bgp := &v1alpha1.BGP{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "new-bgp",
			Namespace: "default",
		},
		Spec: v1alpha1.BGPSpec{
			Site:       "site1",
			NeighborAS: 65000,
		},
	}

	err := w.createBGP(bgp)
	if err != nil {
		t.Errorf("createBGP() error = %v", err)
	}

	// Verify it was created
	list := &v1alpha1.BGPList{}
	err = fakeClient.List(context.Background(), list)
	if err != nil {
		t.Fatalf("failed to list: %v", err)
	}
	if len(list.Items) != 1 {
		t.Errorf("expected 1 item after create, got %d", len(list.Items))
	}
}

func TestUpdateBGP(t *testing.T) {
	scheme := newTestScheme()

	bgp := &v1alpha1.BGP{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "existing-bgp",
			Namespace: "default",
		},
		Spec: v1alpha1.BGPSpec{
			Site:       "site1",
			NeighborAS: 65000,
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, bgp)

	w := &Watcher{
		client: fakeClient,
	}

	bgp.Spec.NeighborAS = 65001
	err := w.updateBGP(bgp)
	if err != nil {
		t.Errorf("updateBGP() error = %v", err)
	}
}

func TestDeleteBGP(t *testing.T) {
	scheme := newTestScheme()

	bgp := &v1alpha1.BGP{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-bgp",
			Namespace: "default",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, bgp)

	w := &Watcher{
		client: fakeClient,
	}

	// Verify it exists
	list := &v1alpha1.BGPList{}
	err := fakeClient.List(context.Background(), list)
	if err != nil {
		t.Fatalf("failed to list: %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(list.Items))
	}

	// Delete it
	err = w.deleteBGP(bgp)
	if err != nil {
		t.Errorf("deleteBGP() error = %v", err)
	}

	// Verify it's gone
	err = fakeClient.List(context.Background(), list)
	if err != nil {
		t.Fatalf("failed to list after delete: %v", err)
	}
	if len(list.Items) != 0 {
		t.Errorf("expected 0 items after delete, got %d", len(list.Items))
	}
}

func TestCreateBGPs(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	w := &Watcher{
		client: fakeClient,
	}

	bgps := []*v1alpha1.BGP{
		{ObjectMeta: metav1.ObjectMeta{Name: "bgp1", Namespace: "default"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "bgp2", Namespace: "default"}},
	}

	errors := w.createBGPs(bgps)
	if len(errors) != 0 {
		t.Errorf("createBGPs() returned %d errors, want 0", len(errors))
	}

	// Verify all created
	list := &v1alpha1.BGPList{}
	err := fakeClient.List(context.Background(), list)
	if err != nil {
		t.Fatalf("failed to list: %v", err)
	}
	if len(list.Items) != 2 {
		t.Errorf("expected 2 items after create, got %d", len(list.Items))
	}
}

func TestUpdateBGPs(t *testing.T) {
	scheme := newTestScheme()

	bgps := []*v1alpha1.BGP{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "bgp1", Namespace: "default"},
			Spec:       v1alpha1.BGPSpec{NeighborAS: 65000},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "bgp2", Namespace: "default"},
			Spec:       v1alpha1.BGPSpec{NeighborAS: 65001},
		},
	}

	objects := make([]runtime.Object, len(bgps))
	for i := range bgps {
		objects[i] = bgps[i]
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, objects...)

	w := &Watcher{
		client: fakeClient,
	}

	// Update ASNs
	bgps[0].Spec.NeighborAS = 65100
	bgps[1].Spec.NeighborAS = 65101

	errors := w.updateBGPs(bgps)
	if len(errors) != 0 {
		t.Errorf("updateBGPs() returned %d errors, want 0", len(errors))
	}
}

func TestDeleteBGPs(t *testing.T) {
	scheme := newTestScheme()

	bgps := []*v1alpha1.BGP{
		{ObjectMeta: metav1.ObjectMeta{Name: "bgp1", Namespace: "default"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "bgp2", Namespace: "default"}},
	}

	objects := make([]runtime.Object, len(bgps))
	for i := range bgps {
		objects[i] = bgps[i]
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, objects...)

	w := &Watcher{
		client: fakeClient,
	}

	errors := w.deleteBGPs(bgps)
	if len(errors) != 0 {
		t.Errorf("deleteBGPs() returned %d errors, want 0", len(errors))
	}

	// Verify all deleted
	list := &v1alpha1.BGPList{}
	err := fakeClient.List(context.Background(), list)
	if err != nil {
		t.Fatalf("failed to list: %v", err)
	}
	if len(list.Items) != 0 {
		t.Errorf("expected 0 items after delete, got %d", len(list.Items))
	}
}

// TestUpdateBGPConfMesh tests the updateBGPConfMesh function
// Note: We can only test the error path (empty/nil bgpConfs) because
// the success path requires a real restClient which would panic
func TestUpdateBGPConfMesh(t *testing.T) {
	tests := []struct {
		name     string
		bgpConfs []*calico.BGPConfiguration
		enabled  bool
		wantErr  bool
	}{
		{
			name:     "empty BGP configurations returns error",
			bgpConfs: []*calico.BGPConfiguration{},
			enabled:  true,
			wantErr:  true,
		},
		{
			name:     "nil BGP configurations returns error",
			bgpConfs: nil,
			enabled:  false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Watcher{
				data: data{
					bgpConfs: tt.bgpConfs,
				},
			}

			err := w.updateBGPConfMesh(tt.enabled)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// mockK8sClient is a mock implementation of K8sClient for testing
type mockK8sClient struct {
	nodes         *v1.NodeList
	patchedNode   *v1.Node
	listNodesErr  error
	patchNodeErr  error
}

func (m *mockK8sClient) ListNodes(ctx context.Context, opts metav1.ListOptions) (*v1.NodeList, error) {
	return m.nodes, m.listNodesErr
}

func (m *mockK8sClient) PatchNode(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions) (*v1.Node, error) {
	if m.patchNodeErr != nil {
		return nil, m.patchNodeErr
	}
	// Return a mock patched node
	if m.patchedNode != nil {
		return m.patchedNode, nil
	}
	return &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: name}}, nil
}

func TestGetNodes(t *testing.T) {
	tests := []struct {
		name      string
		mock      *mockK8sClient
		wantCount int
		wantErr   bool
		errMsg    string
	}{
		{
			name: "returns nodes successfully",
			mock: &mockK8sClient{
				nodes: &v1.NodeList{
					Items: []v1.Node{
						{ObjectMeta: metav1.ObjectMeta{Name: "node1"}},
						{ObjectMeta: metav1.ObjectMeta{Name: "node2"}},
					},
				},
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "returns error when no nodes",
			mock: &mockK8sClient{
				nodes: &v1.NodeList{Items: []v1.Node{}},
			},
			wantErr: true,
			errMsg:  "nodes are missing",
		},
		{
			name: "returns error on list failure",
			mock: &mockK8sClient{
				listNodesErr: fmt.Errorf("connection refused"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Watcher{
				k8sClient: tt.mock,
				data:      data{},
			}

			err := w.getNodes()

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("error = %q, want %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if w.data.nodes == nil {
					t.Fatal("expected nodes to be set")
				}
				if len(w.data.nodes.Items) != tt.wantCount {
					t.Errorf("got %d nodes, want %d", len(w.data.nodes.Items), tt.wantCount)
				}
			}
		})
	}
}

func TestDeleteNodesASNs(t *testing.T) {
	tests := []struct {
		name     string
		mock     *mockK8sClient
		nodes    *v1.NodeList
		asnStart int
		asnEnd   int
		wantErr  bool
	}{
		{
			name: "deletes ASN from node in range",
			mock: &mockK8sClient{},
			nodes: &v1.NodeList{
				Items: []v1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node1",
							Annotations: map[string]string{
								"projectcalico.org/ASNumber": "4230000001",
							},
						},
					},
				},
			},
			asnStart: 4230000000,
			asnEnd:   4239999999,
			wantErr:  false,
		},
		{
			name: "skips node with ASN outside range",
			mock: &mockK8sClient{},
			nodes: &v1.NodeList{
				Items: []v1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node1",
							Annotations: map[string]string{
								"projectcalico.org/ASNumber": "65000",
							},
						},
					},
				},
			},
			asnStart: 4230000000,
			asnEnd:   4239999999,
			wantErr:  false,
		},
		{
			name: "handles node without ASN annotation",
			mock: &mockK8sClient{},
			nodes: &v1.NodeList{
				Items: []v1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:        "node1",
							Annotations: map[string]string{},
						},
					},
				},
			},
			asnStart: 4230000000,
			asnEnd:   4239999999,
			wantErr:  false,
		},
		{
			name: "returns error on patch failure",
			mock: &mockK8sClient{
				patchNodeErr: fmt.Errorf("patch failed"),
			},
			nodes: &v1.NodeList{
				Items: []v1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node1",
							Annotations: map[string]string{
								"projectcalico.org/ASNumber": "4230000001",
							},
						},
					},
				},
			},
			asnStart: 4230000000,
			asnEnd:   4239999999,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Watcher{
				k8sClient: tt.mock,
				data: data{
					nodes:    tt.nodes,
					asnStart: tt.asnStart,
					asnEnd:   tt.asnEnd,
				},
			}

			err := w.deleteNodesASNs()

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestFillNodesASNs(t *testing.T) {
	tests := []struct {
		name     string
		mock     *mockK8sClient
		nodes    *v1.NodeList
		asnStart int
		asnEnd   int
		wantErr  bool
	}{
		{
			name: "fills ASN for node without one",
			mock: &mockK8sClient{},
			nodes: &v1.NodeList{
				Items: []v1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:        "node1",
							Annotations: map[string]string{},
						},
					},
				},
			},
			asnStart: 4230000000,
			asnEnd:   4230000100,
			wantErr:  false,
		},
		{
			name: "skips node that already has ASN",
			mock: &mockK8sClient{},
			nodes: &v1.NodeList{
				Items: []v1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node1",
							Annotations: map[string]string{
								"projectcalico.org/ASNumber": "4230000001",
							},
						},
					},
				},
			},
			asnStart: 4230000000,
			asnEnd:   4230000100,
			wantErr:  false,
		},
		{
			name: "returns error on patch failure",
			mock: &mockK8sClient{
				patchNodeErr: fmt.Errorf("patch failed"),
			},
			nodes: &v1.NodeList{
				Items: []v1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:        "node1",
							Annotations: map[string]string{},
						},
					},
				},
			},
			asnStart: 4230000000,
			asnEnd:   4230000100,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Watcher{
				k8sClient: tt.mock,
				data: data{
					nodes:    tt.nodes,
					asnStart: tt.asnStart,
					asnEnd:   tt.asnEnd,
				},
			}

			err := w.fillNodesASNs()

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestGenerateBGPs(t *testing.T) {
	tests := []struct {
		name         string
		nodesMap     map[string]*nodeIP
		blockSize    int
		serviceCIDRs []string
		site         *site.Site
		clusterCIDR  string
		vnetName     string
		vnetGW       string
		switchName   string
		wantCount    int
		wantErr      bool
	}{
		{
			name:      "empty nodes map",
			nodesMap:  map[string]*nodeIP{},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "single node generates BGP",
			nodesMap: map[string]*nodeIP{
				"node1": {
					IP:   "10.0.0.10/24",
					IPIP: "192.168.0.10",
					ASN:  "4230000001",
				},
			},
			blockSize:    26,
			serviceCIDRs: []string{"10.96.0.0/12"},
			site:         &site.Site{Name: "site1"},
			clusterCIDR:  "192.168.0.0/16",
			vnetName:     "vnet1",
			vnetGW:       "10.0.0.1/24",
			switchName:   "switch1",
			wantCount:    1,
			wantErr:      false,
		},
		{
			name: "multiple nodes generate BGPs",
			nodesMap: map[string]*nodeIP{
				"node1": {IP: "10.0.0.10/24", IPIP: "192.168.0.10", ASN: "4230000001"},
				"node2": {IP: "10.0.0.11/24", IPIP: "192.168.0.11", ASN: "4230000002"},
				"node3": {IP: "10.0.0.12/24", IPIP: "192.168.0.12", ASN: "4230000003"},
			},
			blockSize:    26,
			serviceCIDRs: []string{"10.96.0.0/12"},
			site:         &site.Site{Name: "site1"},
			clusterCIDR:  "192.168.0.0/16",
			vnetName:     "vnet1",
			vnetGW:       "10.0.0.1/24",
			switchName:   "switch1",
			wantCount:    3,
			wantErr:      false,
		},
		{
			name: "invalid ASN returns error",
			nodesMap: map[string]*nodeIP{
				"node1": {IP: "10.0.0.10/24", IPIP: "192.168.0.10", ASN: "invalid"},
			},
			blockSize: 26,
			site:      &site.Site{Name: "site1"},
			wantErr:   true,
		},
		{
			name: "invalid IPIP returns error",
			nodesMap: map[string]*nodeIP{
				"node1": {IP: "10.0.0.10/24", IPIP: "not-an-ip", ASN: "4230000001"},
			},
			blockSize: 26,
			site:      &site.Site{Name: "site1"},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Watcher{
				data: data{
					nodesMap:     tt.nodesMap,
					blockSize:    tt.blockSize,
					serviceCIDRs: tt.serviceCIDRs,
					site:         tt.site,
					clusterCIDR:  tt.clusterCIDR,
					vnetName:     tt.vnetName,
					vnetGW:       tt.vnetGW,
					switchName:   tt.switchName,
				},
			}

			err := w.generateBGPs()

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(w.data.generatedBGPs) != tt.wantCount {
					t.Errorf("generated %d BGPs, want %d", len(w.data.generatedBGPs), tt.wantCount)
				}
			}
		})
	}
}

func TestGenerateBGPsVerifyContent(t *testing.T) {
	w := &Watcher{
		data: data{
			nodesMap: map[string]*nodeIP{
				"worker-node": {
					IP:   "10.0.0.50/24",
					IPIP: "192.168.1.50",
					ASN:  "4230000005",
				},
			},
			blockSize:    26,
			serviceCIDRs: []string{"10.96.0.0/12", "10.100.0.0/16"},
			site:         &site.Site{Name: "production-site", PublicAsn: 65000},
			clusterCIDR:  "192.168.0.0/16",
			vnetName:     "kubernetes-vnet",
			vnetGW:       "10.0.0.1/24",
			switchName:   "spine-switch",
		},
	}

	err := w.generateBGPs()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(w.data.generatedBGPs) != 1 {
		t.Fatalf("expected 1 BGP, got %d", len(w.data.generatedBGPs))
	}

	bgp := w.data.generatedBGPs[0]
	if bgp.Spec.Site != "production-site" {
		t.Errorf("site = %q, want %q", bgp.Spec.Site, "production-site")
	}
	if bgp.Spec.NeighborAS != 4230000005 {
		t.Errorf("neighborAS = %d, want %d", bgp.Spec.NeighborAS, 4230000005)
	}
	if bgp.Spec.Transport.Name != "kubernetes-vnet" {
		t.Errorf("transport name = %q, want %q", bgp.Spec.Transport.Name, "kubernetes-vnet")
	}
	if bgp.Spec.LocalIP != "10.0.0.1/24" {
		t.Errorf("localIP = %q, want %q", bgp.Spec.LocalIP, "10.0.0.1/24")
	}
	if bgp.Spec.RemoteIP != "10.0.0.50/24" {
		t.Errorf("remoteIP = %q, want %q", bgp.Spec.RemoteIP, "10.0.0.50/24")
	}

	// Check annotations
	anns := bgp.GetAnnotations()
	if anns["k8s.netris.ai/calicowatcher"] != "true" {
		t.Errorf("missing calicowatcher annotation")
	}
	if anns["resource.k8s.netris.ai/import"] != "true" {
		t.Errorf("missing import annotation")
	}
}

func TestCreateBGPsWithError(t *testing.T) {
	scheme := newTestScheme()

	// Create a BGP that already exists
	existing := &v1alpha1.BGP{
		ObjectMeta: metav1.ObjectMeta{Name: "existing-bgp", Namespace: "default"},
	}
	fakeClient := fake.NewFakeClientWithScheme(scheme, existing)

	w := &Watcher{client: fakeClient}

	bgps := []*v1alpha1.BGP{
		{ObjectMeta: metav1.ObjectMeta{Name: "existing-bgp", Namespace: "default"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "new-bgp", Namespace: "default"}},
	}

	errors := w.createBGPs(bgps)
	if len(errors) != 1 {
		t.Errorf("createBGPs() returned %d errors, want 1", len(errors))
	}
}

func TestUpdateBGPsWithError(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	w := &Watcher{client: fakeClient}

	// Try to update a non-existent BGP
	bgps := []*v1alpha1.BGP{
		{ObjectMeta: metav1.ObjectMeta{Name: "nonexistent", Namespace: "default"}},
	}

	errors := w.updateBGPs(bgps)
	if len(errors) != 1 {
		t.Errorf("updateBGPs() returned %d errors, want 1", len(errors))
	}
}

func TestDeleteBGPsWithNonexistent(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	w := &Watcher{client: fakeClient}

	// Delete non-existent BGPs will error with the fake client
	bgps := []*v1alpha1.BGP{
		{ObjectMeta: metav1.ObjectMeta{Name: "nonexistent1", Namespace: "default"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "nonexistent2", Namespace: "default"}},
	}

	errors := w.deleteBGPs(bgps)
	// The fake client returns errors for non-existent resources
	if len(errors) != 2 {
		t.Errorf("deleteBGPs() returned %d errors, want 2", len(errors))
	}
}

func TestDeleteNodesProcessing(t *testing.T) {
	mock := &mockK8sClient{
		nodes: &v1.NodeList{
			Items: []v1.Node{
				{ObjectMeta: metav1.ObjectMeta{Name: "node1"}},
			},
		},
	}

	w := &Watcher{
		k8sClient: mock,
		data:      data{},
	}

	err := w.deleteNodesProcessing()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDeleteNodesProcessingError(t *testing.T) {
	mock := &mockK8sClient{
		listNodesErr: fmt.Errorf("failed to list nodes"),
	}

	w := &Watcher{
		k8sClient: mock,
		data:      data{},
	}

	err := w.deleteNodesProcessing()
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestNodesProcessing(t *testing.T) {
	tests := []struct {
		name    string
		nodes   *v1.NodeList
		wantErr bool
		errMsg  string
	}{
		{
			name: "node without IPv4Address annotation is skipped",
			nodes: &v1.NodeList{
				Items: []v1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node1",
							Annotations: map[string]string{
								"projectcalico.org/IPv4IPIPTunnelAddr": "192.168.0.10",
								"projectcalico.org/ASNumber":           "4230000001",
							},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "couldn't find site",
		},
		{
			name: "node without IPv4IPIPTunnelAddr annotation is skipped",
			nodes: &v1.NodeList{
				Items: []v1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node1",
							Annotations: map[string]string{
								"projectcalico.org/IPv4Address": "10.0.0.10/24",
								"projectcalico.org/ASNumber":    "4230000001",
							},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "couldn't find site",
		},
		{
			name: "node without ASNumber annotation returns error",
			nodes: &v1.NodeList{
				Items: []v1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node1",
							Annotations: map[string]string{
								"projectcalico.org/IPv4Address":        "10.0.0.10/24",
								"projectcalico.org/IPv4IPIPTunnelAddr": "192.168.0.10",
							},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "couldn't get as number for node node1",
		},
		{
			name: "node with invalid IP address is skipped",
			nodes: &v1.NodeList{
				Items: []v1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node1",
							Annotations: map[string]string{
								"projectcalico.org/IPv4Address":        "not-an-ip/24",
								"projectcalico.org/IPv4IPIPTunnelAddr": "192.168.0.10",
								"projectcalico.org/ASNumber":           "4230000001",
							},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "couldn't find site",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create storage with empty data
			storage := &netrisstorage.Storage{
				SubnetsStorage: &netrisstorage.SubnetsStorage{},
				SitesStorage:   &netrisstorage.SitesStorage{},
				VNetStorage:    &netrisstorage.VNetStorage{},
			}

			w := &Watcher{
				NStorage: storage,
				data: data{
					nodes: tt.nodes,
				},
			}

			err := w.nodesProcessing()

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("error = %q, want %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestNodesProcessingWithStorage(t *testing.T) {
	// Create storage with proper test data
	storage := &netrisstorage.Storage{
		SubnetsStorage: &netrisstorage.SubnetsStorage{
			Subnets: []*ipam.IPAM{
				{
					Prefix: "10.0.0.0/24",
					Sites:  []ipam.IDName{{ID: 1, Name: "test-site"}},
				},
			},
		},
		SitesStorage: &netrisstorage.SitesStorage{
			Sites: []*site.Site{
				{ID: 1, Name: "test-site", PublicAsn: 65000},
			},
		},
		VNetStorage: &netrisstorage.VNetStorage{
			VNets: []*vnet.VNet{
				{
					ID:   1,
					Name: "test-vnet",
					Gateways: []vnet.VNetGateway{
						{Prefix: "10.0.0.1/24"},
					},
				},
			},
		},
	}

	nodes := &v1.NodeList{
		Items: []v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "worker-node",
					Annotations: map[string]string{
						"projectcalico.org/IPv4Address":        "10.0.0.50/24",
						"projectcalico.org/IPv4IPIPTunnelAddr": "192.168.1.50",
						"projectcalico.org/ASNumber":           "4230000005",
					},
				},
			},
		},
	}

	w := &Watcher{
		NStorage: storage,
		data: data{
			nodes: nodes,
		},
	}

	err := w.nodesProcessing()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the data was set correctly
	if len(w.data.nodesMap) != 1 {
		t.Errorf("nodesMap length = %d, want 1", len(w.data.nodesMap))
	}

	if w.data.site == nil || w.data.site.Name != "test-site" {
		t.Errorf("site = %v, want test-site", w.data.site)
	}

	if w.data.vnetName != "test-vnet" {
		t.Errorf("vnetName = %q, want %q", w.data.vnetName, "test-vnet")
	}

	if w.data.vnetGW != "10.0.0.1/24" {
		t.Errorf("vnetGW = %q, want %q", w.data.vnetGW, "10.0.0.1/24")
	}
}

func TestNodesProcessingNoVnet(t *testing.T) {
	// Create storage with subnet and site but no matching vnet
	storage := &netrisstorage.Storage{
		SubnetsStorage: &netrisstorage.SubnetsStorage{
			Subnets: []*ipam.IPAM{
				{
					Prefix: "10.0.0.0/24",
					Sites:  []ipam.IDName{{ID: 1, Name: "test-site"}},
				},
			},
		},
		SitesStorage: &netrisstorage.SitesStorage{
			Sites: []*site.Site{
				{ID: 1, Name: "test-site", PublicAsn: 65000},
			},
		},
		VNetStorage: &netrisstorage.VNetStorage{
			VNets: []*vnet.VNet{},
		},
	}

	nodes := &v1.NodeList{
		Items: []v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "worker-node",
					Annotations: map[string]string{
						"projectcalico.org/IPv4Address":        "10.0.0.50/24",
						"projectcalico.org/IPv4IPIPTunnelAddr": "192.168.1.50",
						"projectcalico.org/ASNumber":           "4230000005",
					},
				},
			},
		},
	}

	w := &Watcher{
		NStorage: storage,
		data: data{
			nodes: nodes,
		},
	}

	err := w.nodesProcessing()
	if err == nil {
		t.Error("expected error when vnet not found, got nil")
	}
	if err != nil && err.Error() != "couldn't find vnet" {
		t.Errorf("error = %q, want %q", err.Error(), "couldn't find vnet")
	}
}

func TestNodesProcessingVnetWithZeroID(t *testing.T) {
	// Create storage where vnet is found but has ID=0
	storage := &netrisstorage.Storage{
		SubnetsStorage: &netrisstorage.SubnetsStorage{
			Subnets: []*ipam.IPAM{
				{
					Prefix: "10.0.0.0/24",
					Sites:  []ipam.IDName{{ID: 1, Name: "test-site"}},
				},
			},
		},
		SitesStorage: &netrisstorage.SitesStorage{
			Sites: []*site.Site{
				{ID: 1, Name: "test-site", PublicAsn: 65000},
			},
		},
		VNetStorage: &netrisstorage.VNetStorage{
			VNets: []*vnet.VNet{
				{
					ID:   0, // Zero ID
					Name: "test-vnet",
					Gateways: []vnet.VNetGateway{
						{Prefix: "10.0.0.1/24"},
					},
				},
			},
		},
	}

	nodes := &v1.NodeList{
		Items: []v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "worker-node",
					Annotations: map[string]string{
						"projectcalico.org/IPv4Address":        "10.0.0.50/24",
						"projectcalico.org/IPv4IPIPTunnelAddr": "192.168.1.50",
						"projectcalico.org/ASNumber":           "4230000005",
					},
				},
			},
		},
	}

	w := &Watcher{
		NStorage: storage,
		data: data{
			nodes: nodes,
		},
	}

	err := w.nodesProcessing()
	if err == nil {
		t.Error("expected error when vnet ID is 0, got nil")
	}
	if err != nil && err.Error() != "couldn't find vnet" {
		t.Errorf("error = %q, want %q", err.Error(), "couldn't find vnet")
	}
}

func TestNodesProcessingMultipleNodes(t *testing.T) {
	// Create storage with proper test data
	storage := &netrisstorage.Storage{
		SubnetsStorage: &netrisstorage.SubnetsStorage{
			Subnets: []*ipam.IPAM{
				{
					Prefix: "10.0.0.0/24",
					Sites:  []ipam.IDName{{ID: 1, Name: "test-site"}},
				},
			},
		},
		SitesStorage: &netrisstorage.SitesStorage{
			Sites: []*site.Site{
				{ID: 1, Name: "test-site", PublicAsn: 65000},
			},
		},
		VNetStorage: &netrisstorage.VNetStorage{
			VNets: []*vnet.VNet{
				{
					ID:   1,
					Name: "test-vnet",
					Gateways: []vnet.VNetGateway{
						{Prefix: "10.0.0.1/24"},
					},
				},
			},
		},
	}

	nodes := &v1.NodeList{
		Items: []v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node1",
					Annotations: map[string]string{
						"projectcalico.org/IPv4Address":        "10.0.0.10/24",
						"projectcalico.org/IPv4IPIPTunnelAddr": "192.168.1.10",
						"projectcalico.org/ASNumber":           "4230000001",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node2",
					Annotations: map[string]string{
						"projectcalico.org/IPv4Address":        "10.0.0.11/24",
						"projectcalico.org/IPv4IPIPTunnelAddr": "192.168.1.11",
						"projectcalico.org/ASNumber":           "4230000002",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node3",
					// Missing annotations - should be skipped
					Annotations: map[string]string{},
				},
			},
		},
	}

	w := &Watcher{
		NStorage: storage,
		data: data{
			nodes: nodes,
		},
	}

	err := w.nodesProcessing()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have 2 nodes in the map (node3 is skipped)
	if len(w.data.nodesMap) != 2 {
		t.Errorf("nodesMap length = %d, want 2", len(w.data.nodesMap))
	}
}

func TestNodesProcessingInvalidGateway(t *testing.T) {
	// Create storage with vnet that has invalid gateway
	storage := &netrisstorage.Storage{
		SubnetsStorage: &netrisstorage.SubnetsStorage{
			Subnets: []*ipam.IPAM{
				{
					Prefix: "10.0.0.0/24",
					Sites:  []ipam.IDName{{ID: 1, Name: "test-site"}},
				},
			},
		},
		SitesStorage: &netrisstorage.SitesStorage{
			Sites: []*site.Site{
				{ID: 1, Name: "test-site", PublicAsn: 65000},
			},
		},
		VNetStorage: &netrisstorage.VNetStorage{
			VNets: []*vnet.VNet{
				{
					ID:   1,
					Name: "test-vnet",
					Gateways: []vnet.VNetGateway{
						{Prefix: "invalid-gateway"},
					},
				},
			},
		},
	}

	nodes := &v1.NodeList{
		Items: []v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "worker-node",
					Annotations: map[string]string{
						"projectcalico.org/IPv4Address":        "10.0.0.50/24",
						"projectcalico.org/IPv4IPIPTunnelAddr": "192.168.1.50",
						"projectcalico.org/ASNumber":           "4230000005",
					},
				},
			},
		},
	}

	w := &Watcher{
		NStorage: storage,
		data: data{
			nodes: nodes,
		},
	}

	err := w.nodesProcessing()
	// Should return error for invalid gateway
	if err == nil {
		t.Error("expected error for invalid gateway, got nil")
	}
}

func TestNodesProcessingSubnetWithNoSites(t *testing.T) {
	// Create storage where subnet has no sites
	storage := &netrisstorage.Storage{
		SubnetsStorage: &netrisstorage.SubnetsStorage{
			Subnets: []*ipam.IPAM{
				{
					Prefix: "10.0.0.0/24",
					Sites:  []ipam.IDName{}, // Empty sites
				},
			},
		},
		SitesStorage: &netrisstorage.SitesStorage{
			Sites: []*site.Site{
				{ID: 0, Name: "", PublicAsn: 65000}, // ID 0 means no site found
			},
		},
		VNetStorage: &netrisstorage.VNetStorage{},
	}

	nodes := &v1.NodeList{
		Items: []v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "worker-node",
					Annotations: map[string]string{
						"projectcalico.org/IPv4Address":        "10.0.0.50/24",
						"projectcalico.org/IPv4IPIPTunnelAddr": "192.168.1.50",
						"projectcalico.org/ASNumber":           "4230000005",
					},
				},
			},
		},
	}

	w := &Watcher{
		NStorage: storage,
		data: data{
			nodes: nodes,
		},
	}

	err := w.nodesProcessing()
	// Should return an error since vnet won't be found
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// mockCalicoClient is a mock implementation of CalicoClient for testing
type mockCalicoClient struct {
	ipPools               []*calico.IPPool
	bgpConfs              []*calico.BGPConfiguration
	bgpPeer               *calico.BGPPeer
	getIPPoolErr          error
	getBGPConfigurationErr error
	updateBGPConfigurationErr error
	getBGPPeerErr         error
	createBGPPeerErr      error
	updateBGPPeerErr      error
	deleteBGPPeerErr      error
}

func (m *mockCalicoClient) GetIPPool(config *rest.Config) ([]*calico.IPPool, error) {
	return m.ipPools, m.getIPPoolErr
}

func (m *mockCalicoClient) GetBGPConfiguration(config *rest.Config) ([]*calico.BGPConfiguration, error) {
	return m.bgpConfs, m.getBGPConfigurationErr
}

func (m *mockCalicoClient) UpdateBGPConfiguration(bgpConf *calico.BGPConfiguration, config *rest.Config) error {
	return m.updateBGPConfigurationErr
}

func (m *mockCalicoClient) GetBGPPeer(name string, config *rest.Config) (*calico.BGPPeer, error) {
	return m.bgpPeer, m.getBGPPeerErr
}

func (m *mockCalicoClient) CreateBGPPeer(peer *calico.BGPPeer, config *rest.Config) error {
	return m.createBGPPeerErr
}

func (m *mockCalicoClient) UpdateBGPPeer(peer *calico.BGPPeer, config *rest.Config) error {
	return m.updateBGPPeerErr
}

func (m *mockCalicoClient) DeleteBGPPeer(peer *calico.BGPPeer, config *rest.Config) error {
	return m.deleteBGPPeerErr
}

func (m *mockCalicoClient) GenerateBGPPeer(name, namespace, ip string, asn int) *calico.BGPPeer {
	return &calico.BGPPeer{}
}

func TestGetIPPools(t *testing.T) {
	tests := []struct {
		name    string
		mock    *mockCalicoClient
		wantErr bool
		errMsg  string
	}{
		{
			name: "returns IP pools successfully",
			mock: &mockCalicoClient{
				ipPools: []*calico.IPPool{
					{Spec: calico.IPPoolSpec{CIDR: "192.168.0.0/16", BlockSize: 26}},
				},
			},
			wantErr: false,
		},
		{
			name: "returns error on get failure",
			mock: &mockCalicoClient{
				getIPPoolErr: fmt.Errorf("connection refused"),
			},
			wantErr: true,
		},
		{
			name: "returns error when no pools",
			mock: &mockCalicoClient{
				ipPools: []*calico.IPPool{},
			},
			wantErr: true,
			errMsg:  "IPPool is missing",
		},
		{
			name: "returns error when pool is nil",
			mock: &mockCalicoClient{
				ipPools: []*calico.IPPool{nil},
			},
			wantErr: true,
			errMsg:  "IPPool is missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Watcher{
				calicoClient: tt.mock,
			}

			pools, err := w.getIPPools()

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("error = %q, want %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if pools == nil {
					t.Error("expected pools, got nil")
				}
			}
		})
	}
}

func TestGetIPInfo(t *testing.T) {
	tests := []struct {
		name     string
		mock     *mockCalicoClient
		bgpConfs []*calico.BGPConfiguration
		wantErr  bool
	}{
		{
			name: "extracts IP info successfully",
			mock: &mockCalicoClient{
				ipPools: []*calico.IPPool{
					{Spec: calico.IPPoolSpec{CIDR: "192.168.0.0/16", BlockSize: 26}},
				},
			},
			bgpConfs: []*calico.BGPConfiguration{
				{
					Spec: calico.BGPConfigurationSpec{
						ServiceClusterIPs: []calico.ServiceClusterIPBlock{
							{CIDR: "10.96.0.0/12"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "returns error when getIPPools fails",
			mock: &mockCalicoClient{
				getIPPoolErr: fmt.Errorf("connection refused"),
			},
			bgpConfs: []*calico.BGPConfiguration{{}},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Watcher{
				calicoClient: tt.mock,
				data: data{
					bgpConfs: tt.bgpConfs,
				},
			}

			err := w.getIPInfo()

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if w.data.blockSize != 26 {
					t.Errorf("blockSize = %d, want 26", w.data.blockSize)
				}
				if w.data.clusterCIDR != "192.168.0.0/16" {
					t.Errorf("clusterCIDR = %q, want %q", w.data.clusterCIDR, "192.168.0.0/16")
				}
			}
		})
	}
}

func TestUpdateBGPConfMeshWithMock(t *testing.T) {
	enabled := true
	tests := []struct {
		name     string
		mock     *mockCalicoClient
		bgpConfs []*calico.BGPConfiguration
		enabled  bool
		wantErr  bool
	}{
		{
			name: "updates successfully",
			mock: &mockCalicoClient{
				updateBGPConfigurationErr: nil,
			},
			bgpConfs: []*calico.BGPConfiguration{
				{
					Spec: calico.BGPConfigurationSpec{
						NodeToNodeMeshEnabled: &enabled,
					},
				},
			},
			enabled: false,
			wantErr: false,
		},
		{
			name: "returns error on update failure",
			mock: &mockCalicoClient{
				updateBGPConfigurationErr: fmt.Errorf("update failed"),
			},
			bgpConfs: []*calico.BGPConfiguration{
				{
					Spec: calico.BGPConfigurationSpec{
						NodeToNodeMeshEnabled: &enabled,
					},
				},
			},
			enabled: false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Watcher{
				calicoClient: tt.mock,
				data: data{
					bgpConfs: tt.bgpConfs,
				},
			}

			err := w.updateBGPConfMesh(tt.enabled)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

