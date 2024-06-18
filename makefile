.PHONY: build
build:
	go build -o output/bin/jobctl cmd/jobctl/main/main.go

.PHONY: gen
gen:
	go get -u ./...
	go mod tidy
	go generate ./...
	go fmt ./...

# grpc
PROTOCIN=./proto
PROTOCOUT=./pkg/pb
.PHONY: proto
proto:
	rm -rf ${PROTOCOUT}
	mkdir -p ${PROTOCOUT}
	protoc \
    	-I=${PROTOCIN} \
    	--proto_path=${PROTOCIN} \
    	--go_opt=paths=source_relative \
    	--go_out=${PROTOCOUT} \
    	--go-grpc_opt=paths=source_relative \
    	--go-grpc_out=${PROTOCOUT} \
    	$(shell find ${PROTOCIN} -iname "*.proto")

REGISTRY=njpowell
REPO=${REGISTRY}/jobber
.PHONY: image
image:
	docker build -t ${REPO} .

.PHONY: run
run:
	docker run --rm -it --privileged --cgroupns=private ${REPO} 


### TEST

.PHONY: test
test:
	go test -v ./...

.PHONY: image-test
image-test:
	docker build -f Dockerfile.test -t ${REPO}:test .
	docker run --rm -it --privileged --cgroupns=private ${REPO}:test

# HACK start the testing process in a cgroup leaf
# otherwise the testing process is not moved and prevents
# controls from being added
.PHONY: test-script
test-script:
	mkdir /sys/fs/cgroup/test-script
	echo 1 > /sys/fs/cgroup/test-script/cgroup.procs
	-make test
	sh