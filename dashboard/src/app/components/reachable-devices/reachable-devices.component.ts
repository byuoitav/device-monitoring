import { Component, OnInit } from "@angular/core";

import { APIService } from "../../services/api.service";
import { PingResult } from "../../objects";

@Component({
    selector: "reachable-devices",
    templateUrl: "./reachable-devices.component.html",
    styleUrls: ["./reachable-devices.component.scss"],
    standalone: false
})
export class ReachableDevicesComponent implements OnInit {
  public pingResult: Map<string, PingResult>;
  public roomHealth: Map<string, string>;
  public processedPingResult: Array<{key: string, value: PingResult}>;

  constructor(private api: APIService) {}

  async ngOnInit() {
    this.pingResult = await this.api.getRoomPing();
    console.log("ping result:", this.pingResult);

    this.roomHealth = await this.api.getRoomHealth();
    console.log("room health:", this.roomHealth);

    this.processPingResult();
    console.log("processed ping result:", this.processedPingResult);
  }

  processPingResult() {
    const ipMap = new Map<string, PingResult>();
    this.processedPingResult = [];

    this.pingResult.forEach((value, key) => {
      const ip = value.ip;
      if (ip && ipMap.has(ip)){
        // if ip already exists, use the ping value 
        const existingValue = ipMap.get(ip);
        if (existingValue) {
          existingValue.packetsLost += value.packetsLost;
          existingValue.packetsReceived += value.packetsReceived;
        }
      }
    })
    
  }
}
