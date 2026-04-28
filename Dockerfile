FROM node:lts-alpine
MAINTAINER Hongcai Deng <admin@dhchouse.com>

WORKDIR /forsaken-mail

COPY package*.json ./
RUN npm install --production && npm cache clean --force

COPY . /forsaken-mail

EXPOSE 25
EXPOSE 3000
CMD ["npm", "start"]
