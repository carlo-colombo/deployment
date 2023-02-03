FROM node:alpine

ARG TIDDLYWIKI_VERSION

ARG SOURCE_COMMIT

RUN apk add --no-cache tini
RUN npm install -g tiddlywiki@${TIDDLYWIKI_VERSION}

EXPOSE 8080

VOLUME /tiddlywiki
WORKDIR /tiddlywiki

ENTRYPOINT ["/sbin/tini", "--", "tiddlywiki"]
CMD ["--help"]
