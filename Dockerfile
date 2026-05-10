ARG BUILD_HASH="0000000000000000000000000000000000000000"
ARG BUILD_DATE="YYMMDD"
ARG BUILD_ID="r0"

FROM library/golang:1.26 AS base

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0



FROM base AS test

RUN --mount=type=cache,target=/root/.cache/go-build \
	go test -v -trimpath ./...



FROM base AS build

ARG BUILD_HASH
ARG BUILD_DATE
ARG BUILD_ID
RUN BUILD_HASH=${BUILD_HASH} \
	BUILD_DATE=${BUILD_DATE} \
	BUILD_ID=${BUILD_ID} \
	./scripts/gen-version-file.sh

RUN --mount=type=cache,target=/root/.cache/go-build \
	go build \
		-trimpath \
		-ldflags="-s -w" \
		-o a . \
	&& ./a version



FROM scratch AS app

COPY --from=build /app/a /tegra-exporter
ENTRYPOINT ["/tegra-exporter"]
