set -ex

function verify_subm_gateway_label() {
  # Accept cluster context as param
  context=$1
  kubectl config use-context $context

  kubectl get node $context-worker -o jsonpath='{.metadata.labels}' | grep submariner.io/gateway:true
}

function verify_subm_operator() {
  # Accept cluster context as param
  context=$1
  kubectl config use-context $context

  # Verify SubM namespace (ignore SubM Broker ns)
  kubectl get ns | grep -v $subm_broker_ns | grep $subm_ns

  # Verify SubM Operator CRD
  if [[ $operator_helm = true ]]; then
    kubectl get crds | grep submariners.charts.helm.k8s.io
  fi
  if [[ $operator_go = true ]]; then
    kubectl get crds | grep submariners.submariner.io
  fi
  kubectl api-resources | grep submariners

  # Verify SubM Operator SA
  kubectl get sa --namespace=$subm_ns | grep submariner-operator

  # Verify SubM Operator role
  # TODO: Why are these different between go and helm operators?
  if [[ $operator_helm = true ]]; then
    kubectl get clusterroles --namespace=$subm_ns | grep submariner-operator
  fi
  if [[ $operator_go = true ]]; then
    kubectl get roles --namespace=$subm_ns | grep submariner-operator
  fi

  # Verify SubM Operator role binding
  # TODO: Why are these different between go and helm operators?
  if [[ $operator_helm = true ]]; then
    kubectl get clusterrolebindings --namespace=$subm_ns | grep submariner-operator
  fi
  if [[ $operator_go = true ]]; then
    kubectl get rolebindings --namespace=$subm_ns | grep submariner-operator
  fi

  # Verify SubM Operator deployment
  kubectl get deployments --namespace=$subm_ns | grep submariner-operator
}

function verify_subm_crd() {
  # Accept cluster context as param
  context=$1
  kubectl config use-context $context

  crd_name=submariners.submariner.io

  # Verify presence of CRD
  kubectl get crds | grep $crd_name

  # Show full CRD
  kubectl get crd $crd_name -o yaml

  # Verify details of CRD
  kubectl get crd $crd_name -o jsonpath='{.metadata.name}' | grep $crd_name
  kubectl get crd $crd_name -o jsonpath='{.spec.scope}' | grep Namespaced
  kubectl get crd $crd_name -o jsonpath='{.spec.group}' | grep submariner.io
  kubectl get crd $crd_name -o jsonpath='{.spec.version}' | grep v1alpha1
  kubectl get crd $crd_name -o jsonpath='{.spec.names.kind}' | grep Submariner

  if [[ $openapi_checks_enabled = true ]]; then
    kubectl get crd $crd_name -o jsonpath='{.spec.validation.openAPIV3Schema.properties.spec.required}' | grep ceIpsecDebug
    kubectl get crd $crd_name -o jsonpath='{.spec.validation.openAPIV3Schema.properties.spec.required}' | grep ceIpsecPsk
    kubectl get crd $crd_name -o jsonpath='{.spec.validation.openAPIV3Schema.properties.spec.required}' | grep brokerK8sCa
    kubectl get crd $crd_name -o jsonpath='{.spec.validation.openAPIV3Schema.properties.spec.required}' | grep brokerK8sRemotenamespace
    kubectl get crd $crd_name -o jsonpath='{.spec.validation.openAPIV3Schema.properties.spec.required}' | grep brokerK8sApiservertoken
    kubectl get crd $crd_name -o jsonpath='{.spec.validation.openAPIV3Schema.properties.spec.required}' | grep brokerK8sApiserver
    kubectl get crd $crd_name -o jsonpath='{.spec.validation.openAPIV3Schema.properties.spec.required}' | grep submarinerBroker
    kubectl get crd $crd_name -o jsonpath='{.spec.validation.openAPIV3Schema.properties.spec.required}' | grep submarinerNatenabled
    kubectl get crd $crd_name -o jsonpath='{.spec.validation.openAPIV3Schema.properties.spec.required}' | grep submarinerDebug
    kubectl get crd $crd_name -o jsonpath='{.spec.validation.openAPIV3Schema.properties.spec.required}' | grep submarinerColorcodes
    kubectl get crd $crd_name -o jsonpath='{.spec.validation.openAPIV3Schema.properties.spec.required}' | grep submarinerClusterid
    kubectl get crd $crd_name -o jsonpath='{.spec.validation.openAPIV3Schema.properties.spec.required}' | grep submarinerToken
    kubectl get crd $crd_name -o jsonpath='{.spec.validation.openAPIV3Schema.properties.spec.required}' | grep submarinerServicecidr
    kubectl get crd $crd_name -o jsonpath='{.spec.validation.openAPIV3Schema.properties.spec.required}' | grep submarinerClustercidr
    kubectl get crd $crd_name -o jsonpath='{.spec.validation.openAPIV3Schema.properties.spec.required}' | grep submarinerNamespace
    kubectl get crd $crd_name -o jsonpath='{.spec.validation.openAPIV3Schema.properties.spec.required}' | grep count
  fi
}

