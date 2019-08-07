package submariner

import (
	"context"

	submarinerv1alpha1 "github.com/submariner-operator/submariner-operator/pkg/apis/submariner/v1alpha1"
	corev1 "k8s.io/api/core/v1"
  //apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_submariner")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Submariner Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileSubmariner{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("submariner-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Submariner
	err = c.Watch(&source.Kind{Type: &submarinerv1alpha1.Submariner{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Submariner
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &submarinerv1alpha1.Submariner{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileSubmariner implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileSubmariner{}

// ReconcileSubmariner reconciles a Submariner object
type ReconcileSubmariner struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Submariner object and makes changes based on the state read
// and what is in the Submariner.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileSubmariner) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Submariner")

	// Fetch the Submariner instance
	instance := &submarinerv1alpha1.Submariner{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

  // TODO: Add deployment-creation logic here? See memc example

  // TODO: Create submariner-engine SA
  //subm_engine_sa := corev1.ServiceAccount{}
  //subm_engine_sa.Name = "submariner-engine"
  // TODO: debug by checking if this works, as there should already be a submariner SA
  //subm_engine_sa.Name = "submariner"
  //reqLogger.Info("Created a new SA", "SA.Name", subm_engine_sa.Name)

  // FIXME: It seems CRDs need to be created before the Operator starts
  // FIXME: Do all these creations conditionally, only of the resource doesn't exist
  // Define endpoints.submariner.io CRD spec
  //endpoints_crd_spec_names := apiextensions.CustomResourceDefinitionNames{Plural: "endpoints", Singular: "endpoint", ListKind: "EndpointList", Kind: "Endpoint"}
  //endpoints_crd_spec_versions := apiextensions.CustomResourceDefinitionVersion{Name: "v1", Served: true, Storage: true}
  //endpoints_crd_spec_conversion := apiextensions.CustomResourceConversion{Strategy: "None"}
  //endpoints_crd_spec := apiextensions.CustomResourceDefinitionSpec{Group: "submariner.io", Names: endpoints_crd_spec_names, Scope: "Namespaced", Versions: []apiextensions.CustomResourceDefinitionVersion{endpoints_crd_spec_versions}, Version: "v1", Conversion: &endpoints_crd_spec_conversion}

  // Define endpoints.submariner.io CRD status
  //endpoints_crd_status_names := apiextensions.CustomResourceDefinitionNames{Plural: "endpoints", Singular: "endpoint", ListKind: "EndpointList", Kind: "Endpoint"}
  //endpoints_crd_status_storedversions := []string{"v1"}
  //endpoints_crd_status:= apiextensions.CustomResourceDefinitionStatus{AcceptedNames: endpoints_crd_status_names, StoredVersions: endpoints_crd_status_storedversions}

  // Define endpoints.submariner.io CRD
  //endpoints_crd := apiextensions.CustomResourceDefinition{Spec: endpoints_crd_spec, Status: endpoints_crd_status}
  //err = r.client.Create(context.TODO(), &endpoints_crd)
  //if err != nil {
  //  return reconcile.Result{}, err
  //}

  // Define clusters.submariner.io CRD spec
  //clusters_crd_spec_names := apiextensions.CustomResourceDefinitionNames{Plural: "clusters", Singular: "cluster", ListKind: "ClusterList", Kind: "Cluster"}
  //clusters_crd_spec_versions := apiextensions.CustomResourceDefinitionVersion{Name: "v1", Served: true, Storage: true}
  //clusters_crd_spec_conversion := apiextensions.CustomResourceConversion{Strategy: "None"}
  //clusters_crd_spec := apiextensions.CustomResourceDefinitionSpec{Group: "submariner.io", Names: clusters_crd_spec_names, Scope: "Namespaced", Versions: []apiextensions.CustomResourceDefinitionVersion{clusters_crd_spec_versions}, Version: "v1", Conversion: &clusters_crd_spec_conversion}

  // Define clusters.submariner.io CRD status
  //clusters_crd_status_names := apiextensions.CustomResourceDefinitionNames{Plural: "clusters", Singular: "cluster", ListKind: "ClusterList", Kind: "Cluster"}
  //clusters_crd_status_storedversions := []string{"v1"}
  //clusters_crd_status:= apiextensions.CustomResourceDefinitionStatus{AcceptedNames: clusters_crd_status_names, StoredVersions: clusters_crd_status_storedversions}

  // Define clusters.submariner.io CRD
  //clusters_crd := apiextensions.CustomResourceDefinition{Spec: clusters_crd_spec, Status: clusters_crd_status}
  //err = r.client.Create(context.TODO(), &clusters_crd)
  //if err != nil {
  //  return reconcile.Result{}, err
  //}

	// Define a new Pod object
  // TODO: Make this responsive to size
	pod := newPodForCR(instance)

	// Set Submariner instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found := &corev1.Pod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Pod already exists - don't requeue
	reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	return reconcile.Result{}, nil
}

// newPodForCR returns a submariner pod with the same fields as the cr
func newPodForCR(cr *submarinerv1alpha1.Submariner) *corev1.Pod {
	labels := map[string]string{
		"app": "submariner-engine",
	}

  // Create EnvVars for Pod
  subm_namespace_env_var := corev1.EnvVar{Name: "SUBMARINER_NAMESPACE", Value: cr.Spec.SubmarinerNamespace}
  subm_clustercidr_env_var := corev1.EnvVar{Name: "SUBMARINER_CLUSTERCIDR", Value: cr.Spec.SubmarinerClustercidr}
  subm_servicecidr_env_var := corev1.EnvVar{Name: "SUBMARINER_SERVICECIDR", Value: cr.Spec.SubmarinerServicecidr}
  subm_token_env_var := corev1.EnvVar{Name: "SUBMARINER_TOKEN", Value: cr.Spec.SubmarinerToken}
  subm_clusterid_env_var := corev1.EnvVar{Name: "SUBMARINER_CLUSTERID", Value: cr.Spec.SubmarinerClusterid}
  subm_colorcodes_env_var := corev1.EnvVar{Name: "SUBMARINER_COLORCODES", Value: cr.Spec.SubmarinerColorcodes}
  subm_debug_env_var := corev1.EnvVar{Name: "SUBMARINER_DEBUG", Value: cr.Spec.SubmarinerDebug}
  subm_natenabled_env_var := corev1.EnvVar{Name: "SUBMARINER_NATENABLED", Value: cr.Spec.SubmarinerNatenabled}
  subm_broker_env_var := corev1.EnvVar{Name: "SUBMARINER_BROKER", Value: cr.Spec.SubmarinerBroker}
  broker_k8s_apiserver_env_var := corev1.EnvVar{Name: "BROKER_K8S_APISERVER", Value: cr.Spec.BrokerK8sApiserver}
  broker_k8s_apiservertoken_env_var := corev1.EnvVar{Name: "BROKER_K8S_APISERVERTOKEN", Value: cr.Spec.BrokerK8sApiservertoken}
  broker_k8s_remotenamespace_env_var := corev1.EnvVar{Name: "BROKER_K8S_REMOTENAMESPACE", Value: cr.Spec.BrokerK8sRemotenamespace}
  broker_k8s_ca_env_var := corev1.EnvVar{Name: "BROKER_K8S_CA", Value: cr.Spec.BrokerK8sCa}
  ce_ipsec_psk_env_var := corev1.EnvVar{Name: "CE_IPSEC_PSK", Value: cr.Spec.CeIpsecPsk}
  ce_ipsec_debug_env_var := corev1.EnvVar{Name: "CE_IPSEC_DEBUG", Value: cr.Spec.CeIpsecDebug}

  // Create SecurityContext for Pod
  security_context_add_net_admin := corev1.SecurityContext{Capabilities: &corev1.Capabilities{Add: []corev1.Capability{"NET_ADMIN"}}}

  // Create Pod
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "submariner",
          // Use var here
					Image:   "submariner:local",
          // Use var here
					Command: []string{"submariner.sh"},
          SecurityContext: &security_context_add_net_admin,
          Env:     []corev1.EnvVar{subm_namespace_env_var, subm_clustercidr_env_var, subm_servicecidr_env_var, subm_token_env_var, subm_clusterid_env_var, subm_colorcodes_env_var, subm_debug_env_var, subm_natenabled_env_var, subm_broker_env_var, broker_k8s_apiserver_env_var, broker_k8s_apiservertoken_env_var, broker_k8s_remotenamespace_env_var, broker_k8s_ca_env_var, ce_ipsec_psk_env_var, ce_ipsec_debug_env_var},
				},
			},
      // TODO: Use SA submariner-engine or submariner?
      ServiceAccountName: "submariner-operator",
		},
	}
}
