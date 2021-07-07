FROM golang:1.16.3
RUN apt-get update && apt install -y protobuf-compiler
RUN curl -fsSL https://deb.nodesource.com/setup_14.x | bash - && apt-get install -y nodejs
WORKDIR /app
COPY ./Makefile ./Makefile
RUN make migrate-check-deps
