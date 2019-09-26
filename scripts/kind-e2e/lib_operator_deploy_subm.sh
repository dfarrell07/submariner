# This should only be sourced
if [ "${0##*/}" = "lib_operator_deploy_subm.sh" ]; then
    echo "Don't run me, source me" >&2
    exit 1
fi

openapi_checks_enabled=false

# FIXME: Extract these into a setup prereqs function
if ! command -v go; then
  curl https://dl.google.com/go/go1.12.7.linux-amd64.tar.gz -o go.tar.gz
  tar -xf go.tar.gz
  cp go /usr/local/bin/go
fi

if ! command -v dep; then
  # Install dep
  curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

  # Make sure go/bin is in path
  command -v dep
fi

GOPATH=$HOME/go
subm_op_dir=$GOPATH/src/github.com/submariner-operator/submariner-operator
subm_op_src_dir=../operators/go/submariner-operator
mkdir -p $subm_op_dir

cp -a $subm_op_src_dir/. $subm_op_dir/

export GO111MODULE=on

function create_resource_if_missing() {
  resource_type=$1
  resource_name=$2
  resource_yaml=$3
  if ! kubectl get --namespace=$subm_ns $resource_type $resource_name; then
    kubectl create --namespace=$subm_ns -f $resource_yaml
  fi
}

function add_subm_gateway_label() {
  kubectl label node $context-worker "submariner.io/gateway=true" --overwrite
}

function create_subm_clusters_crd() {
  pushd $subm_op_dir

  clusters_crd_file=deploy/crds/submariner_clusters_crd.yaml

  # TODO: Can/should we create this with Op-SDK?
cat <<EOF > $clusters_crd_file
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: clusters.submariner.io
spec:
  group: submariner.io
  version: v1
  names:
    kind: Cluster
    plural: clusters
  scope: Namespaced
EOF

  cat $clusters_crd_file

  # Create clusters CRD
  # NB: This must be done before submariner-engine pod is deployed
  create_resource_if_missing crd clusters.submariner.io $clusters_crd_file

  popd
}

function create_subm_endpoints_crd() {
  pushd $subm_op_dir

  endpoints_crd_file=deploy/crds/submariner_endpoints_crd.yaml

  # TODO: Can/should we create this with Op-SDK?
cat <<EOF > $endpoints_crd_file
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: endpoints.submariner.io
  annotations:
spec:
  group: submariner.io
  version: v1
  names:
    kind: Endpoint
    plural: endpoints
  scope: Namespaced
EOF

  cat $endpoints_crd_file

  # Create endpoints CRD
  # NB: This must be done before submariner-engine pod is deployed
  create_resource_if_missing crd endpoints.submariner.io $endpoints_crd_file

  popd
}

function create_routeagents_crd() {
  pushd $subm_op_dir

  routeagents_crd_file=deploy/crds/submariner.io_routeagents_crd.yaml

  # TODO: Can/should we create this with Op-SDK?
cat <<EOF > $routeagents_crd_file
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: routeagents.submariner.io
  annotations:
spec:
  group: submariner.io
  version: v1alpha1
  names:
    kind: Routeagent
    plural: routeagents
  scope: Namespaced
EOF

  cat $routeagents_crd_file

  # Create routeagents CRD
  create_resource_if_missing crd routeagents.submariner.io $routeagents_crd_file

  popd
}

function deploy_subm_operator() {
  pushd $subm_op_dir

  # If SubM namespace doesn't exist, create it
  if ! kubectl get ns $subm_ns; then
    # Customize namespace definition for subm_ns defined here
    cat deploy/namespace.yaml
    sed -i "s|submariner|$subm_ns|g" deploy/namespace.yaml
    cat deploy/namespace.yaml

    kubectl create -f deploy/namespace.yaml
  fi

  if ! kubectl get crds | grep submariners.submariner.io; then
    kubectl create -f deploy/crds/submariner.io_submariners_crd.yaml
  fi

  # Create SubM Operator service account if it doesn't exist
  create_resource_if_missing sa submariner-operator deploy/service_account.yaml

  # Create SubM Operator role if it doesn't exist
  create_resource_if_missing role submariner-operator deploy/role.yaml

  # Create SubM Operator role binding if it doesn't exist
  create_resource_if_missing rolebinding submariner-operator deploy/role_binding.yaml

  # Create SubM Operator deployment if it doesn't exist
  create_resource_if_missing deployment submariner-operator deploy/operator.yaml

  # Wait for SubM Operator pod to be ready
  kubectl wait --for=condition=Ready pods -l name=submariner-operator --timeout=120s --namespace=$subm_ns

  popd
}

