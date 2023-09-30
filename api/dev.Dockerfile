FROM golang:1.21.1-alpine3.18

# Create development user
ARG CONTAINER_USER_ID=1000
ARG CONTAINER_USER_NAME=api
RUN adduser -D -u ${CONTAINER_USER_ID} ${CONTAINER_USER_NAME}

# Install packages
RUN apk update && apk add bash curl postgresql-client
RUN curl -sSf https://atlasgo.sh | sh

# Create working directory
RUN mkdir -p /opt/time-tracker/api/
RUN chown -R ${CONTAINER_USER_NAME} /opt/time-tracker/api

# Drop down into app user
USER ${CONTAINER_USER_NAME}
WORKDIR /opt/time-tracker/api

# Install dev dependencies
COPY --chown=${CONTAINER_USER_NAME}:${CONTAINER_USER_NAME} ./go.mod ./go.sum ./

RUN go get github.com/Noah-Huppert/goup@term-config && go install github.com/Noah-Huppert/goup

# Run
ENTRYPOINT [ "/bin/bash", "-c" ]
CMD [ "go mod download && GOUP_TERM_SIGNAL=INT goup" ]