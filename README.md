#### This contains the code for both the first and second task
#### I have created a tinyURL like service using redis-server as the key value store

### Steps

-  get driver for communicating with the redis-server. I will be using gomodule/redigo. This can be installed using 
> go get github.com/gomodule/redigo/redis

- run the redis-server using the default port 
- run the server file using
> go run server.go

- add the url using curl -- replace {absolute-url} with your own absolute url
> curl -X POST -d '{"url":"{absolute-url}"}' localhost:8000/add

- retrieving url -- replace {tiny-url} with the tiny url you received
> curl -i localhost:8000/{tiny-url}

- if you use browser to connect instead of curl in the last step, you will be automatically redirected

#### @TODO use go modules
#### the onlu external dependency is the redis-driver. rest has been implemented using the standard library
#### the code is commented for additional information

