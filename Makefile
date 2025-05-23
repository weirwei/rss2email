VERSION ?= $(shell cat ./VERSION)
docker-image-build:
	docker build -t rss2email:${VERSION} --build-arg APP_NAME=rss2email .

docker-run:
	docker run -v /etc/localtime:/etc/localtime:ro -v ./db:/usr/local/bin/db -d rss2email:${VERSION}