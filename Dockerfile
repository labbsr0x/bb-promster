FROM golang:alpine AS BUILDER
RUN mkdir /build 
ADD main.go /build/
ADD alert_rules/alert-rules.yml.tmpl /build/
WORKDIR /build
RUN apk update
RUN apk add git
RUN go get go.etcd.io/etcd/clientv3
RUN go get github.com/sirupsen/logrus
RUN go get github.com/spf13/cobra/cobra
RUN go get github.com/spf13/viper
RUN go build -o main .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main .

FROM flaviostutz/promster:latest

ENV ETCD_URL ""
ENV BB_PROMSTER_LEVEL ""
ENV SCRAPE_INTERVAL ""
ENV SCRAPE_TIMEOUT ""
ENV REGISTRY_SERVICE ""
ENV CLEAR_RR ""
ENV ALERT_MANAGER_URLS ""
ENV REMOTE_WRITE_URL ""
ENV ALERT_MANAGER_SCHEME ""
ENV ALERT_RULES_FILE ""

ENV REGISTRY_TTL 15
ENV RETENTION_TIME 2h

COPY --from=BUILDER /build/main /bin/

ADD run.sh /
ADD rules /etc/prometheus/
ADD alert_rules /etc/prometheus
ADD prometheus.yml.tmpl /

RUN chmod 777 /run.sh

ENTRYPOINT [ "/bin/sh" ]
CMD ["-C", "/run.sh"]   