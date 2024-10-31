IMAGE_BASE = jayden21/
VERSION ?= $(shell git symbolic-ref --short -q HEAD)
REGISTRY = 
TEST_REPORT_UPLOAD_DIR = ftp://ftp.dev.com/ftp/unit_test_report/servers/sev/${REF}/

IMAGE_REPOSITORY = docker.dev.datacloak.com
IMAGE_BASE = $(IMAGE_REPOSITORY)/datacloak/$(PLATFORM)-$(ARCH)
IMAGE_NAME = account:$(VERSION)

#REF ?= $(shell git symbolic-ref --short -q HEAD)-$(shell git rev-parse --short HEAD)
REF ?= $(shell git symbolic-ref --short -q HEAD)

GOOS ?= $(shell go env GOOS )
GOARCH ?= $(shell go env GOARCH )
export MEILIDB_MODE ?= dev

ifeq ($(MEILIDB_MODE), prod)
	export CGO_ENABLED = 1
	GO_BUILD_TAGS = -tags=$(MEILIDB_MODE)
else
	export CGO_ENABLED = 0
endif

PROJECT := ybaccount
OUTPUT := output
TEST_OUT_FILE := ${OUTPUT}/test.log
COVERAGE_OUT_FILE := ${OUTPUT}/cover.out
TEST_REPORT := ${OUTPUT}/${PROJECT}-junit-report.xml
COVERAGE_REPORT := ${OUTPUT}/${PROJECT}-coverage



build: env
	@echo "start build ${PROJECT}"
	GOOS=${GOOS} GOARCH=${GOARCH} go build -trimpath ${GO_BUILD_TAGS} -o ${OUTPUT}/${PROJECT} main.go

# 编译linux平台 请使用 make linux
linux: GOOS=linux
linux: build

test: clean output
	echo "mysql: ${MYSQL_MODE} ${MYSQL_ADDR} ${MYSQL_USER} ${MYSQL_PASS} "
	go test -gcflags=all=-l -v -json -coverprofile=${COVERAGE_OUT_FILE} `go list ./...` | tee ${TEST_OUT_FILE}

test-report:
	cat ${TEST_OUT_FILE} | go-junit-report -parser gojson > ${TEST_REPORT}	# junit report
	#go tool cover -html=${COVERAGE_OUT_FILE} -o ${COVERAGE_REPORT}
	gocov convert ${COVERAGE_OUT_FILE} | gocov-html > ${COVERAGE_REPORT}.html	# coverage report html
	gocov convert ${COVERAGE_OUT_FILE} | gocov-xml > ${COVERAGE_REPORT}.xml		# coverage report cobertura like
	gocov convert ${COVERAGE_OUT_FILE} > ${COVERAGE_REPORT}.json		# coverage report json
	gocov report ${COVERAGE_REPORT}.json  > ${COVERAGE_REPORT}.report 	# coverage report txt

upload-test-report:
	curl -T ${TEST_REPORT} ${TEST_REPORT_UPLOAD_DIR} --ftp-create-dir
	curl -T ${COVERAGE_REPORT}.html ${TEST_REPORT_UPLOAD_DIR} --ftp-create-dir
	curl -T ${COVERAGE_REPORT}.xml ${TEST_REPORT_UPLOAD_DIR} --ftp-create-dir
	curl -T ${COVERAGE_REPORT}.json ${TEST_REPORT_UPLOAD_DIR} --ftp-create-dir
	curl -T ${COVERAGE_REPORT}.report ${TEST_REPORT_UPLOAD_DIR} --ftp-create-dir

upload-test-ci-report:
	cp ${TEST_REPORT} ${CI_TEST_REPORT_DIR}/
	cp ${COVERAGE_REPORT}.* ${CI_TEST_REPORT_DIR}/

output:
	mkdir -p ${OUTPUT}

clean:
	rm -rf ${OUTPUT}

package: env
	cd deploy && \
	/bin/bash -x ./build.sh "datacloak" ${REF} ${IMAGE}

env:
	@echo "STABLE_VERSION =" ${STABLE_VERSION}
	@echo "REF =" ${REF}
	@echo "IMAGE =" ${IMAGE}
	@echo "GOOS =" ${GOOS}
	@echo "GOARCH =" ${GOARCH}
	@echo "CGO_ENABLED =" ${CGO_ENABLED}
	@echo "MEILIDB_MODE =" ${MEILIDB_MODE}

.PHONY: clean output build test linux package env

docker-login:
	docker login -uadmin -pHarbor12345 $(IMAGE_REPOSITORY)

image: docker-login
	docker build --build-arg=OS=$(PLATFORM) --build-arg=ARCH=$(ARCH) -t $(IMAGE_BASE)/$(IMAGE_NAME) .
	docker push $(IMAGE_BASE)/$(IMAGE_NAME)

deploy:
	# 创建公有网络
	docker network create -d overlay yunbay_backend
	# 部署nsq消息集群
	docker stack deploy -c docker-nsq-compose.yml		
	# 部署yunbay服务集群
	docker stack deploy -c docker-yunbay-compose.yml
	