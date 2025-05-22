include .env.local

APP_NAME=vuedemo
VERSION=1.0.0

REPO=ghcr.io/${REPO_USERNAME}

run:
	nodemon --exec go run main.go --signal SIGTERM

login:
	echo ${REPO_SECRET_KEY} | docker login ghcr.io --username ${REPO_USERNAME} --password-stdin

build:
	rm -rf dist
	pnpm build

image: login  build
	docker build -f Dockerfile -t ${REPO}/${APP_NAME}:${VERSION} .
	docker push ${REPO}/${APP_NAME}:${VERSION}
	docker rmi ${REPO}/${APP_NAME}:${VERSION}