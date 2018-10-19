import { Injectable } from "@angular/core";
import { HttpClient } from "@angular/common/http";
import { JsonConvert, OperationMode, ValueCheckingMode } from "json2typescript";

import { DeviceInfo, Status, PingResult, Status } from "../objects";

@Injectable({
  providedIn: "root"
})
export class APIService {
  private jsonConvert: JsonConvert;

  constructor(private http: HttpClient) {
    this.jsonConvert = new JsonConvert();
    this.jsonConvert.ignorePrimitiveChecks = false;
  }

  public async getDeviceInfo() {
    try {
      const data = await this.http.get("device").toPromise();
      const deviceInfo = this.jsonConvert.deserialize(data, DeviceInfo);

      return deviceInfo;
    } catch (e) {
      throw new Error("error getting device info: " + e);
    }
  }

  public async getSoftwareStati() {
    try {
      const data = await this.http.get("device/status").toPromise();
      const stati = this.jsonConvert.deserialize(data, Status);

      return stati;
    } catch (e) {
      throw new Error("error getting device info: " + e);
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
