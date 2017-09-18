FROM golang:1.9

WORKDIR /go/src/app/
COPY . .
RUN go get ... && go install

CMD ["/go/src/app/entrypoint.sh"]
