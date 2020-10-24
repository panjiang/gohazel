
run:
	go run main.go

run-testing:
	MODE=TESTING go run main.go

image:
	./scripts/build-docker-image.sh

run-docker:
	docker run --rm \
		-v /data/gohazel/config.yml:/app/config.yml \
		-v /data/gohazel/assets:/assets \
		-p 8080:8080 \
		panjiang/gohazel