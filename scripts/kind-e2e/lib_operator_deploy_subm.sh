set -ex

subm_op_dir=../operators/helm/0.8.1-1/submariner-operator

# FIXME: Pods route-agent and engine pods are still not deploying

function add_subm_gateway_label() {
  # Accept cluster context to deploy SubM Operator into as param
  context=$1
  kubectl config use-context $context

  kubectl label node $context-worker "submariner.io/gateway=true" --overwrite
}

function deploy_subm_operator() {
  # Accept cluster context to deploy SubM Operator into as param
  context=$1
  kubectl config use-context $context

  subm_ns=submariner

  pushd $subm_op_dir

  # If SubM namespace doesn't exist (ignore SubM Broker ns), create it
  if ! kubectl get ns | grep -v $subm_broker_ns | grep $subm_ns; then
    kubectl create -f deploy/namespace.yaml
  fi

  # Create SubM Operator CRD if it doesn't exist
  if ! kubectl get crds | grep submariners.charts.helm.k8s.io; then
    kubectl create -f deploy/crds/charts_v1alpha1_submariner_crd.yaml
  fi

  # Create SubM Operator service account if it doesn't exist
  if ! kubectl get sa | grep submariner-operator; then
    kubectl create --namespace=$subm_ns -f deploy/service_account.yaml
  fi

  # Create SubM Operator role if it doesn't exist
  if ! kubectl get clusterroles | grep submariner-operator; then
    kubectl create -f deploy/role.yaml
  fi

  # Create SubM Operator role binding if it doesn't exist
  if ! kubectl get clusterrolebindings | grep submariner-operator; then
    kubectl create -f deploy/role_binding.yaml
  fi

  # Create SubM Operator deployment if it doesn't exist
  if ! kubectl get deployments | grep submariner-operator; then
    kubectl create --namespace=$subm_ns -f deploy/operator.yaml
  fi

  popd

  # Wait for SubM Operator pod to be ready
  kubectl get pods --namespace=$subm_ns
  kubectl wait --for=condition=Ready pods -l name=submariner-operator --timeout=60s --namespace=$subm_ns
  kubectl get pods --namespace=$subm_ns
}

function collect_subm_vars() {
  # Accept cluster context to deploy SubM Operator into as param
  context=$1
  kubectl config use-context $context

  clusterCidr=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' $context-worker | head -n 1)/32
  serviceCidr=10.96.0.0/16
  natEnabled=false
  subm_route_agent_image_repo=submariner-route-agent
  subm_route_agent_image_tag=local
  subm_route_agent_image_policy=IfNotPresent
  subm_engine_image_repo=submariner
  subm_engine_image_tag=local
  subm_engine_image_policy=IfNotPresent
}

function create_subm_cr() {
  # Accept cluster context to deploy SubM Operator into as param
  context=$1
  kubectl config use-context $context

  pushd $subm_op_dir

  # Create copy of default SubM CR (from operator-sdk, via Helm values.yaml)
  cr_file=deploy/crds/$context-submariner-cr.yaml
  cp deploy/crds/charts_v1alpha1_submariner_cr.yaml $cr_file

  # Modify the default CR to set correct variables for this context
  sed -i "s|psk: \"\"|psk: \"$SUBMARINER_PSK\"|g" $cr_file
  sed -i "s|server: example.k8s.apiserver|server: $SUBMARINER_BROKER_URL|g" $cr_file
  sed -i "s|token: test|token: $SUBMARINER_BROKER_TOKEN|g" $cr_file
  sed -i "s|namespace: xyz|namespace: $SUBMARINER_BROKER_NS|g" $cr_file
  sed -i "s|ca: \"\"|ca: $SUBMARINER_BROKER_CA|g" $cr_file
  sed -i "s|clusterId: \"\"|clusterId: \"$context\"|g" $cr_file
  sed -i "s|clusterCidr: \"10.42.0.0/16\"|clusterCidr: \"$clusterCidr\"|g" $cr_file
  sed -i "s|serviceCidr: \"10.43.0.0/16\"|serviceCidr: \"$serviceCidr\"|g" $cr_file
  sed -i "s|natEnabled: false|natEnabled: $natEnabled|g" $cr_file
  sed -i "s|repository: rancher/submariner-route-agent|repository: $subm_route_agent_image_repo|g" $cr_file
  sed -i "s|tag: v0.0.1|tag: $subm_route_agent_image_tag|g" $cr_file
  sed -i "s|pullPolicy: Always|pullPolicy: $subm_route_agent_image_policy|g" $cr_file
  sed -i "s|repository: rancher/submariner|repository: $subm_engine_image_repo|g" $cr_file
  sed -i "s|tag: v0.0.1|tag: $subm_engine_image_tag|g" $cr_file
  sed -i "s|pullPolicy: Always|pullPolicy: $subm_engine_image_policy|g" $cr_file

  # Show completed CR file for debugging help
  cat $cr_file

  popd
}

function deploy_subm_cr() {
  # Accept cluster context to deploy SubM Operator into as param
  context=$1
  kubectl config use-context $context

  pushd $subm_op_dir

  cr_file=deploy/crds/$context-submariner-cr.yaml

  # Create SubM CR if it doesn't exist
  if kubectl get submariner 2>&1 | grep -q "No resources found"; then
    kubectl apply --namespace=$subm_ns -f $cr_file
  fi

  popd
}
