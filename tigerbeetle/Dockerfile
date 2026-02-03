# Custom TigerBeetle Image with Healthcheck Support
# This Dockerfile extends the official TigerBeetle image to add netcat (nc) 
# for container healthchecks

FROM ghcr.io/tigerbeetle/tigerbeetle:latest

# Install netcat-openbsd for healthchecks
# TigerBeetle uses Alpine Linux as base, so we use apk
RUN apk add --no-cache netcat-openbsd

# Metadata
LABEL maintainer="wallet-service"
LABEL description="TigerBeetle with healthcheck support"
LABEL version="0.16.70-custom"

# The base image already sets the entrypoint and working directory
# We just add the healthcheck capability

