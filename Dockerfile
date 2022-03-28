FROM ysicing/god AS god

WORKDIR /go/src

ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.mod

COPY go.sum go.sum

RUN go mod download

COPY . .

RUN make build

FROM ysicing/debian

COPY --from=god /go/src/dist/defaultbackend /bin/defaultbackend

RUN chmod +x /bin/defaultbackend

CMD /bin/defaultbackend
