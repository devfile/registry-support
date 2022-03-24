# Index Server build stage
FROM golang:1.17-alpine AS index-builder
WORKDIR /tools
COPY . .
RUN CGO_ENABLED=0 go build -mod=vendor -o index-server main.go

# Application image
FROM registry.access.redhat.com/ubi8/nodejs-14-minimal AS runner
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

WORKDIR /app

# Copy over the registry viewer's resources
COPY --from=registry-viewer --chown=www-data:root /app/next.config.js ./
COPY --from=registry-viewer --chown=www-data:root /app/public ./public
COPY --from=registry-viewer --chown=www-data:root /app/.next ./.next
COPY --from=registry-viewer /app/node_modules ./node_modules
COPY --from=registry-viewer /app/package.json ./package.json
COPY --from=registry-viewer --chown=www-data:root /app/config ./config

# Make sure .next is writable
RUN chmod -R g+rwx /app/.next

# Run the registry viewer in production
ENV NODE_ENV production

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
