import { Component, OnInit } from "@angular/core";

import { APIService } from "../../services/api.service";
import { Status } from "../../objects";

@Component({
    selector: "software-info",
    templateUrl: "./software-info.component.html",
    styleUrls: ["./software-info.component.scss"],
    standalone: false
})
export class SoftwareInfoComponent implements OnInit {
  public stati: Status;

  constructor(private api: APIService) {}

  async ngOnInit() {
    this.stati = await this.api.getSoftwareStati();
    console.log("stati", this.stati);
  }
}
