# shaker - one more http cron manager

## Application config file
```
---
slack:
  enabled: true
  channel: cron
  token: secret-token
jobs:
  http:
    dir: "dist/jobs"
  redis:
    storages:
      default:
        host: 127.0.0.1
        port: 6379
      pubsub:
        host: 127.0.0.1
        port: 6379
    dir: "dist/redis"
users:
  user1:
    user: user1
    password: secret
  user2:
    user: user2
    password: secret
```

* slack - slack configuration
* jobs - http or redis pubsub "publish" jobs
* redis.storages.default - redis for destributed locks
* redis.storages.pubsub - redis for publish to pubsub jobs
* user - user/password for http jobs

## Job configuration

```
{
    "url": "http:/localhost",
    "jobs": [
      {
        "name": "Every minute",
        "cron": "* * * * *",
        "uri": "api/myMethod1"
      },
      {
        "name": "Every minute with 10 second timeout",
        "cron": "* * * * *",
        "uri": "api/myMethod2",
        "timeout": 10
      },
      {
        "name": "Every 4 minutes",
        "cron": "*/4 * * * *",
        "uri": "api/myMethod3"
      },
      {
        "name": "Every day at 1 hour 0 minutes with user user1, method post, content-type header",
        "cron": "0 1 * * *",
        "uri": "api/myMethod4",
        "user": "user1",
        "method": "post",
        "body": {},
        "contentType": "application/json",
        "userAgent": "test-user"
      }
    ]
}
```

# Build

go build 

# For run

export CONFIG="dev/config.yml"
