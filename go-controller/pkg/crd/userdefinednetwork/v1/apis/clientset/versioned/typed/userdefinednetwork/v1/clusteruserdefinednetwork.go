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
// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"context"

	v1 "github.com/ovn-org/ovn-kubernetes/go-controller/pkg/crd/userdefinednetwork/v1"
	userdefinednetworkv1 "github.com/ovn-org/ovn-kubernetes/go-controller/pkg/crd/userdefinednetwork/v1/apis/applyconfiguration/userdefinednetwork/v1"
	scheme "github.com/ovn-org/ovn-kubernetes/go-controller/pkg/crd/userdefinednetwork/v1/apis/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
)

// ClusterUserDefinedNetworksGetter has a method to return a ClusterUserDefinedNetworkInterface.
// A group's client should implement this interface.
type ClusterUserDefinedNetworksGetter interface {
	ClusterUserDefinedNetworks() ClusterUserDefinedNetworkInterface
}

// ClusterUserDefinedNetworkInterface has methods to work with ClusterUserDefinedNetwork resources.
type ClusterUserDefinedNetworkInterface interface {
	Create(ctx context.Context, clusterUserDefinedNetwork *v1.ClusterUserDefinedNetwork, opts metav1.CreateOptions) (*v1.ClusterUserDefinedNetwork, error)
	Update(ctx context.Context, clusterUserDefinedNetwork *v1.ClusterUserDefinedNetwork, opts metav1.UpdateOptions) (*v1.ClusterUserDefinedNetwork, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, clusterUserDefinedNetwork *v1.ClusterUserDefinedNetwork, opts metav1.UpdateOptions) (*v1.ClusterUserDefinedNetwork, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.ClusterUserDefinedNetwork, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.ClusterUserDefinedNetworkList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.ClusterUserDefinedNetwork, err error)
	Apply(ctx context.Context, clusterUserDefinedNetwork *userdefinednetworkv1.ClusterUserDefinedNetworkApplyConfiguration, opts metav1.ApplyOptions) (result *v1.ClusterUserDefinedNetwork, err error)
	// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
	ApplyStatus(ctx context.Context, clusterUserDefinedNetwork *userdefinednetworkv1.ClusterUserDefinedNetworkApplyConfiguration, opts metav1.ApplyOptions) (result *v1.ClusterUserDefinedNetwork, err error)
	ClusterUserDefinedNetworkExpansion
}

// clusterUserDefinedNetworks implements ClusterUserDefinedNetworkInterface
type clusterUserDefinedNetworks struct {
	*gentype.ClientWithListAndApply[*v1.ClusterUserDefinedNetwork, *v1.ClusterUserDefinedNetworkList, *userdefinednetworkv1.ClusterUserDefinedNetworkApplyConfiguration]
}

// newClusterUserDefinedNetworks returns a ClusterUserDefinedNetworks
func newClusterUserDefinedNetworks(c *K8sV1Client) *clusterUserDefinedNetworks {
	return &clusterUserDefinedNetworks{
		gentype.NewClientWithListAndApply[*v1.ClusterUserDefinedNetwork, *v1.ClusterUserDefinedNetworkList, *userdefinednetworkv1.ClusterUserDefinedNetworkApplyConfiguration](
			"clusteruserdefinednetworks",
			c.RESTClient(),
			scheme.ParameterCodec,
			"",
			func() *v1.ClusterUserDefinedNetwork { return &v1.ClusterUserDefinedNetwork{} },
			func() *v1.ClusterUserDefinedNetworkList { return &v1.ClusterUserDefinedNetworkList{} }),
	}
}