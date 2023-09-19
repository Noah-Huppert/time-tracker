FROM golang:1.21.1-alpine3.18

# Create development user
ARG CONTAINER_USER_ID=1000
RUN adduser -D -u ${CONTAINER_USER_ID} api

# Install packages
RUN apk update && apk add bash

# Create working directory
RUN mkdir -p /opt/time-tracker/api/
RUN chown -R api /opt/time-tracker/api

# Drop down into app user
USER api
WORKDIR /opt/time-tracker/api

# Install dev dependencies
COPY --chown=api:api ./go.mod ./go.sum ./

RUN go get github.com/Noah-Huppert/goup@term-config && go install github.com/Noah-Huppert/goup

# Run
ENTRYPOINT [ "/bin/bash", "-c" ]
CMD [ "go mod download && GOUP_TERM_SIGNAL=INT goup" ]