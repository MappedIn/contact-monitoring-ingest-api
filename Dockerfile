# BUILD: build service from source
FROM golang:1.14 as build

WORKDIR /app

# copy dependency manifests (go.mod, go.sum) and install
COPY go.mod .
COPY go.sum .
COPY Makefile .
RUN make install

# copy the rest of our source code and build
COPY . .

# RUN make test

# ENV CGO_ENABLED=0
ENV GOOS=linux
RUN make build


# RUN: copy binary from build and run
FROM alpine
WORKDIR /app
COPY --from=build /app/bin/server /app/server
RUN chmod 0777 /app

ENV PORT=80
ENV GIN_MODE=release
EXPOSE 80

ENTRYPOINT ./server
