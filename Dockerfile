FROM golang:1.9

WORKDIR $GOPATH/src/app/
COPY . .
RUN go get ... && go install

CMD $GOPATH/src/app/entrypoint.sh
