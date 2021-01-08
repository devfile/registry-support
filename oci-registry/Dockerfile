FROM registry:2 as registry

FROM registry.access.redhat.com/ubi8-minimal:8.2
RUN microdnf update -y && rm -rf /var/cache/yum && microdnf install ca-certificates httpd-tools

COPY --from=registry /bin/registry /bin/registry
EXPOSE 5000
ENTRYPOINT ["registry"]
CMD ["serve", "/etc/docker/registry/config.yml"]