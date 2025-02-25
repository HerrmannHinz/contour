// Copyright © 2018 Heptio
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package contour

import (
	"sort"
	"sync"

	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/golang/protobuf/proto"
	"github.com/heptio/contour/internal/dag"
	"github.com/heptio/contour/internal/envoy"
)

// ClusterCache manages the contents of the gRPC CDS cache.
type ClusterCache struct {
	mu     sync.Mutex
	values map[string]*envoy_api_v2.Cluster
	Cond
}

// Update replaces the contents of the cache with the supplied map.
func (c *ClusterCache) Update(v map[string]*envoy_api_v2.Cluster) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.values = v
	c.Cond.Notify()
}

// Contents returns a copy of the cache's contents.
func (c *ClusterCache) Contents() []proto.Message {
	c.mu.Lock()
	defer c.mu.Unlock()
	var values []proto.Message
	for _, v := range c.values {
		values = append(values, v)
	}
	sort.Stable(clusterByName(values))
	return values
}

func (c *ClusterCache) Query(names []string) []proto.Message {
	c.mu.Lock()
	defer c.mu.Unlock()
	var values []proto.Message
	for _, n := range names {
		// if the cluster is not registered we cannot return
		// a blank cluster because each cluster has a required
		// discovery type; DNS, EDS, etc. We cannot determine the
		// correct value for this property from the cluster's name
		// provided by the query so we must not return a blank cluster.
		if v, ok := c.values[n]; ok {
			values = append(values, v)
		}
	}
	sort.Stable(clusterByName(values))
	return values
}

type clusterByName []proto.Message

func (c clusterByName) Len() int      { return len(c) }
func (c clusterByName) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c clusterByName) Less(i, j int) bool {
	return c[i].(*envoy_api_v2.Cluster).Name < c[j].(*envoy_api_v2.Cluster).Name
}

func (*ClusterCache) TypeURL() string { return cache.ClusterType }

type clusterVisitor struct {
	clusters map[string]*envoy_api_v2.Cluster
}

// visitCluster produces a map of *envoy_api_v2.Clusters.
func visitClusters(root dag.Vertex) map[string]*envoy_api_v2.Cluster {
	cv := clusterVisitor{
		clusters: make(map[string]*envoy_api_v2.Cluster),
	}
	cv.visit(root)
	return cv.clusters
}

func (v *clusterVisitor) visit(vertex dag.Vertex) {
	if cluster, ok := vertex.(*dag.Cluster); ok {
		name := envoy.Clustername(cluster)
		if _, ok := v.clusters[name]; !ok {
			c := envoy.Cluster(cluster)
			v.clusters[c.Name] = c
		}
	}

	// recurse into children of v
	vertex.Visit(v.visit)
}
