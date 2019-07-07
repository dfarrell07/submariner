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
  kubectl get crds | grep submariners.charts.helm.k8s.io
  kubectl api-resources | grep submariners

  # Verify SubM Operator SA
  kubectl get sa --namespace=$subm_ns | grep submariner-operator

  # Verify SubM Operator role
  kubectl get clusterroles | grep submariner-operator

  # Verify SubM Operator role binding
  kubectl get clusterrolebindings | grep submariner-operator

  # Verify SubM Operator deployment
  kubectl get deployments --namespace=$subm_ns | grep submariner-operator
}

function verify_subm_cr() {
  # Accept cluster context as param
  context=$1
  kubectl config use-context $context

  # Verify SubM CR presence
  kubectl get submariner --namespace=$subm_ns | grep example-submariner

  # Show full SubM CR JSON
  kubectl get submariner example-submariner --namespace=$subm_ns -o json

  # Verify SubM clusters/endpoints CRDs
  kubectl get crds | grep clusters.submariner.io
  kubectl get crds | grep endpoints.submariner.io

  # Verify SubM namespace
  kubectl get submariner example-submariner --namespace=$subm_ns -o jsonpath='{.metadata.namespace}' | grep submariner

  # Verify SubM PSK
  kubectl get submariner example-submariner --namespace=$subm_ns -o jsonpath='{.spec.ipsec.psk}' | grep $SUBMARINER_PSK

  # Verify SubM's config for SubM Broker server
  kubectl get submariner example-submariner --namespace=$subm_ns -o jsonpath='{.spec.broker.server}' | grep $SUBMARINER_BROKER_URL

  # Verify SubM's config for SubM Broker token
  kubectl get submariner example-submariner --namespace=$subm_ns -o jsonpath='{.spec.broker.token}' | grep $SUBMARINER_BROKER_TOKEN

  # Verify SubM's config for SubM Broker namespace
  kubectl get submariner example-submariner --namespace=$subm_ns -o jsonpath='{.spec.broker.namespace}' | grep $SUBMARINER_BROKER_NS

  # Verify SubM's config for SubM Broker CA
  kubectl get submariner example-submariner --namespace=$subm_ns -o jsonpath='{.spec.broker.ca}' | grep $SUBMARINER_BROKER_CA

  # Verify SubM Cluster ID
  kubectl get submariner example-submariner --namespace=$subm_ns -o jsonpath='{.spec.submariner.clusterId}' | grep $context

  # Verify SubM Cluster CIDR
  kubectl get submariner example-submariner --namespace=$subm_ns -o jsonpath='{.spec.submariner.clusterCidr}' | grep $clusterCidr

  # Verify SubM Service CIDR
  kubectl get submariner example-submariner --namespace=$subm_ns -o jsonpath='{.spec.submariner.serviceCidr}' | grep $serviceCidr

  # Verify SubM NAT config
  kubectl get submariner example-submariner --namespace=$subm_ns -o jsonpath='{.spec.submariner.natEnabled}' | grep $natEnabled

  # Verify SubM RouteAgent container image repo
  kubectl get submariner example-submariner --namespace=$subm_ns -o jsonpath='{.spec.routeAgent.image.repository}' | grep $subm_route_agent_image_repo

  # Verify SubM RouteAgent container image tag
  kubectl get submariner example-submariner --namespace=$subm_ns -o jsonpath='{.spec.routeAgent.image.tag}' | grep $subm_route_agent_image_tag

  # Verify SubM RouteAgent container image pull policy
  kubectl get submariner example-submariner --namespace=$subm_ns -o jsonpath='{.spec.routeAgent.image.pullPolicy}' | grep $subm_route_agent_image_policy
}

function verify_subm_op_pod() {
  # Accept cluster context as param
  context=$1
  kubectl config use-context $context

  subm_operator_pod_name=$(kubectl get pods --namespace=$subm_ns -l name=submariner-operator -o=jsonpath='{.items..metadata.name}')

  kubectl get pod $subm_operator_pod_name --namespace=$subm_ns -o json

  # Verify SubM Operator container image
  kubectl get pod $subm_operator_pod_name --namespace=$subm_ns -o jsonpath='{.spec.containers..image}' | grep docker.io/dfarrell07/submariner-helm-operator:test

  # Verify SubM Operator pod status
  kubectl get pod $subm_operator_pod_name --namespace=$subm_ns -o jsonpath='{.status.phase}' | grep Running

  # Verify logs
  kubectl logs $subm_operator_pod_name --namespace=$subm_ns | grep "Became the leader"
}

function failing_subm_operator_verifcations() {
  kubectl config use-context cluster2
  kubectl get pods --namespace=submariner -l app=submariner-routeagent | grep submariner-routeagent
  kubectl get pods --namespace=submariner -l app=submariner-engine | grep submariner
  kubectl config use-context cluster3
  kubectl get pods --namespace=submariner -l app=submariner-routeagent | grep submariner-routeagent
  kubectl get pods --namespace=submariner -l app=submariner-engine | grep submariner
}
