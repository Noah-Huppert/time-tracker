FROM node:20-alpine3.17

# Install packages
RUN apk update && apk add bash

ARG CONTAINER_USER_NAME=node

# Create working directory
RUN mkdir -p /opt/time-tracker/
RUN chown -R ${CONTAINER_USER_NAME} /opt/time-tracker/

# Drop down into app user
USER ${CONTAINER_USER_NAME}
WORKDIR /opt/time-tracker/frontend

# Run
ENTRYPOINT [ "/bin/bash", "-c" ]
CMD [ "npm install && npm run dev" ]