        FROM alpine:edge

        # Add testing repository (for bgpq4)
        RUN echo "@testing http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories

        # Update and install required packages                
        RUN apk update
        RUN apk add --no-cache bgpq4@testing

        # Install go and build main.go
        RUN apk add --no-cache go


        WORKDIR /go-irr
        COPY . .

        # Install dependencies
        RUN go get github.com/gin-gonic/gin

        RUN go build -o /go-irr-server main.go

        # Clean up
        RUN apk del go
        RUN rm -rf /var/cache/apk/*

        # Start go-irr
        CMD ["/go-irr-server"]

        # Expose port 8080
        EXPOSE 8080