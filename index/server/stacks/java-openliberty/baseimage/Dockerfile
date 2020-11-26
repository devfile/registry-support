FROM adoptopenjdk/openjdk11-openj9:ubi AS maven

RUN yum upgrade --disableplugin=subscription-manager -y \
   && yum clean --disableplugin=subscription-manager packages \
   && echo 'Finished installing dependencies'

USER root

# Dependency install
RUN yum install --disableplugin=subscription-manager -y https://dl.fedoraproject.org/pub/epel/epel-release-latest-7.noarch.rpm \
   && yum install --disableplugin=subscription-manager -y unzip curl ca-certificates wget xmlstarlet

# Maven install
ARG MAVEN_VERSION=3.6.3
ARG USER_HOME_DIR="/root"
ARG SHA=c35a1803a6e70a126e80b2b3ae33eed961f83ed74d18fcd16909b2d44d7dada3203f1ffe726c17ef8dcca2dcaa9fca676987befeadc9b9f759967a8cb77181c0
ARG BASE_URL=https://archive.apache.org/dist/maven/maven-3/${MAVEN_VERSION}/binaries/

RUN mkdir -p /usr/share/maven /usr/share/maven/ref \
  && curl -fsSL -o /tmp/apache-maven.tar.gz ${BASE_URL}/apache-maven-${MAVEN_VERSION}-bin.tar.gz \
  && echo "${SHA}  /tmp/apache-maven.tar.gz" | sha512sum -c - \
  && tar -xzf /tmp/apache-maven.tar.gz -C /usr/share/maven --strip-components=1 \
  && rm -f /tmp/apache-maven.tar.gz \
  && ln -s /usr/share/maven/bin/mvn /usr/bin/mvn

FROM maven AS builder
RUN  mkdir -p /opt/ol \
  && mkdir -p /stacks/java-openliberty/starterapp

RUN chmod -R 777 /stacks/java-openliberty/starterapp

COPY ./LICENSE /licenses/
ADD ./starterapp/src /stacks/java-openliberty/starterapp/src
COPY ./starterapp/pom.xml /stacks/java-openliberty/starterapp/

WORKDIR /stacks/java-openliberty/starterapp
RUN mvn -DserverName=tmp -Dmaven.repo.local=/mvn/repository -Dliberty.runtime.version=20.0.0.9 liberty:create liberty:install-feature dependency:go-offline
RUN rm -rf ./target/liberty/wlp/usr/servers/tmp

FROM maven
RUN mkdir /mvn
RUN mkdir -p /opt/ol
CMD /bin/bash
COPY --from=builder /mvn /mvn
COPY --from=builder /stacks/java-openliberty/starterapp/target/liberty /opt/ol
RUN chmod -R 777 /mvn
RUN chmod -R 777 /opt/ol
ARG stacklabel
LABEL org.opencontainers.image.revision=$stacklabel