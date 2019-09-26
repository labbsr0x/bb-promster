# Big Brother Promster

This project defines custom [Promster](https://github.com/flaviostutz/promster) images to work with specific metrics defined by the [Big Brother](https://github.com/labbsr0x/big-brother) project.

The currently supported metrics are:

```
request_seconds_bucket{type,status, method, addr, le}
request_seconds_count{type, status, method, addr}
request_seconds_sum{type, status, method, addr}
response_size_bytes{type, status, method, addr}
dependency_up{name}
```

These metrics can be easily generated with **Big Brother's** monitor libraries. Please check them out at the the main project.

# How it works

[Promster](https://github.com/flaviostutz/promster) is a powerfull tool to automatically identify new service instancies to scrape. 

It is highly configurable and one can do with it pretty much everything that can be accomplished with Prometheus.

But the knowledge gap is a bit too steep for a developer with no observability training to start using it properly.

The Big Brother Promster, or just BB Promster, comes to solve this issue by aggregating in one place the needed semantics to correctly monitor your application.

The BB Promster should be used in the context of the Big Brother project, where it is assumed that your service:

1. publishes your metrics at a `/metrics` endpoint;

2. have all the metrics listed above exposed;

3. with the help of our [etcd-registrar](https://github.com/flaviostutz/etcd-registrar) or [etcd-registry](https://github.com/flaviostutz/etcd-registry), registers itself at an etcd cluster for automatic scraping ;

# Federation

The BB Promster also leverages the federation features implemented by Promster (and Prometheus), allowing your observability cluster to scale together with your service instances.

Prometheus federation is the concept of clustering prometheus instances to allow the handling of huge metric loads. 

Ultimately, it ends up in a tree layout with a top-layer, any number of middle layers, and a end layer of prometheus hitting your `/metrics` endpoint directly.

In the context of the Big Brother Project, this top-layer will also be federated by Big Brother, meaning that the `/federate` endpoint needs to be exposed to the public internet.

# Configuration

The BB Promster docker image expects at least two different configurations:

1. **BB_PROMSTER_LEVEL**: defines the level where a Promster instance lays on at your own Promster cluster. Level 1 is for Promster that hit your `/metrics` endpoint directly; Level 2+ is for Promsters that federate on each other. The federation happens for scalability issues. Once you have to scale up your app, things need to work a little bit differently. BB Promster comes to solve that issue;

2. **ETCD_URLS**: defines the etcd cluster urls where service discovery is done for monitoring purposes. Here we assume that scraping instances and Promster instances will all register themselves at the same etcd registry. 

If you have a scenario where you have different etcd clusters for scraping instances and Promster instances, you can leave `ETCD_URLS` empty and define the following env:

1. **REGISTRY_ETCD_URL**: the etcd cluster urls where a Promster instance will register itself for federation;

2. **SCRAPE_ETCD_URL**: the etcd cluster urls where a service instance will register itself for scraping; 

All other configurations from Promster itself and Prometheus are still available for use. We recommend, though, to use them with care and always checking for conflicts with our env resolution logic implemented in `run.sh`.

# Example

This repository also comes with an example. Just go to your terminal and type:

```
> docker-compose up
```

This will lauch 5 services:

1. an etcd registry;

2. a node js express service instrumented with our `express-monitor` lib;

3. a level 1 bb-promster instance that will scrape the exposed metrics at the service's `/metrics` endpoint;

4. a level 2 bb-promster instance that will federate the level 1 bb-promster instances;

5. a level 3 bb-promster instance that will federate the level 2 bb-promster instances;

With this setup you can exercise some scenarios, such as:

1. scaling up your service;

2. scaling up level 1 bb-promster;

3. scaling up level 2 bb-promster;


