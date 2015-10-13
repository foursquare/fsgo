# Curator Service Discovery (for Go)
Implement Curator-sytle service discovery in Go -- both server-side registration as well as client-side discovery.


## Server-side

First, create teh service-discovery framework (passing it a connected curator) and call `MaintainRegistrations()` to start its connection monitor:
```go
  s := discovery.NewServiceDiscovery(c, "/path/to/services")
  s.MaintainRegistrations()
```
Then register some services:
```go
  reg := discovery.NewSimpleServiceInstance("baz", "127.0.0.1", 8080)
  ...
  if err := s.Register(reg); err != nil {
    log.Fatal("failed to register: ", err)
  }
  defer s.Unregister(reg)
```


## Client-side
First, create a service-discovery framework (passing it a connected curator) and call `Watch()` to start watching for changes.

Use `Provider(service)` to get a provider, which can be used to get a registered instance (via `GetInstance()`) whenever one is needed.

The default strategy used when creating providers is to randomly select a registed node. Use `ProviderWithStrategy(name, strategy)` to pass a different `ProviderStrategy`, such as the `RoundRobinProvider`.
```go
  s := discovery.NewServiceDiscovery(c, "/path/to/services")
  s.Watch()

  bazProvider := s.Provider("baz")
  ...

  def HandleReq() {
  ...
    hostForThisReq := bazProvider.GetInstance().Spec()
    http.Get(fmt.Sprintf("http://%s/path/to/resource", hostForThisReq))
  ...
  }

```

## Testing
The tests use `zk.StartTestCluster` to get a testing zookeeper instance. This requires you have zookeeper installed locally (it searches a few relative and system paths). On OSX, `brew install zookeeper` is enough to get it working.

## Authors
- [David Taylor](http://github.com/dt)

