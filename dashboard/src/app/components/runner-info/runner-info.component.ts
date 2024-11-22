import { Component, OnInit } from "@angular/core";

import { APIService } from "../../services/api.service";
import { RunnerInfo } from "../../objects";

@Component({
    selector: "runner-info",
    templateUrl: "./runner-info.component.html",
    styleUrls: ["./runner-info.component.scss"],
    standalone: false
})
export class RunnerInfoComponent implements OnInit {
  public infos: RunnerInfo[];

  constructor(private api: APIService) {}

  async ngOnInit() {
    this.infos = await this.api.getRunnerInfo();
    console.log("infos", this.infos);
  }
}
