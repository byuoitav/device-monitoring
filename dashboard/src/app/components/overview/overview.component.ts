import { Component, OnInit } from "@angular/core";

import { APIService } from "../../services/api.service";
import { DeviceInfo, PingResult } from "../../objects";
import { FeedbackService } from "../../services/feedback.service";

@Component({
  selector: "overview",
  templateUrl: "./overview.component.html",
  styleUrls: ["./overview.component.scss"],
  standalone: false
})
export class OverviewComponent implements OnInit {
  public hasDividerSensors: boolean;
  public deviceInfo: DeviceInfo;
  public pingResult: Map<string, PingResult>;
  public dividerSensorStatus: string;
  public dividerSensorAddr: string;

  public isBusy = { resync: false, refresh: false, reboot: false, flush: false };

  constructor(
    public api: APIService,
    private feedback: FeedbackService
  ) {}

  async ngOnInit() {
    try {
      this.deviceInfo = await this.api.getDeviceInfo();
    } catch (e) {
      console.error("error getting device info", e);
      this.feedback.error("Failed to load device info.");
    }

    this.pingResult = await this.api.getRoomPing();
    this.hasDividerSensors = this.getDividerSensors();

    if (this.hasDividerSensors) {
      this.connected();
    }
  }

  public isDefined(test: any): boolean {
    return typeof test !== "undefined" && test !== null;
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

  public async connected(): Promise<void> {
    try {
      const state = await this.api.getDividerSensorsStatus(this.dividerSensorAddr);
      if (state === true) {
        this.dividerSensorStatus = "Connected";
      } else if (state === false) {
        this.dividerSensorStatus = "Disconnected";
      } else {
        this.dividerSensorStatus = "Unknown";
      }
    } catch (e) {
      console.error("error checking divider sensors", e);
      this.dividerSensorStatus = "Unknown";
    }
  }

  // ---------- Button handlers ----------
  public async handleResyncDB(): Promise<void> {
    await this.feedback.run(
      async () => {
        const result = await this.api.reSyncDB();
        if (result === "fail") throw new Error("resync failed");
        return result;
      },
      v => (this.isBusy.resync = v),
      {
        start: "Resyncing database…",
        success: "Database resynced ✓",
        softSuccess: "Resync started. Reconnecting…",
        error: "Resync failed.",
        optimisticOnNetworkError: true,
        onFinally: () => setTimeout(() => this.api.refresh(), 1500)
      }
    );
  }

  public async handleRefreshContainers(): Promise<void> {
    await this.feedback.run(
      async () => {
        const result = await this.api.refreshContainers();
        if (result === "fail") throw new Error("refresh failed");
        return result;
      },
      v => (this.isBusy.refresh = v),
      {
        start: "Refreshing containers…",
        success: "Containers refreshed ✓",
        softSuccess: "Refresh initiated. Reconnecting…",
        error: "Failed to refresh containers.",
        optimisticOnNetworkError: true,
        onFinally: () => setTimeout(() => this.api.refresh(), 1500)
      }
    );
  }

  public async handleReboot(): Promise<void> {
    await this.feedback.run(
      async () => {
        const result = await this.api.reboot();
        if (result === "fail") throw new Error("reboot failed");
        return result;
      },
      v => (this.isBusy.reboot = v),
      {
        start: "Rebooting…",
        success: "Reboot command sent ✓",
        softSuccess: "Rebooting now. Reconnecting…",
        error: "Failed to request reboot.",
        optimisticOnNetworkError: true,
        onFinally: () => setTimeout(() => this.api.refresh(), 3000)
      }
    );
  }

  public async handleFlushDNS(): Promise<void> {
    await this.feedback.run(
      async () => {
        const result = await this.api.flushDNS();
        if (result === "fail") throw new Error("flush failed");
        return result;
      },
      v => (this.isBusy.flush = v),
      {
        start: "Flushing DNS…",
        success: "DNS cache flushed ✓",
        error: "Couldn’t flush DNS.",
        optimisticOnNetworkError: false
      }
    );
  }
}
