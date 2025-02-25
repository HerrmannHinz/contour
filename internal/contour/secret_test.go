package contour

import (
	"reflect"
	"testing"

	envoy_api_v2_auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	envoy_api_v2_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/golang/protobuf/proto"
	"github.com/google/go-cmp/cmp"
	ingressroutev1 "github.com/heptio/contour/apis/contour/v1beta1"
	projcontour "github.com/heptio/contour/apis/projectcontour/v1alpha1"
	"github.com/heptio/contour/internal/dag"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestSecretCacheContents(t *testing.T) {
	tests := map[string]struct {
		contents map[string]*envoy_api_v2_auth.Secret
		want     []proto.Message
	}{
		"empty": {
			contents: nil,
			want:     nil,
		},
		"simple": {
			contents: secretmap(
				secret("default/secret/cd1b506996", "cert", "key"),
			),
			want: []proto.Message{
				secret("default/secret/cd1b506996", "cert", "key"),
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var sc SecretCache
			sc.Update(tc.contents)
			got := sc.Contents()
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestSecretCacheQuery(t *testing.T) {
	tests := map[string]struct {
		contents map[string]*envoy_api_v2_auth.Secret
		query    []string
		want     []proto.Message
	}{
		"exact match": {
			contents: secretmap(
				secret("default/secret/cd1b506996", "cert", "key"),
			),
			query: []string{"default/secret/cd1b506996"},
			want: []proto.Message{
				secret("default/secret/cd1b506996", "cert", "key"),
			},
		},
		"partial match": {
			contents: secretmap(
				secret("default/secret-a/ff2a9f58ca", "cert-a", "key-a"),
				secret("default/secret-b/0a068be4ba", "cert-b", "key-b"),
			),
			query: []string{"default/secret/cd1b506996", "default/secret-b/0a068be4ba"},
			want: []proto.Message{
				secret("default/secret-b/0a068be4ba", "cert-b", "key-b"),
			},
		},
		"no match": {
			contents: secretmap(
				secret("default/secret/cd1b506996", "cert", "key"),
			),
			query: []string{"default/secret-b/0a068be4ba"},
			want:  nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var sc SecretCache
			sc.Update(tc.contents)
			got := sc.Query(tc.query)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestSecretVisit(t *testing.T) {
	tests := map[string]struct {
		objs []interface{}
		want map[string]*envoy_api_v2_auth.Secret
	}{
		"nothing": {
			objs: nil,
			want: map[string]*envoy_api_v2_auth.Secret{},
		},
		"unassociated secrets": {
			objs: []interface{}{
				tlssecret("default", "secret-a", secretdata("cert", "key")),
				tlssecret("default", "secret-b", secretdata("cert", "key")),
			},
			want: map[string]*envoy_api_v2_auth.Secret{},
		},
		"simple ingress with secret": {
			objs: []interface{}{
				&v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kuard",
						Namespace: "default",
					},
					Spec: v1.ServiceSpec{
						Ports: []v1.ServicePort{{
							Name:       "http",
							Protocol:   "TCP",
							Port:       8080,
							TargetPort: intstr.FromInt(8080),
						}},
					},
				},
				&v1beta1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "simple",
						Namespace: "default",
					},
					Spec: v1beta1.IngressSpec{
						TLS: []v1beta1.IngressTLS{{
							Hosts:      []string{"whatever.example.com"},
							SecretName: "secret",
						}},
						Rules: []v1beta1.IngressRule{{
							Host: "whatever.example.com",
							IngressRuleValue: v1beta1.IngressRuleValue{
								HTTP: &v1beta1.HTTPIngressRuleValue{
									Paths: []v1beta1.HTTPIngressPath{{
										Backend: *backend("kuard", 8080),
									}},
								},
							},
						}},
					},
				},
				tlssecret("default", "secret", secretdata("cert", "key")),
			},
			want: secretmap(
				secret("default/secret/cd1b506996", "cert", "key"),
			),
		},
		"multiple ingresses with shared secret": {
			objs: []interface{}{
				&v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kuard",
						Namespace: "default",
					},
					Spec: v1.ServiceSpec{
						Ports: []v1.ServicePort{{
							Name:       "http",
							Protocol:   "TCP",
							Port:       8080,
							TargetPort: intstr.FromInt(8080),
						}},
					},
				},
				&v1beta1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "simple-a",
						Namespace: "default",
					},
					Spec: v1beta1.IngressSpec{
						TLS: []v1beta1.IngressTLS{{
							Hosts:      []string{"whatever.example.com"},
							SecretName: "secret",
						}},
						Rules: []v1beta1.IngressRule{{
							Host: "whatever.example.com",
							IngressRuleValue: v1beta1.IngressRuleValue{
								HTTP: &v1beta1.HTTPIngressRuleValue{
									Paths: []v1beta1.HTTPIngressPath{{
										Backend: *backend("kuard", 8080),
									}},
								},
							},
						}},
					},
				},
				&v1beta1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "simple-b",
						Namespace: "default",
					},
					Spec: v1beta1.IngressSpec{
						TLS: []v1beta1.IngressTLS{{
							Hosts:      []string{"omg.example.com"},
							SecretName: "secret",
						}},
						Rules: []v1beta1.IngressRule{{
							Host: "omg.example.com",
							IngressRuleValue: v1beta1.IngressRuleValue{
								HTTP: &v1beta1.HTTPIngressRuleValue{
									Paths: []v1beta1.HTTPIngressPath{{
										Backend: *backend("kuard", 8080),
									}},
								},
							},
						}},
					},
				},
				tlssecret("default", "secret", secretdata("cert", "key")),
			},
			want: secretmap(
				secret("default/secret/cd1b506996", "cert", "key"),
			),
		},
		"multiple ingresses with different secrets": {
			objs: []interface{}{
				&v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kuard",
						Namespace: "default",
					},
					Spec: v1.ServiceSpec{
						Ports: []v1.ServicePort{{
							Name:       "http",
							Protocol:   "TCP",
							Port:       80,
							TargetPort: intstr.FromInt(8080),
						}},
					},
				},
				&v1beta1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "simple-a",
						Namespace: "default",
					},
					Spec: v1beta1.IngressSpec{
						TLS: []v1beta1.IngressTLS{{
							Hosts:      []string{"whatever.example.com"},
							SecretName: "secret-a",
						}},
						Rules: []v1beta1.IngressRule{{
							Host: "whatever.example.com",
							IngressRuleValue: v1beta1.IngressRuleValue{
								HTTP: &v1beta1.HTTPIngressRuleValue{
									Paths: []v1beta1.HTTPIngressPath{{
										Backend: *backend("kuard", 80),
									}},
								},
							},
						}},
					},
				},
				&v1beta1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "simple-b",
						Namespace: "default",
					},
					Spec: v1beta1.IngressSpec{
						TLS: []v1beta1.IngressTLS{{
							Hosts:      []string{"omg.example.com"},
							SecretName: "secret-b",
						}},
						Rules: []v1beta1.IngressRule{{
							Host: "omg.example.com",
							IngressRuleValue: v1beta1.IngressRuleValue{
								HTTP: &v1beta1.HTTPIngressRuleValue{
									Paths: []v1beta1.HTTPIngressPath{{
										Backend: *backend("kuard", 80),
									}},
								},
							},
						}},
					},
				},
				tlssecret("default", "secret-a", secretdata("cert-a", "key-a")),
				tlssecret("default", "secret-b", secretdata("cert-b", "key-b")),
			},
			want: secretmap(
				secret("default/secret-a/ff2a9f58ca", "cert-a", "key-a"),
				secret("default/secret-b/0a068be4ba", "cert-b", "key-b"),
			),
		},
		"simple ingressroute with secret": {
			objs: []interface{}{
				&v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "backend",
						Namespace: "default",
					},
					Spec: v1.ServiceSpec{
						Ports: []v1.ServicePort{{
							Name:       "http",
							Protocol:   "TCP",
							Port:       80,
							TargetPort: intstr.FromInt(8080),
						}},
					},
				},
				&ingressroutev1.IngressRoute{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "simple",
						Namespace: "default",
					},
					Spec: ingressroutev1.IngressRouteSpec{
						VirtualHost: &projcontour.VirtualHost{
							Fqdn: "www.example.com",
							TLS: &projcontour.TLS{
								SecretName: "secret",
							},
						},
						Routes: []ingressroutev1.Route{{
							Services: []ingressroutev1.Service{{
								Name: "backend",
								Port: 80,
							}},
						}},
					},
				},
				tlssecret("default", "secret", secretdata("cert", "key")),
			},
			want: secretmap(
				secret("default/secret/cd1b506996", "cert", "key"),
			),
		},
		"multiple ingressroutes with shared secret": {
			objs: []interface{}{
				&v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "backend",
						Namespace: "default",
					},
					Spec: v1.ServiceSpec{
						Ports: []v1.ServicePort{{
							Name:       "http",
							Protocol:   "TCP",
							Port:       80,
							TargetPort: intstr.FromInt(8080),
						}},
					},
				},
				&ingressroutev1.IngressRoute{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "simple-a",
						Namespace: "default",
					},
					Spec: ingressroutev1.IngressRouteSpec{
						VirtualHost: &projcontour.VirtualHost{
							Fqdn: "www.example.com",
							TLS: &projcontour.TLS{
								SecretName: "secret",
							},
						},
						Routes: []ingressroutev1.Route{{
							Services: []ingressroutev1.Service{{
								Name: "backend",
								Port: 80,
							}},
						}},
					},
				},
				&ingressroutev1.IngressRoute{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "simple-b",
						Namespace: "default",
					},
					Spec: ingressroutev1.IngressRouteSpec{
						VirtualHost: &projcontour.VirtualHost{
							Fqdn: "www.other.com",
							TLS: &projcontour.TLS{
								SecretName: "secret",
							},
						},
						Routes: []ingressroutev1.Route{{
							Services: []ingressroutev1.Service{{
								Name: "backend",
								Port: 80,
							}},
						}},
					},
				},
				tlssecret("default", "secret", secretdata("cert", "key")),
			},
			want: secretmap(
				secret("default/secret/cd1b506996", "cert", "key"),
			),
		},
		"multiple ingressroutes with different secret": {
			objs: []interface{}{
				&v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "backend",
						Namespace: "default",
					},
					Spec: v1.ServiceSpec{
						Ports: []v1.ServicePort{{
							Name:       "http",
							Protocol:   "TCP",
							Port:       80,
							TargetPort: intstr.FromInt(8080),
						}},
					},
				},
				&ingressroutev1.IngressRoute{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "simple-a",
						Namespace: "default",
					},
					Spec: ingressroutev1.IngressRouteSpec{
						VirtualHost: &projcontour.VirtualHost{
							Fqdn: "www.example.com",
							TLS: &projcontour.TLS{
								SecretName: "secret-a",
							},
						},
						Routes: []ingressroutev1.Route{{
							Services: []ingressroutev1.Service{{
								Name: "backend",
								Port: 80,
							}},
						}},
					},
				},
				&ingressroutev1.IngressRoute{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "simple-b",
						Namespace: "default",
					},
					Spec: ingressroutev1.IngressRouteSpec{
						VirtualHost: &projcontour.VirtualHost{
							Fqdn: "www.other.com",
							TLS: &projcontour.TLS{
								SecretName: "secret-b",
							},
						},
						Routes: []ingressroutev1.Route{{
							Services: []ingressroutev1.Service{{
								Name: "backend",
								Port: 80,
							}},
						}},
					},
				},
				tlssecret("default", "secret-a", secretdata("cert-a", "key-a")),
				tlssecret("default", "secret-b", secretdata("cert-b", "key-b")),
			},
			want: secretmap(
				secret("default/secret-a/ff2a9f58ca", "cert-a", "key-a"),
				secret("default/secret-b/0a068be4ba", "cert-b", "key-b"),
			),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			root := buildDAG(tc.objs...)
			got := visitSecrets(root)
			if !reflect.DeepEqual(tc.want, got) {
				t.Fatalf("expected:\n%+v\ngot:\n%+v", tc.want, got)
			}
		})
	}
}

