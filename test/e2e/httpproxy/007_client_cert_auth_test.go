// Copyright Project Contour Authors
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

// +build e2e

package httpproxy

import (
	"context"
	"crypto/tls"

	certmanagerv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	certmanagermetav1 "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	contourv1 "github.com/projectcontour/contour/apis/projectcontour/v1"
	"github.com/projectcontour/contour/test/e2e"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func testClientCertAuth(fx *e2e.Framework) {
	t := fx.T()
	namespace := "007-client-cert-auth"

	fx.CreateNamespace(namespace)
	defer fx.DeleteNamespace(namespace)

	// Create a self-signed Issuer.
	selfSignedIssuer := &certmanagerv1.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "selfsigned",
		},
		Spec: certmanagerv1.IssuerSpec{
			IssuerConfig: certmanagerv1.IssuerConfig{
				SelfSigned: &certmanagerv1.SelfSignedIssuer{},
			},
		},
	}
	require.NoError(t, fx.Client.Create(context.TODO(), selfSignedIssuer))

	// Using the selfsigned issuer, create a CA signing certificate for the
	// test issuer.
	caSigningCert := &certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "ca-projectcontour-io",
		},
		Spec: certmanagerv1.CertificateSpec{
			IsCA: true,
			Usages: []certmanagerv1.KeyUsage{
				certmanagerv1.UsageSigning,
				certmanagerv1.UsageCertSign,
			},
			Subject: &certmanagerv1.X509Subject{
				OrganizationalUnits: []string{
					"io",
					"projectcontour",
					"testsuite",
				},
			},
			CommonName: "issuer",
			SecretName: "ca-projectcontour-io",
			IssuerRef: certmanagermetav1.ObjectReference{
				Name: "selfsigned",
			},
		},
	}
	require.NoError(t, fx.Client.Create(context.TODO(), caSigningCert))

	// Create a local CA issuer with the CA certificate that the selfsigned
	// issuer gave us.
	localCAIssuer := &certmanagerv1.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "ca-projectcontour-io",
		},
		Spec: certmanagerv1.IssuerSpec{
			IssuerConfig: certmanagerv1.IssuerConfig{
				CA: &certmanagerv1.CAIssuer{
					SecretName: "ca-projectcontour-io",
				},
			},
		},
	}
	require.NoError(t, fx.Client.Create(context.TODO(), localCAIssuer))

	// Using the selfsigned issuer, create a CA signing certificate for another
	// test issuer.
	caSigningCert2 := &certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "ca-notprojectcontour-io",
		},
		Spec: certmanagerv1.CertificateSpec{
			IsCA: true,
			Usages: []certmanagerv1.KeyUsage{
				certmanagerv1.UsageSigning,
				certmanagerv1.UsageCertSign,
			},
			Subject: &certmanagerv1.X509Subject{
				OrganizationalUnits: []string{
					"io",
					"notprojectcontour",
					"testsuite",
				},
			},
			CommonName: "issuer",
			SecretName: "ca-notprojectcontour-io",
			IssuerRef: certmanagermetav1.ObjectReference{
				Name: "selfsigned",
			},
		},
	}
	require.NoError(t, fx.Client.Create(context.TODO(), caSigningCert2))

	// Create a local CA issuer with the CA certificate that the selfsigned
	// issuer gave us.
	localCAIssuer2 := &certmanagerv1.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "ca-notprojectcontour-io",
		},
		Spec: certmanagerv1.IssuerSpec{
			IssuerConfig: certmanagerv1.IssuerConfig{
				CA: &certmanagerv1.CAIssuer{
					SecretName: "ca-notprojectcontour-io",
				},
			},
		},
	}
	require.NoError(t, fx.Client.Create(context.TODO(), localCAIssuer2))

	fx.Fixtures.Echo.Deploy(namespace, "echo-no-auth")

	// Get a server certificate for echo-no-auth.
	echoNoAuthCert := &certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "echo-no-auth-cert",
		},
		Spec: certmanagerv1.CertificateSpec{

			Usages: []certmanagerv1.KeyUsage{
				certmanagerv1.UsageServerAuth,
			},
			DNSNames:   []string{"echo-no-auth.projectcontour.io"},
			SecretName: "echo-no-auth",
			IssuerRef: certmanagermetav1.ObjectReference{
				Name: "ca-projectcontour-io",
			},
		},
	}
	require.NoError(t, fx.Client.Create(context.TODO(), echoNoAuthCert))

	fx.Fixtures.Echo.Deploy(namespace, "echo-with-auth")

	// Get a server certificate for echo-with-auth.
	echoWithAuthCert := &certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "echo-with-auth-cert",
		},
		Spec: certmanagerv1.CertificateSpec{
			Usages: []certmanagerv1.KeyUsage{
				certmanagerv1.UsageServerAuth,
			},
			DNSNames:   []string{"echo-with-auth.projectcontour.io"},
			SecretName: "echo-with-auth",
			IssuerRef: certmanagermetav1.ObjectReference{
				Name: "ca-projectcontour-io",
			},
		},
	}
	require.NoError(t, fx.Client.Create(context.TODO(), echoWithAuthCert))

	fx.Fixtures.Echo.Deploy(namespace, "echo-with-auth-skip-verify")

	// Get a server certificate for echo-with-auth-skip-verify.
	echoWithAuthSkipVerifyCert := &certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "echo-with-auth-skip-verify-cert",
		},
		Spec: certmanagerv1.CertificateSpec{

			Usages: []certmanagerv1.KeyUsage{
				certmanagerv1.UsageServerAuth,
			},
			DNSNames:   []string{"echo-with-auth-skip-verify.projectcontour.io"},
			SecretName: "echo-with-auth-skip-verify",
			IssuerRef: certmanagermetav1.ObjectReference{
				Name: "ca-projectcontour-io",
			},
		},
	}
	require.NoError(t, fx.Client.Create(context.TODO(), echoWithAuthSkipVerifyCert))

	// Get a client certificate.
	clientCert := &certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "echo-client-cert",
		},
		Spec: certmanagerv1.CertificateSpec{
			Usages: []certmanagerv1.KeyUsage{
				certmanagerv1.UsageClientAuth,
			},
			EmailAddresses: []string{
				"client@projectcontour.io",
			},
			CommonName: "client",
			SecretName: "echo-client",
			IssuerRef: certmanagermetav1.ObjectReference{
				Name: "ca-projectcontour-io",
			},
		},
	}
	// Wait for the Cert to be ready since we'll directly download
	// the secret contents for use as a client cert later on.
	fx.Certs.CreateCertAndWaitFor(clientCert, certIsReady)

	// Get another client certificate.
	clientCertInvalid := &certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "echo-client-cert-invalid",
		},
		Spec: certmanagerv1.CertificateSpec{
			Usages: []certmanagerv1.KeyUsage{
				certmanagerv1.UsageClientAuth,
			},
			EmailAddresses: []string{
				"badclient@projectcontour.io",
			},
			CommonName: "badclient",
			SecretName: "echo-client-invalid",
			IssuerRef: certmanagermetav1.ObjectReference{
				Name: "ca-notprojectcontour-io",
			},
		},
	}
	// Wait for the Cert to be ready since we'll directly download
	// the secret contents for use as a client cert later on.
	fx.Certs.CreateCertAndWaitFor(clientCertInvalid, certIsReady)

	// This proxy does not require client certificate auth.
	noAuthProxy := &contourv1.HTTPProxy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "echo-no-auth",
		},
		Spec: contourv1.HTTPProxySpec{
			VirtualHost: &contourv1.VirtualHost{
				Fqdn: "echo-no-auth.projectcontour.io",
				TLS: &contourv1.TLS{
					SecretName: "echo-no-auth",
				},
			},
			Routes: []contourv1.Route{
				{
					Services: []contourv1.Service{
						{
							Name: "echo-no-auth",
							Port: 80,
						},
					},
				},
			},
		},
	}
	fx.CreateHTTPProxyAndWaitFor(noAuthProxy, httpProxyValid)

	// This proxy requires client certificate auth.
	authProxy := &contourv1.HTTPProxy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "echo-with-auth",
		},
		Spec: contourv1.HTTPProxySpec{
			VirtualHost: &contourv1.VirtualHost{
				Fqdn: "echo-with-auth.projectcontour.io",
				TLS: &contourv1.TLS{
					SecretName: "echo-with-auth",
					ClientValidation: &contourv1.DownstreamValidation{
						CACertificate: "echo-with-auth",
					},
				},
			},
			Routes: []contourv1.Route{
				{
					Services: []contourv1.Service{
						{
							Name: "echo-with-auth",
							Port: 80,
						},
					},
				},
			},
		},
	}
	fx.CreateHTTPProxyAndWaitFor(authProxy, httpProxyValid)

	// This proxy requires a client certificate but does not verify it.
	authSkipVerifyProxy := &contourv1.HTTPProxy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "echo-with-auth-skip-verify",
		},
		Spec: contourv1.HTTPProxySpec{
			VirtualHost: &contourv1.VirtualHost{
				Fqdn: "echo-with-auth-skip-verify.projectcontour.io",
				TLS: &contourv1.TLS{
					SecretName: "echo-with-auth-skip-verify",
					ClientValidation: &contourv1.DownstreamValidation{
						SkipClientCertValidation: true,
					},
				},
			},
			Routes: []contourv1.Route{
				{
					Services: []contourv1.Service{
						{
							Name: "echo-with-auth-skip-verify",
							Port: 80,
						},
					},
				},
			},
		},
	}
	fx.CreateHTTPProxyAndWaitFor(authSkipVerifyProxy, httpProxyValid)

	// get the valid & invalid client certs
	validClientCert := fx.Certs.GetTLSCertificate(namespace, clientCert.Spec.SecretName)
	invalidClientCert := fx.Certs.GetTLSCertificate(namespace, clientCertInvalid.Spec.SecretName)

	cases := map[string]struct {
		host       string
		clientCert *tls.Certificate
		wantErr    string
	}{
		"echo-no-auth without a client cert should succeed": {
			host:       noAuthProxy.Spec.VirtualHost.Fqdn,
			clientCert: nil,
			wantErr:    "",
		},
		"echo-no-auth with echo-client-cert should succeed": {
			host:       noAuthProxy.Spec.VirtualHost.Fqdn,
			clientCert: &validClientCert,
			wantErr:    "",
		},
		"echo-no-auth with echo-client-cert-invalid should succeed": {
			host:       noAuthProxy.Spec.VirtualHost.Fqdn,
			clientCert: &invalidClientCert,
			wantErr:    "",
		},

		"echo-with-auth without a client cert should error": {
			host:       authProxy.Spec.VirtualHost.Fqdn,
			clientCert: nil,
			wantErr:    "tls: certificate required",
		},
		"echo-with-auth with echo-client-cert should succeed": {
			host:       authProxy.Spec.VirtualHost.Fqdn,
			clientCert: &validClientCert,
			wantErr:    "",
		},
		"echo-with-auth with echo-client-cert-invalid should error": {
			host:       authProxy.Spec.VirtualHost.Fqdn,
			clientCert: &invalidClientCert,
			wantErr:    "tls: certificate required",
		},

		"echo-with-auth-skip-verify without a client cert should succeed": {
			host:       authSkipVerifyProxy.Spec.VirtualHost.Fqdn,
			clientCert: nil,
			wantErr:    "",
		},
		"echo-with-auth-skip-verify with echo-client-cert should succeed": {
			host:       authSkipVerifyProxy.Spec.VirtualHost.Fqdn,
			clientCert: &validClientCert,
			wantErr:    "",
		},
		"echo-with-auth-skip-verify with echo-client-cert-invalid should succeed": {
			host:       authSkipVerifyProxy.Spec.VirtualHost.Fqdn,
			clientCert: &invalidClientCert,
			wantErr:    "",
		},
	}

	for name, tc := range cases {
		t.Logf("Running test case %s", name)
		opts := &e2e.HTTPSRequestOpts{
			Host: tc.host,
		}
		if tc.clientCert != nil {
			opts.TLSConfigOpts = append(opts.TLSConfigOpts, optUseCert(*tc.clientCert))
		}

		switch {
		case len(tc.wantErr) == 0:
			opts.Condition = e2e.HasStatusCode(200)
			res, ok := fx.HTTP.SecureRequestUntil(opts)
			assert.Truef(t, ok, "expected 200 response code, got %d", res.StatusCode)
		default:
			_, err := fx.HTTP.SecureRequest(opts)
			assert.Contains(t, err.Error(), tc.wantErr)
		}
	}
}

func optUseCert(cert tls.Certificate) func(*tls.Config) {
	return func(c *tls.Config) {
		c.Certificates = append(c.Certificates, cert)
	}
}

func certIsReady(cert *certmanagerv1.Certificate) bool {
	for _, cond := range cert.Status.Conditions {
		if cond.Type == certmanagerv1.CertificateConditionReady && cond.Status == certmanagermetav1.ConditionTrue {
			return true
		}
	}
	return false
}
