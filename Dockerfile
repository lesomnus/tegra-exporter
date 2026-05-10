FROM library/golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 \
	go build \
		-trimpath \
		-ldflags="-s -w" \
		-o a . \
	&& ./a --help



FROM scratch

COPY --from=builder /app/a /tegra-exporter
ENTRYPOINT ["/tegra-exporter"]
