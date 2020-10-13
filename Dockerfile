FROM node:14.13.1-alpine3.12 AS client

COPY client /usr/local/src/moodboard
WORKDIR /usr/local/src/moodboard

ENV NODE_ENV=production
RUN \
	npm install && \
	npm run build

FROM golang:1.15.2-alpine3.12 AS server

COPY server /usr/local/src/moodboard
WORKDIR /usr/local/src/moodboard

RUN CGO_ENABLED=0 go build -o moodboard cmd/moodboard/main.go

FROM alpine:3.12

RUN apk add --no-cache nginx supervisor

COPY --from=client /usr/local/src/moodboard/build /usr/share/nginx/html
COPY --from=server /usr/local/src/moodboard/moodboard /usr/local/bin
COPY nginx.conf /etc/nginx
COPY supervisord.conf /etc

EXPOSE 80

CMD /usr/bin/supervisord --nodaemon --logfile /dev/null --logfile_maxbytes 0
