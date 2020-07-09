package ipam

import (
	"fmt"
	"time"

	"github.com/coreos/go-iptables/iptables"
	"github.com/submariner-io/admiral/pkg/log"
	"github.com/submariner-io/submariner/pkg/routeagent/controllers/route"

	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

func NewController(spec *SubmarinerIpamControllerSpecification, config *InformerConfigStruct, globalCIDR string) (*Controller, error) {
	exclusionMap := make(map[string]bool)
	for _, v := range spec.ExcludeNS {
		exclusionMap[v] = true
	}
	pool, err := NewIPPool(globalCIDR)
	if err != nil {
		return nil, err
	}

	iptableHandler, err := iptables.New()
	if err != nil {
		return nil, err
	}

	ipamController := &Controller{
		kubeClientSet:    config.KubeClientSet,
		serviceWorkqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Services"),
		servicesSynced:   config.ServiceInformer.Informer().HasSynced,
		podWorkqueue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Pods"),
		podsSynced:       config.PodInformer.Informer().HasSynced,
		nodeWorkqueue:    workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Nodes"),
		nodesSynced:      config.NodeInformer.Informer().HasSynced,

		excludeNamespaces: exclusionMap,
		pool:              pool,
		ipt:               iptableHandler,
	}
	klog.Info("Setting up event handlers")
	config.ServiceInformer.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ipamController.enqueueObject(obj, ipamController.serviceWorkqueue)
		},
		UpdateFunc: func(old, newObj interface{}) {
			ipamController.handleUpdateService(old, newObj)
		},
		DeleteFunc: ipamController.handleRemovedService,
	}, handlerResync)
	config.PodInformer.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ipamController.enqueueObject(obj, ipamController.podWorkqueue)
		},
		UpdateFunc: func(old, newObj interface{}) {
			ipamController.handleUpdatePod(old, newObj)
		},
		DeleteFunc: ipamController.handleRemovedPod,
	}, handlerResync)
	config.NodeInformer.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ipamController.enqueueObject(obj, ipamController.nodeWorkqueue)
		},
		UpdateFunc: func(old, newObj interface{}) {
			ipamController.handleUpdateNode(old, newObj)
		},
		DeleteFunc: ipamController.handleRemovedNode,
	}, handlerResync)

	return ipamController, nil
}

func (i *Controller) Run(stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting IPAM Controller")

	err := i.initIPTableChains()
	if err != nil {
		return fmt.Errorf("initIPTableChains returned error. %v", err)
	}

	// Currently submariner global-net implementation works only with kube-proxy.
	if chainExists, _ := i.doesIPTablesChainExist("nat", kubeProxyServiceChainName); !chainExists {
		return fmt.Errorf("%q chain missing, cluster does not seem to use kube-proxy", kubeProxyServiceChainName)
	}

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, i.servicesSynced, i.podsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	go wait.Until(i.runServiceWorker, time.Second, stopCh)
	go wait.Until(i.runPodWorker, time.Second, stopCh)
	go wait.Until(i.runNodeWorker, time.Second, stopCh)
	<-stopCh
	klog.Info("Shutting down workers")
	i.cleanupIPTableRules()
	return nil
}

func (i *Controller) runServiceWorker() {
	for i.processNextObject(i.serviceWorkqueue, i.serviceGetter, i.serviceUpdater) {
	}
}

func (i *Controller) runPodWorker() {
	for i.processNextObject(i.podWorkqueue, i.podGetter, i.podUpdater) {
	}
}

func (i *Controller) runNodeWorker() {
	for i.processNextObject(i.nodeWorkqueue, i.nodeGetter, i.nodeUpdater) {
	}
}

