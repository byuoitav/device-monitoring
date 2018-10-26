import { Component, OnInit } from "@angular/core";

import { APIService } from "../../services/api.service";

@Component({
  selector: "dashboard",
  templateUrl: "./app.component.html",
  styleUrls: ["./app.component.scss"]
})
export class AppComponent implements OnInit {
  public wideSidenav = true;
  public id: string;

  constructor(public api: APIService) {}

  public async ngOnInit() {
    this.id = await this.api.getDeviceID();
  }
}
