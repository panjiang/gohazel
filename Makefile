
run:
	go run main.go

# Jump file download, write a empty file instead.
run-testing:
	MODE=TESTING go run main.go

image:
	./scripts/build-docker-image.sh

release:
	docker push panjiang/gohazel:latest

run-docker:
	docker run --rm -d --name gohazel \
		-v /data/gohazel/config.yml:/app/config.yml \
		-v /data/gohazel/assets:/assets \
		-p 8080:8080 \
		panjiang/gohazel