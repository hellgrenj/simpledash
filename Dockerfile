FROM node:lts-alpine AS node
FROM golang:1.18-alpine AS builder
WORKDIR /app
# copy over node binaries from node layer
COPY --from=node /usr/lib /usr/lib
COPY --from=node /usr/local/share /usr/local/share
COPY --from=node /usr/local/lib /usr/local/lib
COPY --from=node /usr/local/include /usr/local/include
COPY --from=node /usr/local/bin /usr/local/bin

COPY go.mod go.sum ./
# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download
COPY . .
RUN go build -o main .
# install and run uglifyjs on our js scripts
RUN npm install uglify-js -g
RUN uglifyjs ./static/app.js -c -m -o ./static/app.js
RUN uglifyjs ./static/helper.js -c -m -o ./static/helper.js
RUN uglifyjs ./static/ws.js -c -m -o ./static/ws.js

FROM alpine:3.15 as runtime
COPY --from=builder ./app/main main
COPY --from=builder ./app/static static
CMD ["./main"]