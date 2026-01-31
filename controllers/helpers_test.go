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

package controllers

import (
	"strings"

	"github.com/go-logr/logr"
	k8sv1alpha1 "github.com/netrisai/netris-operator/api/v1alpha1"
	"github.com/netrisai/netris-operator/netrisstorage"
	"github.com/netrisai/netriswebapi/v2/types/inventory"
	"github.com/netrisai/netriswebapi/v2/types/ipam"
	"github.com/netrisai/netriswebapi/v2/types/site"
	"github.com/netrisai/netriswebapi/v2/types/vpc"
	"k8s.io/apimachinery/pkg/runtime"
)

// containsSubstr checks if s contains substr.
func containsSubstr(s, substr string) bool {
	return strings.Contains(s, substr)
}

// testLogger is a no-op logger for testing.
// logr v0.1.0 doesn't have Discard(), so we implement our own.
type testLogger struct{}

func (t testLogger) Info(_ string, _ ...interface{})           {}
func (t testLogger) Enabled() bool                             { return false }
func (t testLogger) Error(_ error, _ string, _ ...interface{}) {}
func (t testLogger) V(_ int) logr.InfoLogger                   { return t }
func (t testLogger) WithValues(_ ...interface{}) logr.Logger   { return t }
func (t testLogger) WithName(_ string) logr.Logger             { return t }

// newTestLogger returns a no-op logger for testing.
func newTestLogger() logr.Logger {
	return testLogger{}
}

// newTestScheme creates a scheme with the operator's CRDs registered.
func newTestScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = k8sv1alpha1.AddToScheme(scheme)
	return scheme
}

// newTestStorage creates a Storage with pre-populated sites for testing.
// Also initializes all storage types to avoid nil pointer issues.
func newTestStorage(sites []*site.Site) *netrisstorage.Storage {
	sitesStorage := netrisstorage.NewSitesStorage()
	sitesStorage.Sites = sites
	return &netrisstorage.Storage{
		SitesStorage:            sitesStorage,
		PortsStorage:            netrisstorage.NewPortStorage(),
		TenantsStorage:          netrisstorage.NewTenantsStorage(),
		SubnetsStorage:          netrisstorage.NewSubnetsStorage(),
		HWsStorage:              netrisstorage.NewHWsStorage(),
		VPCStorage:              netrisstorage.NewVPCStorage(),
		VNetStorage:             netrisstorage.NewVNetStorage(),
		BGPStorage:              netrisstorage.NewBGPStorage(),
		L4LBStorage:             netrisstorage.NewL4LBStorage(),
		LinksStorage:            netrisstorage.NewLinksStorage(),
		NATStorage:              netrisstorage.NewNATStorage(),
		InventoryProfileStorage: netrisstorage.NewInventoryProfileStorage(),
	}
}

// newTestStorageWithHWs creates a Storage with pre-populated sites and hardware for testing.
func newTestStorageWithHWs(sites []*site.Site, hws []*inventory.HW) *netrisstorage.Storage {
	sitesStorage := netrisstorage.NewSitesStorage()
	sitesStorage.Sites = sites
	hwsStorage := netrisstorage.NewHWsStorage()
	hwsStorage.HWs = hws
	return &netrisstorage.Storage{
		SitesStorage:            sitesStorage,
		PortsStorage:            netrisstorage.NewPortStorage(),
		TenantsStorage:          netrisstorage.NewTenantsStorage(),
		SubnetsStorage:          netrisstorage.NewSubnetsStorage(),
		HWsStorage:              hwsStorage,
		VPCStorage:              netrisstorage.NewVPCStorage(),
		VNetStorage:             netrisstorage.NewVNetStorage(),
		BGPStorage:              netrisstorage.NewBGPStorage(),
		L4LBStorage:             netrisstorage.NewL4LBStorage(),
		LinksStorage:            netrisstorage.NewLinksStorage(),
		NATStorage:              netrisstorage.NewNATStorage(),
		InventoryProfileStorage: netrisstorage.NewInventoryProfileStorage(),
	}
}

// newTestStorageWithSubnets creates a Storage with pre-populated subnets for testing.
// This avoids the API download attempt when FindByID doesn't find a match.
func newTestStorageWithSubnets(subnets []*ipam.IPAM) *netrisstorage.Storage {
	subnetsStorage := netrisstorage.NewSubnetsStorage()
	subnetsStorage.Subnets = subnets
	return &netrisstorage.Storage{
		SitesStorage:            netrisstorage.NewSitesStorage(),
		PortsStorage:            netrisstorage.NewPortStorage(),
		TenantsStorage:          netrisstorage.NewTenantsStorage(),
		SubnetsStorage:          subnetsStorage,
		HWsStorage:              netrisstorage.NewHWsStorage(),
		VPCStorage:              netrisstorage.NewVPCStorage(),
		VNetStorage:             netrisstorage.NewVNetStorage(),
		BGPStorage:              netrisstorage.NewBGPStorage(),
		L4LBStorage:             netrisstorage.NewL4LBStorage(),
		LinksStorage:            netrisstorage.NewLinksStorage(),
		NATStorage:              netrisstorage.NewNATStorage(),
		InventoryProfileStorage: netrisstorage.NewInventoryProfileStorage(),
	}
}

// newTestStorageWithVPCs creates a Storage with pre-populated VPCs for testing.
func newTestStorageWithVPCs(vpcs []*vpc.VPC) *netrisstorage.Storage {
	vpcStorage := netrisstorage.NewVPCStorage()
	vpcStorage.VPCs = vpcs
	return &netrisstorage.Storage{
		SitesStorage:            netrisstorage.NewSitesStorage(),
		PortsStorage:            netrisstorage.NewPortStorage(),
		TenantsStorage:          netrisstorage.NewTenantsStorage(),
		SubnetsStorage:          netrisstorage.NewSubnetsStorage(),
		HWsStorage:              netrisstorage.NewHWsStorage(),
		VPCStorage:              vpcStorage,
		VNetStorage:             netrisstorage.NewVNetStorage(),
		BGPStorage:              netrisstorage.NewBGPStorage(),
		L4LBStorage:             netrisstorage.NewL4LBStorage(),
		LinksStorage:            netrisstorage.NewLinksStorage(),
		NATStorage:              netrisstorage.NewNATStorage(),
		InventoryProfileStorage: netrisstorage.NewInventoryProfileStorage(),
	}
}
