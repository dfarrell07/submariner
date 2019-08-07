set -ex

subm_broker_op_dir=../operators/helm/0.8.1-1/submariner-k8s-broker-operator

function deploy_subm_broker_operator() {
  # Accept cluster context as param
  context=$1
  kubectl config use-context $context

  pushd $subm_broker_op_dir

  # Create SubM Broker namespace
  if ! kubectl get ns | grep $subm_broker_ns; then
    kubectl create -f deploy/namespace.yaml
  fi

  # Create SubM Broker CRD
  if ! kubectl get crds | grep submarinerk8sbrokers.charts.helm.k8s.io; then
    kubectl create -f deploy/crds/charts_v1alpha1_submarinerk8sbroker_crd.yaml
  fi

  # Create SubM Broker Operator service account
  if ! kubectl get sa | grep submariner-k8s-broker-operator; then
    kubectl create --namespace=$subm_broker_ns -f deploy/service_account.yaml
  fi

  # Create SubM Broker Operator role
  if ! kubectl get clusterroles | grep submariner-k8s-broker-operator; then
    kubectl create -f deploy/role.yaml
  fi

  # Create SubM Broker Operator role binding
  if ! kubectl get clusterrolebindings | grep submariner-k8s-broker-operator; then
    kubectl create -f deploy/role_binding.yaml
  fi

  # Create SubM Broker Operator deployment
  if ! kubectl get deployments | grep submariner-k8s-broker-operator; then
    kubectl create --namespace=$subm_broker_ns -f deploy/operator.yaml
  fi

  # Wait for SubM Operator pod to be ready
  kubectl get pods --namespace=$subm_broker_ns
  kubectl wait --for=condition=Ready pods -l name=submariner-k8s-broker-operator --timeout=180s --namespace=$subm_broker_ns
  kubectl get pods --namespace=$subm_broker_ns

  popd
}

function deploy_subm_broker_cr() {
  # Accept cluster context to deploy SubM Operator into as param
  context=$1
  kubectl config use-context $context

  pushd $subm_broker_op_dir

  #cr_file=deploy/crds/$context-submariner-cr.yaml
  cr_file=deploy/crds/charts_v1alpha1_submarinerk8sbroker_cr.yaml

  # Create SubM CR if it doesn't exist
  if kubectl get submarinerk8sbroker 2>&1 | grep -q "No resources found"; then
    # Create the SubM Op CR, which will create SubM endpoints and clusters CRDs if they don't already exist
    kubectl apply --namespace=$subm_broker_ns -f $cr_file
  fi

  popd
}

function collect_subm_broker_vars() {
  # Accept cluster context to deploy SubM Operator into as param
  context=$1
  kubectl config use-context $context

  # Need to reuse this, gen only if not already set
  if [ -z ${SUBMARINER_PSK+x} ]; then
    SUBMARINER_PSK=$(cat /dev/urandom | LC_CTYPE=C tr -dc 'a-zA-Z0-9' | fold -w 64 | head -n 1)
    echo "New PSK $SUBMARINER_PSK"
  else
    echo "Reusing PSK $SUBMARINER_PSK"
  fi

  # Get URL of SubM broker endpoint
  # TODO: Is the default ns right here?
  SUBMARINER_BROKER_URL=$(kubectl -n default get endpoints kubernetes -o jsonpath="{.subsets[0].addresses[0].ip}:{.subsets[0].ports[?(@.name=='https')].port}")
  curl $SUBMARINER_BROKER_URL

  # Get CA cert for SubM Broker Operator service account
  SUBMARINER_BROKER_CA=$(kubectl get secrets --namespace=$subm_broker_ns -o jsonpath="{.items[?(@.metadata.annotations['kubernetes\.io/service-account\.name']=='submariner-k8s-broker-operator')].data['ca\.crt']}")
  # TODO: Verify not-null

  # Get token for SubM Broker Operator service account
  SUBMARINER_BROKER_TOKEN=$(kubectl get secrets --namespace=$subm_broker_ns -o jsonpath="{.items[?(@.metadata.annotations['kubernetes\.io/service-account\.name']=='submariner-k8s-broker-operator')].data.token}")
  # TODO: Verify not-null
}
