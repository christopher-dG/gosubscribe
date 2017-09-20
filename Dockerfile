FROM golang:1.9

WORKDIR $GOPATH/src/app/
COPY . .
RUN go get ... && go install

ENTRYPOINT ["/go/src/app/entrypoint.sh"]
