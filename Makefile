go:
	docker run -it --rm \
		-v $(PWD):$(PWD) \
		-w $(PWD) \
		--env-file .env \
		golang:1.14 \
		bash
