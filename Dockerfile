FROM golang:1.11

WORKDIR /go/src/spawn/
EXPOSE 3100/tcp
EXPOSE 3101/tcp
COPY . .

RUN go install

CMD [ "spawn" ]




