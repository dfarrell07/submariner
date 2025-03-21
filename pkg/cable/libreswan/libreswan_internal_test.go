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

package libreswan

import (
	"fmt"
	"os"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	fakecommand "github.com/submariner-io/admiral/pkg/command/fake"
	subv1 "github.com/submariner-io/submariner/pkg/apis/submariner.io/v1"
	"github.com/submariner-io/submariner/pkg/endpoint"
	"github.com/submariner-io/submariner/pkg/natdiscovery"
	"github.com/submariner-io/submariner/pkg/types"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/scheme"
	k8snet "k8s.io/utils/net"
)

var _ = Describe("Libreswan", func() {
	Describe("IPsec port configuration", testIPsecPortConfiguration)
	Describe("trafficStatusRE", testTrafficStatusRE)
	Describe("ConnectToEndpoint", testConnectToEndpoint)
	Describe("DisconnectFromEndpoint", testDisconnectFromEndpoint)
	Describe("GetConnections", testGetConnections)
	Describe("Preferred server config", testPreferredServerConfig)
})

func testTrafficStatusRE() {
	When("Parsing a normal connection", func() {
		It("should match", func() {
			matches := trafficStatusRE.FindStringSubmatch("006 #3: \"submariner-cable-cluster3-172-17-0-8-v4-0-0\", " +
				"type=ESP, add_time=1590508783, inBytes=0, outBytes=0, id='172.17.0.8'\n")
			Expect(matches).NotTo(BeNil())
		})
	})

	When("Parsing a server-side connection", func() {
		It("should match", func() {
			matches := trafficStatusRE.FindStringSubmatch("006 #2: \"submariner-cable-cluster3-172-17-0-8-v4-0-0\"[1] 3.139.75.179," +
				" type=ESP, add_time=1617195756, inBytes=0, outBytes=0, id='@10.0.63.203-0-0'\n")
			Expect(matches).NotTo(BeNil())
		})
	})
	When("Parsing a normal v6 connection", func() {
		It("should match", func() {
			matches := trafficStatusRE.FindStringSubmatch("006 #3: \"submariner-cable-cluster3-fd12:3456:789a:1::1-v6-0-0\", " +
				" type=ESP, add_time=1590508783, inBytes=0, outBytes=0, id='@10.0.63.203-0-0'/n")
			Expect(matches).NotTo(BeNil())
		})
	})

	When("Parsing a server-side v6 connection", func() {
		It("should match", func() {
			matches := trafficStatusRE.FindStringSubmatch("006 #2: \"submariner-cable-cluster3-fd12:3456:789a:1::1-v6-0-0\"[1] 3.139.75.179," +
				" type=ESP, add_time=1617195756, inBytes=0, outBytes=0, id='@10.0.63.203-0-0'\n")
			Expect(matches).NotTo(BeNil())
		})
	})
}

func testIPsecPortConfiguration() {
	t := newTestDriver()

	When("NewLibreswan is called with no port environment variables set", func() {
		It("should set the port fields from the defaults in the specification definition", func() {
			Expect(t.driver.ipSecNATTPort).To(Equal("4500"))
		})
	})

	When("NewLibreswan is called with port environment variables set", func() {
		const (
			NATTPort       = "4555"
			NATTPortEnvVar = "CE_IPSEC_NATTPORT"
		)

		BeforeEach(func() {
			os.Setenv(NATTPortEnvVar, NATTPort)
		})

		AfterEach(func() {
			os.Unsetenv(NATTPortEnvVar)
		})

		It("should set the port fields from the environment variables", func() {
			Expect(t.driver.ipSecNATTPort).To(Equal(NATTPort))
		})
	})
}