# FIXME: Call this submariner-engine vs submariner?
function create_subm_cr() {
  pushd $subm_op_dir

  cr_file_base=deploy/crds/submariner.io_v1alpha1_submariner_cr.yaml
  cr_file=deploy/crds/submariner-cr-$context.yaml

  # Create copy of default SubM CR (from operator-sdk)
  cp $cr_file_base $cr_file

  # Show base CR file
  cat $cr_file

  # Verify CR file exists
  [ -f $cr_fil_go ]

  # TODO: Use $engine_deployment_name here?
  sed -i "s|name: example-submariner|name: $deployment_name|g" $cr_file

  sed -i "/spec:/a \ \ size: $subm_engine_size" $cr_file

  # These all need to end up in pod container/environment vars
  sed -i "/spec:/a \ \ submariner_namespace: $subm_ns" $cr_file
  if [[ $context = cluster2 ]]; then
    sed -i "/spec:/a \ \ submariner_servicecidr: $serviceCidr_cluster2" $cr_file
    sed -i "/spec:/a \ \ submariner_clustercidr: $clusterCidr_cluster2" $cr_file
  elif [[ $context = cluster3 ]]; then
    sed -i "/spec:/a \ \ submariner_servicecidr: $serviceCidr_cluster3" $cr_file
    sed -i "/spec:/a \ \ submariner_clustercidr: $clusterCidr_cluster3" $cr_file
  fi
  sed -i "/spec:/a \ \ submariner_token: $subm_token" $cr_file
  sed -i "/spec:/a \ \ submariner_clusterid: $context" $cr_file
  sed -i "/spec:/a \ \ submariner_colorcodes: $subm_colorcodes" $cr_file
  # NB: Quoting bool-like vars is required or Go will type as bool and fail when set as env vars as strs
  sed -i "/spec:/a \ \ submariner_debug: \"$subm_debug\"" $cr_file
  # NB: Quoting bool-like vars is required or Go will type as bool and fail when set as env vars as strs
  sed -i "/spec:/a \ \ submariner_natenabled: \"$natEnabled\"" $cr_file
  sed -i "/spec:/a \ \ submariner_broker: $subm_broker" $cr_file
  sed -i "/spec:/a \ \ broker_k8s_apiserver: $SUBMARINER_BROKER_URL" $cr_file
  sed -i "/spec:/a \ \ broker_k8s_apiservertoken: $SUBMARINER_BROKER_TOKEN" $cr_file
  sed -i "/spec:/a \ \ broker_k8s_remotenamespace: $SUBMARINER_BROKER_NS" $cr_file
  sed -i "/spec:/a \ \ broker_k8s_ca: $SUBMARINER_BROKER_CA" $cr_file
  sed -i "/spec:/a \ \ ce_ipsec_psk: $SUBMARINER_PSK" $cr_file
  # NB: Quoting bool-like vars is required or Go will type as bool and fail when set as env vars as strs
  sed -i "/spec:/a \ \ ce_ipsec_debug: \"$ce_ipsec_debug\"" $cr_file
  sed -i "/spec:/a \ \ image: $subm_engine_image_repo:$subm_engine_image_tag" $cr_file

  # Show completed CR file for debugging help
  cat $cr_file

  popd
}

function create_routeagent_cr() {
  pushd $subm_op_dir

  cr_file=deploy/crds/routeagent-cr-$context.yaml

  cp deploy/crds/submariner.io_v1alpha1_routeagent_cr.yaml $cr_file

  sed -i "s|name: example-routeagent|name: $routeagent_deployment_name|g" $cr_file

  # These all need to end up in pod container/environment vars
  sed -i "/spec:/a \ \ submariner_namespace: $subm_ns" $cr_file
  sed -i "/spec:/a \ \ submariner_clusterid: $context" $cr_file
  sed -i "/spec:/a \ \ submariner_debug: \"$subm_debug\"" $cr_file

  # These all need to end up in pod containers/submariner vars
  sed -i "/spec:/a \ \ image: $subm_routeagent_image_repo:$subm_routeagent_image_tag" $cr_file

  # Show completed CR file for debugging help
  cat $cr_file

  popd
}

function deploy_subm_cr() {
  pushd $subm_op_dir

  # FIXME: This must match cr_file value used in create_subm_cr fn
  cr_file=deploy/crds/submariner-cr-$context.yaml

  # Create SubM CR if it doesn't exist
  if kubectl get submariner 2>&1 | grep -q "No resources found"; then
    kubectl apply --namespace=$subm_ns -f $cr_file
  fi

  popd
}

function deploy_routeagent_cr() {
  pushd $subm_op_dir

  # FIXME: This must match cr_file value used in create_routeagent_cr fn
  cr_file=deploy/crds/routeagent-cr-$context.yaml

  # Create SubM CR if it doesn't exist
  if kubectl get routeagent 2>&1 | grep -q "No resources found"; then
    kubectl apply --namespace=$subm_ns -f $cr_file
  fi

  popd
}

function deploy_netshoot_cluster2() {
    kubectl config use-context cluster2
    echo Deploying netshoot on cluster2 worker: ${worker_ip}
    kubectl apply -f ./kind-e2e/netshoot.yaml
    echo Waiting for netshoot pods to be Ready on cluster2.
    kubectl rollout status deploy/netshoot --timeout=120s

    # TODO: Add verifications
}

function deploy_nginx_cluster3() {
    kubectl config use-context cluster3
    echo Deploying nginx on cluster3 worker: ${worker_ip}
    kubectl apply -f ./kind-e2e/nginx-demo.yaml
    echo Waiting for nginx-demo deployment to be Ready on cluster3.
    kubectl rollout status deploy/nginx-demo --timeout=120s

    # TODO: Add verifications
    # TODO: Do this with nginx operator?
}