function verify_endpoints_crd() {
  # Accept cluster context as param
  context=$1
  kubectl config use-context $context

  crd_name=endpoints.submariner.io

  # Verify presence of CRD
  kubectl get crds | grep $crd_name

  # Show full CRD
  kubectl get crd endpoints.submariner.io -o yaml

  # Verify details of CRD
  kubectl get crd $crd_name -o jsonpath='{.metadata.name}' | grep $crd_name
  kubectl get crd $crd_name -o jsonpath='{.spec.scope}' | grep Namespaced
  kubectl get crd $crd_name -o jsonpath='{.spec.group}' | grep submariner.io
  # TODO: Should this version really be v1, or maybe v1alpha1?
  kubectl get crd $crd_name -o jsonpath='{.spec.version}' | grep v1
  kubectl get crd $crd_name -o jsonpath='{.spec.names.kind}' | grep Endpoint
}

function verify_clusters_crd() {
  # Accept cluster context as param
  context=$1
  kubectl config use-context $context

  crd_name=clusters.submariner.io

  # Verify presence of CRD
  kubectl get crds | grep clusters.submariner.io

  # Show full CRD
  kubectl get crd clusters.submariner.io -o yaml

  # Verify details of CRD
  kubectl get crd $crd_name -o jsonpath='{.metadata.name}' | grep $crd_name
  kubectl get crd $crd_name -o jsonpath='{.spec.scope}' | grep Namespaced
  kubectl get crd $crd_name -o jsonpath='{.spec.group}' | grep submariner.io
  # TODO: Should this version really be v1, or maybe v1alpha1?
  kubectl get crd $crd_name -o jsonpath='{.spec.version}' | grep v1
  kubectl get crd $crd_name -o jsonpath='{.spec.names.kind}' | grep Cluster
}

function verify_routeagents_crd() {
  # Accept cluster context as param
  context=$1
  kubectl config use-context $context

  crd_name=routeagents.submariner.io

  # Verify presence of CRD
  kubectl get crds | grep routeagents.submariner.io

  # Show full CRD
  kubectl get crd routeagents.submariner.io -o yaml

  # Verify details of CRD
  kubectl get crd $crd_name -o jsonpath='{.metadata.name}' | grep $crd_name
  kubectl get crd $crd_name -o jsonpath='{.spec.scope}' | grep Namespaced
  kubectl get crd $crd_name -o jsonpath='{.spec.group}' | grep submariner.io
  # TODO: Should this version really be v1, or maybe v1alpha1?
  #kubectl get crd $crd_name -o jsonpath='{.spec.version}' | grep v1
  kubectl get crd $crd_name -o jsonpath='{.spec.version}' | grep v1alpha1
  kubectl get crd $crd_name -o jsonpath='{.spec.names.kind}' | grep Routeagent

  # TODO: Add additional checks
}

