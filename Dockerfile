FROM golang AS build

WORKDIR /usr/src
COPY . .
RUN rm -f terratag && go build -ldflags "-linkmode external -extldflags -static"

FROM hashicorp/terraform:light
COPY --from=build /usr/src/terratag /bin/terratag
ENTRYPOINT ["/bin/terratag"]
