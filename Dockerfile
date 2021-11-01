FROM golang:1.16-alpine AS build
COPY . /src/
WORKDIR /src
RUN CGO_ENABLED=0 go build -o /src/docker_dns

FROM scratch
COPY --from=build /src/docker_dns /bin/docker_dns
# Copy certs over from base image too
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
CMD ["/bin/docker_dns"]
