set -ex

function verify_subm_broker_operator() {
  # Accept cluster context as param
  context=$1
  kubectl config use-context $context

  # Verify SubM Broker namespace
  kubectl get ns | grep $subm_broker_ns

  # Verify SubM Broker CRD
  kubectl get crds | grep submarinerk8sbrokers.charts.helm.k8s.io

  # Verify SubM Broker Operator service account
  kubectl get sa --namespace=$subm_broker_ns | grep submariner-k8s-broker-operator

  # Verify SubM Broker Operator role
  kubectl get clusterroles | grep submariner-k8s-broker-operator

  # Verify SubM Broker Operator role binding
  kubectl get clusterrolebindings | grep submariner-k8s-broker-operator

  # Verify SubM Broker Operator deployment
  kubectl get deployments --namespace=$subm_broker_ns | grep submariner-k8s-broker-operator
}

function verify_subm_broker_cr() {
  # Accept cluster context as param
  context=$1
  kubectl config use-context $context

  # Verify SubM CR presence
  kubectl get submarinerk8sbrokers --namespace=$subm_broker_ns | grep example-submarinerk8sbroker

  # Show full SubM CR JSON
  kubectl get submarinerk8sbrokers example-submarinerk8sbroker --namespace=$subm_broker_ns -o json
}

function verify_subm_broker_op_pod() {
  # Accept cluster context as param
  context=$1
  kubectl config use-context $context

  subm_broker_op_pod_name=$(kubectl get pods --namespace=$subm_broker_ns -l name=submariner-k8s-broker-operator -o=jsonpath='{.items..metadata.name}')

  kubectl get pod $subm_broker_op_pod_name --namespace=$subm_broker_ns -o json
  # TODO: Add checks

  # Verify SubM Operator container image
  kubectl get pod $subm_broker_op_pod_name --namespace=$subm_broker_ns -o jsonpath='{.spec.containers..image}' | grep dfarrell07/submariner-broker-helm-operator:test

  # Verify SubM Operator pod status
  kubectl get pod $subm_broker_op_pod_name --namespace=$subm_broker_ns -o jsonpath='{.status.phase}' | grep Running

  # Verify logs
  kubectl logs $subm_broker_op_pod_name --namespace=$subm_broker_ns | grep "Became the leader"
}
