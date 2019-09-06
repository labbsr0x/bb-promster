# Big Brother Promster

This project defines custom [Promster](https://github.com/flaviostutz/promster) images to work with specific metrics defined by the Big Brother project.

The currently supported metrics are:

```
http_requests_second{status, method, url, le}
http_requests_second_count{status, method, url}
http_requests_second_sum{status, method, url}
http_response_size_bytes{status, method, url}
dependency_up{name}
```

They can be easily generated with our monitor libraries. For now, we only have a [Node JS Express library](https://github.com/labbsr0x/express-monitor), but we'll soon support other stacks.

# How it works

[Promster](https://github.com/flaviostutz/promster) is a powerfull tool to automatically identify new service instancies to scrape. 

It is highly configurable and one can do with it pretty much everything that can be accomplished with Prometheus.

But the knowledge gap is a bit too steep for a developer with no observability training to start using it properly.

The Big Brother Promster, or just BB Promster, comes to solve this issue by aggregating in one place the needed semantics to correctly monitor your application.

The BB Promster should be used in the context of the Big Brother project, where it is assumed that your service:

1. publishes your metrics at a `/metrics` endpoint;

2. have all the metrics listed above exposed;

3. registers itself at an etcd cluster for automatic scraping;

# Configuration

The BB Promster docker image expects at least two different configurations:

1. **BB_PROMSTER_LEVEL**: defines the level where a Promster instance lays on at your own Promster cluster. Level 1 is for Promster that hit your `/metrics` endpoint directly; Level 2+ is for Promsters that federate on each other. The federation happens for scalability issues. Once you have to scale up your app, things need to work a little bit differently. BB Promster comes to solve that issue;

2. **ETCD_URLS**: defines the etcd cluster urls where service discovery is done for monitoring purposes. Here we assume that scraping instances and Promster instances will all register themselves at the same etcd registry. 

If you have a scenario where you have different etcd clusters for scraping instances and Promster instances, you can leave `ETCD_URLS` empty and define the following env:

1. **REGISTRY_ETCD_URL**: the etcd cluster urls where a Promster instance will register itself for federation;

2. **SCRAPE_ETCD_URL**: the etcd cluster urls where a service instance will register itself for scraping; 

All other configurations from Promster itself and 



