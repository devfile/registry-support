FROM registry:2 as registry

FROM registry.access.redhat.com/ubi8-minimal:8.2
RUN microdnf update -y && rm -rf /var/cache/yum && microdnf install ca-certificates httpd-tools

# Create a non-root user to run the server as
RUN set -x ; \
    adduser registry -u 1001 -G root && exit 0

COPY --from=registry --chown=registry:0 /bin/registry /bin/registry
USER 1001
EXPOSE 5000
ENTRYPOINT ["registry"]
CMD ["serve", "/etc/docker/registry/config.yml"]