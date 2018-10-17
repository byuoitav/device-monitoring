import { JsonObject, JsonProperty, Any } from "json2typescript";

@JsonObject("MStatus")
class MStatus {
  @JsonProperty("name", String)
  name: string = undefined;

  @JsonProperty("bin", String)
  bin: string = undefined;

  @JsonProperty("statuscode", String)
  statuscode: string = undefined;

  @JsonProperty("version", String)
  version: string = undefined;

  @JsonProperty("info", Any, true)
  info: any = undefined;
}

@JsonObject("DeviceInfo")
export class DeviceInfo {
  @JsonProperty("hostname", String)
  hostname: string = undefined;

  @JsonProperty("id", String)
  id: string = undefined;

  @JsonProperty("internet-connectivity", Boolean)
  internetConnectivity: boolean = undefined;

  @JsonProperty("ip", String)
  ip: string = undefined;

  @JsonProperty("mstatus", [MStatus], true)
  mstatus: MStatus[] = Array<MStatus>();
}
