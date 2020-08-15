go:
	docker run -it --rm \
		-v $(PWD):$(PWD) \
		-w $(PWD) \
		--env-file .env \
		golang:1.14 \
		bash

build-bin:
	docker run -it --rm \
		-v $(PWD):$(PWD) \
		-w $(PWD) \
		golang:1.14 \
		go build -o build/main main.go

build: build-bin
	docker run -it --rm \
		-v $(PWD)/build:/build \
		-w /build \
		alpine \
		sh -c 'apk add --no-cache zip && zip camelid_payload.zip main'

test:
	docker run -it --rm \
		-v $(PWD):$(PWD):ro \
		-w $(PWD) \
		golang:1.14 \
		go test ./...