func (i *Controller) processNextObject(objWorkqueue workqueue.RateLimitingInterface, objGetter func(namespace, name string) (runtime.Object, error), objUpdater func(runtimeObj runtime.Object, key string) error) bool {
	obj, shutdown := objWorkqueue.Get()
	if shutdown {
		return false
	}
	err := func() error {
		defer objWorkqueue.Done(obj)

		key := obj.(string)
		ns, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			objWorkqueue.Forget(obj)
			return fmt.Errorf("error while splitting meta namespace key %s: %v", key, err)
		}

		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			// Retrieve the latest version of object before attempting update
			// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
			runtimeObj, err := objGetter(ns, name)

			if err != nil {
				if errors.IsNotFound(err) {
					// already deleted Forget and return
					objWorkqueue.Forget(obj)
				} else {
					// could be API error, so we want to requeue
					logAndRequeue(key, objWorkqueue)
				}
				return fmt.Errorf("error retrieving submariner-ipam-controller object %s: %v", key, err)
			}

			switch runtimeObj := runtimeObj.(type) {
			case *k8sv1.Service:
				switch i.evaluateService(runtimeObj) {
				case Ignore:
					objWorkqueue.Forget(obj)
					return nil
				case Requeue:
					objWorkqueue.AddRateLimited(obj)
					return fmt.Errorf("service %s requeued %d times", key, objWorkqueue.NumRequeues(obj))
				}
			case *k8sv1.Pod:
				// Process pod event only when it has an ipaddress. Pods skipped here will be handled subsequently
				// during podUpdate event when an ipaddress is assigned to it.
				if runtimeObj.Status.PodIP == "" {
					objWorkqueue.Forget(obj)
					return nil
				}

				// Privileged pods that use hostNetwork will be ignored.
				if runtimeObj.Spec.HostNetwork {
					klog.V(log.DEBUG).Infof("Ignoring pod %q on host %q as it uses hostNetworking", key, runtimeObj.Status.PodIP)
					return nil
				}
			case *k8sv1.Node:
				switch i.evaluateNode(runtimeObj) {
				case Ignore:
					objWorkqueue.Forget(obj)
					return nil
				case Requeue:
					objWorkqueue.AddRateLimited(obj)
					return fmt.Errorf("Node %s requeued %d times", key, objWorkqueue.NumRequeues(obj))
				}
			}

			return objUpdater(runtimeObj, key)
		})
		if retryErr != nil {
			return retryErr
		}

		objWorkqueue.Forget(obj)
		return nil
	}()

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (i *Controller) enqueueObject(obj interface{}, workqueue workqueue.RateLimitingInterface) {
	if key := i.getEnqueueKey(obj); key != "" {
		klog.V(log.TRACE).Infof("Enqueueing %v for ipam controller", key)
		workqueue.AddRateLimited(key)
	}
}

func (i *Controller) getEnqueueKey(obj interface{}) string {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return ""
	}
	ns, _, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(err)
		return ""
	}
	if i.excludeNamespaces[ns] {
		return ""
	}
	return key
}

func (i *Controller) handleUpdateService(old interface{}, newObj interface{}) {
	//TODO: further minimize duplication between this and handleUpdatePod
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(newObj); err != nil {
		utilruntime.HandleError(err)
		return
	}

	if i.excludeNamespaces[newObj.(*k8sv1.Service).Namespace] {
		klog.V(log.DEBUG).Infof("In handleUpdateService, skipping Service %q as it belongs to excluded namespace.", key)
		return
	}

	oldGlobalIP := old.(*k8sv1.Service).GetAnnotations()[submarinerIpamGlobalIP]
	newGlobalIP := newObj.(*k8sv1.Service).GetAnnotations()[submarinerIpamGlobalIP]
	if oldGlobalIP != newGlobalIP && newGlobalIP != i.pool.GetAllocatedIP(key) {
		klog.V(log.DEBUG).Infof("GlobalIP changed from %s to %s for Service %q", oldGlobalIP, newGlobalIP, key)
		i.serviceWorkqueue.Add(key)
		return
	}

	if newGlobalIP == "" {
		klog.Warningf("In handleUpdateService, Service %s does not have globalIP annotation yet.", key)
	}
}

