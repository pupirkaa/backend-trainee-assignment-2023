# syntax=docker/dockerfile:1
FROM golang:1.21-alpine AS build
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o /bin/assignment ./main.go

FROM scratch
COPY create.sql .
COPY --from=build /bin/assignment /bin/assignment
CMD ["/bin/assignment"]


