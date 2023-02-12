FROM portainer/base as production

COPY dist/docker /
COPY dist/docker-compose /
COPY dist/helm /
COPY dist/kubectl /
COPY dist/mustache-templates /mustache-templates/
COPY dist/pod-security-policy /pod-security-policy/
COPY dist/portainer /
COPY dist/public /public/

# storybook exists only in portainerci builds
COPY dist/storybook* /storybook/

VOLUME /data
WORKDIR /

EXPOSE 9000
EXPOSE 9443
EXPOSE 8000

LABEL io.portainer.server true

ENTRYPOINT ["/portainer"]

