FROM golang:1.19
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
WORKDIR /app
# Avoid invalidating the `go mod download` cache when only code has changed.
COPY go.mod go.sum cmd/main.go ./
RUN go mod download
RUN apt-get update && apt-get install -yqq libdevmapper1.02.1
COPY . ./
RUN go build -o /bin/ecr-image-sync ./main.go

# Run as UID for nobody
USER 65534

ENTRYPOINT ["/bin/ecr-image-sync"]
