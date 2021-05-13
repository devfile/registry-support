# Builder image
FROM golang:alpine3.11 AS builder
WORKDIR /tools
COPY . .
RUN CGO_ENABLED=0 go build -mod=vendor -o index-server main.go

# Application image
FROM registry.access.redhat.com/ubi8-minimal:8.2
COPY nginx.repo  /etc/yum.repos.d/nginx.repo

# Install and configure dependencies
RUN microdnf update -y && rm -rf /var/cache/yum && microdnf install nginx shadow-utils findutils git

COPY nginx.conf /etc/nginx/nginx.conf
COPY entrypoint.sh /

# Load index server
COPY --from=builder /tools/index-server /registry/index-server
RUN chgrp -R 0 /registry && \
    chmod -R g=u /registry

# Create a non-root user to run the nginx server as
RUN set -x ; \
    adduser www-data -u 82 -G root && exit 0

RUN touch /var/run/nginx.pid
RUN mkdir -p /www/data && mkdir -p /var/cache/nginx

# Modify the permissions on the necessary files to allow the container to properly run as a non-root UID
RUN chown -R www-data:root /var/run/nginx.pid && \
    chown -R www-data:root /var/cache/nginx && \
    chown www-data:root /etc/nginx/conf.d /etc/nginx/nginx.conf
RUN chmod g+rwx /var/run/nginx.pid && \
    chmod -R g+rwx /var/cache/nginx && chmod -R g+rwx /etc/nginx && chmod -R g+rwx /www/data

USER www-data

# Set env vars for the locations of the devfile stacks and index.json
ENV DEVFILE_STACKS /stacks
ENV DEVFILE_INDEX /index.json
ENV DEVFILE_SAMPLE_INDEX /www/data/sample_index.json
ENV DEVFILE_STACK_INDEX /www/data/stack_index.json

EXPOSE 8080 7071
ENTRYPOINT ["/entrypoint.sh"]
