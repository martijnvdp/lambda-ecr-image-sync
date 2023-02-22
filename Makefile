IMAGE := ccv-hosting/ecr-image-sync
TAG := $(shell git describe --abbrev=0 --tags)
SHELL := /bin/bash
export PATH := /home/ubuntu/.local/bin:/home/ubuntu/go/bin:$(PATH)

.PHONY: bench
bench:
	go test -run ^$$ -bench . ./...

.PHONY: cyclo
cyclo:
	which gocyclo || go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	gocyclo -over 15 -ignore 'external' .

.PHONY: fix-typos
fix-typos:
	which codespell || pip install codespell
	codespell -S .git,.terraform --ignore-words .codespellignore -f -w -i1

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: image
image:
	go build ./cmd/main.go
	which docker-slim || sudo curl https://downloads.dockerslim.com/releases/1.37.2/dist_linux.tar.gz -o ./dist_linux.tar.gz && sudo tar xf ./dist_linux.tar.gz -C . && sudo mv dist_linux/* /usr/bin && sudo chmod +x /usr/bin/docker-slim && sudo chmod +x /usr/bin/docker-slim-sensor && rm dist_linux.tar.gz
	docker-slim build --tag $(IMAGE):${TAG} --tag $(IMAGE):latest --include-cert-all --http-probe=false --continue-after 60 --dockerfile ./Dockerfile --dockerfile-context . --delete-generated-fat-image

.PHONY: init
init:
	which gcc || sudo apt-get install gcc --fix-missing 
	sudo apt install libdevmapper-dev 
	sudo apt install libbtrfs-dev
	sudo apt install pkg-config

.PHONY: make-tests
make-tests:
	which gotests ||go install github.com/cweill/gotests/gotests@latest
	gotests --all pkg/lambda

.PHONY: pre-pr
pre-pr: quality typos test fmt

.PHONY: quality
quality: cyclo vet

.PHONY: tagger
tagger:
	@git checkout master
	@git fetch --tags
	@echo "the most recent tag was `git describe --tags --abbrev=0`"
	@echo ""
	read -p "Tag number: " TAG; \
	 git tag -a "$${TAG}" -m "$${TAG}"; \
	 git push origin "$${TAG}"

.PHONY: test
test:
	which gotestsum || (pushd /tmp && go install gotest.tools/gotestsum@latest && popd)
	gotestsum --format testname -- --mod=readonly -bench=^$$ -race ./...

.PHONY: typos
typos:
	which codespell || pip install codespell
	codespell -S .terraform,.git --ignore-words .codespellignore -f

.PHONY: vet
vet:
	go vet ./...
