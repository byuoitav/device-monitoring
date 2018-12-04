import { Component, OnInit } from "@angular/core";

import { APIService } from "../../services/api.service";
import { DeviceInfo, PingResult } from "../../objects";

@Component({
  selector: "overview",
  templateUrl: "./overview.component.html",
  styleUrls: ["./overview.component.scss"]
})
export class OverviewComponent implements OnInit {
  public deviceInfo: DeviceInfo;
  public pingResult: PingResult;
  public maintenanceMode: boolean;

  constructor(public api: APIService) {}

  async ngOnInit() {
    this.deviceInfo = await this.api.getDeviceInfo();
    console.log("deviceInfo", this.deviceInfo);

    this.pingResult = await this.api.getRoomPing();
    console.log("pingResult", this.pingResult);

    this.maintenanceMode = await this.api.getMaintenanceMode();
    console.log("maintenanceMode", this.maintenanceMode);
  }

  public isDefined(test: any): boolean {
    return typeof test !== "undefined" && test !== null;
  }

  public async toggleMaintenanceMode() {
    console.log("toggling maintenance mode");

    this.maintenanceMode = await this.api.toggleMaintenanceMode();
    console.log("maintenanceMode", this.maintenanceMode);
  }
}
