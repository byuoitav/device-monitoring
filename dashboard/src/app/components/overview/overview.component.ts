import { Component, OnInit } from "@angular/core";

import { APIService } from "../../services/api.service";

@Component({
  selector: "overview",
  templateUrl: "./overview.component.html",
  styleUrls: ["./overview.component.scss"]
})
export class OverviewComponent implements OnInit {
  constructor(private api: APIService) {}

  ngOnInit() {
    console.log("here");
    this.api.getDeviceInfo().subscribe(data => {});
  }
}
