# Start by building the application.
FROM golang:1.20-bullseye as build

WORKDIR /src
COPY . .

ENV CGO_ENABLED=1

RUN go mod download
RUN go build -o bootstrap

## Now copy it into our base image.
FROM gcr.io/distroless/base-debian11

COPY --from=build /src/bootstrap /

CMD ["/bootstrap"]