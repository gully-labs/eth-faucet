# node:22 bundles yarn 1; node:lts-alpine (Node 24+) no longer does, so a fresh pull fails at `yarn install`.
FROM node:22-alpine as frontend

WORKDIR /frontend-build

COPY web/package.json web/yarn.lock ./
RUN yarn install

COPY web ./
RUN yarn build

FROM golang:1.17-alpine as backend

RUN apk add --no-cache gcc musl-dev linux-headers

WORKDIR /backend-build

COPY go.* ./
RUN go mod download

COPY . .
COPY --from=frontend /frontend-build/dist web/dist

RUN go build -o eth-faucet -ldflags "-s -w"

FROM alpine

RUN apk add --no-cache ca-certificates

COPY --from=backend /backend-build/eth-faucet /app/eth-faucet

EXPOSE 8080

ENTRYPOINT ["/app/eth-faucet"]
