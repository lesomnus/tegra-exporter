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

RUN --mount=type=cache,target=/root/.cache/go-build \
	go build \
		-trimpath \
		-ldflags="-s -w" \
		-o a . \
	&& ./a version



FROM scratch AS app

COPY --from=build /app/a /tegra-exporter
ENTRYPOINT ["/tegra-exporter"]
