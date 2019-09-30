FROM flaviostutz/promster:latest

ENV ETCD_URL ""
ENV BB_PROMSTER_LEVEL ""
ENV SCRAPE_INTERVAL ""
ENV SCRAPE_TIMEOUT ""

ENV REGISTRY_ETCD_BASE /bb-promster
ENV REGISTRY_TTL 15
ENV RETENTION_TIME 2h

ADD run.sh /
ADD rules /etc/prometheus/
ADD prometheus.yml.tmpl /

RUN chmod 777 /run.sh

CMD ["-C", "/run.sh"]