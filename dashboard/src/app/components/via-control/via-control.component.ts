import { Component, OnInit } from "@angular/core";

import { APIService } from "../../services/api.service";
import { ViaInfo } from "../../objects";

@Component({
  selector: "via-control",
  templateUrl: "./via-control.component.html",
  styleUrls: ["./via-control.component.scss"]
})
export class ViaControlComponent implements OnInit {
  public viainfo: ViaInfo[] = [];

  constructor(private api: APIService) {}

  async ngOnInit() {
    this.viainfo = await this.api.getViaInfo();
    console.log("via info", this.viainfo);
  }

  reset(via: ViaInfo) {
    this.api.resetVia(via.address);
  }

  reboot(via: ViaInfo) {
    this.api.rebootVia(via.address);
  }
}
