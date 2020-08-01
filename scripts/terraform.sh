#!/bin/sh

docker run -it --rm \
    --env-file .env \
    -v $(pwd)/terraform:/work \
    -w /work \
    hashicorp/terraform:0.12.29 \
    "$@"
