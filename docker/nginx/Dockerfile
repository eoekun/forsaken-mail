FROM nginx:1.27-alpine

RUN apk add --no-cache apache2-utils

COPY default.conf /etc/nginx/conf.d/default.conf
COPY docker-entrypoint.sh /docker-entrypoint.d/99-generate-htpasswd.sh
RUN chmod +x /docker-entrypoint.d/99-generate-htpasswd.sh
