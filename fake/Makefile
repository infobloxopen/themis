.PHONY docker: docker-build
docker-build:
	docker build -t infobloxopen/themis:fake-v1 -f ./docker/Dockerfile ..

.PHONY k3dimport: k3dimport-build
k3dimport-build: docker-build
	k3d import-images infobloxopen/themis:fake-v1
