# Rate Limiter

An HTTP service that limits the number of requests from a single IPv4 subnet. It produces the same static content until the subnet reaches the limit. If limit is reached - 429 Too Many Requests.
IPs are extracted from the header X-Forwarded-For.

Rate Limiter's host, subnet's prefix length, number of requests, limit time for requests and cooldown time can be defined in the config file or environment variables.

Example of config.yaml:
```sh
rateLimiterHost: :8083
subnetPrefixLength: 24
timeCooldownSec: 60
rateLimits:
  requestsCount: 100
  timeSec: 60
```

To run service in docker container:
```sh
docker-compose up --build
```
URL: http://localhost:8080
The request should be with method "GET" and header "X-Forwarded-For".