VERSION ?= $(shell cat ./VERSION)
APP_NAME ?= rss2email
IMAGE_NAME ?= rss2email
IMAGE_TAG ?= $(VERSION)
IMAGE ?= $(IMAGE_NAME):$(IMAGE_TAG)
CONTAINER_NAME ?= rss2email
WEB_CONTAINER_NAME ?= rss2email-web
WEB_PORT ?= 8081
DB_DIR ?= $(CURDIR)/db

.PHONY: docker-image-build docker-run docker-run-web docker-logs docker-logs-web docker-stop docker-stop-web docker-rm docker-rm-web

docker-image-build:
	docker build -t $(IMAGE) --build-arg APP_NAME=$(APP_NAME) .

docker-run:
	docker run -d \
		--name $(CONTAINER_NAME) \
		--restart unless-stopped \
		-v /etc/localtime:/etc/localtime:ro \
		-v $(DB_DIR):/usr/local/bin/db \
		$(IMAGE)

docker-run-web:
	docker run -d \
		--name $(WEB_CONTAINER_NAME) \
		--restart unless-stopped \
		-p $(WEB_PORT):$(WEB_PORT) \
		-v /etc/localtime:/etc/localtime:ro \
		-v $(DB_DIR):/usr/local/bin/db \
		$(IMAGE) sh -c "/usr/local/bin/$(APP_NAME) web --addr :$(WEB_PORT) | tee /usr/local/bin/app.log"

docker-logs:
	docker logs -f --tail 200 $(CONTAINER_NAME)

docker-logs-web:
	docker logs -f --tail 200 $(WEB_CONTAINER_NAME)

docker-stop:
	-docker stop $(CONTAINER_NAME)

docker-stop-web:
	-docker stop $(WEB_CONTAINER_NAME)

docker-rm:
	-docker rm $(CONTAINER_NAME)

docker-rm-web:
	-docker rm $(WEB_CONTAINER_NAME)
