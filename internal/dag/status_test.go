// Copyright © 2017 Heptio
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
	"testing"

	"github.com/google/go-cmp/cmp"
	ingressroutev1 "github.com/heptio/contour/apis/contour/v1beta1"
	projcontour "github.com/heptio/contour/apis/projectcontour/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestDAGIngressRouteStatus(t *testing.T) {
	sec1 := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ssl-cert",
			Namespace: "roots",
		},
		Type: v1.SecretTypeTLS,
		Data: secretdata("certificate", "key"),
	}

	// ir1 is a valid ingressroute
	ir1 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "example",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "example.com",
			},
			Routes: []ingressroutev1.Route{{
				Match: "/foo",
				Services: []ingressroutev1.Service{{
					Name: "home",
					Port: 8080,
				}},
			}, {
				Match: "/prefix",
				Delegate: &ingressroutev1.Delegate{
					Name: "delegated",
				}},
			},
		},
	}

	// ir2 is invalid because it contains a service with negative port
	ir2 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "example",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "example.com",
			},
			Routes: []ingressroutev1.Route{{
				Match: "/foo",
				Services: []ingressroutev1.Service{{
					Name: "home",
					Port: -80,
				}},
			}},
		},
	}

	// ir3 is invalid because it lives outside the roots namespace
	ir3 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "finance",
			Name:      "example",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "example.com",
			},
			Routes: []ingressroutev1.Route{{
				Match: "/foobar",
				Services: []ingressroutev1.Service{{
					Name: "home",
					Port: 8080,
				}},
			}},
		},
	}

	// ir4 is invalid because its match prefix does not match its parent's (ir1)
	ir4 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "delegated",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			Routes: []ingressroutev1.Route{{
				Match: "/doesnotmatch",
				Services: []ingressroutev1.Service{{
					Name: "home",
					Port: 8080,
				}},
			}},
		},
	}

	// ir6 is invalid because it delegates to itself, producing a cycle
	ir6 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "self",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "example.com",
			},
			Routes: []ingressroutev1.Route{{
				Match: "/foo",
				Delegate: &ingressroutev1.Delegate{
					Name: "self",
				},
			}},
		},
	}

	// ir7 delegates to ir8, which is invalid because it delegates back to ir7
	ir7 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "parent",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "example.com",
			},
			Routes: []ingressroutev1.Route{{
				Match: "/foo",
				Delegate: &ingressroutev1.Delegate{
					Name: "child",
				},
			}},
		},
	}

	ir8 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "child",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			Routes: []ingressroutev1.Route{{
				Match: "/foo",
				Delegate: &ingressroutev1.Delegate{
					Name: "child",
				},
			}},
		},
	}

	// ir9 is invalid because it has a route that both delegates and has a list of services
	ir9 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "parent",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "example.com",
			},
			Routes: []ingressroutev1.Route{{
				Match: "/foo",
				Delegate: &ingressroutev1.Delegate{
					Name: "child",
				},
				Services: []ingressroutev1.Service{{
					Name: "kuard",
					Port: 8080,
				}},
			}},
		},
	}

	// ir10 delegates to ir11 and ir 12.
	ir10 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "parent",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "example.com",
			},
			Routes: []ingressroutev1.Route{{
				Match: "/foo",
				Delegate: &ingressroutev1.Delegate{
					Name: "validChild",
				},
			}, {
				Match: "/bar",
				Delegate: &ingressroutev1.Delegate{
					Name: "invalidChild",
				},
			}},
		},
	}

	ir11 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "validChild",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			Routes: []ingressroutev1.Route{{
				Match: "/foo",
				Services: []ingressroutev1.Service{{
					Name: "foo2",
					Port: 8080,
				}},
			}},
		},
	}

	// ir12 is invalid because it contains an invalid port
	ir12 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "invalidChild",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			Routes: []ingressroutev1.Route{{
				Match: "/bar",
				Services: []ingressroutev1.Service{{
					Name: "foo3",
					Port: 12345678,
				}},
			}},
		},
	}

	// ir13 is invalid because it does not specify and FQDN
	ir13 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "parent",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &projcontour.VirtualHost{},
			Routes: []ingressroutev1.Route{{
				Match: "/foo",
				Services: []ingressroutev1.Service{{
					Name: "foo",
					Port: 8080,
				}},
			}},
		},
	}

	// ir14 delegates tp ir15 but it is invalid because it is missing fqdn
	ir14 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "invalidParent",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &projcontour.VirtualHost{},
			Routes: []ingressroutev1.Route{{
				Match: "/foo",
				Delegate: &ingressroutev1.Delegate{
					Name: "validChild",
				},
			}},
		},
	}

	// ir15 is invalid because it contains a wildcarded fqdn
	ir15 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "example",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "example.*.com",
			},
			Routes: []ingressroutev1.Route{{
				Match: "/foo",
				Services: []ingressroutev1.Service{{
					Name: "home",
					Port: 8080,
				}},
			}},
		},
	}

	// ir16 is invalid because it references an invalid service
	ir16 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "invalidir",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "example.com",
			},
			Routes: []ingressroutev1.Route{{
				Match: "/foo",
				Services: []ingressroutev1.Service{{
					Name: "invalid",
					Port: 8080,
				}},
			}},
		},
	}

	ir17 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-com",
			Namespace: "roots",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "example.com",
			},
			Routes: []ingressroutev1.Route{{
				Match: "/",
				Services: []ingressroutev1.Service{{
					Name: "kuard",
					Port: 8080,
				}},
			}},
		},
	}

	// ir18 reuses the fqdn used in ir17
	ir18 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "other-example",
			Namespace: "roots",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "example.com",
			},
			Routes: []ingressroutev1.Route{{
				Match: "/",
				Services: []ingressroutev1.Service{{
					Name: "kuard",
					Port: 8080,
				}},
			}},
		},
	}

	s1 := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kuard",
			Namespace: ir17.Namespace,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{{
				Name:       "http",
				Protocol:   "TCP",
				Port:       8080,
				TargetPort: intstr.FromInt(8080),
			}},
		},
	}

	s4 := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "home",
			Namespace: "roots",
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{{
				Name:     "http",
				Protocol: "TCP",
				Port:     8080,
			}},
		},
	}

	s5 := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "parent",
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{{
				Name:     "http",
				Protocol: "TCP",
				Port:     8080,
			}},
		},
	}

	s6 := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "foo2",
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{{
				Name:     "http",
				Protocol: "TCP",
				Port:     8080,
			}},
		},
	}

	s7 := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo3",
			Namespace: "roots",
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{{
				Name:     "http",
				Protocol: "TCP",
				Port:     12345678,
			}},
		},
	}

	ir20 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "root-blog",
			Namespace: "roots",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "blog.containersteve.com",
				TLS: &projcontour.TLS{
					SecretName: "blog-containersteve-com",
				},
			},
			Routes: []ingressroutev1.Route{{
				Match: "/",
				Delegate: &ingressroutev1.Delegate{
					Name:      "blog",
					Namespace: "marketing",
				},
			}},
		},
	}

	ir21 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "blog",
			Namespace: "marketing",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "blog.containersteve.com",
				TLS: &projcontour.TLS{
					SecretName: "blog-containersteve-com",
				},
			},
			Routes: []ingressroutev1.Route{{
				Match: "/",
				Services: []ingressroutev1.Service{{
					Name: "green",
					Port: 80,
				}},
			}},
		},
	}

	ir22 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "root-blog",
			Namespace: "roots",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "blog.containersteve.com",
			},
			Routes: []ingressroutev1.Route{{
				Match: "/",
				Delegate: &ingressroutev1.Delegate{
					Name:      "blog",
					Namespace: "marketing",
				},
			}},
		},
	}

	ir23 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "blog",
			Namespace: "marketing",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "www.containersteve.com",
			},
			Routes: []ingressroutev1.Route{{
				Match: "/",
				Services: []ingressroutev1.Service{{
					Name: "green",
					Port: 80,
				}},
			}},
		},
	}

	s8 := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "green",
			Namespace: "marketing",
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{{
				Name:     "http",
				Protocol: "TCP",
				Port:     80,
			}},
		},
	}

	s9 := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx",
			Namespace: "roots",
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{{
				Protocol: "TCP",
				Port:     80,
			}},
		},
	}

	i1 := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx",
			Namespace: "roots",
		},
		Spec: v1beta1.IngressSpec{
			TLS: []v1beta1.IngressTLS{{
				Hosts:      []string{"example.com"},
				SecretName: sec1.Name,
			}},
			Rules: []v1beta1.IngressRule{{
				Host:             "example.com",
				IngressRuleValue: ingressrulevalue(backend(s9.Name, intstr.FromInt(80))),
			}},
		},
	}

	ir24 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx",
			Namespace: "roots",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "example.com",
				TLS: &projcontour.TLS{
					SecretName: sec1.Name,
				},
			},
			TCPProxy: &ingressroutev1.TCPProxy{
				Services: []ingressroutev1.Service{{
					Name: s9.Name,
					Port: 80,
				}},
			},
		},
	}

	sec2 := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-ssl-cert",
			Namespace: "heptio-contour",
		},
		Type: v1.SecretTypeTLS,
		Data: secretdata("certificate", "key"),
	}

	ir25 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample-app",
			Namespace: "roots",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "127.0.0.1.nip.io",
				TLS: &projcontour.TLS{
					SecretName: sec2.Namespace + "/" + sec2.Name,
				},
			},
			TCPProxy: &ingressroutev1.TCPProxy{
				Services: []ingressroutev1.Service{{
					Name: "sample-app",
					Port: 80,
				}},
			},
		},
	}

	ir26 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app-with-tls-delegation",
			Namespace: "roots",
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "app-with-tls-delegation.127.0.0.1.nip.io",
				TLS: &projcontour.TLS{
					SecretName: sec2.Namespace + "/" + sec2.Name,
				},
			},
			Routes: []ingressroutev1.Route{{
				Services: []ingressroutev1.Service{{
					Name: "sample-app",
					Port: 80,
				}},
			}},
		},
	}

	s10 := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tls-passthrough",
			Namespace: "roots",
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{{
				Name:       "https",
				Protocol:   "TCP",
				Port:       443,
				TargetPort: intstr.FromInt(443),
			}, {
				Name:       "http",
				Protocol:   "TCP",
				Port:       80,
				TargetPort: intstr.FromInt(80),
			}},
		},
	}

	ir27 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kuard-tcp",
			Namespace: s10.Namespace,
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "kuard.example.com",
				TLS: &projcontour.TLS{
					Passthrough: true,
				},
			},
			Routes: []ingressroutev1.Route{{
				Match: "/",
				Services: []ingressroutev1.Service{{
					Name: s10.Name,
					Port: 80, // proxy non secure traffic to port 80
				}},
			}},
			TCPProxy: &ingressroutev1.TCPProxy{
				Services: []ingressroutev1.Service{{
					Name: s10.Name,
					Port: 443, // ssl passthrough to secure port
				}},
			},
		},
	}

	ir28 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app-with-tls-delegation",
			Namespace: s10.Namespace,
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "app-with-tls-delegation.127.0.0.1.nip.io",
				TLS: &projcontour.TLS{
					SecretName: "heptio-contour/ssl-cert", // not delegated
				},
			},
			Routes: []ingressroutev1.Route{{
				Match: "/",
				Services: []ingressroutev1.Service{{
					Name: s10.Name,
					Port: 80,
				}},
			}},
		},
	}

	ir29 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "site1",
			Namespace: s1.Namespace,
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "site1.com",
			},
			Routes: []ingressroutev1.Route{{
				Match: "/",
				Delegate: &ingressroutev1.Delegate{
					Name:      "www",
					Namespace: s1.Namespace,
				},
			}},
		},
	}

	ir30 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "site2",
			Namespace: s1.Namespace,
		},
		Spec: ingressroutev1.IngressRouteSpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "site2.com",
			},
			Routes: []ingressroutev1.Route{{
				Match: "/foo",
				Delegate: &ingressroutev1.Delegate{
					Name:      "www",
					Namespace: s1.Namespace,
				},
			}},
		},
	}

	ir31 := &ingressroutev1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "www",
			Namespace: s1.Namespace,
		},
		Spec: ingressroutev1.IngressRouteSpec{
			Routes: []ingressroutev1.Route{{
				Match: "/foo",
				Services: []ingressroutev1.Service{{
					Name: "kuard",
					Port: 8080,
				}},
			}},
		},
	}

	tests := map[string]struct {
		objs []interface{}
		want map[Meta]Status
	}{
		"valid ingressroute": {
			objs: []interface{}{ir1, s4},
			want: map[Meta]Status{
				{name: ir1.Name, namespace: ir1.Namespace}: {Object: ir1, Status: "valid", Description: "valid IngressRoute", Vhost: "example.com"},
			},
		},
		"invalid port in service": {
			objs: []interface{}{ir2},
			want: map[Meta]Status{
				{name: ir2.Name, namespace: ir2.Namespace}: {Object: ir2, Status: "invalid", Description: `route "/foo": service "home": port must be in the range 1-65535`, Vhost: "example.com"},
			},
		},
		"root ingressroute outside of roots namespace": {
			objs: []interface{}{ir3},
			want: map[Meta]Status{
				{name: ir3.Name, namespace: ir3.Namespace}: {Object: ir3, Status: "invalid", Description: "root IngressRoute cannot be defined in this namespace"},
			},
		},
		"delegated route's match prefix does not match parent's prefix": {
			objs: []interface{}{ir1, ir4, s4},
			want: map[Meta]Status{
				{name: ir1.Name, namespace: ir1.Namespace}: {Object: ir1, Status: "valid", Description: "valid IngressRoute", Vhost: "example.com"},
				{name: ir4.Name, namespace: ir4.Namespace}: {Object: ir4, Status: "invalid", Description: `the path prefix "/doesnotmatch" does not match the parent's path prefix "/prefix"`},
			},
		},
		"root ingressroute does not specify FQDN": {
			objs: []interface{}{ir13},
			want: map[Meta]Status{
				{name: ir13.Name, namespace: ir13.Namespace}: {Object: ir13, Status: "invalid", Description: "Spec.VirtualHost.Fqdn must be specified"},
			},
		},
		"self-edge produces a cycle": {
			objs: []interface{}{ir6},
			want: map[Meta]Status{
				{name: ir6.Name, namespace: ir6.Namespace}: {
					Object:      ir6,
					Status:      "invalid",
					Description: "root ingressroute cannot delegate to another root ingressroute",
					Vhost:       "example.com",
				},
			},
		},
		"child delegates to parent, producing a cycle": {
			objs: []interface{}{ir7, ir8},
			want: map[Meta]Status{
				{name: ir7.Name, namespace: ir7.Namespace}: {
					Object:      ir7,
					Status:      "valid",
					Description: "valid IngressRoute",
					Vhost:       "example.com",
				},
				{name: ir8.Name, namespace: ir8.Namespace}: {
					Object:      ir8,
					Status:      "invalid",
					Description: "route creates a delegation cycle: roots/parent -> roots/child -> roots/child",
				},
			},
		},
		"route has a list of services and also delegates": {
			objs: []interface{}{ir9},
			want: map[Meta]Status{
				{name: ir9.Name, namespace: ir9.Namespace}: {Object: ir9, Status: "invalid", Description: `route "/foo": cannot specify services and delegate in the same route`, Vhost: "example.com"},
			},
		},
		"ingressroute is an orphaned route": {
			objs: []interface{}{ir8},
			want: map[Meta]Status{
				{name: ir8.Name, namespace: ir8.Namespace}: {Object: ir8, Status: "orphaned", Description: "this IngressRoute is not part of a delegation chain from a root IngressRoute"},
			},
		},
		"ingressroute delegates to multiple ingressroutes, one is invalid": {
			objs: []interface{}{ir10, ir11, ir12, s6, s7},
			want: map[Meta]Status{
				{name: ir11.Name, namespace: ir11.Namespace}: {Object: ir11, Status: "valid", Description: "valid IngressRoute"},
				{name: ir12.Name, namespace: ir12.Namespace}: {Object: ir12, Status: "invalid", Description: `route "/bar": service "foo3": port must be in the range 1-65535`},
				{name: ir10.Name, namespace: ir10.Namespace}: {Object: ir10, Status: "valid", Description: "valid IngressRoute", Vhost: "example.com"},
			},
		},
		"invalid parent orphans children": {
			objs: []interface{}{ir14, ir11},
			want: map[Meta]Status{
				{name: ir14.Name, namespace: ir14.Namespace}: {Object: ir14, Status: "invalid", Description: "Spec.VirtualHost.Fqdn must be specified"},
				{name: ir11.Name, namespace: ir11.Namespace}: {Object: ir11, Status: "orphaned", Description: "this IngressRoute is not part of a delegation chain from a root IngressRoute"},
			},
		},
		"multi-parent children is not orphaned when one of the parents is invalid": {
			objs: []interface{}{ir14, ir11, ir10, s5, s6},
			want: map[Meta]Status{
				{name: ir14.Name, namespace: ir14.Namespace}: {Object: ir14, Status: "invalid", Description: "Spec.VirtualHost.Fqdn must be specified"},
				{name: ir11.Name, namespace: ir11.Namespace}: {Object: ir11, Status: "valid", Description: "valid IngressRoute"},
				{name: ir10.Name, namespace: ir10.Namespace}: {Object: ir10, Status: "valid", Description: "valid IngressRoute", Vhost: "example.com"},
			},
		},
		"invalid FQDN contains wildcard": {
			objs: []interface{}{ir15},
			want: map[Meta]Status{
				{name: ir15.Name, namespace: ir15.Namespace}: {Object: ir15, Status: "invalid", Description: `Spec.VirtualHost.Fqdn "example.*.com" cannot use wildcards`, Vhost: "example.*.com"},
			},
		},
		"missing service shows invalid status": {
			objs: []interface{}{ir16},
			want: map[Meta]Status{
				{name: ir16.Name, namespace: ir16.Namespace}: {
					Object:      ir16,
					Status:      "invalid",
					Description: `Service [invalid:8080] is invalid or missing`,
					Vhost:       ir16.Spec.VirtualHost.Fqdn,
				},
			},
		},
		"insert ingressroute": {
			objs: []interface{}{s1, ir17},
			want: map[Meta]Status{
				{name: ir17.Name, namespace: ir17.Namespace}: {
					Object:      ir17,
					Status:      StatusValid,
					Description: "valid IngressRoute",
					Vhost:       "example.com",
				},
			},
		},
		"insert conflicting ingressroutes due to fqdn reuse": {
			objs: []interface{}{ir17, ir18},
			want: map[Meta]Status{
				{name: ir17.Name, namespace: ir17.Namespace}: {
					Object:      ir17,
					Status:      StatusInvalid,
					Description: `fqdn "example.com" is used in multiple IngressRoutes: roots/example-com, roots/other-example`,
					Vhost:       "example.com",
				},
				{name: ir18.Name, namespace: ir18.Namespace}: {
					Object:      ir18,
					Status:      StatusInvalid,
					Description: `fqdn "example.com" is used in multiple IngressRoutes: roots/example-com, roots/other-example`,
					Vhost:       "example.com",
				},
			},
		},
		"root ingress delegating to another root": {
			objs: []interface{}{ir20, ir21},
			want: map[Meta]Status{
				{name: ir20.Name, namespace: ir20.Namespace}: {
					Object:      ir20,
					Status:      StatusInvalid,
					Description: `fqdn "blog.containersteve.com" is used in multiple IngressRoutes: marketing/blog, roots/root-blog`,
					Vhost:       "blog.containersteve.com",
				},
				{name: ir21.Name, namespace: ir21.Namespace}: {
					Object:      ir21,
					Status:      StatusInvalid,
					Description: `fqdn "blog.containersteve.com" is used in multiple IngressRoutes: marketing/blog, roots/root-blog`,
					Vhost:       "blog.containersteve.com",
				},
			},
		},
		"root ingress delegating to another root w/ different hostname": {
			objs: []interface{}{ir22, ir23, s8},
			want: map[Meta]Status{
				{name: ir22.Name, namespace: ir22.Namespace}: {
					Object:      ir22,
					Status:      StatusInvalid,
					Description: "root ingressroute cannot delegate to another root ingressroute",
					Vhost:       "blog.containersteve.com",
				},
				{name: ir23.Name, namespace: ir23.Namespace}: {
					Object:      ir23,
					Status:      StatusValid,
					Description: `valid IngressRoute`,
					Vhost:       "www.containersteve.com",
				},
			},
		},
		// issue 1399
		"service shared across ingress and ingressroute tcpproxy": {
			objs: []interface{}{
				sec1, s9, i1, ir24,
			},
			want: map[Meta]Status{
				{name: ir24.Name, namespace: ir24.Namespace}: {
					Object:      ir24,
					Status:      StatusValid,
					Description: `valid IngressRoute`,
					Vhost:       "example.com",
				},
			},
		},
		// issue 1347
		"check status set when tcpproxy combined with tls delegation failure": {
			objs: []interface{}{
				sec2,
				ir25,
			},
			want: map[Meta]Status{
				{name: ir25.Name, namespace: ir25.Namespace}: {
					Object:      ir25,
					Status:      StatusInvalid,
					Description: sec2.Namespace + "/" + sec2.Name + ": certificate delegation not permitted",
					Vhost:       ir25.Spec.VirtualHost.Fqdn,
				},
			},
		},
		// issue 1348
		"check status set when routes combined with tls delegation failure": {
			objs: []interface{}{
				sec2,
				ir26,
			},
			want: map[Meta]Status{
				{name: ir26.Name, namespace: ir26.Namespace}: {
					Object:      ir26,
					Status:      StatusInvalid,
					Description: sec2.Namespace + "/" + sec2.Name + ": certificate delegation not permitted",
					Vhost:       ir26.Spec.VirtualHost.Fqdn,
				},
			},
		},
		// issue 910
		"non tls routes can be combined with tcp proxy": {
			objs: []interface{}{
				s10,
				ir27,
			},
			want: map[Meta]Status{
				{name: ir27.Name, namespace: ir27.Namespace}: {
					Object:      ir27,
					Status:      StatusValid,
					Description: `valid IngressRoute`,
					Vhost:       ir27.Spec.VirtualHost.Fqdn,
				},
			},
		},
		// issue 1452
		"ingressroute with missing secret delegation should be invalid": {
			objs: []interface{}{
				s10,
				ir28,
			},
			want: map[Meta]Status{
				{name: ir28.Name, namespace: ir28.Namespace}: {
					Object:      ir28,
					Status:      StatusInvalid,
					Description: "TLS Secret [heptio-contour/ssl-cert] not found or is malformed",
					Vhost:       ir28.Spec.VirtualHost.Fqdn,
				},
			},
		},
		"two root ingressroutes delegated to the same object should not conflict on hostname": {
			objs: []interface{}{
				s1, ir29, ir30, ir31,
			},
			want: map[Meta]Status{
				{name: ir29.Name, namespace: ir29.Namespace}: {
					Object:      ir29,
					Status:      "valid",
					Description: "valid IngressRoute",
					Vhost:       "site1.com",
				},
				{name: ir30.Name, namespace: ir30.Namespace}: {
					Object:      ir30,
					Status:      "valid",
					Description: "valid IngressRoute",
					Vhost:       "site2.com",
				},
				{name: ir31.Name, namespace: ir31.Namespace}: {
					Object:      ir31,
					Status:      "valid",
					Description: "valid IngressRoute",
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			builder := Builder{
				Source: KubernetesCache{
					RootNamespaces: []string{"roots", "marketing"},
					FieldLogger:    testLogger(t),
				},
			}
			for _, o := range tc.objs {
				builder.Source.Insert(o)
			}
			dag := builder.Build()

			got := dag.Statuses()

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestDAGHTTPProxyStatus(t *testing.T) {

	s1 := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "green",
			Namespace: "roots",
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{{
				Name:     "http",
				Protocol: "TCP",
				Port:     80,
			}},
		},
	}

	// proxy1 is a valid proxy
	proxy1 := &projcontour.HTTPProxy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "example",
		},
		Spec: projcontour.HTTPProxySpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "example.com",
			},
			Includes: []projcontour.Include{{
				Name:      "delegated",
				Namespace: "roots",
				Condition: projcontour.Condition{
					Prefix: "/prefix",
				},
			}},
			Routes: []projcontour.Route{{
				Condition: &projcontour.Condition{
					Prefix: "/foo",
				},
				Services: []projcontour.Service{{
					Name: "home",
					Port: 8080,
				}},
			}},
		},
	}

	// proxy2 is invalid because it contains a service with negative port
	proxy2 := &projcontour.HTTPProxy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "example",
		},
		Spec: projcontour.HTTPProxySpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "example.com",
			},
			Routes: []projcontour.Route{{
				Condition: &projcontour.Condition{
					Prefix: "/foo",
				},
				Services: []projcontour.Service{{
					Name: "home",
					Port: -80,
				}},
			}},
		},
	}

	// proxy3 is invalid because it lives outside the roots namespace
	proxy3 := &projcontour.HTTPProxy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "finance",
			Name:      "example",
		},
		Spec: projcontour.HTTPProxySpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "example.com",
			},
			Routes: []projcontour.Route{{
				Condition: &projcontour.Condition{
					Prefix: "/foobar",
				},
				Services: []projcontour.Service{{
					Name: "home",
					Port: 8080,
				}},
			}},
		},
	}

	// proxy6 is invalid because it delegates to itself, producing a cycle
	proxy6 := &projcontour.HTTPProxy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "self",
			Namespace: "roots",
		},
		Spec: projcontour.HTTPProxySpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "example.com",
			},
			Includes: []projcontour.Include{{
				Name:      "self",
				Namespace: "roots",
				Condition: projcontour.Condition{
					Prefix: "/foo",
				},
			}},
			Routes: []projcontour.Route{{
				Services: []projcontour.Service{{
					Name: "green",
					Port: 80,
				}},
			}},
		},
	}

	// proxy7 delegates to proxy8, which is invalid because proxy8 delegates back to proxy8
	proxy7 := &projcontour.HTTPProxy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "parent",
			Namespace: "roots",
		},
		Spec: projcontour.HTTPProxySpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "example.com",
			},
			Includes: []projcontour.Include{{
				Name:      "child",
				Namespace: "roots",
				Condition: projcontour.Condition{
					Prefix: "/foo",
				},
			}},
		},
	}

	proxy8 := &projcontour.HTTPProxy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "child",
			Namespace: "roots",
		},
		Spec: projcontour.HTTPProxySpec{
			Includes: []projcontour.Include{{
				Name:      "child",
				Namespace: "roots",
				Condition: projcontour.Condition{
					Prefix: "/foo",
				},
			}},
		},
	}

	//// proxy9 is invalid because it has a route that both delegates and has a list of services
	//proxy9 := &projcontour.HTTPProxy{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Namespace: "roots",
	//		Name:      "parent",
	//	},
	//	Spec: projcontour.HTTPProxySpec{
	//		VirtualHost: &projcontour.VirtualHost{
	//			Fqdn: "example.com",
	//		},
	//		Includes: []projcontour.Include{{
	//			Name:      "child",
	//			Namespace: "roots",
	//			Condition: projcontour.Condition{
	//				Prefix: "/foo",
	//			},
	//		}},
	//		Routes: []projcontour.Route{{
	//			Condition: &projcontour.Condition{
	//				Prefix: "/foo",
	//			},
	//			Services: []projcontour.Service{{
	//				Name: "kuard",
	//				Port: 8080,
	//			}},
	//		}},
	//	},
	//}

	//// proxy10 delegates to proxy11 and proxy12.
	//proxy10 := &projcontour.HTTPProxy{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Namespace: "roots",
	//		Name:      "parent",
	//	},
	//	Spec: projcontour.HTTPProxySpec{
	//		VirtualHost: &projcontour.VirtualHost{
	//			Fqdn: "example.com",
	//		},
	//		Includes: []projcontour.Include{{
	//			Name:      "validChild",
	//			Namespace: "roots",
	//			Condition: projcontour.Condition{
	//				Prefix: "/foo",
	//			},
	//		}, {
	//			Name:      "invalidChild",
	//			Namespace: "roots",
	//			Condition: projcontour.Condition{
	//				Prefix: "/bar",
	//			},
	//		}},
	//	},
	//}

	proxy11 := &projcontour.HTTPProxy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "validChild",
			Namespace: "roots",
		},
		Spec: projcontour.HTTPProxySpec{
			Routes: []projcontour.Route{{
				Services: []projcontour.Service{{
					Name: "foo2",
					Port: 8080,
				}},
			}},
		},
	}

	//// proxy12 is invalid because it contains an invalid port
	//proxy12 := &projcontour.HTTPProxy{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      "invalidChild",
	//		Namespace: "roots",
	//	},
	//	Spec: projcontour.HTTPProxySpec{
	//		Routes: []projcontour.Route{{
	//			Services: []projcontour.Service{{
	//				Name: "foo3",
	//				Port: 12345678,
	//			}},
	//		}},
	//	},
	//}

	// proxy13 is invalid because it does not specify and FQDN
	proxy13 := &projcontour.HTTPProxy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "parent",
		},
		Spec: projcontour.HTTPProxySpec{
			VirtualHost: &projcontour.VirtualHost{},
			Routes: []projcontour.Route{{
				Condition: &projcontour.Condition{
					Prefix: "/foo",
				},
				Services: []projcontour.Service{{
					Name: "foo",
					Port: 8080,
				}},
			}},
		},
	}

	// proxy14 delegates tp ir15 but it is invalid because it is missing fqdn
	proxy14 := &projcontour.HTTPProxy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "invalidParent",
		},
		Spec: projcontour.HTTPProxySpec{
			VirtualHost: &projcontour.VirtualHost{},
			Includes: []projcontour.Include{{
				Name:      "validChild",
				Namespace: "roots",
				Condition: projcontour.Condition{
					Prefix: "/foo",
				},
			}},
		},
	}

	// proxy15 is invalid because it contains a wildcarded fqdn
	proxy15 := &projcontour.HTTPProxy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "example",
		},
		Spec: projcontour.HTTPProxySpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "example.*.com",
			},
			Routes: []projcontour.Route{{
				Condition: &projcontour.Condition{
					Prefix: "/foo",
				},
				Services: []projcontour.Service{{
					Name: "home",
					Port: 8080,
				}},
			}},
		},
	}

	// proxy16 is invalid because it references an invalid service
	proxy16 := &projcontour.HTTPProxy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "roots",
			Name:      "invalidir",
		},
		Spec: projcontour.HTTPProxySpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "example.com",
			},
			Routes: []projcontour.Route{{
				Condition: &projcontour.Condition{
					Prefix: "/foo",
				},
				Services: []projcontour.Service{{
					Name: "invalid",
					Port: 8080,
				}},
			}},
		},
	}

	proxy17 := &projcontour.HTTPProxy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-com",
			Namespace: "roots",
		},
		Spec: projcontour.HTTPProxySpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "example.com",
			},
			Routes: []projcontour.Route{{
				Condition: &projcontour.Condition{
					Prefix: "/foo",
				},
				Services: []projcontour.Service{{
					Name: "kuard",
					Port: 8080,
				}},
			}},
		},
	}

	// proxy18 reuses the fqdn used in proxy17
	proxy18 := &projcontour.HTTPProxy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "other-example",
			Namespace: "roots",
		},
		Spec: projcontour.HTTPProxySpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "example.com",
			},
			Routes: []projcontour.Route{{
				Services: []projcontour.Service{{
					Name: "kuard",
					Port: 8080,
				}},
			}},
		},
	}

	proxy20 := &projcontour.HTTPProxy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "root-blog",
			Namespace: "roots",
		},
		Spec: projcontour.HTTPProxySpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "blog.containersteve.com",
				TLS: &projcontour.TLS{
					SecretName: "blog-containersteve-com",
				},
			},
			Includes: []projcontour.Include{{
				Name:      "blog",
				Namespace: "marketing",
				Condition: projcontour.Condition{
					Prefix: "/",
				},
			}},
		},
	}

	proxy21 := &projcontour.HTTPProxy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "blog",
			Namespace: "marketing",
		},
		Spec: projcontour.HTTPProxySpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "blog.containersteve.com",
				TLS: &projcontour.TLS{
					SecretName: "blog-containersteve-com",
				},
			},
			Routes: []projcontour.Route{{
				Services: []projcontour.Service{{
					Name: "green",
					Port: 80,
				}},
			}},
		},
	}

	proxy22 := &projcontour.HTTPProxy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "root-blog",
			Namespace: "roots",
		},
		Spec: projcontour.HTTPProxySpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "blog.containersteve.com",
			},
			Includes: []projcontour.Include{{
				Name:      "blog",
				Namespace: "marketing",
				Condition: projcontour.Condition{
					Prefix: "/",
				},
			}},
		},
	}

	proxy23 := &projcontour.HTTPProxy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "blog",
			Namespace: "marketing",
		},
		Spec: projcontour.HTTPProxySpec{
			VirtualHost: &projcontour.VirtualHost{
				Fqdn: "www.containersteve.com",
			},
			Routes: []projcontour.Route{{
				Services: []projcontour.Service{{
					Name: "green",
					Port: 80,
				}},
			}},
		},
	}

	s2 := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kuard",
			Namespace: "roots",
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{{
				Name:     "http",
				Protocol: "TCP",
				Port:     8080,
			}},
		},
	}

	s4 := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "home",
			Namespace: "roots",
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{{
				Name:     "http",
				Protocol: "TCP",
				Port:     8080,
			}},
		},
	}

	s8 := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "green",
			Namespace: "marketing",
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{{
				Name:     "http",
				Protocol: "TCP",
				Port:     80,
			}},
		},
	}

	tests := map[string]struct {
		objs []interface{}
		want map[Meta]Status
	}{
		"valid proxy": {
			objs: []interface{}{proxy1, s4},
			want: map[Meta]Status{
				{name: proxy1.Name, namespace: proxy1.Namespace}: {Object: proxy1, Status: "valid", Description: "valid HTTPProxy", Vhost: "example.com"},
			},
		},
		"invalid port in service": {
			objs: []interface{}{proxy2},
			want: map[Meta]Status{
				{name: proxy2.Name, namespace: proxy2.Namespace}: {Object: proxy2, Status: "invalid", Description: `route "/foo": service "home": port must be in the range 1-65535`, Vhost: "example.com"},
			},
		},
		"root proxy outside of roots namespace": {
			objs: []interface{}{proxy3},
			want: map[Meta]Status{
				{name: proxy3.Name, namespace: proxy3.Namespace}: {Object: proxy3, Status: "invalid", Description: "root HTTPProxy cannot be defined in this namespace"},
			},
		},
		"root proxy does not specify FQDN": {
			objs: []interface{}{proxy13},
			want: map[Meta]Status{
				{name: proxy13.Name, namespace: proxy13.Namespace}: {Object: proxy13, Status: "invalid", Description: "Spec.VirtualHost.Fqdn must be specified"},
			},
		},
		"self-edge produces a cycle": {
			objs: []interface{}{proxy6, s1},
			want: map[Meta]Status{
				{name: proxy6.Name, namespace: proxy6.Namespace}: {
					Object:      proxy6,
					Status:      "invalid",
					Description: "root httpproxy cannot delegate to another root httpproxy",
					Vhost:       "example.com",
				},
			},
		},
		"child delegates to parent, producing a cycle": {
			objs: []interface{}{proxy7, proxy8},
			want: map[Meta]Status{
				{name: proxy7.Name, namespace: proxy7.Namespace}: {
					Object:      proxy7,
					Status:      "valid",
					Description: "valid HTTPProxy",
					Vhost:       "example.com",
				},
				{name: proxy8.Name, namespace: proxy8.Namespace}: {
					Object:      proxy8,
					Status:      "invalid",
					Description: "include creates a delegation cycle: roots/parent -> roots/child -> roots/child",
					Vhost:       "example.com",
				},
			},
		},
		"orphaned route": {
			objs: []interface{}{proxy8},
			want: map[Meta]Status{
				{name: proxy8.Name, namespace: proxy8.Namespace}: {Object: proxy8, Status: "orphaned", Description: "this HTTPProxy is not part of a delegation chain from a root HTTPProxy"},
			},
		},
		"invalid parent orphans children": {
			objs: []interface{}{proxy14, proxy11},
			want: map[Meta]Status{
				{name: proxy14.Name, namespace: proxy14.Namespace}: {Object: proxy14, Status: "invalid", Description: "Spec.VirtualHost.Fqdn must be specified"},
				{name: proxy11.Name, namespace: proxy11.Namespace}: {Object: proxy11, Status: "orphaned", Description: "this HTTPProxy is not part of a delegation chain from a root HTTPProxy"},
			},
		},
		"invalid FQDN contains wildcard": {
			objs: []interface{}{proxy15},
			want: map[Meta]Status{
				{name: proxy15.Name, namespace: proxy15.Namespace}: {Object: proxy15, Status: "invalid", Description: `Spec.VirtualHost.Fqdn "example.*.com" cannot use wildcards`, Vhost: "example.*.com"},
			},
		},
		"missing service shows invalid status": {
			objs: []interface{}{proxy16},
			want: map[Meta]Status{
				{name: proxy16.Name, namespace: proxy16.Namespace}: {
					Object:      proxy16,
					Status:      "invalid",
					Description: `Service [invalid:8080] is invalid or missing`,
					Vhost:       proxy16.Spec.VirtualHost.Fqdn,
				},
			},
		},
		"insert proxy": {
			objs: []interface{}{s2, proxy17},
			want: map[Meta]Status{
				{name: proxy17.Name, namespace: proxy17.Namespace}: {
					Object:      proxy17,
					Status:      StatusValid,
					Description: "valid HTTPProxy",
					Vhost:       "example.com",
				},
			},
		},
		"insert conflicting proxies due to fqdn reuse": {
			objs: []interface{}{proxy17, proxy18},
			want: map[Meta]Status{
				{name: proxy17.Name, namespace: proxy17.Namespace}: {
					Object:      proxy17,
					Status:      StatusInvalid,
					Description: `fqdn "example.com" is used in multiple HTTPProxies: roots/example-com, roots/other-example`,
					Vhost:       "example.com",
				},
				{name: proxy18.Name, namespace: proxy18.Namespace}: {
					Object:      proxy18,
					Status:      StatusInvalid,
					Description: `fqdn "example.com" is used in multiple HTTPProxies: roots/example-com, roots/other-example`,
					Vhost:       "example.com",
				},
			},
		},
		"root proxy delegating to another root": {
			objs: []interface{}{proxy20, proxy21},
			want: map[Meta]Status{
				{name: proxy20.Name, namespace: proxy20.Namespace}: {
					Object:      proxy20,
					Status:      StatusInvalid,
					Description: `fqdn "blog.containersteve.com" is used in multiple HTTPProxies: marketing/blog, roots/root-blog`,
					Vhost:       "blog.containersteve.com",
				},
				{name: proxy21.Name, namespace: proxy21.Namespace}: {
					Object:      proxy21,
					Status:      StatusInvalid,
					Description: `fqdn "blog.containersteve.com" is used in multiple HTTPProxies: marketing/blog, roots/root-blog`,
					Vhost:       "blog.containersteve.com",
				},
			},
		},
		"root ingress delegating to another root w/ different hostname": {
			objs: []interface{}{proxy22, proxy23, s8},
			want: map[Meta]Status{
				{name: proxy22.Name, namespace: proxy22.Namespace}: {
					Object:      proxy22,
					Status:      StatusInvalid,
					Description: "root httpproxy cannot delegate to another root httpproxy",
					Vhost:       "blog.containersteve.com",
				},
				{name: proxy23.Name, namespace: proxy23.Namespace}: {
					Object:      proxy23,
					Status:      StatusValid,
					Description: `valid HTTPProxy`,
					Vhost:       "www.containersteve.com",
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			builder := Builder{
				Source: KubernetesCache{
					RootNamespaces: []string{"roots", "marketing"},
					FieldLogger:    testLogger(t),
				},
			}
			for _, o := range tc.objs {
				builder.Source.Insert(o)
			}

			dag := builder.Build()
			got := dag.Statuses()

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
