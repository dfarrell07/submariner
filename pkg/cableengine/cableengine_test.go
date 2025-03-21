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

package cableengine_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/submariner-io/admiral/pkg/gomega"
	"github.com/submariner-io/admiral/pkg/log/kzerolog"
	subv1 "github.com/submariner-io/submariner/pkg/apis/submariner.io/v1"
	"github.com/submariner-io/submariner/pkg/cable"
	"github.com/submariner-io/submariner/pkg/cable/fake"
	"github.com/submariner-io/submariner/pkg/cableengine"
	submendpoint "github.com/submariner-io/submariner/pkg/endpoint"
	"github.com/submariner-io/submariner/pkg/natdiscovery"
	"github.com/submariner-io/submariner/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/scheme"
	k8snet "k8s.io/utils/net"
)

func init() {
	kzerolog.AddFlags(nil)
}

var fakeDriver *fake.Driver

var _ = BeforeSuite(func() {
	kzerolog.InitK8sLogging()
	cable.AddDriver(fake.DriverName, func(_ *submendpoint.Local, _ *types.SubmarinerCluster) (cable.Driver, error) {
		return fakeDriver, nil
	})
})

var _ = Describe("Cable Engine", func() {
	const localClusterID = "local"
	const remoteClusterID = "remote"

	var (
		engine         cableengine.Engine
		natDiscovery   *fakeNATDiscovery
		localEndpoint  *subv1.Endpoint
		remoteEndpoint *subv1.Endpoint
		skipStart      bool
	)

	BeforeEach(func() {
		skipStart = false

		localEndpoint = &subv1.Endpoint{
			ObjectMeta: metav1.ObjectMeta{
				CreationTimestamp: metav1.Now(),
			},
			Spec: subv1.EndpointSpec{
				ClusterID:  localClusterID,
				CableName:  fmt.Sprintf("submariner-cable-%s-1.1.1.1", localClusterID),
				PrivateIPs: []string{"1.1.1.1", "FDC8:BF8B:E62C:ABCD:1111:2222:3333:4444"},
				PublicIPs:  []string{"2.2.2.2", "FDC8:BF8B:E62C:ABCD:1111:2222:3333:5555"},
				Backend:    fake.DriverName,
			},
		}

		remoteEndpoint = &subv1.Endpoint{
			ObjectMeta: metav1.ObjectMeta{
				CreationTimestamp: metav1.Now(),
			},
			Spec: subv1.EndpointSpec{
				ClusterID:     remoteClusterID,
				CableName:     fmt.Sprintf("submariner-cable-%s-1.1.1.1", remoteClusterID),
				PrivateIPs:    []string{"1.1.1.1", "FDC7:BF8B:E62C:ABCD:1111:2222:3333:4444"},
				PublicIPs:     []string{"2.2.2.2", "FDC7:BF8B:E62C:ABCD:1111:2222:3333:5555"},
				BackendConfig: map[string]string{"port": "1234"},
			},
		}

		fakeDriver = fake.New()
		engine = cableengine.NewEngine(&types.SubmarinerCluster{
			ID: localClusterID,
			Spec: subv1.ClusterSpec{
				ClusterID: localClusterID,
			},
		}, submendpoint.NewLocal(&localEndpoint.Spec, dynamicfake.NewSimpleDynamicClient(scheme.Scheme), ""))

		natDiscovery = &fakeNATDiscovery{removeEndpoint: make(chan string, 20), readyChannel: make(chan *natdiscovery.NATEndpointInfo, 100)}
		engine.SetupNATDiscovery(natDiscovery)
	})

	JustBeforeEach(func() {
		if skipStart {
			return
		}

		err := engine.StartEngine()
		if fakeDriver.ErrOnInit != nil {
			Expect(err).To(ContainErrorSubstring(fakeDriver.ErrOnInit))
		} else {
			Expect(err).To(Succeed())
			fakeDriver.AwaitInit()
		}
	})

	It("should return the local endpoint when queried", func() {
		Expect(engine.GetLocalEndpoint()).To(Equal(&types.SubmarinerEndpoint{Spec: localEndpoint.Spec}))
	})

	When("install cable for a remote endpoint", func() {
		Context("and no endpoint was previously installed for the cluster", func() {
			It("should connect to the endpoint", func() {
				Expect(engine.InstallCable(remoteEndpoint, k8snet.IPv4)).To(Succeed())
				fakeDriver.AwaitConnectToEndpoint(natEndpointInfoFor(remoteEndpoint, k8snet.IPv4))
			})
			It("should connect to endpoint with IPv6", func() {
				Expect(engine.InstallCable(remoteEndpoint, k8snet.IPv6)).To(Succeed())
				fakeDriver.AwaitConnectToEndpoint(natEndpointInfoFor(remoteEndpoint, k8snet.IPv6))
			})
		})

		Context("and an endpoint was previously installed for the cluster", func() {
			var prevEndpoint *subv1.Endpoint
			var newEndpoint *subv1.Endpoint

			var ipFamily k8snet.IPFamily

			BeforeEach(func() {
				c := *remoteEndpoint
				newEndpoint = &c
				prevEndpoint = remoteEndpoint
				ipFamily = k8snet.IPv4
			})

			JustBeforeEach(func() {
				Expect(engine.InstallCable(prevEndpoint, ipFamily)).To(Succeed())
				fakeDriver.AwaitConnectToEndpoint(natEndpointInfoFor(prevEndpoint, ipFamily))

				Expect(engine.InstallCable(newEndpoint, ipFamily)).To(Succeed())
			})

			testTimestamps := func(family k8snet.IPFamily) {
				Context("and older creation timestamp", func() {
					BeforeEach(func() {
						time.Sleep(100 * time.Millisecond)
						newEndpoint.CreationTimestamp = metav1.Now()
					})

					It("should disconnect from the previous endpoint and connect to the new one", func() {
						fakeDriver.AwaitDisconnectFromEndpoint(&prevEndpoint.Spec, family)
						fakeDriver.AwaitConnectToEndpoint(natEndpointInfoFor(newEndpoint, family))
					})
				})

				Context("and newer creation timestamp", func() {
					BeforeEach(func() {
						newEndpoint.CreationTimestamp = metav1.Now()
						time.Sleep(100 * time.Millisecond)
						prevEndpoint.CreationTimestamp = metav1.Now()
					})

					It("should not disconnect from the previous endpoint nor connect to the new one", func() {
						fakeDriver.AwaitNoDisconnectFromEndpoint()
						fakeDriver.AwaitNoConnectToEndpoint()
					})
				})
			}

			// Run the tests for both IPv4 and IPv6
			Context("with a different cable name", func() {
				BeforeEach(func() {
					newEndpoint.Spec.CableName = "new cable"
				})

				for _, family := range []k8snet.IPFamily{k8snet.IPv4, k8snet.IPv6} {
					Context(fmt.Sprintf("using IPv%v", family), func() {
						BeforeEach(func() {
							ipFamily = family
						})
						testTimestamps(family)
					})
				}
			})

			Context("with the same cable name", func() {
				testTimestamps(k8snet.IPv4)

				Context("but different endpoint IP", func() {
					BeforeEach(func() {
						newEndpoint.Spec.PublicIPs = []string{"3.3.3.3", "FDC7:BF8B:E62C:ABCD:1111:2222:3333:7777"}
					})

					for _, family := range []k8snet.IPFamily{k8snet.IPv4, k8snet.IPv6} {
						Context(fmt.Sprintf("using IPv%v", family), func() {
							BeforeEach(func() {
								ipFamily = family
							})
							It("should disconnect from the previous endpoint and connect to the new one", func() {
								fakeDriver.AwaitDisconnectFromEndpoint(&prevEndpoint.Spec, family)
								fakeDriver.AwaitConnectToEndpoint(natEndpointInfoFor(newEndpoint, family))
							})
						})
					}
				})

				Context("but different backend configuration", func() {
					BeforeEach(func() {
						newEndpoint.Spec.BackendConfig = map[string]string{"port": "6789"}
					})

					for _, family := range []k8snet.IPFamily{k8snet.IPv4, k8snet.IPv6} {
						Context(fmt.Sprintf("using IPv%v", family), func() {
							BeforeEach(func() {
								ipFamily = family
							})
							It("should disconnect from the previous endpoint and connect to the new one", func() {
								fakeDriver.AwaitDisconnectFromEndpoint(&prevEndpoint.Spec, family)
								fakeDriver.AwaitConnectToEndpoint(natEndpointInfoFor(newEndpoint, family))
							})
						})
					}
				})

				Context("and connection info", func() {
					It("should not disconnect from the previous endpoint nor connect to the new one", func() {
						fakeDriver.AwaitNoDisconnectFromEndpoint()
						fakeDriver.AwaitNoConnectToEndpoint()
					})
				})
			})
		})

		Context("and an endpoint was previously installed for another cluster", func() {
			It("should connect to the new endpoint and not disconnect from the previous one", func() {
				otherEndpoint := subv1.Endpoint{Spec: subv1.EndpointSpec{
					ClusterID: "other",
					CableName: "submariner-cable-other-1.1.1.1",
				}}

				Expect(engine.InstallCable(&otherEndpoint, k8snet.IPv4)).To(Succeed())
				fakeDriver.AwaitConnectToEndpoint(natEndpointInfoFor(&otherEndpoint, k8snet.IPv4))

				Expect(engine.InstallCable(remoteEndpoint, k8snet.IPv4)).To(Succeed())
				fakeDriver.AwaitConnectToEndpoint(natEndpointInfoFor(remoteEndpoint, k8snet.IPv4))
				fakeDriver.AwaitNoDisconnectFromEndpoint()
			})
		})

		Context("followed by remove cable before NAT discovery is complete", func() {
			BeforeEach(func() {
				natDiscovery.captureAddEndpoint = make(chan *subv1.Endpoint, 10)
			})

			It("should not connect to the endpoint", func() {
				Expect(engine.InstallCable(remoteEndpoint, k8snet.IPv4)).To(Succeed())
				Eventually(natDiscovery.captureAddEndpoint).Should(Receive())

				Expect(engine.RemoveCable(remoteEndpoint, k8snet.IPv4)).To(Succeed())
				Eventually(natDiscovery.removeEndpoint).Should(Receive(Equal(remoteEndpoint.Spec.GetFamilyCableName(k8snet.IPv4))))
				fakeDriver.AwaitNoDisconnectFromEndpoint()

				natDiscovery.notifyReady(remoteEndpoint, k8snet.IPv4)
				fakeDriver.AwaitNoConnectToEndpoint()
			})
		})
	})

	When("install cable for a local endpoint", func() {
		It("should not connect to the endpoint", func() {
			Expect(engine.InstallCable(localEndpoint, k8snet.IPv4)).To(Succeed())
			fakeDriver.AwaitNoConnectToEndpoint()
		})
	})

	When("remove cable for a remote endpoint", func() {
		JustBeforeEach(func() {
			Expect(engine.InstallCable(remoteEndpoint, k8snet.IPv4)).To(Succeed())
			fakeDriver.AwaitConnectToEndpoint(natEndpointInfoFor(remoteEndpoint, k8snet.IPv4))
		})

		It("should disconnect from the endpoint", func() {
			Expect(engine.RemoveCable(remoteEndpoint, k8snet.IPv4)).To(Succeed())
			fakeDriver.AwaitDisconnectFromEndpoint(&remoteEndpoint.Spec, k8snet.IPv4)
			Eventually(natDiscovery.removeEndpoint).Should(Receive(Equal(remoteEndpoint.Spec.GetFamilyCableName(k8snet.IPv4))))
		})

		Context("and the driver fails to disconnect from the endpoint", func() {
			JustBeforeEach(func() {
				fakeDriver.ErrOnDisconnectFromEndpoint = errors.New("fake disconnect error")
			})

			It("should return an error", func() {
				Expect(engine.RemoveCable(remoteEndpoint, k8snet.IPv4)).To(HaveOccurred())
			})
		})
	})

	When("remove cable for a local endpoint", func() {
		JustBeforeEach(func() {
			Expect(engine.InstallCable(remoteEndpoint, k8snet.IPv4)).To(Succeed())
			fakeDriver.AwaitConnectToEndpoint(natEndpointInfoFor(remoteEndpoint, k8snet.IPv4))
		})

		It("should not disconnect from the endpoint", func() {
			Expect(engine.RemoveCable(localEndpoint, k8snet.IPv4)).To(Succeed())
			fakeDriver.AwaitNoDisconnectFromEndpoint()
			Consistently(natDiscovery.removeEndpoint).ShouldNot(Receive())
		})
	})

	When("list cable connections", func() {
		BeforeEach(func() {
			fakeDriver.Connections = []subv1.Connection{{Endpoint: remoteEndpoint.Spec}}
		})

		It("should retrieve the connections from the driver", func() {
			Expect(engine.ListCableConnections()).To(Equal(fakeDriver.Connections))
		})

		Context("and retrieval of the driver's connections fails", func() {
			JustBeforeEach(func() {
				fakeDriver.Connections = errors.New("fake connections error")
			})

			It("should return an error", func() {
				_, err := engine.ListCableConnections()
				Expect(err).To(ContainErrorSubstring(fakeDriver.Connections.(error)))
			})
		})
	})

	When("the HA status is queried", func() {
		It("should return active", func() {
			Expect(engine.GetHAStatus()).To(Equal(subv1.HAStatusActive))
		})
	})

	When("driver initialization fails", func() {
		BeforeEach(func() {
			fakeDriver.ErrOnInit = errors.New("fake init error")
		})

		It("should fail to start", func() {
		})
	})

	When("not started", func() {
		BeforeEach(func() {
			skipStart = true
		})

		Context("and the HA status is queried", func() {
			It("should return passive", func() {
				Expect(engine.GetHAStatus()).To(Equal(subv1.HAStatusPassive))
			})
		})

		Context("and list of cable connections is queried", func() {
			It("should return non-nil", func() {
				Expect(engine.ListCableConnections()).ToNot(BeNil())
			})
		})
	})

	When("after Stop is called", func() {
		BeforeEach(func() {
			fakeDriver.Connections = []subv1.Connection{{Endpoint: remoteEndpoint.Spec}}
		})

		JustBeforeEach(func() {
			engine.Stop()
		})

		Context("and the HA status is queried", func() {
			It("should return passive", func() {
				Expect(engine.GetHAStatus()).To(Equal(subv1.HAStatusPassive))
			})
		})

		Context("and list of cable connections is queried", func() {
			It("should return empty", func() {
				Expect(engine.ListCableConnections()).To(BeEmpty())
			})
		})
	})
})

