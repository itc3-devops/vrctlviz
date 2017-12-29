FROM mhart/alpine-node

# Create app directory
RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app

# Install app dependencies
COPY package.json /usr/src/app/
COPY viz_data.json /.shared/configs/viz_data.json
RUN npm install

# Bundle app source
COPY . /usr/src/app

EXPOSE 8080

ENV TRAFFIC_URL=/.shared/configs/viz_data.json
ENV TRAFFIC_INTERVAL=300
CMD [ "npm", "run", "dev" ]
