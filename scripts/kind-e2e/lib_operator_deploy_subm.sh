set -ex

# FIXME: Pods route-agent and engine pods are still not deploying (for helm op)

#operator_helm=true
operator_go=true
also_engine=true
also_routeagent=true
openapi_checks_enabled=false

if [[ $operator_helm = true ]]; then
  subm_op_dir=../operators/helm/0.8.1-1/submariner-operator
  subm_ns=submariner
  subm_broker_ns=submariner-k8s-broker
fi
if [[ $operator_go = true ]]; then
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
  mkdir -p $subm_op_dir

  cp -r ../operators/go/submariner-operator/ $subm_op_dir/..

  #subm_ns=default
  subm_ns=submariner
  subm_broker_ns=submariner-k8s-broker

  export GO111MODULE=on
fi

function add_subm_gateway_label() {
  # Accept cluster context to deploy SubM Operator into as param
  context=$1
  kubectl config use-context $context

  kubectl label node $context-worker "submariner.io/gateway=true" --overwrite
}

function create_subm_clusters_crd() {
  # Accept cluster context as param
  context=$1
  kubectl config use-context $context

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
  if ! kubectl get crds | grep clusters.submariner.io; then
    kubectl create -f $clusters_crd_file
  fi

  popd
}

function create_subm_endpoints_crd() {
  # Accept cluster context as param
  context=$1
  kubectl config use-context $context

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
  if ! kubectl get crds | grep endpoints.submariner.io; then
    kubectl create -f $endpoints_crd_file
  fi

  popd
}

function create_routeagents_crd() {
  # Accept cluster context as param
  context=$1
  kubectl config use-context $context

  pushd $subm_op_dir

  routeagents_crd_file=deploy/crds/submariner_routeagents_crd.yaml

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
  if ! kubectl get crds | grep routeagents.submariner.io; then
    kubectl create -f $routeagents_crd_file
  fi

  popd
}

function deploy_subm_operator() {
  # Accept cluster context to deploy SubM Operator into as param
  context=$1
  kubectl config use-context $context

  pushd $subm_op_dir

  # If SubM namespace doesn't exist (ignore SubM Broker ns), create it
  if ! kubectl get ns | grep -v $subm_broker_ns | grep $subm_ns; then
    # TODO: Make this dynamically use any $subm_ns
    kubectl create -f deploy/namespace.yaml
  fi

  if [[ $operator_helm = true ]]; then
    # Create SubM Operator CRD if it doesn't exist
    if ! kubectl get crds | grep submariners.charts.helm.k8s.io; then
      kubectl create -f deploy/crds/charts_v1alpha1_submariner_crd.yaml
    fi
  fi
  if [[ $operator_go = true ]]; then
    if ! kubectl get crds | grep submariners.submariner.io; then
      kubectl create -f deploy/crds/submariner_v1alpha1_submariner_crd.yaml
    fi
  fi

  # Create SubM Operator service account if it doesn't exist
  if ! kubectl get sa --namespace=$subm_ns | grep submariner-operator; then
    kubectl create --namespace=$subm_ns -f deploy/service_account.yaml
  fi

  # TODO: Why are these different between go and helm operators?
  if [[ $operator_helm = true ]]; then
    # Create SubM Operator role if it doesn't exist
    if ! kubectl get clusterroles --namespace=$subm_ns | grep submariner-operator; then
      kubectl create --namespace=$subm_ns -f deploy/role.yaml
    fi
  fi
  if [[ $operator_go = true ]]; then
    # Create SubM Operator role if it doesn't exist
    if ! kubectl get roles --namespace=$subm_ns | grep submariner-operator; then
      kubectl create --namespace=$subm_ns -f deploy/role.yaml
    fi
  fi

  # TODO: Why are these different between go and helm operators?
  if [[ $operator_go = true ]]; then
    # Create SubM Operator role binding if it doesn't exist
    if ! kubectl get rolebindings --namespace=$subm_ns | grep submariner-operator; then
      kubectl create --namespace=$subm_ns -f deploy/role_binding.yaml
    fi
  fi
  if [[ $operator_helm = true ]]; then
    # Create SubM Operator role binding if it doesn't exist
    if ! kubectl get clusterrolebindings --namespace=$subm_ns | grep submariner-operator; then
      kubectl create --namespace=$subm_ns -f deploy/role_binding.yaml
    fi
  fi

  # Create SubM Operator deployment if it doesn't exist
  if ! kubectl get deployments --namespace=$subm_ns | grep submariner-operator; then
    kubectl create --namespace=$subm_ns -f deploy/operator.yaml
  fi

  popd

  # Wait for SubM Operator pod to be ready
  kubectl get pods --namespace=$subm_ns
  kubectl wait --for=condition=Ready pods -l name=submariner-operator --timeout=120s --namespace=$subm_ns
  kubectl get pods --namespace=$subm_ns
}

