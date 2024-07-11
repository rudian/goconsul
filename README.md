# goconsul

Docker up a consul instance
```dockerfile
docker run --name consul1 -d -p 8500:8500 -p 8300:8300 -p 8301:8301 -p 8302:8302 -p 8600:8600 hashicorp/consul agent -server -bootstrap-expect=1 -ui -bind=0.0.0.0 -client=0.0.0.0
```

### Register Service
If you dont inject consul server address, system will get the default one
```go
//1. Start consul and inject the consul service address
consulService, err := consul.NewService("127.0.0.1:8500")
if err != nil {
    log.Fatal(err)
    return
}

//2. Register a service
err := consulService.RegisterService(
	        consul.RegisterService{
                ServiceName:   "hello_service",
                Address:       "127.0.0.1",
                Port:          3000,
                HeathCheckTTL: 10 * time.Second,
        })
if err != nil {
    log.Fatalln(err)
    return
}
```
### Service Discovery
```go
//1. Start consul and inject the consul service address
consulService, err := consul.NewService("127.0.0.1:8500")
if err != nil {
    log.Fatal(err)
    return
}

//2. Discover a service, and get service address
serviceAddress, err := consulService.GetServiceAddress("hello_service")
if err != nil {
    log.Fatal(err)
    return
}
```