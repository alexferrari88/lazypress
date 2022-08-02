FROM golang AS builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN go build -v -o /usr/local/bin/lazypress cmd/main.go

# Expose default server's port
EXPOSE 3444

FROM chromedp/headless-shell:latest

COPY --from=builder /usr/local/bin/lazypress .

ENTRYPOINT ["./lazypress"]