function collect_subm_vars() {
  # Accept cluster context to deploy SubM Operator into as param
  context=$1
  kubectl config use-context $context

  # FIXME A better name might be submariner-engine, but just kinda-matching submariner-<random hash> name used by Helm/upstream tests
  deployment_name=submariner
  engine_deployment_name=submariner-engine
  routeagent_deployment_name=submariner-routeagent

  clusterCidr=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' $context-worker | head -n 1)/32
  # FIXME: This non-cluster-specific serviceCidr should be removed
  serviceCidr=10.96.0.0/16
  serviceCidr_cluster2=100.95.0.0/16
  serviceCidr_cluster3=100.96.0.0/16
  natEnabled=false
  subm_routeagent_image_repo=submariner-route-agent
  subm_routeagent_image_tag=local
  subm_routeagent_image_policy=IfNotPresent
  subm_engine_image_repo=submariner
  subm_engine_image_tag=local
  subm_engine_image_policy=IfNotPresent
  subm_colorcodes=blue
  subm_debug=false
  subm_broker=k8s
  ce_ipsec_debug=false
}

# FIXME: Call this submariner-engine vs submariner?
function create_subm_cr() {
  # Accept cluster context as param
  context=$1
  kubectl config use-context $context

  pushd $subm_op_dir

  if [[ $operator_helm = true ]]; then
    # NB: Need to have CR file per-operator so they don't clobber each other
    cr_file_helm_base=deploy/crds/charts_v1alpha1_submariner_cr.yaml
    cr_file_helm=deploy/crds/helm-$context-submariner-cr.yaml

    # Create copy of default SubM CR (from operator-sdk, via Helm values.yaml)
    cp $cr_file_helm_base $cr_file_helm

    # Modify the default CR to set correct variables for this context
    sed -i "s|psk: \"\"|psk: \"$SUBMARINER_PSK\"|g" $cr_file_helm
    sed -i "s|server: example.k8s.apiserver|server: $SUBMARINER_BROKER_URL|g" $cr_file_helm
    sed -i "s|token: test|token: $SUBMARINER_BROKER_TOKEN|g" $cr_file_helm
    sed -i "s|namespace: xyz|namespace: $SUBMARINER_BROKER_NS|g" $cr_file_helm
    sed -i "s|ca: \"\"|ca: $SUBMARINER_BROKER_CA|g" $cr_file_helm
    sed -i "s|clusterId: \"\"|clusterId: \"$context\"|g" $cr_file_helm
    sed -i "s|clusterCidr: \"10.42.0.0/16\"|clusterCidr: \"$clusterCidr\"|g" $cr_file_helm
    sed -i "s|serviceCidr: \"10.43.0.0/16\"|serviceCidr: \"$serviceCidr\"|g" $cr_file_helm
    sed -i "s|natEnabled: false|natEnabled: $natEnabled|g" $cr_file_helm
    sed -i "s|repository: rancher/submariner-route-agent|repository: $subm_routeagent_image_repo|g" $cr_file_helm
    sed -i "s|tag: v0.0.1|tag: $subm_routeagent_image_tag|g" $cr_file_helm
    sed -i "s|pullPolicy: Always|pullPolicy: $subm_routeagent_image_policy|g" $cr_file_helm
    sed -i "s|repository: rancher/submariner|repository: $subm_engine_image_repo|g" $cr_file_helm
    sed -i "s|tag: v0.0.1|tag: $subm_engine_image_tag|g" $cr_file_helm
    sed -i "s|pullPolicy: Always|pullPolicy: $subm_engine_image_policy|g" $cr_file_helm

    cr_file=$cr_file_helm
  fi
  if [[ $operator_go = true ]]; then
    # NB: Need to have CR file per-operator so they don't clobber each other
    cr_file_go_base=deploy/crds/submariner_v1alpha1_submariner_cr.yaml
    cr_file_go=deploy/crds/go-$context-submariner-cr.yaml

    # Create copy of default SubM CR (from operator-sdk)
    cp $cr_file_go_base $cr_file_go

    # Show base CR file
    cat $cr_file_go

    # Verify CR file exists
    [ -f $cr_fil_go ]

    # TODO: Use $engine_deployment_name here?
    sed -i "s|name: example-submariner|name: $deployment_name|g" $cr_file_go

    sed -i "/spec:/a \ \ Count: 1" $cr_file_go

    # These all need to end up in pod container/environment vars
    sed -i "/spec:/a \ \ submariner_namespace: $subm_ns" $cr_file_go
    sed -i "/spec:/a \ \ submariner_clustercidr: $clusterCidr" $cr_file_go
    if [[ $context = cluster2 ]]; then
      sed -i "/spec:/a \ \ submariner_servicecidr: $serviceCidr_cluster2" $cr_file_go
    elif [[ $context = cluster3 ]]; then
      sed -i "/spec:/a \ \ submariner_servicecidr: $serviceCidr_cluster3" $cr_file_go
    fi
    # This should be stored in a secret volume
    sed -i "/spec:/a \ \ submariner_token: FIXME_FOO_SUB_TOKEN_TODO" $cr_file_go
    sed -i "/spec:/a \ \ submariner_clusterid: $context" $cr_file_go
    sed -i "/spec:/a \ \ submariner_colorcodes: $subm_colorcodes" $cr_file_go
    # NB: Quoting bool-like vars is required or Go will type as bool and fail when set as env vars as strs
    sed -i "/spec:/a \ \ submariner_debug: \"$subm_debug\"" $cr_file_go
    # NB: Quoting bool-like vars is required or Go will type as bool and fail when set as env vars as strs
    sed -i "/spec:/a \ \ submariner_natenabled: \"$natEnabled\"" $cr_file_go
    sed -i "/spec:/a \ \ submariner_broker: $subm_broker" $cr_file_go
    sed -i "/spec:/a \ \ broker_k8s_apiserver: $SUBMARINER_BROKER_URL" $cr_file_go
    sed -i "/spec:/a \ \ broker_k8s_apiservertoken: $SUBMARINER_BROKER_TOKEN" $cr_file_go
    sed -i "/spec:/a \ \ broker_k8s_remotenamespace: $SUBMARINER_BROKER_NS" $cr_file_go
    sed -i "/spec:/a \ \ broker_k8s_ca: $SUBMARINER_BROKER_CA" $cr_file_go
    sed -i "/spec:/a \ \ ce_ipsec_psk: $SUBMARINER_PSK" $cr_file_go
    # NB: Quoting bool-like vars is required or Go will type as bool and fail when set as env vars as strs
    sed -i "/spec:/a \ \ ce_ipsec_debug: \"$ce_ipsec_debug\"" $cr_file_go
    sed -i "/spec:/a \ \ image: $subm_engine_image_repo:$subm_engine_image_tag" $cr_file_go

    cr_file=$cr_file_go
  fi

  # Show completed CR file for debugging help
  cat $cr_file

  popd
}

