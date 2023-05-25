FROM golang:1.20.4 as builder

ENV PATH $PATH:/usr/local/go/bin
ENV GOROOT=/usr/local/go
ENV GOPATH=$HOME/gowork
RUN mkdir $HOME/gowork
ENV CGO_LDFLAGS_ALLOW="-Wl,-rpath.*"
ENV GO111MODULE="on"
ENV GIT_SSL_NO_VERIFY=true

ENV PATH="${PATH}:/opt/mqm"
ENV PATH="${PATH}:/opt/mqm/inc"
ENV PATH="${PATH}:/opt/mqm/lib64"

ADD https://ibm.biz/IBM-MQC-Redist-LinuxX64targz /opt/mqm/
RUN tar -xvf /opt/mqm/IBM-MQC-Redist-LinuxX64targz -C /opt/mqm/
RUN rm -rf /opt/mqm/IBM-MQC-Redist-LinuxX64targz

RUN git clone https://github.com/felix-lessoer/qbeat.git ${GOPATH}/src/github.com/felix-lessoer/qbeat

WORKDIR $GOPATH/src/github.com/felix-lessoer/qbeat
RUN go get -u
RUN go build

FROM debian:bullseye-slim

ENV PATH="${PATH}:/opt/mqm/inc"
ENV PATH="${PATH}:/opt/mqm/lib64"

COPY --from=builder /opt/mqm /opt/mqm
COPY --from=builder /gowork/src/github.com/felix-lessoer/qbeat/qbeat /usr/share/qbeat/qbeat
COPY --from=builder /gowork/src/github.com/felix-lessoer/qbeat/qbeat.yml /usr/share/qbeat/qbeat.yml

WORKDIR /usr/share/qbeat
ENTRYPOINT ["/usr/share/qbeat/qbeat"]