func testConnectToEndpoint() {
	t := newTestDriver()

	var (
		natInfo          *natdiscovery.NATEndpointInfo
		natInfoIPv6      *natdiscovery.NATEndpointInfo
		natInfoDualStack *natdiscovery.NATEndpointInfo
	)

	BeforeEach(func() {
		natInfo = &natdiscovery.NATEndpointInfo{
			Endpoint: subv1.Endpoint{
				Spec: subv1.EndpointSpec{
					ClusterID:  "east",
					CableName:  "submariner-cable-east-192-68-2-1",
					PrivateIPs: []string{"192.68.2.1"},
					Subnets:    []string{"20.0.0.0/16"},
				},
			},
			UseIP:     "172.93.2.1",
			UseNAT:    true,
			UseFamily: k8snet.IPv4,
		}
		natInfoIPv6 = &natdiscovery.NATEndpointInfo{
			Endpoint: subv1.Endpoint{
				Spec: subv1.EndpointSpec{
					ClusterID:  "east",
					CableName:  "submariner-cable-east-192-68-2-1",
					PrivateIPs: []string{"2002::1234:abcd:ffff:c0a8:101"},
					Subnets:    []string{"2001::1234:abcd:ffff:c0a8:101/64"},
				},
			},
			UseIP:     "2003:db8:3333:4444:5555:6666:7777:8888",
			UseNAT:    true,
			UseFamily: k8snet.IPv6,
		}

		natInfoDualStack = &natdiscovery.NATEndpointInfo{
			Endpoint: subv1.Endpoint{
				Spec: subv1.EndpointSpec{
					ClusterID:  "east",
					CableName:  "submariner-cable-east-192-68-2-1",
					PrivateIPs: []string{"192.68.2.1", "2002::1234:abcd:ffff:c0a8:101"},
					Subnets:    []string{"20.0.0.0/16", "2001::1234:abcd:ffff:c0a8:101/64"},
				},
			},
			UseIP:     "172.93.2.1",
			UseNAT:    true,
			UseFamily: k8snet.IPv4,
		}
	})

	testBiDirectionalMode := func(natInfoParam *natdiscovery.NATEndpointInfo) {
		family := k8snet.IPFamilyOfString(natInfoParam.UseIP)
		ip, err := t.driver.ConnectToEndpoint(natInfoParam)

		Expect(err).To(Succeed())
		Expect(ip).To(Equal(natInfoParam.UseIP))

		t.assertActiveConnection(natInfoParam)
		t.cmdExecutor.AwaitCommand(nil, "whack", t.endpointSpec.GetPrivateIP(family), natInfoParam.UseIP,
			t.endpointSpec.ParseSubnets(family)[0].String(), natInfoParam.Endpoint.Spec.ParseSubnets(family)[0].String())
		t.cmdExecutor.AwaitCommand(nil, "whack", "--initiate")
	}

	testCreateConnection := func(natInfoParam *natdiscovery.NATEndpointInfo) {
		family := k8snet.IPFamilyOfString(natInfoParam.UseIP)
		ip, err := t.driver.ConnectToEndpoint(natInfoParam)

		Expect(err).To(Succeed())
		Expect(ip).To(Equal(natInfoParam.UseIP))

		t.assertActiveConnection(natInfoParam)
		t.cmdExecutor.AwaitCommand(nil, "whack", t.endpointSpec.GetPrivateIP(family), natInfoParam.UseIP,
			t.endpointSpec.ParseSubnets(family)[0].String(), natInfoParam.Endpoint.Spec.ParseSubnets(family)[0].String())
		t.cmdExecutor.AwaitCommand(nil, "whack", "--initiate")
	}

	testServerMode := func(natInfoParam *natdiscovery.NATEndpointInfo) {
		family := k8snet.IPFamilyOfString(natInfoParam.UseIP)
		ip, err := t.driver.ConnectToEndpoint(natInfoParam)

		Expect(err).To(Succeed())
		Expect(ip).To(Equal(natInfoParam.UseIP))

		t.assertActiveConnection(natInfoParam)
		t.cmdExecutor.AwaitCommand(nil, "whack", t.endpointSpec.GetPrivateIP(family), t.endpointSpec.ParseSubnets(family)[0].String(),
			natInfoParam.Endpoint.Spec.ParseSubnets(family)[0].String())
		t.cmdExecutor.EnsureNoCommand("whack", "--initiate")
	}

	When("only the local side prefers to be a server", func() {
		BeforeEach(func() {
			t.endpointSpec.BackendConfig = map[string]string{subv1.PreferredServerConfig: "true"}
			natInfo.Endpoint.Spec.BackendConfig = map[string]string{subv1.PreferredServerConfig: "false"}
			natInfoIPv6.Endpoint.Spec.BackendConfig = map[string]string{subv1.PreferredServerConfig: "false"}
			natInfoDualStack.Endpoint.Spec.BackendConfig = map[string]string{subv1.PreferredServerConfig: "false"}
		})

		It("should create a IPv4 server Connection", func() {
			testServerMode(natInfo)
		})
		It("should create a IPv6 server Connection", func() {
			testServerMode(natInfoIPv6)
		})
		It("should create a IPv4 server Connection on dualstack endpoint", func() {
			testServerMode(natInfoDualStack)
		})
	})

	When("only the remote side prefers to be a server", func() {
		BeforeEach(func() {
			t.endpointSpec.BackendConfig = map[string]string{subv1.PreferredServerConfig: "false"}
			natInfo.Endpoint.Spec.BackendConfig = map[string]string{subv1.PreferredServerConfig: "true"}
			natInfoIPv6.Endpoint.Spec.BackendConfig = map[string]string{subv1.PreferredServerConfig: "true"}
			natInfoDualStack.Endpoint.Spec.BackendConfig = map[string]string{subv1.PreferredServerConfig: "true"}
		})

		It("should create a IPv4 client Connection", func() {
			testCreateConnection(natInfo)
		})
		It("should create a IPv6 client Connection", func() {
			testCreateConnection(natInfoIPv6)
		})
		It("should create a IPv4 client Connection on dualstack endpoint", func() {
			testCreateConnection(natInfoDualStack)
		})
	})

	When("neither side prefers to be a server", func() {
		BeforeEach(func() {
			t.endpointSpec.BackendConfig = map[string]string{subv1.PreferredServerConfig: "false"}
			natInfo.Endpoint.Spec.BackendConfig = map[string]string{subv1.PreferredServerConfig: "false"}
			natInfoIPv6.Endpoint.Spec.BackendConfig = map[string]string{subv1.PreferredServerConfig: "false"}
			natInfoDualStack.Endpoint.Spec.BackendConfig = map[string]string{subv1.PreferredServerConfig: "false"}
		})

		It("should create a bi-directional IPv4 Connection", func() {
			testBiDirectionalMode(natInfo)
		})
		It("should create a bi-directional IPv6 Connection", func() {
			testBiDirectionalMode(natInfoIPv6)
		})
		It("should create a bi-directional IPv4 Connection on dualstack endpoint", func() {
			testBiDirectionalMode(natInfoDualStack)
		})
	})

	When("no preferred server is configured", func() {
		It("should default to a bi-directional IPv4 Connection", func() {
			testBiDirectionalMode(natInfo)
		})
		It("should default to a bi-directional IPv6 Connection", func() {
			testBiDirectionalMode(natInfoIPv6)
		})
		It("should default to a bi-directional IPv4 Connection on dualstack endpoint", func() {
			testBiDirectionalMode(natInfoDualStack)
		})
	})

	When("both sides prefer to be a server", func() {
		BeforeEach(func() {
			t.endpointSpec.BackendConfig = map[string]string{subv1.PreferredServerConfig: "true"}
			natInfo.Endpoint.Spec.BackendConfig = map[string]string{subv1.PreferredServerConfig: "true"}
			natInfoIPv6.Endpoint.Spec.BackendConfig = map[string]string{subv1.PreferredServerConfig: "true"}
			natInfoDualStack.Endpoint.Spec.BackendConfig = map[string]string{subv1.PreferredServerConfig: "true"}
		})

		It("should create a server IPv4 Connection due to comparison of the cable names", func() {
			testServerMode(natInfo)
		})
		It("should create a server IPv6 Connection due to comparison of the cable names", func() {
			testServerMode(natInfoIPv6)
		})
		It("should create a server IPv4 Connection on dualstack ep due to comparison of the cable names", func() {
			testServerMode(natInfoDualStack)
		})
	})
}