function create_routeagent_cr() {
  # Accept cluster context as param
  context=$1
  kubectl config use-context $context

  pushd $subm_op_dir

  if [[ $operator_go = true ]]; then
    # NB: Need to have CR file per-operator so they don't clobber each other
    cr_file_go=deploy/crds/go-$context-routeagent-cr.yaml

    cp deploy/crds/submariner_v1alpha1_routeagent_cr.yaml $cr_file_go

    sed -i "s|name: example-routeagent|name: $routeagent_deployment_name|g" $cr_file_go

    # These all need to end up in pod container/environment vars
    sed -i "/spec:/a \ \ submariner_namespace: $subm_ns" $cr_file_go
    sed -i "/spec:/a \ \ submariner_clusterid: $context" $cr_file_go
    sed -i "/spec:/a \ \ submariner_debug: \"$subm_debug\"" $cr_file_go

    # These all need to end up in pod containers/submariner vars
    # FIXME: This seems like it should break, as should use engine image repo
    sed -i "/spec:/a \ \ image: $subm_routeagent_image_repo:$subm_routeagent_image_tag" $cr_file_go

    cr_file=$cr_file_go
  fi

  # Show completed CR file for debugging help
  cat $cr_file

  popd
}

function deploy_subm_cr() {
  # Accept cluster context as param
  context=$1
  kubectl config use-context $context

  pushd $subm_op_dir

  # FIXME: This must match cr_file value used in create_subm_cr fn
  if [[ $operator_helm = true ]]; then
    # NB: Need to have CR file per-operator so they don't clobber each other
    cr_file=deploy/crds/helm-$context-submariner-cr.yaml
  fi
  if [[ $operator_go = true ]]; then
    # NB: Need to have CR file per-operator so they don't clobber each other
    cr_file=deploy/crds/go-$context-submariner-cr.yaml
  fi

  # Create SubM CR if it doesn't exist
  if kubectl get submariner 2>&1 | grep -q "No resources found"; then
    kubectl apply --namespace=$subm_ns -f $cr_file
  fi

  popd
}

function deploy_routeagent_cr() {
  # Accept cluster context as param
  context=$1
  kubectl config use-context $context

  pushd $subm_op_dir

  # FIXME: This must match cr_file value used in create_routeagent_cr fn
  if [[ $operator_go = true ]]; then
    # NB: Need to have CR file per-operator so they don't clobber each other
    cr_file=deploy/crds/go-$context-routeagent-cr.yaml
  fi

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
