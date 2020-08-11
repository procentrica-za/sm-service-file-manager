FROM golang:alpine
RUN apk add --no-cache git &&\
go get "github.com/gorilla/mux" &&\
go get "github.com/google/uuid"
COPY /src/resources/ /app/imageResources/
ADD /src/ /app/
WORKDIR /app/
RUN go build -o main .
CMD ["/app/main"]

