import { Component, OnInit } from "@angular/core";

import { APIService } from "../../services/api.service";
import { DeviceInfo, PingResult } from "../../objects";
import { AlertService } from "../../_alert";

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
  // public maintenanceMode: boolean;

  public isBusy = { resync: false, refresh: false, reboot: false, flush: false };

  options = {
    autoClose: true,
    keepAfterRouteChange: true 
  };

  constructor(public api: APIService, public alertService: AlertService) {}

  async ngOnInit() {
    try {
      this.deviceInfo = await this.api.getDeviceInfo();
    } catch (e) {
      console.error("error getting device info", e);
      this.alertService.error("Failed to load device info.", this.options);
    }

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
    if (this.isBusy.resync) return;
    this.isBusy.resync = true;
    try {
      const result = await this.api.reSyncDB();
      if (result === "success") {
        this.alertService.success("Successfully ReSync DB!", this.options);
      } else {
        this.alertService.error("Failed to ReSync DB.", this.options);
      }
    } catch (e) {
      console.error(e);
      this.alertService.error("Error during ReSync DB.", this.options);
    } finally {
      this.isBusy.resync = false;
    }
  }

  public async handleRefreshContainers(): Promise<void> {
    if (this.isBusy.refresh) return;
    this.isBusy.refresh = true;
    try {
      const result = await this.api.refreshContainers(); // returns "success" | "fail"
      if (result === "success") {
        this.alertService.success("Successfully refreshed containers.", this.options);
      } else {
        this.alertService.error("Failed to refresh containers.", this.options);
      }
    } catch (e) {
      console.error(e);
      this.alertService.error("Error while refreshing containers.", this.options);
    } finally {
      this.isBusy.refresh = false;
    }
  }

  public async handleReboot(): Promise<void> {
    if (this.isBusy.reboot) return;
    this.isBusy.reboot = true;
    try {
      const result = await this.api.reboot(); // "success" | "fail"
      if (result === "success") {
        this.alertService.success("Reboot requested. Device will restart shortly.", this.options);
      } else {
        this.alertService.error("Failed to request reboot.", this.options);
      }
    } catch (e) {
      console.error(e);
      this.alertService.error("Error requesting reboot.", this.options);
    } finally {
      this.isBusy.reboot = false;
    }
  }

  public async handleFlushDNS(): Promise<void> {
    if (this.isBusy.flush) return;
    this.isBusy.flush = true;
    try {
      const result = await this.api.flushDNS(); // "success" | "fail"
      if (result === "success") {
        this.alertService.success("Successfully flushed DNS cache.", this.options);
      } else {
        this.alertService.error("Failed to flush DNS cache.", this.options);
      }
    } catch (e) {
      console.error(e);
      this.alertService.error("Error flushing DNS cache.", this.options);
    } finally {
      this.isBusy.flush = false;
    }
  }
}
