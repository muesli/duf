DOCKER_IMAGE_TAG ?= duf

duf-run:
	docker build \
  	-t $(DOCKER_IMAGE_TAG) \
  	-f Dockerfile . && \
	docker run --name duf-run -e GOMAXPROCS=1 --rm  $(DOCKER_IMAGE_TAG)