func (i *Controller) handleUpdatePod(old interface{}, newObj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(newObj); err != nil {
		utilruntime.HandleError(err)
		return
	}

	if i.excludeNamespaces[newObj.(*k8sv1.Pod).Namespace] {
		klog.V(log.DEBUG).Infof("In handleUpdatePod, skipping pod %q as it belongs to excluded namespace.", key)
		return
	}

	// Ignore privileged pods that use hostNetwork
	if newObj.(*k8sv1.Pod).Spec.HostNetwork {
		klog.V(log.DEBUG).Infof("Pod %q on host %q uses hostNetwork, ignoring", key, newObj.(*k8sv1.Pod).Status.HostIP)
		return
	}

	oldPodIP := old.(*k8sv1.Pod).Status.PodIP
	updatedPodIP := newObj.(*k8sv1.Pod).Status.PodIP
	// When the POD is getting terminated, sometimes we get pod update event with podIP removed.
	if oldPodIP != "" && updatedPodIP == "" {
		klog.V(log.DEBUG).Infof("Pod %q with ip %s is being terminated", key, oldPodIP)
		i.handleRemovedPod(old)
		return
	}

	newGlobalIP := newObj.(*k8sv1.Pod).GetAnnotations()[submarinerIpamGlobalIP]
	if newGlobalIP == "" {
		// Pod events that are skipped during addEvent are handled here when they are assigned an ipaddress.
		if updatedPodIP != "" {
			klog.V(log.DEBUG).Infof("In handleUpdatePod, pod %q is now assigned %s address, enqueing", key, updatedPodIP)
			i.podWorkqueue.Add(key)
			return
		} else {
			klog.Warningf("In handleUpdatePod, waiting for K8s to assign an IPAddress to pod %s, state %s", key, newObj.(*k8sv1.Pod).Status.Phase)
			return
		}
	}

	oldGlobalIP := old.(*k8sv1.Pod).GetAnnotations()[submarinerIpamGlobalIP]
	if oldGlobalIP != newGlobalIP && newGlobalIP != i.pool.GetAllocatedIP(key) {
		klog.V(log.DEBUG).Infof("GlobalIP changed from %s to %s for %s", oldGlobalIP, newGlobalIP, key)
		i.podWorkqueue.Add(key)
		return
	}
}

func (i *Controller) handleUpdateNode(old interface{}, newObj interface{}) {
	// Todo minimize the duplication
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(newObj); err != nil {
		utilruntime.HandleError(err)
		return
	}

	if i.isControlNode(newObj.(*k8sv1.Node)) {
		klog.V(log.TRACE).Infof("Node %q is a control node, skip processing.", newObj.(*k8sv1.Node).Name)
		return
	}

	oldCniIfaceIPOnNode := old.(*k8sv1.Node).GetAnnotations()[route.CniInterfaceIP]
	newCniIfaceIPOnNode := newObj.(*k8sv1.Node).GetAnnotations()[route.CniInterfaceIP]
	if oldCniIfaceIPOnNode == "" && newCniIfaceIPOnNode == "" {
		klog.V(log.DEBUG).Infof("In handleUpdateNode, node %q is not yet annotated with cniIfaceIP, enqueing", newObj.(*k8sv1.Node).Name)
		i.enqueueObject(newObj, i.nodeWorkqueue)
		return
	}

	oldGlobalIP := old.(*k8sv1.Node).GetAnnotations()[submarinerIpamGlobalIP]
	newGlobalIP := newObj.(*k8sv1.Node).GetAnnotations()[submarinerIpamGlobalIP]
	if oldGlobalIP != newGlobalIP && newGlobalIP != i.pool.GetAllocatedIP(key) {
		klog.V(log.DEBUG).Infof("GlobalIP changed from %s to %s for %s", oldGlobalIP, newGlobalIP, key)
		i.enqueueObject(newObj, i.nodeWorkqueue)
	}
}

