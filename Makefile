include env/config.env

APP_NAME=eseed
VERSION=v1.0.0

REPO=ghcr.io/${REPO_USERNAME}

run:
	nodemon --exec go run main.go --signal SIGTERM

login:
	echo ${REPO_SECRET_KEY} | docker login ghcr.io --username ${REPO_USERNAME} --password-stdin

image: login
	docker build -f Dockerfile -t ${REPO}/${APP_NAME}:${VERSION} .
	docker push ${REPO}/${APP_NAME}:${VERSION}
	docker rmi ${REPO}/${APP_NAME}:${VERSION}