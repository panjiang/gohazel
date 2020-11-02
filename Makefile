
run:
	go run main.go

run-private:
	MODE=TESTING go run main.go -config=config.private.yml

# Jump file download, write a empty file instead.
run-testing:
	MODE=TESTING go run main.go

docker-image:
	./scripts/build-docker-image.sh

docker-push:
	docker push panjiang/gohazel:latest

docker-run:
	docker run --rm -d --name gohazel \
		-v /data/gohazel/assets:/assets \
		-p 8400:8400 \
		panjiang/gohazel

staticcheck:
	@hash staticcheck > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u honnef.co/go/tools/cmd/staticcheck; \
	fi
	staticcheck ./...

.PHONY: test
test:
	go test ./...

release:
	goreleaser --rm-dist 