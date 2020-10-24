
run:
	go run main.go

# Jump file download, write a empty file instead.
run-testing:
	MODE=TESTING go run main.go

docker-image:
	./scripts/build-docker-image.sh

docker-push:
	docker push panjiang/gohazel:latest

docker-run:
	docker run --rm -d --name gohazel \
		-v /data/gohazel/config.yml:/app/config.yml \
		-v /data/gohazel/assets:/assets \
		-p 8080:8080 \
		panjiang/gohazel

# VERSION=1.0.1 make release
.PHONY: release
release:
	./scripts/release.sh