ARG GO_VERSION=1.26.2

FROM golang:${GO_VERSION}-bookworm AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG GO_BUILD_FLAGS="-trimpath"
ARG GO_LDFLAGS="-s -w"

RUN CGO_ENABLED=0 go build ${GO_BUILD_FLAGS} -ldflags "${GO_LDFLAGS}" -o /out/dsl ./cmd/dsl-cli
RUN CGO_ENABLED=0 go build ${GO_BUILD_FLAGS} -ldflags "${GO_LDFLAGS}" -tags dsl_lsp -o /out/dsl-full ./cmd/dsl-cli
RUN CGO_ENABLED=0 go build ${GO_BUILD_FLAGS} -ldflags "${GO_LDFLAGS}" -o /out/dsl-lsp ./cmd/dsl-lsp

FROM scratch

COPY --from=build /out/dsl /usr/local/bin/dsl
COPY --from=build /out/dsl-full /usr/local/bin/dsl-full
COPY --from=build /out/dsl-lsp /usr/local/bin/dsl-lsp

WORKDIR /workspace
USER 65532:65532

CMD ["/usr/local/bin/dsl", "--help"]
