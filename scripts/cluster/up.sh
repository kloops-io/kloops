#!/bin/bash

set -e

NO_KIND=false
NO_GITEA_INIT=false
NO_KLOOPS_INIT=false
CLUSTER_NAME=kloops
HELM_BINARY=helm
GITEA_ADMIN_USER=gitea
GITEA_ADMIN_PASSWORD=admin
GITEA_USER=my-user
GITEA_PASSWORD=My-User#1
GITEA_ORG=my-org
GITEA_REPO=my-repo
GITEA_KLOOPS_USER=kloops-bot
GITEA_KLOOPS_PASSWORD=KLoopsBot#1
KLOOPS_HMAC=this-is-secret

function message() {
    echo "--------------------------------------------------------------------------------------------------------------"
    echo "$1"
    echo "--------------------------------------------------------------------------------------------------------------"
}

function verify_supported() {
  if [ "$NO_KIND" == "false" ] && ! type "kind" > /dev/null 2>&1; then
    echo "kind is required"
    exit 1
  fi
  if [ "$NO_INIT" == "false" ] && ! type "jq" > /dev/null 2>&1; then
    echo "jq is required"
    exit 1
  fi
  if ! type "kubectl" > /dev/null 2>&1; then
    echo "kubectl is required"
    exit 1
  fi
  if ! type "$HELM_BINARY" > /dev/null 2>&1; then
    echo "helm is required"
    exit 1
  fi
}

function create_kind_cluster() {
    message "Starting kind cluster $CLUSTER_NAME ..."
    kind create cluster --name $CLUSTER_NAME --config ./scripts/cluster/kind-config.yaml
}

function add_helm_repos() {
    message "Adding helm repositories ..."
    $HELM_BINARY repo add stable https://kubernetes-charts.storage.googleapis.com
    $HELM_BINARY repo add minio https://helm.min.io/
    $HELM_BINARY repo add banzaicloud-stable https://kubernetes-charts.banzaicloud.com
    $HELM_BINARY repo add gitea-charts https://dl.gitea.io/charts/
    $HELM_BINARY repo add kubernetes-dashboard https://kubernetes.github.io/dashboard
}

function update_helm_repos() {
    message "Updating helm repositories ..."
    $HELM_BINARY repo update
}

function deploy_nginx_ingress() {
    message "Deploying NGINX ingress controller ..."
    kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-0.32.0/deploy/static/provider/kind/deploy.yaml
    sleep 30s
    kubectl wait -n ingress-nginx --for=condition=ready pod --selector=app.kubernetes.io/component=controller --timeout=120s
}

function deploy_metrics_server() {
    message "Deploying metrics server ..."
    $HELM_BINARY upgrade --install --wait --create-namespace --namespace tools metrics-server stable/metrics-server --values ./scripts/cluster/metrics-server-config.yaml
}

function deploy_kubernetes_dashboard() {
    message "Deploying Kubernetes dashboard ..."
    $HELM_BINARY upgrade --install --wait --create-namespace --namespace tools dashboard kubernetes-dashboard/kubernetes-dashboard --values ./scripts/cluster/kubernetes-dashboard-config.yaml
}

function deploy_minio() {
    message "Deploying minio storage ..."
    $HELM_BINARY upgrade --install --version 6.3.1 --wait --create-namespace --namespace tools minio minio/minio --values ./scripts/cluster/minio-config.yaml
}

function deploy_logging_operator() {
    message "Deploying logging operator ..."
    $HELM_BINARY upgrade --install --version 3.6.0 --wait --create-namespace --namespace tools logging-operator banzaicloud-stable/logging-operator --set createCustomResource=false
}

function deploy_logging_pipeline() {
    message "Deploying logging pipeline ..."
    kubectl apply -n tools -f ./scripts/cluster/logging-pipeline.yaml
}

function deploy_logs_server() {
    message "Deploying logs server ..."
    kubectl apply -n tools -f ./scripts/cluster/logs-server.yaml
}

function deploy_tekton_pipelines() {
    message "Deploying Tekton pipelines ..."
    kubectl apply -f https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml
}

function deploy_tekton_dashboard() {
    message "Deploying Tekton dashboard ..."
    curl -sL https://raw.githubusercontent.com/tektoncd/dashboard/master/scripts/release-installer | \
        bash -s -- install latest --ingress-url tekton-dashboard.127.0.0.1.nip.io
    kubectl patch deployment tekton-dashboard -n tekton-pipelines --type='json' \
        --patch='[{"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--external-logs=http://logs-server.tools.svc.cluster.local:3000/logs"}]'
}

function deploy_gitea() {
    message "Deploying Gitea ..."
    $HELM_BINARY upgrade --install --namespace tools --create-namespace --wait gitea gitea-charts/gitea --values ./scripts/cluster/gitea-config.yaml \
        --set gitea.admin.username=$GITEA_ADMIN_USER --set gitea.admin.password=$GITEA_ADMIN_PASSWORD
}

