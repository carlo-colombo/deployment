FROM hayd/deno

RUN apk add --no-cache tini

ADD deps.ts .

RUN deno cache deps.ts

ADD . .

RUN deno cache *.ts

ENTRYPOINT ["/sbin/tini", "--", "deno"]

CMD ["run", "--allow-net", "--allow-env", "server.ts"]