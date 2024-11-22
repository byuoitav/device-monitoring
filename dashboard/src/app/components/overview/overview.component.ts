import { Component, OnInit } from "@angular/core";

import { APIService } from "../../services/api.service";
import { DeviceInfo, PingResult } from "../../objects";
import { AlertService } from "../../_alert";

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

  options = {
    autoClose: true,
    keepAfterRouteChange: true 
  };

  constructor(public api: APIService, public alertService: AlertService) {}

  async ngOnInit() {
    this.deviceInfo = await this.api.getDeviceInfo();
    console.log("device info", this.deviceInfo);

    this.pingResult = await this.api.getRoomPing();
    console.log("ping result", this.pingResult);
    this.hasDividerSensors = this.getDividerSensors();
    console.log("hasDividerSensors", this.hasDividerSensors);
    if (this.hasDividerSensors) {
      this.connected();
      console.log("dividerSensorStatus", this.dividerSensorStatus);
    }

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
      console.log("k value from overview component is: ", k); 
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