function verify_subm_cr() {
  # Accept cluster context as param
  context=$1
  kubectl config use-context $context

  # TODO: Use $engine_deployment_name here?

  # Verify SubM CR presence
  kubectl get submariner --namespace=$subm_ns | grep $deployment_name

  # Show full SubM CR JSON
  kubectl get submariner $deployment_name --namespace=$subm_ns -o json

  # Verify SubM namespace
  kubectl get submariner $deployment_name --namespace=$subm_ns -o jsonpath='{.metadata.namespace}' | grep submariner



  if [[ $operator_helm = true ]]; then
    # Verify SubM PSK
    kubectl get submariner $deployment_name --namespace=$subm_ns -o jsonpath='{.spec.ipsec.psk}' | grep $SUBMARINER_PSK

    # Verify SubM's config for SubM Broker server
    kubectl get submariner $deployment_name --namespace=$subm_ns -o jsonpath='{.spec.broker.server}' | grep $SUBMARINER_BROKER_URL

    # Verify SubM's config for SubM Broker token
    kubectl get submariner $deployment_name --namespace=$subm_ns -o jsonpath='{.spec.broker.token}' | grep $SUBMARINER_BROKER_TOKEN

    # Verify SubM's config for SubM Broker namespace
    kubectl get submariner $deployment_name --namespace=$subm_ns -o jsonpath='{.spec.broker.namespace}' | grep $SUBMARINER_BROKER_NS

    # Verify SubM's config for SubM Broker CA
    kubectl get submariner $deployment_name --namespace=$subm_ns -o jsonpath='{.spec.broker.ca}' | grep $SUBMARINER_BROKER_CA

    # Verify SubM Cluster ID
    kubectl get submariner $deployment_name --namespace=$subm_ns -o jsonpath='{.spec.submariner.clusterId}' | grep $context

    # Verify SubM Cluster CIDR
    kubectl get submariner $deployment_name --namespace=$subm_ns -o jsonpath='{.spec.submariner.clusterCidr}' | grep $clusterCidr

    # Verify SubM Service CIDR
    kubectl get submariner $deployment_name --namespace=$subm_ns -o jsonpath='{.spec.submariner.serviceCidr}' | grep $serviceCidr

    # Verify SubM NAT config
    kubectl get submariner $deployment_name --namespace=$subm_ns -o jsonpath='{.spec.submariner.natEnabled}' | grep $natEnabled

    # Verify SubM RouteAgent container image repo
    kubectl get submariner $deployment_name --namespace=$subm_ns -o jsonpath='{.spec.routeAgent.image.repository}' | grep $subm_routeagent_image_repo

    # Verify SubM RouteAgent container image tag
    kubectl get submariner $deployment_name --namespace=$subm_ns -o jsonpath='{.spec.routeAgent.image.tag}' | grep $subm_routeagent_image_tag

    # Verify SubM RouteAgent container image pull policy
    kubectl get submariner $deployment_name --namespace=$subm_ns -o jsonpath='{.spec.routeAgent.image.pullPolicy}' | grep $subm_routeagent_image_policy
  fi
  if [[ $operator_go = true ]]; then
    # Verify SubM PSK
    kubectl get submariner $deployment_name --namespace=$subm_ns -o jsonpath='{.spec.ce_ipsec_psk}' | grep $SUBMARINER_PSK

    # Verify SubM's config for SubM Broker server
    kubectl get submariner $deployment_name --namespace=$subm_ns -o jsonpath='{..spec.broker_k8s_apiserver}' | grep $SUBMARINER_BROKER_URL

    # Verify SubM's config for SubM Broker token
    kubectl get submariner $deployment_name --namespace=$subm_ns -o jsonpath='{.spec.broker_k8s_apiservertoken}' | grep $SUBMARINER_BROKER_TOKEN

    # Verify SubM's config for SubM Broker namespace
    kubectl get submariner $deployment_name --namespace=$subm_ns -o jsonpath='{.spec.broker_k8s_remotenamespace}' | grep $SUBMARINER_BROKER_NS

    # TODO: Convert more of the Helm-based checks above into correct JSON paths for go
  fi
}

function verify_routeagent_cr() {
  # Accept cluster context as param
  context=$1
  kubectl config use-context $context

  # Verify Routeagent CR presence
  kubectl get routeagent --namespace=$subm_ns | grep $routeagent_deployment_name

  # Show full Routeagent CR JSON
  kubectl get routeagent $routeagent_deployment_name --namespace=$subm_ns -o json

  # Verify Routeagent CR
  kubectl get routeagent $routeagent_deployment_name --namespace=$subm_ns -o jsonpath='{.metadata.namespace}' | grep $subm_ns
  kubectl get routeagent $routeagent_deployment_name --namespace=$subm_ns -o jsonpath='{.apiVersion}' | grep submariner.io/v1alpha1
  kubectl get routeagent $routeagent_deployment_name --namespace=$subm_ns -o jsonpath='{.kind}' | grep Routeagent
  kubectl get routeagent $routeagent_deployment_name --namespace=$subm_ns -o jsonpath='{.metadata.name}' | grep $routeagent_deployment_name
  kubectl get routeagent $routeagent_deployment_name --namespace=$subm_ns -o jsonpath='{.spec.submariner_clusterid}' | grep $context
  kubectl get routeagent $routeagent_deployment_name --namespace=$subm_ns -o jsonpath='{.spec.image}' | grep $subm_routeagent_image_repo:$subm_routeagent_image_tag
  kubectl get routeagent $routeagent_deployment_name --namespace=$subm_ns -o jsonpath='{.spec.submariner_namespace}' | grep $subm_ns
  kubectl get routeagent $routeagent_deployment_name --namespace=$subm_ns -o jsonpath='{.spec.submariner_debug}' | grep $subm_debug
}

