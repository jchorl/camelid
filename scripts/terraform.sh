#!/bin/sh

docker run -it --rm \
    --env-file .env \
    -v $(pwd):/work \
    -w /work/terraform \
    hashicorp/terraform:0.13.0 \
    "$@"
