import {
  JsonObject,
  JsonProperty,
  Any,
  JsonConverter,
  JsonCustomConvert
} from "json2typescript";

@JsonObject("Status")
export class Status {
  @JsonProperty("name", String)
  name: string = undefined;

  @JsonProperty("bin", String)
  bin: string = undefined;

  @JsonProperty("statuscode", String)
  statuscode: string = undefined;

  @JsonProperty("version", String)
  version: string = undefined;

  @JsonProperty("uptime", String)
  uptime: string = undefined;

  @JsonProperty("info", Any, true)
  info: any = undefined;
}

@JsonObject("DHCPInfo")
export class DHCPInfo {
  @JsonProperty("error", Any, true)
  error: any = undefined;

  @JsonProperty("enabled", Boolean, true)
  enabled: boolean = undefined;

  @JsonProperty("toggleable", Boolean, true)
  toggleable = false;
}

@JsonObject("DeviceInfo")
export class DeviceInfo {
  @JsonProperty("hostname", String)
  hostname: string = undefined;

  @JsonProperty("id", String)
  id: string = undefined;

  @JsonProperty("internet-connectivity", Boolean, true)
  internetConnectivity: boolean = undefined;

  @JsonProperty("ip", String, true)
  ip: string = undefined;

  @JsonProperty("status", [Status], true)
  status: Status[] = Array<Status>();

  @JsonProperty("dhcp", DHCPInfo, true)
  dhcp: DHCPInfo = undefined;

  @JsonProperty("error", Any, true)
  error: any = undefined;
}

@JsonObject("DevicePingResult")
class DevicePingResult {
  @JsonProperty("deviceID", String)
  deviceID: string = undefined;

  @JsonProperty("error", String)
  error: string = undefined;

  @JsonProperty("packets-received", Number)
  packetsReceived: number = undefined;

  @JsonProperty("packets-sent", Number)
  packetsSent: number = undefined;

  @JsonProperty("packet-loss", Number)
  packetLoss: number = undefined;

  @JsonProperty("ip", String)
  ip: string = undefined;

  @JsonProperty("address", String)
  address: string = undefined;

  @JsonProperty("average-round-trip", String)
  averageRoundTrip: string = undefined;
}

@JsonObject("PingResult")
export class PingResult {
  @JsonProperty("successful", [DevicePingResult], true)
  successful: DevicePingResult[] = Array<DevicePingResult>();

  @JsonProperty("unsuccessful", [DevicePingResult], true)
  unsuccessful: DevicePingResult[] = Array<DevicePingResult>();
}

@JsonConverter
class DateConverter implements JsonCustomConvert<Date> {
  serialize(date: Date): any {
    function pad(n) {
      return n < 10 ? "0" + n : n;
    }

    return (
      date.getUTCFullYear() +
      "-" +
      pad(date.getUTCMonth() + 1) +
      "-" +
      pad(date.getUTCDate()) +
      "T" +
      pad(date.getUTCHours()) +
      ":" +
      pad(date.getUTCMinutes()) +
      ":" +
      pad(date.getUTCSeconds()) +
      "Z"
    );
  }

  deserialize(date: any): Date {
    if (date == null) {
      return undefined;
    }

    return new Date(date);
  }
}

@JsonObject("Trigger")
export class Trigger {
  @JsonProperty("type", String)
  tType: String = undefined;

  @JsonProperty("at", String, true)
  at: String = undefined;

  @JsonProperty("every", String, true)
  every: String = undefined;

  @JsonProperty("match", Any, true)
  match: any = undefined;

  constructor(tType: string) {
    if (tType === undefined) {
      return;
    }
  }
}

@JsonObject("RunnerInfo")
export class RunnerInfo {
  @JsonProperty("id", String)
  id: string = undefined;

  @JsonProperty("trigger", Trigger)
  trigger = new Trigger(undefined);

  @JsonProperty("context", Any, true)
  context: any = undefined;

  @JsonProperty("last-run-start-time", DateConverter, true)
  lastRunTime: Date = undefined;

  @JsonProperty("last-run-duration", String, true)
  lastRunDuration: string = undefined;

  @JsonProperty("last-run-error", String, true)
  lastRunError: String = undefined;

  @JsonProperty("currently-running", Boolean, true)
  currentlyRunning: boolean = undefined;

  @JsonProperty("run-count", Number)
  runCount: number = undefined;
}
