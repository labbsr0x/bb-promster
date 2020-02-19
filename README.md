# Big Brother Promster
[![Docker Pulls](https://img.shields.io/docker/pulls/labbsr0x/bb-promster.svg)](https://hub.docker.com/r/labbsr0x/bb-promster)

This project defines custom [Promster](https://github.com/flaviostutz/promster) images to work with specific metrics defined by the [Big Brother](https://github.com/labbsr0x/big-brother) project.

The out of the box supported metrics are:

```
request_seconds_bucket{type, status, isError, method, addr, le}
request_seconds_count{type, status, isError, method, addr}
request_seconds_sum{type, status, isError, method, addr}
response_size_bytes{type, status, isError, method, addr}
dependency_up{name}
application_info{version}
```

These metrics can be easily generated with **Big Brother's** monitor libraries. Please check them out at the the [main project](https://github.com/labbsr0x/big-brother).

# How it works

BB Promster is an extension to Flavio Stutz's [Promster](https://github.com/flaviostutz/promster), a powerfull tool to automatically identify new service instances to scrape.

It is highly configurable and one can do with it pretty much anything that can be accomplished with [Prometheus](https://github.com/prometheus/prometheus).

On the other side, the knowledge gap is a bit too steep for a professional with no observability training to start using Promster properly.

The Big Brother Promster, or just BB Promster, comes to solve this issue by aggregating in one place the needed semantics to correctly monitor your application.

The BB Promster should be used in the context of the Big Brother project, where it is assumed that your service:

1. publishes your metrics at a `/metrics` endpoint;

2. have all the big-brother metrics listed above exposed;

3. with the help of our [etcd-registrar](https://github.com/flaviostutz/etcd-registrar) or [etcd-registry](https://github.com/flaviostutz/etcd-registry), registers itself at an etcd cluster for automatic scraping;

![how it works](https://raw.githubusercontent.com/labbsr0x/bb-promster/assets/how-it-works.png "how it works")

# Federation

The BB Promster also leverages the federation features implemented by Promster (and Prometheus), allowing your observability cluster to scale together with your service instances.

Prometheus federation is the concept of clustering prometheus instances to allow the handling of huge metric loads. 

Ultimately, it ends up in a tree layout with a top-layer, any number of middle layers, and an end layer of prometheus hitting your `/metrics` endpoint directly.

# Configuration

## What BB Promster expects
The BB Promster docker image expects at least four different configurations:

1. **BB_PROMSTER_LEVEL**: defines the level where a Promster instance lays on at your own Promster cluster topology. Level 1 is for BB-Promsters that hit your `/metrics` endpoint directly; Level 2+ is for Promsters that federate on each other. The federation happens for scalability issues. Once you have to scale up your app, things need to work a little bit differently. BB-Promster comes to solve those issues setting up by the default the appropriate recording rules;

2. **ETCD_URLS**: defines the etcd cluster urls where service discovery is done for monitoring purposes. Here we assume that scraping instances and Promster instances will all register themselves at the same etcd registry. **Important: all registered IPs or addresses must provide only the host name, without schema or paths. The metric paths and schema should be configured by other environment variables**;

3. **REGISTRY_ETCD_BASE**: defines in which base path of the ETCD is grouped all the components to observe a specific application;

4. **REGISTRY_SERVICE**: defines the name of the service you are observing as defined at the appropriate ETCD record;

5. **SCRAPE_ETCD_PATH**: this information tells level 1 BB-Promsters where to find the targets IP addresses at the provided ETCD installation. **Important: Mandatory only for level 1 BB-Promsters**;

## Multiple ETCDs
If you have a scenario where you have different ETCD clusters, one for registering Scraping instances and other for registering Promster instances, you can leave `ETCD_URLS` empty and define the following ENVs:

1. **REGISTRY_ETCD_URL**: the etcd cluster urls where a Promster instance will register itself for federation;

2. **SCRAPE_ETCD_URL**: the etcd cluster urls where a service instance will register itself for scraping; 

## Other Optional Configs

1. **TLS_INSECURE**: informs Prometheus to ignore TLS verification;

2. **SCHEME**: `http` (default) or `https`. Configure your level 1 BB-Promsters if your targets are only exposed at `https` endpoints and do not have automatic redirection from `http` to `https`;

3. **SCRAPE_PATHS**: if your metrics path does not follow the default `/metrics`, you'll need to configure this variable to point to the exact path where your metrics are exposed;

4. **CLEAR_RR**: the recording rules can be all removed if you wish. Just set the `CLEAR_RR` env to `true` and `bb-promster` will have it's set of recording rules deleted; 

5. **ALERT_MANAGER_URLS**: if you have a configured alertmanager at your disposal, you can set BB Promster to leverage it by using the provided `ALERT_MANAGER_URLS` environment variable. Only Level 1 BB Promster will have the alerting rules installed and the alert manager urls properly configured. This is to disable redundant alerting;

All other configurations from [Promster](https://github.com/flaviostutz/promster) itself and Prometheus are still available for use. We recommend, though, to use them with care and always checking for conflicts with our env resolution logic implemented in `run.sh`.

# Example

This repository also comes with an example. Just go to your terminal and type:

```
> docker-compose up
```

This will lauch 4 services:

1. an etcd registry;

2. two metrics generator services with IPs registered at the `/metrics-generator/` ETCD path;

3. two a level 1 bb-promster instances that will scrape the exposed metrics at the services' `/metrics` endpoint;

4. two level 2 bb-promster instances that will federate the right level 1 bb-promster instances;

With this setup you can exercise some scenarios, such as:

1. scaling up your service;

2. scaling up level 1 bb-promster;

# Cortex

A cortex example was added to the `cortex` folder as a proof-of-concept.

To experiment with it, you need to have a valid/configured cassandra instance running on your host. You can run it by executing:

```bash
docker run -d --name cassandra --rm -p 9042:9042 cassandra:3.11
```

Wait a bit for Cassandra to normalize (usually `30s`). After that, configure your cortex cassandra `KEYSPACE` by first entering a valid `CQLSH` session with `docker exec -it cassandra cqlsh` and then executing:

```sql
CREATE KEYSPACE cortex WITH replication = {'class':'SimpleStrategy', 'replication_factor' : 1};
```

After successfully configuring your cassandra, you can just `docker-compose up -d` from the `cortex` folder.
 
The cortex startup time can be really slow due to Ingester ring synchronization routines, so you should wait a bit (from 2 minutes up until 10).

You will know cortex is up and running when hitting `http://localhost:9001/ring` at your browser lists the available Ingesters.

You can also check your cortex logs for the following message:

```log
level=info ts=2020-02-19T12:51:33.3486187Z caller=main.go:100 msg="Starting Cortex" version="(version=, branch=, revision=)"
``` 

After everything is up and running, go to your web browser at `http://localhost:3000`, add the Prometheus Cortex datasource (`http://cortex1:9009/api/prom`) and import our Big Brother grafana dashboard with ID `11544`.

You should see something like the following:

![Health Overview](https://raw.githubusercontent.com/labbsr0x/bb-grafana/assets/screenshot.jpg "Health Overview")

# Big Brother

This is part of a more large application called [Big Brother](https://github.com/labbsr0x/big-brother).