func (i *Controller) handleRemovedService(obj interface{}) {
	//TODO: further minimize duplication between this and handleRemovedPod
	var service *k8sv1.Service
	var ok bool
	var key string
	var err error
	if service, ok = obj.(*k8sv1.Service); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			klog.Errorf("Could not convert object %v to Service", obj)
			return
		}
		service, ok = tombstone.Obj.(*k8sv1.Service)
		if !ok {
			klog.Errorf("Could not convert object tombstone %v to Service", tombstone.Obj)
			return
		}
	}
	if !i.excludeNamespaces[service.Namespace] {
		globalIP := service.Annotations[submarinerIpamGlobalIP]
		if globalIP != "" {
			if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
				utilruntime.HandleError(err)
				return
			}
			i.pool.Release(key)
			klog.V(log.DEBUG).Infof("Released ip %s for service %s", globalIP, key)
			err = i.syncServiceRules(service, globalIP, DeleteRules)
			if err != nil {
				klog.Errorf("Error while cleaning up Service %q ingress rules. %v", key, err)
			}
		} else {
			klog.Errorf("Error: handleRemovedService called for %q, but globalIP annotation is missing.", key)
		}
	}
}

func (i *Controller) handleRemovedPod(obj interface{}) {
	var pod *k8sv1.Pod
	var ok bool
	var key string
	var err error
	if pod, ok = obj.(*k8sv1.Pod); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			klog.Errorf("Could not convert object %v to pod", obj)
			return
		}
		pod, ok = tombstone.Obj.(*k8sv1.Pod)
		if !ok {
			klog.Errorf("Could not convert object tombstone %v to Pod", tombstone.Obj)
			return
		}
	}
	if !i.excludeNamespaces[pod.Namespace] {
		globalIP := pod.Annotations[submarinerIpamGlobalIP]
		if globalIP != "" && pod.Status.PodIP != "" {
			if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
				utilruntime.HandleError(err)
				return
			}
			i.pool.Release(key)
			klog.V(log.DEBUG).Infof("Released globalIP %s for pod %q", globalIP, key)
			err = i.syncPodRules(pod.Status.PodIP, globalIP, DeleteRules)
			if err != nil {
				klog.Errorf("Error while cleaning up Pod egress rules. %v", err)
			}
		} else {
			podKey := pod.Namespace + "/" + pod.Name
			klog.V(log.DEBUG).Infof("handleRemovedPod called for %q, that has globalIP %s and PodIP %s", podKey, globalIP, pod.Status.PodIP)
		}
	}
}

func (i *Controller) handleRemovedNode(obj interface{}) {
	// TODO: further minimize duplication between this and handleRemovedPod
	var node *k8sv1.Node
	var ok bool
	var key string
	var err error
	if node, ok = obj.(*k8sv1.Node); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			klog.Errorf("Could not convert object %v to Node", obj)
			return
		}
		node, ok = tombstone.Obj.(*k8sv1.Node)
		if !ok {
			klog.Errorf("Could not convert object tombstone %v to Node", tombstone.Obj)
			return
		}
	}

	if i.isControlNode(node) {
		klog.V(log.TRACE).Infof("Node %q is a control node, skip processing.", node.Name)
		return
	}

	globalIP := node.Annotations[submarinerIpamGlobalIP]
	cniIfaceIP := node.Annotations[route.CniInterfaceIP]
	if globalIP != "" && cniIfaceIP != "" {
		if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
			utilruntime.HandleError(err)
			return
		}
		i.pool.Release(key)
		klog.V(log.DEBUG).Infof("Released ip %s for Node %s", globalIP, key)
		err = i.syncNodeRules(cniIfaceIP, globalIP, DeleteRules)
		if err != nil {
			klog.Errorf("Error while cleaning up HostNetwork egress rules. %v", err)
		}
	} else {
		klog.V(log.DEBUG).Infof("handleRemovedNode called for %q, that has globalIP %s and cniIfaceIP %s", key, globalIP, cniIfaceIP)
	}
}

