ARG src=/tmp/src

# image with the project dependencies installed in the Go cache.
FROM golang:1.14-buster AS with-deps
ARG src
WORKDIR ${src}
COPY go.mod .
COPY go.sum .
RUN go mod download

# image with the sources, useful for CI testing and building
FROM with-deps AS with-sources
COPY . .
RUN ls -la 

# image with the compiled static binary to run the app.
FROM with-sources AS build
RUN CGO_ENABLED=0 go build \
    -o /bin/build \
    -ldflags '-extldflags "-static"' \
    -tags timetzdata \
    .

# Creates an empty image with certs and other things required
# to run Go static binaries.
FROM scratch AS with-certs
COPY --from=with-deps \
    /etc/ssl/certs/ca-certificates.crt \
    /etc/ssl/certs/ca-certificates.crt

# Creates a single layer image to run the app.
FROM with-certs AS run
COPY --from=build /bin/build /bin/runme
ENTRYPOINT ["/bin/runme"]
