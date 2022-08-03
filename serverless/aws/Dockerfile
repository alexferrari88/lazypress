FROM golang AS builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN make build

FROM chromedp/headless-shell:latest

COPY --from=builder /usr/local/bin/app .

CMD ["app"]