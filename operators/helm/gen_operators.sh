set -ex

# NB: There must be a running K8s cluster pointed at by the exported KUBECONFIG
# for operator-sdk to work (although this dependency doesn't make sense)
export KUBECONFIG=/home/daniel/.kube/config
if ! minikube status | grep Running; then
  minikube start
fi
kubectl config use-context minikube

version=0.8.1-1

if [ ! -d ./$version ]; then
  mkdir $version
fi
pushd $version

function create_subm_operator() {
  if [ -d ./submariner-operator ]; then
    rm -rf ./submariner-operator
  fi
  operator-sdk new submariner-operator --type=helm --verbose --helm-chart=submariner-latest/submariner --skip-git-init
  pushd submariner-operator
  sed -i 's|REPLACE_IMAGE|docker.io/dfarrell07/submariner-helm-operator:test|g' deploy/operator.yaml
  sed -i 's|REPLACE_NAMESPACE|submariner|g' deploy/role_binding.yaml
cat <<EOF >> deploy/role.yaml
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get
EOF
cat <<EOF > deploy/namespace.yaml
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

  operator-sdk build docker.io/dfarrell07/submariner-helm-operator:test
  docker push docker.io/dfarrell07/submariner-helm-operator:test
  popd
}

function create_subm_broker_operator() {
  if [ -d ./submariner-k8s-broker-operator ]; then
    rm -rf ./submariner-k8s-broker-operator
  fi
  operator-sdk new submariner-k8s-broker-operator --type=helm --verbose --helm-chart=submariner-latest/submariner-k8s-broker --skip-git-init
  pushd submariner-k8s-broker-operator
  sed -i 's|REPLACE_IMAGE|dfarrell07/submariner-broker-helm-operator:test|g' deploy/operator.yaml
  sed -i 's|REPLACE_NAMESPACE|submariner-k8s-broker|g' deploy/role_binding.yaml
  # Add RBAC permissions to do 'get pods' commands
cat <<EOF >> deploy/role.yaml
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get
EOF
  # Create a namespace for the SubM Broker
  # FIXME: Should we do this? Maybe it should run in default?
cat <<EOF > deploy/namespace.yaml
{
  "apiVersion": "v1",
  "kind": "Namespace",
  "metadata": {
    "name": "submariner-k8s-broker",
    "labels": {
      "name": "submariner-k8s-broker"
    }
  }
}
EOF
  operator-sdk build docker.io/dfarrell07/submariner-broker-helm-operator:test
  docker push docker.io/dfarrell07/submariner-broker-helm-operator:test
  popd
}

function create_nginx_operator() {
  if [ -d ./nginx-operator ]; then
    #echo "WARNING: About to delete $PWD/nginx-operator! 3 seconds..."
    #sleep 3
    rm -rf ./nginx-operator
  fi
  operator-sdk new nginx-operator --api-version=example.com/v1alpha1 --kind=Nginx --type=helm --skip-git-init
  pushd nginx-operator
  sed -i 's|REPLACE_IMAGE|dfarrell07/nginx-operator:test|g' deploy/operator.yaml
  operator-sdk build docker.io/dfarrell07/nginx-operator:test
  docker push docker.io/dfarrell07/nginx-operator:test
  popd
}

function deploy_nginx_operator() {
  pushd nginx-operator
  # Deploy Nginx Operator
  kubectl create -f deploy/crds/example_v1alpha1_nginx_crd.yaml
  kubectl create -f deploy/service_account.yaml
  kubectl create -f deploy/role.yaml
  kubectl create -f deploy/role_binding.yaml
  kubectl create -f deploy/operator.yaml
}

function deploy_nginx() {
  # Create Nginx CR, causing Operator to create Nginx pod
  kubectl apply -f deploy/crds/example_v1alpha1_nginx_cr.yaml
  # TODO: Does this wait actually wait?
  kubectl wait --for=condition=Ready pods -l app=nginx --timeout=60s
  kubectl get pods
  sleep 5
  popd
}

function verify_nginx() {
  deployment_name=$(kubectl get deployments | grep example-nginx | awk '{print $1}')
  kubectl expose deployment $deployment_name --type=LoadBalancer --name=nginx-service
  sleep 5
  minikube service nginx-service
  nginx_url=$(minikube service list | grep nginx-service | cut -d "|" -f4 | xargs)
  curl $nginx_url
}

function cleanup_nginx_operator() {
  pushd nginx-operator
  kubectl delete -f deploy/crds/example_v1alpha1_nginx_cr.yaml
  kubectl delete -f deploy/operator.yaml
  kubectl delete -f deploy/role_binding.yaml
  kubectl delete -f deploy/role.yaml
  kubectl delete -f deploy/service_account.yaml
  kubectl delete -f deploy/crds/example_v1alpha1_nginx_crd.yaml
  kubectl delete service nginx-service
  popd
}

# Create SubM Operator
create_subm_operator

# Create SubM Broker Operator
create_subm_broker_operator

# Nginx work is just verification/experiment
# Create Nginx Operator
#create_nginx_operator

# Verify Nginx Operator
#deploy_nginx_operator
#deploy_nginx
#verify_nginx
#cleanup_nginx_operator
