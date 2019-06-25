FROM golang:alpine as builder
WORKDIR /go/src/github.com/weibaohui/podInteractive/
COPY . .
RUN ls
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-d -w -s ' -a -installsuffix cgo -o app .
RUN ls

FROM busybox
WORKDIR /app/
COPY --from=builder /go/src/github.com/weibaohui/podInteractive/app .
COPY view ./view/
COPY static ./static/
RUN ls

CMD ["./app","--kubeconfig="]