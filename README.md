# device-monitoring-microservice
The Device Monitoring Microservice houses functionality to monitor device information for a given room, a dashboard to visualize that information, and API's for divider sensor integration.

Pings all the devices in a given room (according to the ```configuration-database-microservice```). The building and room are specified based on the ```PI_HOSTNAME``` environment variable, e.g. ```ITB-1101-CP1```. After each device is pinged, a heartbeat event is sent to a Logstash shipper which must be specified by the ```ELASTIC_API_EVENTS``` environment variable.

<img width="799" height="480" alt="image" src="https://github.com/user-attachments/assets/ac593320-4dc1-4c73-b4ad-1e06dd55e267" />

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

## API Endpoints

| Method | Path | Handler / Notes |
| --- | --- | --- |
| GET | /dash | Redirects to `/dashboard` |
| GET | / | Redirects to `/dashboard` |
| GET | /dashboard | Dashboard App |
| GET | /dashboard/* | SPA fallback to `dashboard/index.html` |
| GET | /ping | Health check (plain text) |
| GET | /device | Returns basic device info like hostname, ip, etc |
| GET | /device/hostname | Returns the device hostname ex: ITB-1106-CP2 |
| GET | /device/id | Returns the device id ex: ITB-1106-CP2 |
| GET | /device/ip | Returns the device ip ex: 0.0.0.0 |
| GET | /device/network | Returns a boolean indicating if connected to internet |
| GET | /device/dhcp | returns two booleans for if DHCP is enabled or toggleable  |
| GET | /device/screenshot | Returns a screenshot of the device display |
| GET | /device/hardwareinfo | Returns hardware information of the device |
| PUT | /device/health | Returns the health status of the device services |
| GET | /room/ping | Pings all devices in the room |
| GET | /room/state | Returns the current state of the room for each display and audioDevice |
| GET | /room/activesignal | Returns booleans for each display indicating if it has an active signal |
| GET | /room/hardwareinfo | Returns hardware information of the room |
| GET | /room/health | Returns the health status of the room |
| PUT | /device/reboot | Reboots the device |
| PUT | /device/dhcp/:state | Sets the DHCP state of the device |
| POST | /event | Sends an event |
| GET | /divider/state | returns the status of the divider sensor {"connected":[],"disconnected":[""]}  |
| GET | /divider/preset/:hostname | Returns the preset for a given hostname |
| GET | /divider/pins/:systemID | Returns the divider pins for a given system ID |
| GET | /actions | Action manager info (Echo handler adapter) |
| GET | /actions/trigger/:trigger | Action manager trigger config (Echo handler adapter) |
| GET | /dns | Flushes the DNS cache |
| GET | /resyncDB | Resyncs the database |
| GET | /refreshContainers | Refreshes the containers |
| GET | /api/v1/monitoring | Returns device health information |
