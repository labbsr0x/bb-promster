# Big Brother Promster

This project defines custom [Promster](https://github.com/flaviostutz/promster) images to work with specific metrics defined by the [Big Brother](https://github.com/labbsr0x/big-brother) project.

The currently supported metrics are:

```
request_seconds_bucket{type,status, method, addr, isError, version, le}
request_seconds_count{type, status, method, addr, isError, version}
request_seconds_sum{type, status, method, addr, isError, version}
response_size_bytes{type, status, method, addr, isError, version}
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

3. **SCRAPE_PATHS**: and, if your metrics path does not follow the default `/metrics`, you'll need to configure this variable to point to the exact path where your metrics are exposed;

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



