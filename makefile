include .env
.PHONY: pull push dev down log conf
BRANCH := ${shell git rev-parse --symbolic-full-name --abbrev-ref HEAD}

pull:
	git pull origin ${BRANCH}

push:
	git push origin ${BRANCH}

dev:
	docker-compose up -d

down:
	docker-compose down

log:
	docker-compose logs -f

conf:
	aws eks update-kubeconfig --name ${AWS_CLUSTER_NAME}