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
	"github.com/netrisai/netris-operator/calicowatcher/calico"
	"k8s.io/client-go/rest"
)

// CalicoClient is an interface for Calico operations that can be mocked for testing
type CalicoClient interface {
	GetIPPool(config *rest.Config) ([]*calico.IPPool, error)
	GetBGPConfiguration(config *rest.Config) ([]*calico.BGPConfiguration, error)
	UpdateBGPConfiguration(bgpConf *calico.BGPConfiguration, config *rest.Config) error
	GetBGPPeer(name string, config *rest.Config) (*calico.BGPPeer, error)
	CreateBGPPeer(peer *calico.BGPPeer, config *rest.Config) error
	UpdateBGPPeer(peer *calico.BGPPeer, config *rest.Config) error
	DeleteBGPPeer(peer *calico.BGPPeer, config *rest.Config) error
	GenerateBGPPeer(name, namespace, ip string, asn int) *calico.BGPPeer
}

// calicoClientWrapper wraps the calico.Calico to implement CalicoClient interface
type calicoClientWrapper struct {
	calico *calico.Calico
}

// NewCalicoClient creates a new CalicoClient wrapper
func NewCalicoClient(c *calico.Calico) CalicoClient {
	return &calicoClientWrapper{calico: c}
}

func (w *calicoClientWrapper) GetIPPool(config *rest.Config) ([]*calico.IPPool, error) {
	return w.calico.GetIPPool(config)
}

func (w *calicoClientWrapper) GetBGPConfiguration(config *rest.Config) ([]*calico.BGPConfiguration, error) {
	return w.calico.GetBGPConfiguration(config)
}

func (w *calicoClientWrapper) UpdateBGPConfiguration(bgpConf *calico.BGPConfiguration, config *rest.Config) error {
	return w.calico.UpdateBGPConfiguration(bgpConf, config)
}

func (w *calicoClientWrapper) GetBGPPeer(name string, config *rest.Config) (*calico.BGPPeer, error) {
	return w.calico.GetBGPPeer(name, config)
}

func (w *calicoClientWrapper) CreateBGPPeer(peer *calico.BGPPeer, config *rest.Config) error {
	return w.calico.CreateBGPPeer(peer, config)
}

func (w *calicoClientWrapper) UpdateBGPPeer(peer *calico.BGPPeer, config *rest.Config) error {
	return w.calico.UpdateBGPPeer(peer, config)
}

func (w *calicoClientWrapper) DeleteBGPPeer(peer *calico.BGPPeer, config *rest.Config) error {
	return w.calico.DeleteBGPPeer(peer, config)
}

func (w *calicoClientWrapper) GenerateBGPPeer(name, namespace, ip string, asn int) *calico.BGPPeer {
	return w.calico.GenerateBGPPeer(name, namespace, ip, asn)
}
