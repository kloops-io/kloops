
# Image URL to use all building/pushing image targets
IMG ?= eddycharly/kloops
TAG ?= latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: chatbot dashboard

# Run tests
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out

chatbot: generate fmt vet
	go build -o bin/chatbot cmd/chatbot/main.go

dashboard: generate fmt vet
	go build -o bin/chatbot cmd/dashboard/main.go

dashboard-front:
	cd dashboard && CI=true npm install && npm run build

chatbot-linux: generate fmt vet
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o bin/chatbot cmd/chatbot/main.go

dashboard-linux: generate fmt vet
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o bin/dashboard cmd/dashboard/main.go

run-chatbot: generate fmt vet manifests
	go run ./main.go

run-dashboard: generate fmt vet manifests
	go run ./cmd/dashboard/main.go --namespace tools

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	kustomize build config/crd | kubectl delete -f -

# # Deploy controller in the configured Kubernetes cluster in ~/.kube/config
# deploy: manifests
# 	cd config/manager && kustomize edit set image controller=${IMG}
# 	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

docker-chatbot-build: chatbot-linux
	docker build . -t ${IMG}-chatbot:${TAG} -f .docker/Dockerfile.chatbot

docker-chatbot-push: docker-chatbot-build
	docker push ${IMG}-chatbot:${TAG}

docker-dashboard-build: dashboard-linux dashboard-front
	docker build . -t ${IMG}-dashboard:${TAG} -f .docker/Dockerfile.dashboard

docker-dashboard-push: docker-dashboard-build
	docker push ${IMG}-dashboard:${TAG}

docker-build: docker-chatbot-build docker-dashboard-build

docker-push: docker-chatbot-push docker-dashboard-push

e2e-tests:
	E2E_GIT_SERVER=http://gitea.127.0.0.1.nip.io \
	E2E_KLOOPS_URL=http://kloops-chatbot.tools.svc.cluster.local/hook \
	go test -v ./test/e2e/...

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

gen-docs-commands:
	go run ./scripts/gen-docs/commands/main.go > ./docs/plugins/COMMANDS.md

gen-docs-plugins:
	go run ./scripts/gen-docs/plugins/main.go > ./docs/plugins/PLUGINS.md

gen-docs-dashboard:
	go run ./scripts/gen-docs/dashboard/main.go > ./docs/dev/DASHBOARD-BACKEND.md

gen-docs: gen-docs-commands gen-docs-plugins gen-docs-dashboard
