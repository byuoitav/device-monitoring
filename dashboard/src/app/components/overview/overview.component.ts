import { Component, OnInit } from "@angular/core";

import { APIService } from "../../services/api.service";
import { DeviceInfo, PingResult } from "../../objects";

@Component({
  selector: "overview",
  templateUrl: "./overview.component.html",
  styleUrls: ["./overview.component.scss"]
})
export class OverviewComponent implements OnInit {
  public hasDividerSensors: boolean;
  public deviceInfo: DeviceInfo;
  public pingResult: Map<string, PingResult>;
  public dividerSensorStatus: string;
  public dividerSensorAddr: string;
  // public maintenanceMode: boolean;

  constructor(public api: APIService) {}

  async ngOnInit() {
    this.deviceInfo = await this.api.getDeviceInfo();
    console.log("device info", this.deviceInfo);

    this.pingResult = await this.api.getRoomPing();
    console.log("ping result", this.pingResult);
    this.hasDividerSensors = await this.getDividerSensors();
    this.connected();
    setInterval(() => {
      this.connected();
    }, 2000);

    /*
    this.maintenanceMode = await this.api.getMaintenanceMode();
    console.log("maintenanceMode", this.maintenanceMode);
     */
  }

  public isDefined(test: any): boolean {
    return typeof test !== "undefined" && test !== null;
  }

  public async toggleMaintenanceMode() {
    /*
    console.log("toggling maintenance mode");

    this.maintenanceMode = await this.api.toggleMaintenanceMode();
    console.log("maintenanceMode", this.maintenanceMode);
     */
  }

  public reachable(): number {
    if (!this.pingResult) {
      return 0;
    }

    return Array.from(this.pingResult.values()).filter(r => r.packetsLost === 0)
      .length;
  }

  public unreachable(): number {
    if (!this.pingResult) {
      return 0;
    }

    return Array.from(this.pingResult.values()).filter(r => r.packetsLost > 0)
      .length;
  }

  public getDividerSensors() {
    for (const k of Array.from(this.pingResult.keys())) {
      if (k.includes("DS1")) {
        this.dividerSensorAddr = k + ".byu.edu";
        return true;
      }
    }
    return false;
  }

  public async connected() {
    if (
      (await this.api.getDividerSensorsStatus(this.dividerSensorAddr)) == true
    ) {
      this.dividerSensorStatus = "Connected";
    } else {
      this.dividerSensorStatus = "Disconnected";
    }
  }
}
