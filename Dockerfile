FROM golang:1.11

WORKDIR /go/src/github.pie.apple.com/privatecloud/spawn/
EXPOSE 3100/tcp
EXPOSE 3101/tcp
COPY . .

RUN go install

CMD [ "spawn" ]