function verify_subm_op_pod() {
  # Accept cluster context as param
  context=$1
  kubectl config use-context $context

  # TODO: Support (dynamic) pod name when running with Helm
  subm_operator_pod_name=$(kubectl get pods --namespace=$subm_ns -l name=submariner-operator -o=jsonpath='{.items..metadata.name}')

  # Show SubM Operator pod info
  kubectl get pod $subm_operator_pod_name --namespace=$subm_ns -o json

  # Verify SubM Operator pod status
  kubectl get pod $subm_operator_pod_name --namespace=$subm_ns -o jsonpath='{.status.phase}' | grep Running

  # Show SubM Operator pod logs
  kubectl logs $subm_operator_pod_name --namespace=$subm_ns
  # TODO: Verify logs

  # Show SubM Operator pod environment variables
  kubectl exec -it $subm_operator_pod_name -n submariner -- env
  kubectl exec -it $subm_operator_pod_name -n submariner -- env | grep "OPERATOR_NAME=submariner-operator"
  # TODO: Verify additional expected env vars (from CR) are set in running operator container
}

function verify_subm_routeagent_pod() {
  # Accept cluster context as param
  context=$1
  kubectl config use-context $context

  # FIXME: These are TDD expected-fails for the Operator currently, as there is no Go logic to read the CR and create a Pod

  kubectl wait --for=condition=Ready pods -l app=submariner-routeagent --timeout=120s --namespace=$subm_ns

  # Loop tests over all routeagent pods
  subm_routeagent_pod_names=$(kubectl get pods --namespace=$subm_ns -l app=submariner-routeagent -o=jsonpath='{.items..metadata.name}')
  # Globing-safe method, but -a flag gives me trouble in ZSH for some reason
  read -ra subm_routeagent_pod_names_array <<<"$subm_routeagent_pod_names"
  # TODO: Fail if there are zero routeagent pods
  for subm_routeagent_pod_name in "${subm_routeagent_pod_names_array[@]}"; do
    echo "Testing Submariner routeagent pod $subm_routeagent_pod_name"
    kubectl get pod $subm_routeagent_pod_name --namespace=$subm_ns -o json
    kubectl get pod $subm_routeagent_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..image}' | grep $subm_routeagent_image_repo:$subm_engine_image_tag
    kubectl get pod $subm_routeagent_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..securityContext.capabilities.add}' | grep ALL
    kubectl get pod $subm_routeagent_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..securityContext.allowPrivilegeEscalation}' | grep "true"
    kubectl get pod $subm_routeagent_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..securityContext.privileged}' | grep "true"
    kubectl get pod $subm_routeagent_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..command}' | grep submariner-route-agent.sh
    kubectl get pod $subm_routeagent_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..env}'
    kubectl get pod $subm_routeagent_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..env}' | grep "name:SUBMARINER_NAMESPACE value:$subm_ns"
    kubectl get pod $subm_routeagent_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..env}' | grep "name:SUBMARINER_CLUSTERID value:$context"
    kubectl get pod $subm_routeagent_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..env}' | grep "name:SUBMARINER_DEBUG value:$subm_debug"
    # TODO: Use submariner-routeagent SA vs submariner-operator
    #kubectl get pod $subm_routeagent_pod_name --namespace=$subm_ns -o jsonpath='{.spec.serviceAccount}' | grep submariner-routeagent
    kubectl get pod $subm_routeagent_pod_name --namespace=$subm_ns -o jsonpath='{.spec.serviceAccount}' | grep submariner-operator
    kubectl get pod $subm_routeagent_pod_name --namespace=$subm_ns -o jsonpath='{.status.phase}' | grep Running
    kubectl get pod $subm_routeagent_pod_name --namespace=$subm_ns -o jsonpath='{.metadata.namespace}' | grep $subm_ns
  done
}

