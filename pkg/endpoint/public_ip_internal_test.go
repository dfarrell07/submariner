/*
SPDX-License-Identifier: Apache-2.0

Copyright Contributors to the Submariner project.

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

package endpoint

import (
	"net"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/submariner-io/submariner/pkg/types"
	v1 "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	k8snet "k8s.io/utils/net"
)

var _ = Describe("firstIPv4InString", func() {
	When("the content has an IPv4", func() {
		const testIP = "1.2.3.4"
		const jsonIP = "{\"ip\": \"" + testIP + "\"}"

		It("should return the IP", func() {
			ip, err := firstIPv4InString(jsonIP)
			Expect(err).ToNot(HaveOccurred())
			Expect(ip).To(Equal(testIP))
		})
	})

	When("the content doesn't have an IPv4", func() {
		It("should result in error", func() {
			ip, err := firstIPv4InString("no IPs here")
			Expect(err).To(HaveOccurred())
			Expect(ip).To(Equal(""))
		})
	})
})

const (
	testServiceName = "my-loadbalancer"
	testNamespace   = "namespace"
)

var _ = Describe("public ip resolvers", func() {
	var submSpec *types.SubmarinerSpecification
	var backendConfig map[string]string

	const (
		publicIPConfig = "public-ip"
		testIPDNS      = "4.3.2.1"
		testIP         = "1.2.3.4"
		dnsHost        = testIPDNS + ".nip.io"
		ipv4PublicIP   = "ipv4:" + testIP
		lbPublicIP     = "lb:" + testServiceName
	)

	BeforeEach(func() {
		submSpec = &types.SubmarinerSpecification{
			Namespace: testNamespace,
		}

		backendConfig = map[string]string{}
	})

	When("a LoadBalancer with Ingress IP is specified", func() {
		It("should return the IP", func() {
			backendConfig[publicIPConfig] = lbPublicIP
			client := fake.NewClientset(serviceWithIngress(v1.LoadBalancerIngress{Hostname: "", IP: testIP}))
			ip, resolver, err := getPublicIP(k8snet.IPv4, submSpec, client, backendConfig, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(ip).To(Equal(testIP))
			Expect(resolver).To(Equal(lbPublicIP))
		})
	})

	When("a LoadBalancer with Ingress hostname is specified", func() {
		It("should return the IP", func() {
			backendConfig[publicIPConfig] = lbPublicIP
			client := fake.NewClientset(serviceWithIngress(v1.LoadBalancerIngress{
				Hostname: dnsHost,
				IP:       "",
			}))
			ip, resolver, err := getPublicIP(k8snet.IPv4, submSpec, client, backendConfig, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(ip).To(Equal(testIPDNS))
			Expect(resolver).To(Equal(lbPublicIP))
		})
	})

	When("a LoadBalancer with no ingress is specified", func() {
		It("should return error", func() {
			loadBalancerRetryConfig.Cap = 1 * time.Second
			loadBalancerRetryConfig.Duration = 50 * time.Millisecond
			loadBalancerRetryConfig.Steps = 1
			backendConfig[publicIPConfig] = lbPublicIP
			client := fake.NewClientset(serviceWithIngress())
			_, _, err := getPublicIP(k8snet.IPv4, submSpec, client, backendConfig, false)
			Expect(err).To(HaveOccurred())
		})
	})

	When("an IPv4 entry specified", func() {
		It("should return the IP", func() {
			backendConfig[publicIPConfig] = ipv4PublicIP
			client := fake.NewClientset()
			ip, resolver, err := getPublicIP(k8snet.IPv4, submSpec, client, backendConfig, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(ip).To(Equal(testIP))
			Expect(resolver).To(Equal(ipv4PublicIP))
		})
	})

	When("an IPv4 entry specified in air-gapped deployment", func() {
		It("should return the IP and not an empty value", func() {
			backendConfig[publicIPConfig] = ipv4PublicIP
			client := fake.NewClientset()
			ip, resolver, err := getPublicIP(k8snet.IPv4, submSpec, client, backendConfig, true)
			Expect(err).ToNot(HaveOccurred())
			Expect(ip).To(Equal(testIP))
			Expect(resolver).To(Equal(ipv4PublicIP))
		})
	})

	When("a DNS entry specified", func() {
		It("should return the IP", func() {
			backendConfig[publicIPConfig] = "dns:" + dnsHost
			client := fake.NewClientset()
			ip, resolver, err := getPublicIP(k8snet.IPv4, submSpec, client, backendConfig, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(ip).To(Equal(testIPDNS))
			Expect(resolver).To(Equal(backendConfig[publicIPConfig]))
		})
	})

	When("an API entry specified", func() {
		It("should return some IP", func() {
			backendConfig[publicIPConfig] = "api:4.icanhazip.com/"
			client := fake.NewClientset()
			ip, resolver, err := getPublicIP(k8snet.IPv4, submSpec, client, backendConfig, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(net.ParseIP(ip)).NotTo(BeNil())
			Expect(resolver).To(Equal(backendConfig[publicIPConfig]))
		})
	})

	When("multiple entries are specified", func() {
		It("should return the first working one", func() {
			backendConfig[publicIPConfig] = ipv4PublicIP + ",dns:" + dnsHost
			client := fake.NewClientset()
			ip, resolver, err := getPublicIP(k8snet.IPv4, submSpec, client, backendConfig, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(ip).To(Equal(testIP))
			Expect(resolver).To(Equal(ipv4PublicIP))
		})
	})

	When("multiple entries are specified and the first one doesn't succeed", func() {
		It("should return the first working one", func() {
			backendConfig[publicIPConfig] = "dns:thisdomaindoesntexistforsure.badbadbad," + ipv4PublicIP
			client := fake.NewClientset()
			ip, resolver, err := getPublicIP(k8snet.IPv4, submSpec, client, backendConfig, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(ip).To(Equal(testIP))
			Expect(resolver).To(Equal(ipv4PublicIP))
		})
	})
})

func serviceWithIngress(ingress ...v1.LoadBalancerIngress) *v1.Service {
	return &v1.Service{
		ObjectMeta: v1meta.ObjectMeta{
			Name:      testServiceName,
			Namespace: testNamespace,
		},
		Status: v1.ServiceStatus{
			LoadBalancer: v1.LoadBalancerStatus{
				Ingress: ingress,
			},
		},
	}
}