func testDisconnectFromEndpoint() {
	t := newTestDriver()

	It("should remove the Connection", func() {
		natInfo1 := &natdiscovery.NATEndpointInfo{
			Endpoint: subv1.Endpoint{
				Spec: subv1.EndpointSpec{
					ClusterID:  "remote1",
					CableName:  "submariner-cable-remote1-192-68-2-1",
					PrivateIPs: []string{"192.68.2.1"},
					Subnets:    []string{"20.0.0.0/16"},
				},
			},
			UseIP:     "172.93.2.1",
			UseFamily: k8snet.IPv4,
		}

		_, err := t.driver.ConnectToEndpoint(natInfo1)
		Expect(err).To(Succeed())

		natInfo2 := &natdiscovery.NATEndpointInfo{
			Endpoint: subv1.Endpoint{
				Spec: subv1.EndpointSpec{
					ClusterID:  "remote2",
					CableName:  "submariner-cable-remote2-192-68-3-1",
					PrivateIPs: []string{"192.68.3.1"},
					Subnets:    []string{"30.0.0.0/16"},
				},
			},
			UseIP:     "173.93.2.1",
			UseFamily: k8snet.IPv4,
		}

		_, err = t.driver.ConnectToEndpoint(natInfo2)
		Expect(err).To(Succeed())

		natInfoIPv6 := &natdiscovery.NATEndpointInfo{
			Endpoint: subv1.Endpoint{
				Spec: subv1.EndpointSpec{
					ClusterID:  "remote3",
					CableName:  "submariner-cable-east-192-68-4-1",
					PrivateIPs: []string{"2002::1234:abcd:ffff:c0a8:101"},
					Subnets:    []string{"2001::1234:abcd:ffff:c0a8:101/64"},
				},
			},
			UseIP:     "2003:db8:3333:4444:5555:6666:7777:8888",
			UseNAT:    true,
			UseFamily: k8snet.IPv6,
		}

		_, err = t.driver.ConnectToEndpoint(natInfoIPv6)
		Expect(err).To(Succeed())

		Expect(t.driver.DisconnectFromEndpoint(&types.SubmarinerEndpoint{Spec: natInfo1.Endpoint.Spec}, k8snet.IPv4)).To(Succeed())
		t.assertNoActiveConnection(natInfo1)
		t.cmdExecutor.AwaitCommand(nil, "whack", "--delete")
		t.cmdExecutor.Clear()

		Expect(t.driver.DisconnectFromEndpoint(&types.SubmarinerEndpoint{Spec: natInfoIPv6.Endpoint.Spec}, k8snet.IPv6)).To(Succeed())
		t.assertNoActiveConnection(natInfoIPv6)
		t.cmdExecutor.AwaitCommand(nil, "whack", "--delete")
		t.assertActiveConnection(natInfo2)
		t.cmdExecutor.Clear()

		Expect(t.driver.DisconnectFromEndpoint(&types.SubmarinerEndpoint{Spec: natInfo2.Endpoint.Spec}, k8snet.IPv4)).To(Succeed())
		t.assertNoActiveConnection(natInfo2)
		t.cmdExecutor.AwaitCommand(nil, "whack", "--delete")
		t.cmdExecutor.Clear()
	})
}

