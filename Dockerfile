FROM node:lts-alpine
MAINTAINER Hongcai Deng <admin@dhchouse.com>

WORKDIR /forsaken-mail

COPY . /forsaken-mail

RUN npm install --production && npm cache clean --force

EXPOSE 25
EXPOSE 3000
CMD ["npm", "start"]