// buildDAG produces a dag.DAG from the supplied objects.
func buildDAG(objs ...interface{}) *dag.DAG {
	var builder dag.Builder
	for _, o := range objs {
		builder.Source.Insert(o)
	}
	return builder.Build()
}

func secretmap(secrets ...*envoy_api_v2_auth.Secret) map[string]*envoy_api_v2_auth.Secret {
	m := make(map[string]*envoy_api_v2_auth.Secret)
	for _, s := range secrets {
		m[s.Name] = s
	}
	return m
}

func secret(name, cert, key string) *envoy_api_v2_auth.Secret {
	return &envoy_api_v2_auth.Secret{
		Name: name,
		Type: &envoy_api_v2_auth.Secret_TlsCertificate{
			TlsCertificate: &envoy_api_v2_auth.TlsCertificate{
				PrivateKey: &envoy_api_v2_core.DataSource{
					Specifier: &envoy_api_v2_core.DataSource_InlineBytes{
						InlineBytes: []byte(key),
					},
				},
				CertificateChain: &envoy_api_v2_core.DataSource{
					Specifier: &envoy_api_v2_core.DataSource_InlineBytes{
						InlineBytes: []byte(cert),
					},
				},
			},
		},
	}
}

// tlssecert creates a new v1.Secret object of type kubernetes.io/tls.
func tlssecret(namespace, name string, data map[string][]byte) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: v1.SecretTypeTLS,
		Data: data,
	}
}

func backend(name string, port int) *v1beta1.IngressBackend {
	return &v1beta1.IngressBackend{
		ServiceName: name,
		ServicePort: intstr.FromInt(port),
	}
}
