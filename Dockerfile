FROM golang:1.16.3
RUN apt-get update && apt install -y protobuf-compiler
WORKDIR /app
COPY ./Makefile ./Makefile
RUN make migrate-check-deps
RUN curl -fsSL https://deb.nodesource.com/setup_14.x | bash - && apt-get install -y nodejs
