import { Injectable } from "@angular/core";
import { HttpClient } from "@angular/common/http";
import { JsonConvert, OperationMode, ValueCheckingMode } from "json2typescript";

import { DeviceInfo, Status, PingResult } from "../objects";

@Injectable({
  providedIn: "root"
})
export class APIService {
  public theme = "default";

  private jsonConvert: JsonConvert;
  private urlParams: URLSearchParams;

  constructor(private http: HttpClient) {
    this.jsonConvert = new JsonConvert();
    this.jsonConvert.ignorePrimitiveChecks = false;

    this.urlParams = new URLSearchParams(window.location.search);
    if (this.urlParams.has("theme")) {
      this.theme = this.urlParams.get("theme");
    }
  }

  public switchToUI() {
    window.location.assign("http://" + window.location.hostname + ":8888/");
  }

  public refresh() {
    window.location.reload(true);
  }

  public switchTheme(name: string) {
    console.log("switching theme to", name);

    this.theme = name;
    this.urlParams.set("theme", name);
    window.history.replaceState(
      null,
      "System Health Dashboard",
      window.location.pathname + "?" + this.urlParams.toString()
    );
  }

  public async reboot() {
    try {
      const data = await this.http
        .put("device/reboot", {
          responseType: "text"
        })
        .toPromise();
    } catch (e) {
      // bug where responseType doesn't actually work
      if (e.status === 200) {
        console.log(e.error.text);
        return e.error.text;
      }

      throw new Error("error getting rebooting device: " + e);
    }
  }

  public async getDeviceInfo() {
    try {
      const data = await this.http.get("device").toPromise();
      const deviceInfo = this.jsonConvert.deserialize(data, DeviceInfo);

      return deviceInfo;
    } catch (e) {
      const deviceInfo = this.jsonConvert.deserialize(e.error, DeviceInfo);

      console.error("error getting device info:", e);
      return deviceInfo;
    }
  }

  public async getSoftwareStati() {
    try {
      const data = await this.http.get("device/status").toPromise();
      const stati = this.jsonConvert.deserialize(data, Status);

      return stati;
    } catch (e) {
      throw new Error("error getting software status': " + e);
    }
  }

  public async getDeviceID() {
    try {
      const data = await this.http
        .get("device/id", { responseType: "text" })
        .toPromise();

      return data;
    } catch (e) {
      throw new Error("error getting device id: " + e);
    }
  }

  public async getRoomPing() {
    try {
      const data = await this.http.get("room/ping").toPromise();
      const pingResult = this.jsonConvert.deserialize(data, PingResult);

      return pingResult;
    } catch (e) {
      throw new Error("error getting room ping info: " + e);
    }
  }
}
