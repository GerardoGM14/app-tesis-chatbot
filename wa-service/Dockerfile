# syntax=docker/dockerfile:1.7

FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY go.su[m] ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/wa-service ./cmd/wa-service

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/wa-service /wa-service
USER nonroot:nonroot
EXPOSE 8001
HEALTHCHECK --interval=15s --timeout=3s --start-period=10s --retries=3 CMD ["/wa-service", "healthcheck"]
ENTRYPOINT ["/wa-service"]
CMD ["serve"]
