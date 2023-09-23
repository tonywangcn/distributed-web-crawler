include .env
.PHONY: pull push dev down log conf
BRANCH := ${shell git rev-parse --symbolic-full-name --abbrev-ref HEAD}

BRANCH := ${shell git rev-parse --symbolic-full-name --abbrev-ref HEAD}
SRV_NAME=distributed-web-crawler
REPO=${DOCKER_IMAGE_REGISTRY}

TAG=$(shell date +%Y%m%d%H%M%S)
FIXTAG?=prod
NAME=${REPO}/${REGISTRY_NAME_SPACE}/${SRV_NAME}
NODE=node-crawler
GO=go-crawler

build-go:
	echo build ${GO}:latest
	cp ./docker/go/Dockerfile .
	docker build -t ${GO}:latest .
	rm Dockerfile
	docker tag ${GO}:latest ${NAME}:${GO}-latest
	docker tag ${GO}:latest ${NAME}:${GO}-${TAG}
	docker push ${NAME}:${GO}-latest
	docker push ${NAME}:${GO}-${TAG}

build-node:
	echo build ${NODE}:latest
	cp ./docker/node/Dockerfile .
	docker build --progress=plain -t ${NODE}:latest .
	rm Dockerfile
	docker tag ${NODE}:latest ${NAME}:${NODE}-latest
	docker tag ${NODE}:latest ${NAME}:${NODE}-${TAG}
	docker push ${NAME}:${NODE}-latest
	docker push ${NAME}:${NODE}-${TAG}

pull:
	git pull origin ${BRANCH}

push:
	git push origin ${BRANCH}

dev:
	docker-compose up -d

prod:
	docker-compose -f docker-compose-prod.yml up -d

prod-down:
	docker-compose -f docker-compose-prod.yml down

down:
	docker-compose down

log:
	docker-compose logs -f

conf:
	aws eks update-kubeconfig --name ${AWS_CLUSTER_NAME}