# Index Server build stage
FROM golang:1.17-alpine AS index-builder
WORKDIR /tools
COPY . .
RUN CGO_ENABLED=0 go build -mod=vendor -o index-server main.go

# Registry Viewer build stage
FROM registry.access.redhat.com/ubi8/nodejs-14-minimal AS viewer-builder
USER root

# Install yarn global dependency
RUN npm install -g yarn

# Copy devfile-web repository from index-builder
COPY --from=index-builder /tools/devfile-web /app

WORKDIR /app

# Install project dependencies
RUN $(npm get prefix)/bin/yarn install --frozen-lockfile --ignore-optional
# Build production registry-viewer
RUN $(npm get prefix)/bin/yarn nx build registry-viewer --configuration=production
# Export static site
RUN $(npm get prefix)/bin/yarn nx run registry-viewer:export

# Application image
FROM registry.access.redhat.com/ubi8/ubi-minimal AS runner
USER root

# Install and configure dependencies
RUN microdnf update -y && microdnf install shadow-utils findutils && rm -rf /var/cache/yum
COPY entrypoint.sh /

# Copy index server
COPY --from=index-builder /tools/index-server /registry/index-server
RUN chgrp -R 0 /registry && \
    chmod -R g=u /registry

# Create a non-root user to run the server as
RUN set -x ; \
    adduser www-data -u 82 -G root && exit 0

# Modify the permissions on the necessary files to allow the container to properly run as a non-root UID
RUN mkdir -p /www/data && chmod -R g+rwx /www/data

WORKDIR /www

# Copy over the registry viewer's resources
COPY --from=viewer-builder /app/dist/apps/registry-viewer/exported ./viewer

ENV DEVFILE_VIEWER /www/viewer
ENV DEVFILE_VIEWER_ROOT /viewer
ENV DEVFILE_COMMUNITY_HOST false

# Set env vars for the locations of the devfile stacks and index.json
ENV DEVFILE_STACKS /registry/stacks
ENV DEVFILE_SAMPLES /registry/samples
ENV DEVFILE_INDEX /registry/index.json
ENV DEVFILE_BASE64_INDEX /www/data/index_base64.json
ENV DEVFILE_SAMPLE_INDEX /www/data/sample_index.json
ENV DEVFILE_SAMPLE_BASE64_INDEX /www/data/sample_base64_index.json
ENV DEVFILE_STACK_INDEX /www/data/stack_index.json
ENV DEVFILE_STACK_BASE64_INDEX /www/data/stack_base64_index.json

USER www-data

EXPOSE 8080

ENTRYPOINT ["/entrypoint.sh"]
