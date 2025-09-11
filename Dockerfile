# Build FE
FROM node:20-alpine AS fe-build

WORKDIR /app

COPY ./frontend/package*.json ./

RUN npm install

COPY ./frontend/ .

RUN npm run build

# Build BE
FROM golang:alpine AS be-build

WORKDIR /app

COPY ./backend/ ./

RUN go get -d -v ./...

RUN CGO_ENABLED=0 go build -o /bin/app ./cmd/

# Run
FROM gcr.io/distroless/static

COPY --from=be-build /bin/app /
COPY --from=fe-build /app/dist /frontend

WORKDIR /
CMD ["/app"]
