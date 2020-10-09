# Build the registry index generator
FROM golang:alpine3.11 AS build
WORKDIR /tools
COPY generator .
RUN ./build.sh

FROM docker.io/httpd:2.4.43-alpine AS registry

# Install and configure dependencies
RUN apk add --no-cache bash git && \
    # Allow htaccess
    sed -i 's|    AllowOverride None|    AllowOverride All|' /usr/local/apache2/conf/httpd.conf && \
    sed -i 's|Listen 80|Listen 8080|' /usr/local/apache2/conf/httpd.conf && \
    mkdir -m 777 /usr/local/apache2/htdocs/devfiles && \
    mkdir -p /var/www && ln -s /usr/local/apache2/htdocs /var/www/html && \
    chmod -R g+rwX /usr/local/apache2 && \
    echo "ServerName localhost" >> /usr/local/apache2/conf/httpd.conf && \
    apk add --no-cache coreutils
    
# Install ORAS
RUN wget https://github.com/deislabs/oras/releases/download/v0.8.1/oras_0.8.1_linux_amd64.tar.gz && \
  mkdir -p oras-install/ && \
  tar -zxf oras_0.8.1_*.tar.gz -C oras-install/ && \
  mv oras-install/oras /usr/local/bin/ && \
  rm -rf oras_0.8.1_*.tar.gz oras-install/
  
 
COPY .htaccess /usr/local/apache2/htdocs/
RUN chgrp -R 0 /usr/local/apache2/htdocs && \
      chmod -R g=u /usr/local/apache2/htdocs


COPY entrypoint.sh /scripts/entrypoint.sh

# Git clone the devfiles
# ToDo: Switch to github.com/devfile/catalog/
RUN git clone https://github.com/odo-devfiles/registry /registry
COPY --from=build /tools/index-generator /registry/index-generator
RUN chgrp -R 0 /registry && \
    chmod -R g=u /registry

# Load the devfiles into the registry
ENTRYPOINT ["/scripts/entrypoint.sh"]
CMD ["httpd-foreground"]