go:
	docker run -it --rm \
		-v $(PWD):$(PWD) \
		-w $(PWD) \
		--env-file .env \
		golang:1.14 \
		bash

test:
	docker run -it --rm \
		-v $(PWD):$(PWD):ro \
		-w $(PWD) \
		golang:1.14 \
		go test ./...
