FROM golang:alpine
RUN apk add --no-cache git &&\
go get "github.com/gorilla/mux" &&\
go get "github.com/google/uuid"
ADD /src/ /app/
WORKDIR /app/
COPY resources/* /app/resources/
RUN go build -o main .
CMD ["/app/main"]

