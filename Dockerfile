FROM golang:1.17 AS base
ENTRYPOINT [ "/main" ]
RUN apt-get update && apt-get install -yqq libdevmapper1.02.1
COPY /main /