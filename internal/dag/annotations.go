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

package dag

import (
	"fmt"
	"strconv"
	"strings"

	"k8s.io/api/extensions/v1beta1"
)

const (
	// set docs/annotations.md for details of how these annotations
	// are applied by Contour.

	annotationRequestTimeout     = "contour.heptio.com/request-timeout"
	annotationWebsocketRoutes    = "contour.heptio.com/websocket-routes"
	annotationUpstreamProtocol   = "contour.heptio.com/upstream-protocol"
	annotationMaxConnections     = "contour.heptio.com/max-connections"
	annotationMaxPendingRequests = "contour.heptio.com/max-pending-requests"
	annotationMaxRequests        = "contour.heptio.com/max-requests"
	annotationMaxRetries         = "contour.heptio.com/max-retries"
	annotationRetryOn            = "contour.heptio.com/retry-on"
	annotationNumRetries         = "contour.heptio.com/num-retries"
	annotationPerTryTimeout      = "contour.heptio.com/per-try-timeout"
)

// parseUInt32 parses the supplied string as if it were a uint32.
// If the value is not present, or malformed, or outside uint32's range, zero is returned.
func parseUInt32(s string) uint32 {
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0
	}
	return uint32(v)
}

// parseUpstreamProtocols parses the annotations map for a contour.heptio.com/upstream-protocol.{protocol}
// where 'protocol' identifies which protocol must be used in the upstream.
// If the value is not present, or malformed, then an empty map is returned.
func parseUpstreamProtocols(annotations map[string]string, annotation string, protocols ...string) map[string]string {
	up := make(map[string]string)
	for _, protocol := range protocols {
		ports := annotations[fmt.Sprintf("%s.%s", annotation, protocol)]
		for _, v := range strings.Split(ports, ",") {
			port := strings.TrimSpace(v)
			if port != "" {
				up[port] = protocol
			}
		}
	}
	return up
}

// httpAllowed returns true unless the kubernetes.io/ingress.allow-http annotation is
// present and set to false.
func httpAllowed(i *v1beta1.Ingress) bool {
	return !(i.Annotations["kubernetes.io/ingress.allow-http"] == "false")
}

// tlsRequired returns true if the ingress.kubernetes.io/force-ssl-redirect annotation is
// present and set to true.
func tlsRequired(i *v1beta1.Ingress) bool {
	return i.Annotations["ingress.kubernetes.io/force-ssl-redirect"] == "true"
}

func websocketRoutes(i *v1beta1.Ingress) map[string]bool {
	routes := make(map[string]bool)
	for _, v := range strings.Split(i.Annotations[annotationWebsocketRoutes], ",") {
		route := strings.TrimSpace(v)
		if route != "" {
			routes[route] = true
		}
	}
	return routes
}

// getIngressClassAnnotation checks for the acceptable ingress class annotations
// 1. contour.heptio.com/ingress.class
// 2. kubernetes.io/ingress.class
//
// it returns the first matching ingress annotation (in the above order) with test
func getIngressClassAnnotation(annotations map[string]string) string {
	class, ok := annotations["contour.heptio.com/ingress.class"]
	if ok {
		return class
	}

	class, ok = annotations["kubernetes.io/ingress.class"]
	if ok {
		return class
	}

	return ""
}
