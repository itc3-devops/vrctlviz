FROM mhart/alpine-node

# Create app directory
RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app

# Install app dependencies
COPY package.json /usr/src/app/
RUN npm install

# Bundle app source
COPY . /usr/src/app

EXPOSE 8080

ENV TRAFFIC_URL=sample_data.json
ENV TRAFFIC_INTERVAL=5
CMD [ "npm", "run", "dev" ]
