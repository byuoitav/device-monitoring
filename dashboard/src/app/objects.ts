import { JsonObject, JsonProperty } from "json2typescript";

@JsonObject("deviceInfo")
export class DeviceInfo {
  @JsonProperty("hostname", String)
  hostname: string = undefined;

  @JsonProperty("id", String)
  id: string = undefined;

  @JsonProperty("internet-connectivity", Boolean)
  internetConnectivity: boolean = undefined;

  @JsonProperty("ip", String)
  ip: string = undefined;
}
