FROM node:alpine as build-stage

WORKDIR '/app'

# Install some depenendencies
COPY /package.json .
RUN yarn install
COPY /.  .

RUN npm run build:release

# production  stage
FROM openresty/openresty:1.15.8.2-7-centos AS pc-front
# FROM nginx:stable-alpine as pc-front
# COPY nginx.conf  /etc/nginx/conf.d/default.conf
COPY --from=build-stage /app/dist /usr/share/nginx/html/

EXPOSE 8080
CMD ["nginx", "-g", "daemon off;"]

