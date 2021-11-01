FROM golang:1.16-alpine AS build
COPY . /src/
WORKDIR /src
RUN CGO_ENABLED=0 go build -o /src/docker_dns

FROM scratch
COPY --from=build /src/docker_dns /bin/docker_dns
CMD ["/bin/docker_dns"]