function verify_subm_engine_pod() {
  # Accept cluster context as param
  context=$1
  kubectl config use-context $context

  # TODO: Is it better to fail if pod doesn't go to condition Ready or to continue and let the tests run/fail/give details?
  #kubectl wait --for=condition=Ready pods -l app=submariner-engine --timeout=120s --namespace=$subm_ns
  # TODO: For now, don't fail/end execution if pod doesn't go to Ready
  #kubectl wait --for=condition=Ready pods -l app=submariner-engine --timeout=120s --namespace=$subm_ns || true

  subm_engine_pod_name=$(kubectl get pods --namespace=$subm_ns -l app=submariner-engine -o=jsonpath='{.items..metadata.name}')

  kubectl get pod $subm_engine_pod_name --namespace=$subm_ns -o json
  kubectl get pod $subm_engine_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..image}' | grep submariner:local
  if [[ $helm = true ]]; then
    kubectl get pod $subm_engine_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..securityContext.capabilities.add}' | grep ALL
  fi
  if [[ $operator  = true ]]; then
    kubectl get pod $subm_engine_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..securityContext.capabilities.add}' | grep NET_ADMIN
  fi
  kubectl get pod $subm_engine_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..command}' | grep submariner.sh
  kubectl get pod $subm_engine_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..env}'
  kubectl get pod $subm_engine_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..env}' | grep "name:SUBMARINER_NAMESPACE value:$subm_ns"
  kubectl get pod $subm_engine_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..env}' | grep "name:SUBMARINER_CLUSTERCIDR value:$clusterCidr"
  if [[ $context = cluster2 ]]; then
    kubectl get pod $subm_engine_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..env}' | grep "name:SUBMARINER_SERVICECIDR value:$serviceCidr_cluster2"
  elif [[ $context = cluster3 ]]; then
    kubectl get pod $subm_engine_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..env}' | grep "name:SUBMARINER_SERVICECIDR value:$serviceCidr_cluster3"
  fi
  # FIXME: This token value isn't getting set
  kubectl get pod $subm_engine_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..env}' | grep "name:SUBMARINER_TOKEN"
  kubectl get pod $subm_engine_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..env}' | grep "name:SUBMARINER_CLUSTERID value:$context"
  kubectl get pod $subm_engine_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..env}' | grep "name:SUBMARINER_COLORCODES value:$subm_colorcodes"
  kubectl get pod $subm_engine_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..env}' | grep "name:SUBMARINER_DEBUG value:$subm_debug"
  kubectl get pod $subm_engine_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..env}' | grep "name:SUBMARINER_NATENABLED value:$natEnabled"
  kubectl get pod $subm_engine_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..env}' | grep "name:SUBMARINER_BROKER value:$subm_broker"
  kubectl get pod $subm_engine_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..env}' | grep "name:BROKER_K8S_APISERVER value:$SUBMARINER_BROKER_URL"
  kubectl get pod $subm_engine_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..env}' | grep "name:BROKER_K8S_APISERVERTOKEN value:$SUBMARINER_BROKER_TOKEN"
  kubectl get pod $subm_engine_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..env}' | grep "name:BROKER_K8S_REMOTENAMESPACE value:$SUBMARINER_BROKER_NS"
  kubectl get pod $subm_engine_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..env}' | grep "name:BROKER_K8S_CA value:$SUBMARINER_BROKER_CA"
  # FIXME: This changes between some deployment runs and causes failures
  kubectl get pod $subm_engine_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..env}' | grep "name:CE_IPSEC_PSK value:$SUBMARINER_PSK" || true
  kubectl get pod $subm_engine_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..env}' | grep "name:CE_IPSEC_DEBUG value:$ce_ipsec_debug"
  kubectl get pod $subm_engine_pod_name --namespace=$subm_ns -o jsonpath='{.status.phase}' | grep Running
  kubectl get pod $subm_engine_pod_name --namespace=$subm_ns -o jsonpath='{.metadata.namespace}' | grep $subm_ns
}

function failing_subm_operator_verifications() {
  kubectl config use-context cluster2
  kubectl get pods --namespace=submariner -l app=submariner-routeagent | grep submariner-routeagent
  kubectl get pods --namespace=submariner -l app=submariner-engine | grep submariner
  kubectl config use-context cluster3
  kubectl get pods --namespace=submariner -l app=submariner-routeagent | grep submariner-routeagent
  kubectl get pods --namespace=submariner -l app=submariner-engine | grep submariner
}
