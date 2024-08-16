        # Build environment
        FROM golang:alpine AS builder
        WORKDIR /go/src/app

        COPY go.mod ./
        RUN go mod download

        COPY . .        
        RUN go build -o /go-irr-server main.go

        # Runtime environment
        FROM alpine:edge AS runtime

        # Add testing repository (for bgpq4)
        RUN echo "@testing http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories
        RUN apk add --no-cache bgpq4@testing
        COPY --from=builder /go-irr-server /go-irr-server
        CMD ["/go-irr-server"]          
        EXPOSE 8080