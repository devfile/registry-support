FROM nginx:stable-alpine

# Install and configure dependencies
RUN apk add --no-cache bash git curl coreutils
RUN wget https://github.com/deislabs/oras/releases/download/v0.8.1/oras_0.8.1_linux_amd64.tar.gz && \
  mkdir -p oras-install/ && \
  tar -zxf oras_0.8.1_*.tar.gz -C oras-install/ && \
  mv oras-install/oras /usr/local/bin/ && \
  rm -rf oras_0.8.1_*.tar.gz oras-install/

COPY nginx.conf /etc/nginx/nginx.conf
COPY entrypoint.sh /

# Create a non-root user to run the nginx server as
RUN set -x ; \
  adduser -u 82 -D -S -G root www-data && exit 0 ; exit 1

RUN touch /var/run/nginx.pid
RUN mkdir -p /www/data



# Modify the permissions on the necessary files to allow the container to properly run as a non-root UID
RUN chown -R www-data:root /var/run/nginx.pid && \
  chown -R www-data:root /var/cache/nginx && \
  chown www-data:root /etc/nginx/conf.d /etc/nginx/nginx.conf
RUN chmod g+rwx /var/run/nginx.pid && \
  chmod -R g+rwx /var/cache/nginx && chmod -R g+rwx /etc/nginx && chmod -R g+rwx /www/data

USER www-data

EXPOSE 8080
ENTRYPOINT ["/entrypoint.sh"]
CMD ["nginx", "-g", "daemon off;"]