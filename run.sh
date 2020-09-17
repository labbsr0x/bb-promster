#!/bin/sh
set -e
set -x

if [[ "$ETCD_URLS" != "" ]]; then
    export REGISTRY_ETCD_URL=$ETCD_URLS
    export SCRAPE_ETCD_URL=$ETCD_URLS
fi

if [[ "$REGISTRY_ETCD_URL" == "" ]]; then
    echo "REGISTRY_ETCD_URL must NOT be empty" 1>&2
    exit 1
fi

if [[ "$SCRAPE_ETCD_URL" == "" ]]; then
    echo "SCRAPE_ETCD_URL must NOT be empty" 1>&2
    exit 2
fi

if [[ "$REGISTRY_ETCD_BASE" == "" ]]; then
    echo "REGISTRY_ETCD_BASE must NOT be empty" 1>&2
    exit 3
fi

if [[ "$REGISTRY_SERVICE" == "" ]]; then
    echo "REGISTRY_SERVICE must NOT be empty" 1>&2
    exit 3
fi

if [[ "$BB_PROMSTER_LEVEL" == "" ]]; then
    echo "BB_PROMSTER_LEVEL must NOT be empty" 1>&2
    exit 4
fi

if [[ $((BB_PROMSTER_LEVEL)) -lt 1 ]]; then
    echo "BB_PROMSTER_LEVEL must be a valid integer greater than 0" 1>&2
    exit 5
fi


ll=$((BB_PROMSTER_LEVEL - 1))
if [[ $ll -ne 0 ]]; then # if true, we properly configure bb-promster to behave like a level-n promster
    export SCRAPE_MATCH_REGEX="l$ll" # configure federation to look for (BB_PROMSTER_LVEL - 1) metrics
    export SCRAPE_ETCD_PATH="${REGISTRY_ETCD_BASE}-promster-${SCRAPE_MATCH_REGEX}/${REGISTRY_SERVICE}" # configure where to look for federation targets
    export SCRAPE_PATHS="/federate" # path to federate
else
    if [[ "$SCRAPE_ETCD_PATH" == "" ]]; then 
        echo "SCRAPE_ETCD_PATH cannot be empty for a Level 1 BB-PROMSTER instance" 1>&2
        exit 6
    fi
    if [[ "$SCRAPE_PATHS" == "" ]]; then 
        # defaults to /metrics when not set by user
        export SCRAPE_PATHS="/metrics"
    fi
fi

if [[ "$SCRAPE_INTERVAL" == "" ]]; then
    export SCRAPE_INTERVAL="$((BB_PROMSTER_LEVEL * 30))s"
fi

if [[ "$SCRAPE_TIMEOUT" == "" ]]; then
    export SCRAPE_TIMEOUT="$((BB_PROMSTER_LEVEL * 15))s"
fi

if [[ "$ALERT_MANAGER_SCHEME" == "" ]]; then
    export ALERT_MANAGER_SCHEME="https"
fi

sed -i -e 's/$BB_PROMSTER_LEVEL/'"l${BB_PROMSTER_LEVEL}"'/g' "/etc/prometheus/rules-ln.yml"

# We need to register the BB-Promster in a different etcd base then the one informed by the user
# If we don't do this, PRSN will get overwritten in a federation and we'll loose that information
export REGISTRY_ETCD_BASE="${REGISTRY_ETCD_BASE}-promster-l${BB_PROMSTER_LEVEL}"

# with the CLEAR_RR env, the user has the option to clear the recording rules
if [[ "$CLEAR_RR" == "true" ]]; then 
    rm /etc/prometheus/rules-l1.yml
    rm /etc/prometheus/rules-ln.yml
fi

sed -i -e 's;$REMOTE_WRITE_URL;'"${REMOTE_WRITE_URL}"';g' "/prometheus.yml.tmpl";

if [[ "$ALERT_RULES_FILE" == "" ]]; then
    export ALERT_RULES_FILE="/etc/prometheus/alert-rules.yml"
fi

if [[ "$ALERT_MANAGER_URLS" != "" ]]; then
    sed -i -e 's;$ALERT_RULES_FILE;'"- $ALERT_RULES_FILE"';g' "/prometheus.yml.tmpl"
	sed -i "/alerting/,/targets: /d" "/prometheus.yml.tmpl"
	cat >> "/prometheus.yml.tmpl" <<- EOM

alerting:
  alertmanagers:
  - scheme: $ALERT_MANAGER_SCHEME
    static_configs:
    - targets: ['$ALERT_MANAGER_URLS']
EOM
	
else
    sed -i -e 's/$ALERT_RULES_FILE/'""'/g' "/prometheus.yml.tmpl"
fi

# the user can also configure a remote url to send prometheus metrics to 
if [[ "$REMOTE_WRITE_URL" != "" ]]; then
    cat >> "/prometheus.yml.tmpl" <<- EOM

remote_write:
  - url: $REMOTE_WRITE_URL
EOM
fi

sh /startup.sh & # inherited from flaviostutz/promster 
/bin/main