func testGetConnections() {
	t := newTestDriver()

	It("should return the correct Connections", func() {
		natInfo1 := &natdiscovery.NATEndpointInfo{
			Endpoint: subv1.Endpoint{
				Spec: subv1.EndpointSpec{
					ClusterID:  "remote1",
					CableName:  "submariner-cable-remote1-192-68-2-1",
					PrivateIPs: []string{"192.68.2.1"},
					Subnets:    []string{"20.0.0.0/16", "30.0.0.0/16"},
				},
			},
			UseIP:     "172.93.2.1",
			UseFamily: k8snet.IPv4,
		}

		_, err := t.driver.ConnectToEndpoint(natInfo1)
		Expect(err).To(Succeed())

		natInfo2 := &natdiscovery.NATEndpointInfo{
			Endpoint: subv1.Endpoint{
				Spec: subv1.EndpointSpec{
					ClusterID:  "remote2",
					CableName:  "submariner-cable-remote2-192-68-3-1",
					PrivateIPs: []string{"192.68.3.1"},
					Subnets:    []string{"11.0.0.0/16"},
				},
			},
			UseIP:     "173.93.3.1",
			UseFamily: k8snet.IPv4,
		}

		_, err = t.driver.ConnectToEndpoint(natInfo2)
		Expect(err).To(Succeed())

		t.cmdExecutor.SetupCommandStdOut(
			fmt.Sprintf(" \"%s-v4-0-0\", type=ESP, add_time=1590508783, inBytes=10, outBytes=20, id='192.68.2.1'",
				natInfo1.Endpoint.Spec.CableName),
			nil, "whack", "--trafficstatus")

		conn, err := t.driver.GetConnections()
		Expect(err).To(Succeed())

		Expect(conn).To(HaveExactElements(subv1.Connection{
			Status:   subv1.Connected,
			Endpoint: natInfo1.Endpoint.Spec,
			UsingIP:  natInfo1.UseIP,
			UsingNAT: natInfo1.UseNAT,
		}, subv1.Connection{
			Status:   subv1.Connecting,
			Endpoint: natInfo2.Endpoint.Spec,
			UsingIP:  natInfo2.UseIP,
			UsingNAT: natInfo2.UseNAT,
		}))
	})
}

