#!/usr/bin/env bash
docker run -p 9999:9999 -e DOCKER_API_ADDRESS="134.44.36.120:2376" -it --rm weibh/podinteractive