func TestCableEngine(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cable Engine Suite")
}

type fakeNATDiscovery struct {
	removeEndpoint     chan string
	captureAddEndpoint chan *subv1.Endpoint
	readyChannel       chan *natdiscovery.NATEndpointInfo
}

func (n *fakeNATDiscovery) Run(_ <-chan struct{}) error {
	return nil
}

func (n *fakeNATDiscovery) AddEndpoint(endpoint *subv1.Endpoint, family k8snet.IPFamily) {
	if n.captureAddEndpoint != nil {
		n.captureAddEndpoint <- endpoint
		return
	}

	n.notifyReady(endpoint, family)
}

func (n *fakeNATDiscovery) RemoveEndpoint(endpointName string) {
	n.removeEndpoint <- endpointName
}

func (n *fakeNATDiscovery) GetReadyChannel() chan *natdiscovery.NATEndpointInfo {
	return n.readyChannel
}

func (n *fakeNATDiscovery) notifyReady(endpoint *subv1.Endpoint, family k8snet.IPFamily) {
	n.readyChannel <- natEndpointInfoFor(endpoint, family)
}

func natEndpointInfoFor(endpoint *subv1.Endpoint, family k8snet.IPFamily) *natdiscovery.NATEndpointInfo {
	return &natdiscovery.NATEndpointInfo{
		UseIP:     endpoint.Spec.GetPublicIP(family),
		UseNAT:    true,
		Endpoint:  *endpoint,
		UseFamily: family,
	}
}
