FROM golang:1.24.2-alpine AS build
WORKDIR /app
COPY ./docs/ ./docs/
COPY ./env/ ./env/
COPY ./pkg/ ./pkg/
COPY ./route/ ./route/
COPY ./store/ ./store/
COPY ./go.mod ./
COPY ./go.sum ./
COPY ./main.go ./
RUN go mod download
RUN go build -o /app.bin

FROM golang:1.24.2-alpine
WORKDIR /app
COPY ./views/ ./views/
COPY --from=build /app.bin /app.bin
ENTRYPOINT ["/app.bin"]
# CMD ["/bin/sh", "-c", "/app.bin >> /app/log/eseed.log 2>&1"]