func (i *Controller) annotateGlobalIP(key string, globalIP string) (string, error) {
	var ip string
	var err error
	if globalIP == "" {
		ip, err = i.pool.Allocate(key)
		if err != nil {
			return "", err
		}
		return ip, nil
	}
	givenIP, err := i.pool.RequestIP(key, globalIP)
	if err != nil {
		return "", err
	}
	if globalIP != givenIP {
		// This resource has been allocated a different IP
		klog.Warningf("Updating globalIP for %q from %s to %s", key, globalIP, givenIP)
		return givenIP, nil
	}
	// globalIP on the resource is now updated in the local pool
	// This case will be hit either when the gateway is migrated or when the globalNet Pod is restarted
	return "", nil
}

func (i *Controller) serviceGetter(namespace, name string) (runtime.Object, error) {
	return i.kubeClientSet.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
}

func (i *Controller) podGetter(namespace, name string) (runtime.Object, error) {
	return i.kubeClientSet.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
}

func (i *Controller) nodeGetter(namespace, name string) (runtime.Object, error) {
	return i.kubeClientSet.CoreV1().Nodes().Get(name, metav1.GetOptions{})
}

func (i *Controller) serviceUpdater(obj runtime.Object, key string) error {
	service := obj.(*k8sv1.Service)
	existingGlobalIP := service.GetAnnotations()[submarinerIpamGlobalIP]
	allocatedIP, err := i.annotateGlobalIP(key, existingGlobalIP)
	if err != nil { // failed to get globalIP or failed to update, we want to retry
		logAndRequeue(key, i.serviceWorkqueue)
		return fmt.Errorf("failed to annotate GlobalIP to service %s: %v", key, err)
	}

	// This case is hit in one of the two situations
	// 1. when the Service does not have the globalIP annotation and a new globalIP is allocated
	// 2. when the current globalIP annotation on the Service does not match with the info maintained by ipPool
	if allocatedIP != "" {
		klog.V(log.DEBUG).Infof("Allocating globalIP %s to Service %q ", allocatedIP, key)
		err = i.syncServiceRules(service, allocatedIP, AddRules)
		if err != nil {
			logAndRequeue(key, i.serviceWorkqueue)
			return err
		}

		annotations := service.GetAnnotations()
		if annotations == nil {
			annotations = map[string]string{}
		}
		annotations[submarinerIpamGlobalIP] = allocatedIP
		service.SetAnnotations(annotations)
		_, err := i.kubeClientSet.CoreV1().Services(service.Namespace).Update(service)
		if err != nil {
			logAndRequeue(key, i.serviceWorkqueue)
			return err
		}
	} else if existingGlobalIP != "" {
		klog.V(log.DEBUG).Infof("Service %q already has globalIP %s annotation, syncing rules", key, existingGlobalIP)
		// When Globalnet Controller is migrated, we get notification for all the existing Services.
		// For Services that already have the annotation, we update the local ipPool cache and sync
		// the iptable rules on the new GatewayNode.
		// Note: This case will also be hit when Globalnet Pod is restarted
		err = i.syncServiceRules(service, existingGlobalIP, AddRules)
		if err != nil {
			logAndRequeue(key, i.serviceWorkqueue)
			return err
		}
	}
	return nil
}