func testPreferredServerConfig() {
	t := newTestDriver()

	AfterEach(func() {
		os.Unsetenv("CE_IPSEC_PREFERREDSERVER")
	})

	When("the preferred server setting is present in the local endpoint's BackendConfig", func() {
		BeforeEach(func() {
			os.Setenv("CE_IPSEC_PREFERREDSERVER", strconv.FormatBool(false))
			t.endpointSpec.BackendConfig = map[string]string{
				subv1.PreferredServerConfig: strconv.FormatBool(true),
				"other":                     "xyz",
			}
		})

		It("should correctly update the BackendConfig", func() {
			Expect(t.localEndpoint.Spec().BackendConfig).To(HaveKeyWithValue(subv1.PreferredServerConfig, strconv.FormatBool(true)))
			Expect(t.localEndpoint.Spec().BackendConfig).To(HaveKey(subv1.PreferredServerConfig + "-timestamp"))
			Expect(t.localEndpoint.Spec().BackendConfig).To(HaveKeyWithValue("other", "xyz"))
		})
	})

	When("the preferred server setting is present in the env variable", func() {
		BeforeEach(func() {
			os.Setenv("CE_IPSEC_PREFERREDSERVER", strconv.FormatBool(true))
		})

		It("should correctly update the BackendConfig", func() {
			Expect(t.localEndpoint.Spec().BackendConfig).To(HaveKeyWithValue(subv1.PreferredServerConfig, strconv.FormatBool(true)))
			Expect(t.localEndpoint.Spec().BackendConfig).To(HaveKey(subv1.PreferredServerConfig + "-timestamp"))
		})
	})

	When("the preferred server setting isn't present", func() {
		It("should correctly update the BackendConfig", func() {
			Expect(t.localEndpoint.Spec().BackendConfig).To(HaveKeyWithValue(subv1.PreferredServerConfig, strconv.FormatBool(false)))
			Expect(t.localEndpoint.Spec().BackendConfig).ToNot(HaveKey(subv1.PreferredServerConfig + "-timestamp"))
		})
	})
}

type testDriver struct {
	endpointSpec  subv1.EndpointSpec
	localEndpoint *endpoint.Local
	cmdExecutor   *fakecommand.Executor
	driver        *libreswan
}

func newTestDriver() *testDriver {
	t := &testDriver{}

	BeforeEach(func() {
		t.cmdExecutor = fakecommand.New()
		t.endpointSpec = subv1.EndpointSpec{
			ClusterID:  "local",
			CableName:  "submariner-cable-local-192-68-1-1",
			PrivateIPs: []string{"192.68.1.1", "2002::4321:abcd:ffff:c0a8:101"},
			Subnets:    []string{"10.0.0.0/16", "2005::1234:abcd:ffff:c0a8:101/64"},
		}
	})

	JustBeforeEach(func() {
		t.localEndpoint = endpoint.NewLocal(&t.endpointSpec, dynamicfake.NewSimpleDynamicClient(scheme.Scheme), "")
		ls, err := NewLibreswan(t.localEndpoint, &types.SubmarinerCluster{})
		Expect(err).NotTo(HaveOccurred())

		t.driver = ls.(*libreswan)
		t.driver.plutoStarted = true
	})

	return t
}

func (t *testDriver) assertActiveConnection(natInfo *natdiscovery.NATEndpointInfo) {
	conn, err := t.driver.GetActiveConnections()
	Expect(err).To(Succeed())
	Expect(conn).To(HaveExactElements(subv1.Connection{
		Status:   subv1.Connected,
		Endpoint: natInfo.Endpoint.Spec,
		UsingIP:  natInfo.UseIP,
		UsingNAT: natInfo.UseNAT,
	}))
}

func (t *testDriver) assertNoActiveConnection(natInfo *natdiscovery.NATEndpointInfo) {
	conn, err := t.driver.GetActiveConnections()
	Expect(err).To(Succeed())
	Expect(conn).ToNot(HaveExactElements(subv1.Connection{
		Status:   subv1.Connected,
		Endpoint: natInfo.Endpoint.Spec,
		UsingIP:  natInfo.UseIP,
		UsingNAT: natInfo.UseNAT,
	}))
}
