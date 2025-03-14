FROM golang:1.24 AS build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 CGOOS=linux GOARCH=amd64 go build -o serverrl cmd/main.go

FROM scratch
WORKDIR /app
COPY --from=build /app/serverrl .
COPY --from=build /app/.env .
ENTRYPOINT ["./serverrl"]
EXPOSE 8080