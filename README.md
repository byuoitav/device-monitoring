# device-monitoring-microservice
A microservice that pings all the devices in a given room (according to the ```configuration-database-microservice```). The building and room are specified based on the ```PI_HOSTNAME``` environment variable, e.g. ```ITB-1101-CP1```. After each device is pinged, a heartbeat event is sent to a Logstash shipper which must be specified by the ```ELASTIC_API_EVENTS``` environment variable.

## Building

To build the microservice, you can use the provided Makefile. The default target is `all`, which builds the web and local components. You can also build the binaries for multiple platforms by running:

```bash
make build-binaries
```

This will create binaries for the specified platforms in the makefile. This is useful if you want to deploy the microservice on different architectures (e.g., Linux, Windows, macOS).

The `deploy` target can be used to build and package the microservice for deployment. It creates a tarball containing the binaries and other necessary files. This also can be uploaded to the database for deployment.


```bash
make deploy
```

## Go Project dependencies

The project on its current state does not use the tagged version of some packages. To install the dependencies directly from the `byuoitav` repo, you can use the following command:

```bash
go mod edit -require=github.com/byuoitav/shipwright@DMM-Prod-Working
go mod tidy
```

This will ensure that the project uses the correct version of the `shipwright` package, which is necessary for the microservice to function correctly.

The `go.mod` file also needed the following changes to work with the `shipwright` package:

```go
replace github.com/byuoitav/wso2services => /home/user/Desktop/wso2services

replace github.com/byuoitav/endpoint-authorization-controller => /home/user/Desktop/endpoint-authorization-controller
```
This packages are not available in the public repositories, so you need to replace them with the local paths where you have them cloned.

