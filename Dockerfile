FROM ysicing/god AS god

COPY . /go/src

WORKDIR /go/src

RUN make build

FROM ysicing/debian

COPY --from=god /go/src/dist/defaultbackend /bin/defaultbackend

RUN chmod +x /bin/defaultbackend

CMD /bin/defaultbackend
