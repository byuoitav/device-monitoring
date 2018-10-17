import { Component, OnInit } from "@angular/core";

import { APIService } from "../../services/api.service";
import { DeviceInfo } from "../../objects";

@Component({
  selector: "overview",
  templateUrl: "./overview.component.html",
  styleUrls: ["./overview.component.scss"]
})
export class OverviewComponent implements OnInit {
  private deviceInfo: DeviceInfo;

  constructor(private api: APIService) {}

  async ngOnInit() {
    this.deviceInfo = await this.api.getDeviceInfo();
    console.log("deviceInfo", this.deviceInfo);
  }
}
