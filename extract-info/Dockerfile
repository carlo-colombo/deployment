FROM node:alpine

RUN apk add --no-cache git tini

ADD . .

RUN npm install

CMD ["/sbin/tini", "--", "node", "index.mjs"]