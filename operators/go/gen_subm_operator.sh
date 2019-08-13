#!/bin/bash
set -ex

# TODO: Integrate this into e2e tests
# TODO: Add size-scaling logic to controller
# TODO: Add deployment handling logic to controller
# TODO: Add logic to accept env vars from CR and set in pod spec to controller
# TODO: 
# TODO: 

version=0.0.1

# Work around https://github.com/operator-framework/operator-sdk/issues/1675
GOROOT="$(go env GOROOT)"
export GOROOT

function setup_prereqs(){
  if ! command -v dep; then
    # Install dep
    curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh                           

    # Make sure go/bin is in path
    command -v dep
  fi

  GOPATH=$HOME/go
  op_dir=$GOPATH/src/github.com/submariner-operator

  rm -rf $op_dir
  mkdir -p $op_dir
  #if ! [ -d $op_dir ]; then
  #  mkdir $op_dir
  #fi

  export GO111MODULE=on 

  # NB: There must be a running K8s cluster pointed at by the exported KUBECONFIG
  # for operator-sdk to work (although this dependency doesn't make sense)
  export KUBECONFIG=$HOME/.kube/config
  #minikube delete
  #minikube start
  if ! minikube status | grep Running; then
    minikube start
  fi
  kubectl config use-context minikube
}

function create_subm_operator() {
  pushd $op_dir
  operator-sdk new submariner-operator --verbose
  pushd submariner-operator
  api_version=submariner.io/v1alpha1
  kind=Submariner
  operator-sdk add api --api-version=$api_version --kind=$kind
  types_file=pkg/apis/submariner/v1alpha1/submariner_types.go
  sed -i '/SubmarinerSpec struct/a \ \ Count int32 `json:"size"`' $types_file
  sed -i '/SubmarinerSpec struct/a \ \ SubmarinerNamespace string `json:"submariner_namespace"`' $types_file
  sed -i '/SubmarinerSpec struct/a \ \ SubmarinerClustercidr string `json:"submariner_clustercidr"`' $types_file
  sed -i '/SubmarinerSpec struct/a \ \ SubmarinerServicecidr string `json:"submariner_servicecidr"`' $types_file
  sed -i '/SubmarinerSpec struct/a \ \ SubmarinerToken string `json:"submariner_token"`' $types_file
  sed -i '/SubmarinerSpec struct/a \ \ SubmarinerClusterid string `json:"submariner_clusterid"`' $types_file
  sed -i '/SubmarinerSpec struct/a \ \ SubmarinerColorcodes string `json:"submariner_colorcodes"`' $types_file
  sed -i '/SubmarinerSpec struct/a \ \ SubmarinerDebug string `json:"submariner_debug"`' $types_file
  sed -i '/SubmarinerSpec struct/a \ \ SubmarinerNatenabled string `json:"submariner_natenabled"`' $types_file
  sed -i '/SubmarinerSpec struct/a \ \ SubmarinerBroker string `json:"submariner_broker"`' $types_file
  sed -i '/SubmarinerSpec struct/a \ \ BrokerK8sApiserver string `json:"broker_k8s_apiserver"`' $types_file
  sed -i '/SubmarinerSpec struct/a \ \ BrokerK8sApiservertoken string `json:"broker_k8s_apiservertoken"`' $types_file
  sed -i '/SubmarinerSpec struct/a \ \ BrokerK8sRemotenamespace string `json:"broker_k8s_remotenamespace"`' $types_file
  sed -i '/SubmarinerSpec struct/a \ \ BrokerK8sCa string `json:"broker_k8s_ca"`' $types_file
  sed -i '/SubmarinerSpec struct/a \ \ CeIpsecPsk string `json:"ce_ipsec_psk"`' $types_file
  sed -i '/SubmarinerSpec struct/a \ \ CeIpsecDebug string `json:"ce_ipsec_debug"`' $types_file

  sed -i '/SubmarinerStatus struct/a \ \ PodNames []string `json:"nodes"`' $types_file

  operator-sdk generate k8s

  operator-sdk add controller --api-version=$api_version --kind=$kind

  controller_file_src=$GOPATH/src/github.com/submariner-io/submariner/operators/go/submariner_controller.go
  controller_file_dst=pkg/controller/submariner/submariner_controller.go

  cp $controller_file_src $controller_file_dst

  sed -i 's|REPLACE_NAMESPACE|submariner|g' deploy/role_binding.yaml

  # Create a namespace for SubM
  ns_file=deploy/namespace.yaml
cat <<EOF > $ns_file
{
  "apiVersion": "v1",
  "kind": "Namespace",
  "metadata": {
    "name": "submariner",
    "labels": {
      "name": "submariner"
    }
  }
}
EOF

  sed -i "s|REPLACE_IMAGE|quay.io/submariner/submariner-operator:$version|g" deploy/operator.yaml

  go mod vendor
  # This seems like a bug in operator-sdk, that this is needed?
  go get k8s.io/kube-state-metrics/pkg/collector
  go get k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1
  go get github.com/coreos/prometheus-operator/pkg/apis/monitoring

  operator-sdk build quay.io/submariner/submariner-operator:$version --verbose
  docker push quay.io/submariner/submariner-operator:$version
  popd
  popd
}

function export_subm_op() {
  cp -r $op_dir/submariner-operator $HOME/submariner/operators/go/
}

function deploy_subm_operator() {
  pushd $op_dir/submariner-operator
  kubectl create -f deploy/crds/submariner_v1alpha1_submariner_crd.yaml
  # TODO: Add wait-for here?
  kubectl get crd

  kubectl create -f deploy/service_account.yaml
  kubectl create -f deploy/role.yaml
  kubectl create -f deploy/role_binding.yaml
  kubectl create -f deploy/operator.yaml
  # TODO: Add a wait-for here?
  kubectl get deploy

  cr_file=deploy/crds/submariner_v1alpha1_submariner_cr.yaml

  # TODO: Likely need to pass more vars via CR (then to Pod via controller) here
  #sed -i '/spec:/a \ \ count: 3' $cr_file

  cat $cr_file

  kubectl apply -f $cr_file

  kubectl get submariner
  kubectl describe submariner example-submariner

  popd
}

function verify_subm() {
  kubectl expose deployment example-submariner --type=LoadBalancer --name=example-service
  minikube service example-service
  nginx_url=$(minikube service list | grep example-service | cut -d "|" -f4 | xargs)
  curl $nginx_url
}

# Make sure prereqs are installed
setup_prereqs

# Create SubM Operator
create_subm_operator

#deploy_subm_operator

#verify_subm

export_subm_op