function init_gitea() {
    message "Initializing Gitea ..."
    sleep 30s
    # TODO: make better use of env vars
    # create user
    curl -s --output /dev/null -X POST "http://$GITEA_ADMIN_USER:$GITEA_ADMIN_PASSWORD@gitea.127.0.0.1.nip.io/api/v1/admin/users" \
        -H "Content-Type: application/json" --data '{ "email": "my-user@local.com", "username": "my-user", "password": "My-User#1" }'
    # create kloops-bot
    curl -s --output /dev/null -X POST "http://$GITEA_ADMIN_USER:$GITEA_ADMIN_PASSWORD@gitea.127.0.0.1.nip.io/api/v1/admin/users" \
        -H "Content-Type: application/json" --data '{ "email": "kloops-bot@local.com", "username": "kloops-bot", "password": "KLoopsBot#1" }'
    # create org
    curl -s --output /dev/null -X POST "http://$GITEA_ADMIN_USER:$GITEA_ADMIN_PASSWORD@gitea.127.0.0.1.nip.io/api/v1/admin/users/$GITEA_USER/orgs" \
        -H "Content-Type: application/json" --data '{ "username": "my-org" }'
    # add kloops-bot in org
    curl -s --output /dev/null -X PUT "http://$GITEA_ADMIN_USER:$GITEA_ADMIN_PASSWORD@gitea.127.0.0.1.nip.io/api/v1/teams/2/members/$GITEA_KLOOPS_USER"
    # create repo
    curl -s --output /dev/null -X POST "http://$GITEA_ADMIN_USER:$GITEA_ADMIN_PASSWORD@gitea.127.0.0.1.nip.io/api/v1/admin/users/$GITEA_ORG/repos" \
        -H "Content-Type: application/json" --data '{ "default_branch": "master", "name": "my-repo" }'
    # create kloops-bot token
    KLOOPS_TOKEN=$(curl -s -X POST "http://$GITEA_ADMIN_USER:$GITEA_ADMIN_PASSWORD@gitea.127.0.0.1.nip.io/api/v1/users/$GITEA_KLOOPS_USER/tokens" -H "Content-Type: application/json" --data '{ "name": "kloops" }' | jq -r '.sha1')
}

function deploy_kloops() {
    message "Deploying KLoops ..."
    $HELM_BINARY upgrade --install --namespace tools --create-namespace --wait kloops ./charts/kloops --values ./scripts/cluster/kloops-config.yaml
}

function init_kloops() {
    message "Initializing KLoops ..."
    kubectl apply -n tools -f - <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: gitea-kloops-bot
type: Opaque
stringData:
  hmac: $KLOOPS_HMAC
  token: $KLOOPS_TOKEN
EOF
    kubectl apply -n tools -f ./scripts/cluster/kloops-default-pluginconfig.yaml
    kubectl apply -n tools -f - <<EOF
apiVersion: config.kloops.io/v1alpha1
kind: RepoConfig
metadata:
  name: gitea-$GITEA_ORG-$GITEA_REPO
spec:
  gitea:
    owner: $GITEA_ORG
    repo: $GITEA_REPO
    server: http://gitea-http.tools.svc.cluster.local:3000
    token:
      valueFrom:
        secretKeyRef:
          name: gitea-kloops-bot
          key: token
    hmacToken:
      valueFrom:
        secretKeyRef:
          name: gitea-kloops-bot
          key: hmac
  autoMerge:
    batchSizeLimit: 5
    mergeType: squash
    labels:
      - lgtm
      - approve
    missingLabels:
      - ok-to-test
    reviewApprovedRequired: true
  pluginConfig:
    ref: default
    plugins:
      - assign
      - branchcleaner
      - cat
      - dog
      - goose
      - pony
      - shrug
      - stage
      - welcome
      - wip
      - yuks
EOF
}

function cluster_up() {
    message "Cluster up !"
    # echo "Kubernetes dashboard URL   : http://k8s-dashboard.127.0.0.1.nip.io"
    echo "Minio URL                  : http://minio.127.0.0.1.nip.io:12080"
    echo "Logs server URL            : http://logs.127.0.0.1.nip.io:12080"
    echo "Tekton dashboard URL       : http://tekton-dashboard.127.0.0.1.nip.io:12080"
    echo "KLoops hooks URL           : http://kloops-hooks.127.0.0.1.nip.io:12080"
    echo "KLoops dashboard URL       : http://kloops-dashboard.127.0.0.1.nip.io:12080"
    echo "Gitea                      : http://gitea.127.0.0.1.nip.io:12080"
    echo "Gitea admin user/password  : $GITEA_ADMIN_USER/$GITEA_ADMIN_PASSWORD"
    if [ "$NO_GITEA_INIT" == "false" ]; then
    echo "Gitea org                  : $GITEA_ORG"
    echo "Gitea repo                 : $GITEA_REPO"
    echo "Gitea user/password        : $GITEA_USER/$GITEA_PASSWORD"
    echo "Gitea kloops user/password : $GITEA_KLOOPS_USER/$GITEA_KLOOPS_PASSWORD"
    echo "Gitea kloops bot token     : $KLOOPS_TOKEN"
    echo "KLoops HMAC                : $KLOOPS_HMAC"
    fi
}

set -u

while [[ $# -gt 0 ]]; do
    case $1 in
        '--no-kind')
            NO_KIND="true"
            ;;
        '--no-init')
            NO_GITEA_INIT="true"
            NO_KLOOPS_INIT="true"
            ;;
        '--name')
            shift
            CLUSTER_NAME="${1}"
            ;;
        '--helm')
            shift
            HELM_BINARY="${1}"
            ;;
        *)
            echo "ERROR: Unknown option $1"
            exit 1
            ;;
    esac
    shift
done

set +u

verify_supported
if [ "$NO_KIND" == "false" ]; then
    create_kind_cluster
fi
add_helm_repos
update_helm_repos
deploy_nginx_ingress
# deploy_metrics_server
# deploy_kubernetes_dashboard
deploy_minio
deploy_logging_operator
deploy_logging_pipeline
deploy_logs_server
deploy_tekton_pipelines
# deploy_tekton_dashboard
deploy_gitea
if [ "$NO_GITEA_INIT" == "false" ]; then
    init_gitea
fi
# deploy_kloops
if [ "$NO_KLOOPS_INIT" == "false" ]; then
    init_kloops
fi
cluster_up
