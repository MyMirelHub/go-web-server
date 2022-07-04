# WebServer

A dummy app which exposes an HTTP endpoint which, given an input string, returns the amount of characters appearing at least 2 times in it.

It consists of a 
- go web server 
- some basic prometheus instrumentation
- docker-compose file where it is served by traefik
- gh actions build pipeline.

## Quick Run
To run the application, you can use the following command:
- `docker-compose up -build`

Send through a PUT request to the following endpoint:
- `curl -X PUT -id "test11112222" -H Host:kata.docker.localhost http://127.0.0.1` 

And you should get the following response:

```
HTTP/1.1 200 OK
Content-Length: 12
Content-Type: text/plain; charset=utf-8
Date: Thu, 09 Jun 2022 19:55:35 GMT

1 = 4
2 = 4
```