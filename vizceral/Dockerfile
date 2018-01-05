FROM mhart/alpine-node

COPY rootfs/ /
# Create app directory
RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app

# Install app dependencies
RUN npm install

RUN apk add --no-cache bash runit \
&& chmod +x /sbin/runit-docker \
&& sed -i 's/\r//g' /sbin/runit-docker

EXPOSE 8080

ENV TRAFFIC_URL=sample_data.json
ENV TRAFFIC_INTERVAL=1
ENTRYPOINT ["/sbin/runit-docker"]