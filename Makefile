NAME = jayden21/
VERSION = 1.1.1
REGISTRY = 

.PHONY: build start push

build: build-version

build-app:
	go build -o output/account account/main.go
	go build -o output/ybapi ybapi/main.go	
	go build -o output/ybasset ybasset/main.go	
	go build -o output/ybgoods ybgoods/main.go
	go build -o output/yborder yborder/main.go
	go build -o output/ybim ybim/main.go
	go build -o output/ybcron ybcron/main.go
	go build -o output/ybnsq ybnsq/main.go
	go build -o output/ybpay ybpay/main.go
	go build -o output/ybsearch ybsearch/main.go
	#go build -o ybnsq/ybnsq ybnsq/main.go
	

build-version:
	docker build -t ${NAME}account:${VERSION}  account/
	docker build -t ${NAME}ybapi:${VERSION}  ybapi/	
	docker build -t ${NAME}ybgoods:${VERSION}  ybgoods/
	docker build -t ${NAME}yborder:${VERSION}  yborder/	
	docker build -t ${NAME}ybasset:${VERSION}  ybasset/
	
	docker build -t ${NAME}ybnsq:${VERSION}  ybnsq/
	docker build -t ${NAME}ybcron:${VERSION}  ybcron/
	docker build -t ${NAME}ybeos:${VERSION}  ybeos/
	docker build -t ${NAME}ybpay:${VERSION}  ybpay/
	docker build -t ${NAME}ybsearch:${VERSION}  ybsearch/
	docker build -t ${NAME}ybim:${VERSION}  ybim/

tag-latest:
	docker tag ${NAME}account:${VERSION} ${REGISTRY}${NAME}account:latest
	docker tag ${NAME}ybapi:${VERSION} ${REGISTRY}${NAME}ybapi:latest	
	docker tag ${NAME}ybgoods:${VERSION} ${REGISTRY}${NAME}ybgoods:latest
	docker tag ${NAME}yborder:${VERSION} ${REGISTRY}${NAME}yborder:latest	
	docker tag ${NAME}ybasset:${VERSION} ${REGISTRY}${NAME}ybasset:latest	

	docker tag ${NAME}ybcron:${VERSION} ${REGISTRY}${NAME}ybcron:latest
	docker tag ${NAME}ybnsq:${VERSION} ${REGISTRY}${NAME}ybnsq:latest
	docker tag ${NAME}ybeos:${VERSION} ${REGISTRY}${NAME}ybeos:latest
	docker tag ${NAME}ybpay:${VERSION} ${REGISTRY}${NAME}ybpay:latest
	docker tag ${NAME}ybsearch:${VERSION} ${REGISTRY}${NAME}ybsearch:latest
	docker tag ${NAME}ybim:${VERSION} ${REGISTRY}${NAME}ybim:latest
	

push:	build-version tag-latest
	docker push ${REGISTRY}${NAME}account:latest
	docker push ${REGISTRY}${NAME}ybapi:latest	
	docker push ${REGISTRY}${NAME}ybgoods:latest
	docker push ${REGISTRY}${NAME}yborder:latest	
	docker push ${REGISTRY}${NAME}ybasset:latest

	docker push ${REGISTRY}${NAME}ybcron:latest
	docker push ${REGISTRY}${NAME}ybnsq:latest
	docker push ${REGISTRY}${NAME}ybeos:latest
	docker push ${REGISTRY}${NAME}ybpay:latest
	docker push ${REGISTRY}${NAME}ybsearch:latest
	docker push ${REGISTRY}${NAME}ybim:latest


deploy:
	# 创建公有网络
	docker network create -d overlay yunbay_backend
	# 部署nsq消息集群
	docker stack deploy -c docker-nsq-compose.yml		
	# 部署yunbay服务集群
	docker stack deploy -c docker-yunbay-compose.yml
	