func (i *Controller) podUpdater(obj runtime.Object, key string) error {
	pod := obj.(*k8sv1.Pod)
	pod.GetSelfLink()
	existingGlobalIP := pod.GetAnnotations()[submarinerIpamGlobalIP]
	allocatedIP, err := i.annotateGlobalIP(key, existingGlobalIP)
	if err != nil { // failed to get globalIP or failed to update, we want to retry
		logAndRequeue(key, i.podWorkqueue)
		return fmt.Errorf("failed to annotate globalIP to Pod %q: %+v", key, err)
	}

	// This case is hit in one of the two situations
	// 1. when the POD does not have the globalIP annotation and a new globalIP is allocated
	// 2. when the current globalIP annotation on the POD does not match with the info maintained by ipPool
	if allocatedIP != "" {
		klog.V(log.DEBUG).Infof("Allocating globalIP %s to Pod %q ", allocatedIP, key)
		err = i.syncPodRules(pod.Status.PodIP, allocatedIP, AddRules)
		if err != nil {
			logAndRequeue(key, i.podWorkqueue)
			return err
		}

		annotations := pod.GetAnnotations()
		if annotations == nil {
			annotations = map[string]string{}
		}
		annotations[submarinerIpamGlobalIP] = allocatedIP
		pod.SetAnnotations(annotations)
		_, err := i.kubeClientSet.CoreV1().Pods(pod.Namespace).Update(pod)
		if err != nil {
			logAndRequeue(key, i.podWorkqueue)
			return err
		}
	} else if existingGlobalIP != "" {
		klog.V(log.DEBUG).Infof("Pod %q already has globalIP %s annotation, syncing rules", key, existingGlobalIP)
		// When Globalnet Controller is migrated, we get notification for all the existing PODs.
		// For PODs that already have the annotation, we update the local ipPool cache and sync
		// the iptable rules on the new GatewayNode.
		// Note: This case will also be hit when Globalnet Pod is restarted
		err = i.syncPodRules(pod.Status.PodIP, existingGlobalIP, AddRules)
		if err != nil {
			logAndRequeue(key, i.podWorkqueue)
			return err
		}
	}
	return nil
}

func (i *Controller) nodeUpdater(obj runtime.Object, key string) error {
	node := obj.(*k8sv1.Node)
	cniIfaceIP := node.GetAnnotations()[route.CniInterfaceIP]
	existingGlobalIP := node.GetAnnotations()[submarinerIpamGlobalIP]
	allocatedIP, err := i.annotateGlobalIP(key, existingGlobalIP)
	if err != nil { // failed to get globalIP or failed to update, we want to retry
		logAndRequeue(key, i.nodeWorkqueue)
		return fmt.Errorf("failed to annotate globalIP to node %q: %+v", key, err)
	}

	// This case is hit in one of the two situations
	// 1. when the Worker Node does not have the globalIP annotation and a new globalIP is allocated
	// 2. when the current globalIP annotation on the Node does not match with the info maintained by ipPool
	if allocatedIP != "" {
		klog.V(log.DEBUG).Infof("Allocating globalIP %s to Node %q ", allocatedIP, key)
		err = i.syncNodeRules(cniIfaceIP, allocatedIP, AddRules)
		if err != nil {
			logAndRequeue(key, i.nodeWorkqueue)
			return err
		}

		annotations := node.GetAnnotations()
		if annotations == nil {
			annotations = map[string]string{}
		}
		annotations[submarinerIpamGlobalIP] = allocatedIP
		node.SetAnnotations(annotations)

		_, err := i.kubeClientSet.CoreV1().Nodes().Update(node)
		if err != nil {
			logAndRequeue(key, i.nodeWorkqueue)
			return err
		}
	} else if existingGlobalIP != "" {
		klog.V(log.DEBUG).Infof("Node %q already has globalIP %s annotation, syncing rules", key, existingGlobalIP)
		// When Globalnet Controller is migrated, we get notification for all the existing Nodes.
		// For Worker Nodes that already have the annotation, we update the local ipPool cache and sync
		// the iptable rules on the new GatewayNode.
		// Note: This case will also be hit when Globalnet Pod is restarted
		err = i.syncNodeRules(cniIfaceIP, existingGlobalIP, AddRules)
		if err != nil {
			logAndRequeue(key, i.nodeWorkqueue)
			return err
		}
	}
	return nil
}

func logAndRequeue(key string, workqueue workqueue.RateLimitingInterface) {
	klog.V(log.DEBUG).Infof("%s enqueued %d times", key, workqueue.NumRequeues(key))
	workqueue.AddRateLimited(key)
}
