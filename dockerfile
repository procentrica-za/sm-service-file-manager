FROM golang:alpine
RUN apk add --no-cache git &&\
go get "github.com/gorilla/mux"
ADD /src/ /app/
WORKDIR /app/
RUN go build -o main .
CMD ["/app/main"]

