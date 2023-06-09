# Accept the Go version for the image to be set as a build argument.
# Default to Go 1.11
ARG GO_VERSION=1.18
#---------------- First stage: build the executable.
FROM golang:${GO_VERSION}-alpine AS builder

# Mofify URL default Alpine to Nexus Alpine Repo
COPY apk-repositories /etc/apk/repositories

# Create the user and group files that will be used in the running container to
# run the process as an unprivileged user.
RUN mkdir /user && \
    echo 'nobody:x:65534:65534:nobody:/:' > /user/passwd && \
    echo 'nobody:x:65534:' > /user/group
#RUN export GOPROXY=https://artifact.vnpay.vn/nexus/repository/go-proxy/
RUN go env -w GOPROXY=https://artifact.vnpay.vn/nexus/repository/go-proxy,direct
RUN apk add --no-cache ca-certificates git

# Set the working directory outside $GOPATH to enable the support for modules.
WORKDIR /src

# Fetch dependencies first; they are less susceptible to change on every build
# and will therefore be cached for speeding up the next build
COPY ./go.mod ./go.sum ./
RUN go mod download

# Import the code from the context.
COPY ./ ./

# Build the executable to `/executable`. Mark the build as statically linked.
RUN CGO_ENABLED=0 go build \
    -installsuffix 'static' \
    -o ./executable .
#---------------- Final stage: the running container.
FROM alpine:3.16 AS final

# Mofify URL default Alpine to Nexus Alpine Repo
COPY apk-repositories /etc/apk/repositories

RUN apk --no-cache add tzdata && apk --update --no-cache add busybox-extras && apk add curl
# Import the user and group files from the first stage.
COPY --from=builder /user/group /user/passwd /etc/

# Import the Certificate-Authority certificates for enabling HTTPS.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Import the compiled executable from the first stage.
COPY --from=builder /src/executable /src/executable
COPY --from=builder /src/config/development.json /src/config/development.json
COPY --from=builder /src/prv.rsa /src/prv.rsa
COPY --from=builder /src/pub.rsa /src/pub.rsa
EXPOSE 4000
WORKDIR /src
RUN chmod +x ./executable
USER nobody:nobody
ENTRYPOINT GIN_MODE=release ./executable -TEST=false -ENV=development