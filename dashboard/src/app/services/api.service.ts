import { Injectable } from "@angular/core";
import { HttpClient } from "@angular/common/http";
import { firstValueFrom } from "rxjs";
import { MatDialog } from "@angular/material/dialog";

import {
  DeviceInfo,
  Status,
  PingResult,
  RunnerInfo,
  ViaInfo
} from "../objects";
import { RebootComponent } from "../popups/reboot/reboot.component";


@Injectable({
  providedIn: "root"
})
export class APIService {
  public theme = "default";

  private urlParams: URLSearchParams;

  constructor(private http: HttpClient, private dialog: MatDialog) {
    this.urlParams = new URLSearchParams(window.location.search);
    if (this.urlParams.has("theme")) {
      this.theme = this.urlParams.get("theme") ?? 'default';
    }
  }

  private api(path: string): string {
    if (/^https?:\/\//i.test(path)) return path;
    return path.startsWith("/") ? path : `/${path}`;
  }

  public switchToUI() {
    window.location.assign(`${location.protocol}//${location.hostname}:8888/`);
  }

  public refresh() {
    window.location.reload();
  }

  public switchTheme(name: string) {
    console.log("switching theme to", name);

    this.theme = name;
    this.urlParams.set("theme", name);
    window.history.replaceState(
      null,
      "System Health Dashboard",
      window.location.pathname + "?" + this.urlParams.toString()
    );
  }

  public async reboot(): Promise<"success" | "fail"> {
    try {
      // optional: keep your dialog
      this.dialog.open(RebootComponent, { disableClose: true });

      const text = await firstValueFrom(
        this.http.put(this.api("device/reboot"), null, { responseType: "text" as const })
      );

      // treat any 200 as success; some backends just return plain text
      if (typeof text === "string" && text.trim().length >= 0) {
        return "success";
      }
      return "fail";
    } catch (e: any) {
      // Some backends send the payload via the "error" field with 200 status
      if (e?.status === 200 && typeof e?.error === "string") {
        return "success";
      }
      console.error("error rebooting device:", e);
      return "fail";
    }
  }

  public async getDeviceInfo(): Promise<DeviceInfo> {
    try {
      const data = await firstValueFrom(
        this.http.get<DeviceInfo>(this.api("device"))
      );
      return data;
    } catch (e: any) {
      // Some backends send the actual payload on error paths
      if (e?.error) {
        return e.error as DeviceInfo;
      }
      console.error('error getting device info:', e);
      throw new Error('error getting device info: ' + (e?.message ?? e));
    }
  }

  public async getMaintenanceMode() {
    return false;
    /*
    try {
      const data = await firstValueFrom(this.http.get<boolean>('maintenance'));
      return data;
    } catch (e: any) {
      throw new Error('error getting maintenance mode: ' + (e?.message ?? e));
    }
    */
  }

  public async toggleMaintenanceMode() {
    return false;
    /*
    try {
      const data = await firstValueFrom(this.http.put<boolean>('maintenance', null));
      return data;
    } catch (e: any) {
      throw new Error('error toggling maintenance mode: ' + (e?.message ?? e));
    }
    */
  }

  public async getSoftwareStati(): Promise<Status> {
    try {
      const data = await firstValueFrom(
        this.http.get<Status>(this.api("device/status"))
      );
      return data;
    } catch (e: any) {
      throw new Error("error getting software status': " + (e?.message ?? e));
    }
  }

  public async getDeviceID(): Promise<string> {
    try {
      const data = await firstValueFrom(
        this.http.get(this.api("device/id"), { responseType: "text" as const })
      );
      return data;
    } catch (e: any) {
      throw new Error('error getting device id: ' + (e?.message ?? e));
    }
  }

  public async getRoomPing(): Promise<Map<string, PingResult>> {
    try {
      const data = await firstValueFrom(
        this.http.get<Record<string, PingResult>>(this.api("room/ping"))
      );

      console.log('room ping data:', data);

      const result = new Map<string, PingResult>();
      for (const [key, val] of Object.entries(data ?? {})) {
        if (key && val) {
          result.set(key, val);
        }
      }
      return result;
    } catch (e: any) {
      throw new Error('error getting room ping info: ' + (e?.message ?? e));
    }
  }

  public async getRoomHealth(): Promise<Map<string, string>> {
    try {
      const data = await firstValueFrom(
        this.http.get<Record<string, string>>(this.api("room/health"))
      );

      const result = new Map<string, string>();
      for (const [key, val] of Object.entries(data ?? {})) {
        if (key && val) {
          result.set(key, val);
        }
      }
      return result;
    } catch (e: any) {
      throw new Error('error getting room health info: ' + (e?.message ?? e));
    }
  }

  public async getRunnerInfo(): Promise<RunnerInfo[]> {
    try {
      const data = await firstValueFrom(
        this.http.get<RunnerInfo[]>(this.api("device/runners"))
      );
      return data;
    } catch (e: any) {
      throw new Error('error getting device runner info: ' + (e?.message ?? e));
    }
  }

  public async getViaInfo(): Promise<ViaInfo[]> {
    try {
      const data = await firstValueFrom(
        this.http.get<ViaInfo[]>(this.api("room/viainfo"))
      );
      return data;
    } catch (e: any) {
      throw new Error('error getting via info: ' + (e?.message ?? e));
    }
  }

  public async resetVia(address: string): Promise<void> {
    try {
      await firstValueFrom(
        this.http.get(`http://${location.hostname}:8014/via/${address}/reset`)
      );
    } catch (e: any) {
      throw new Error('error resetting via: ' + (e?.message ?? e));
    }
  }

  public async rebootVia(address: string): Promise<void> {
    try {
      await firstValueFrom(
        this.http.get(`http://${location.hostname}:8014/via/${address}/reboot`)
      );
    } catch (e: any) {
      throw new Error('error rebooting via: ' + (e?.message ?? e));
    }
  }

  public async getDividerSensorsStatus(address: string): Promise<boolean | undefined> {
    try {
      const data = await firstValueFrom(
        this.http.get<Record<string, unknown>>(`http://${address}:10000/divider/state`)
      );

      for (const [key] of Object.entries(data ?? {})) {
        if (key.includes('disconnected')) {
          return false;
        }
        if (key.includes('connected')) {
          return true;
        }
      }
      return undefined; // if neither key is present
    } catch (e: any) {
      throw new Error('error getting divider sensors connection status: ' + (e?.message ?? e));
    }
  }

  public async getHardwareInfo(): Promise<unknown> {
    try {
      const data = await firstValueFrom(
        this.http.get<unknown>(this.api("/device/hardwareinfo"))
      );
      return data;
    } catch (e: any) {
      throw new Error('error getting hardware info: ' + (e?.message ?? e));
    }
  }


  public async flushDNS(): Promise<"success" | "fail"> {
    try {
      const data = await firstValueFrom(
        this.http.get(this.api("/dns"), { responseType: "text" as const })
      );
      if (typeof data === "string" && data.toLowerCase().includes("success")) {
        console.log("%c successfully flushed the dns cache", "color: green; font-size: 20px");
        return "success";
      }
      console.log("%c failed to flush the dns cache", "color: red; font-size: 20px");
      return "fail";
    } catch (e) {
      console.error("error flushing dns:", e);
      return "fail";
    }
  }


  // reSyncDB (Swab)
  public async reSyncDB(): Promise<"success" | "fail"> {
    try {
      const data = await firstValueFrom(
        this.http.get(this.api("/resyncDB"), { responseType: "text" as const })
      );
      if (data && data.toLowerCase().includes("success")) {
        console.log("%c successfully resynced the database", "color: green; font-size: 20px");
        return "success";
      }
      console.log("%c failed to resync the database", "color: red; font-size: 20px");
      return "fail";
    } catch (e) {
      console.log("%c failed to resync the database", "color: red; font-size: 20px");
      return "fail";
    }
  }

  // refreshContainers (Float)
  public async refreshContainers(): Promise<"success" | "fail"> {
    try {
      const data = await firstValueFrom(
        this.http.get(this.api("/refreshContainers"), { responseType: "text" as const })
      );
      if (data && data.toLowerCase().includes("success")) {
        console.log("%c successfully refreshed the containers", "color: green; font-size: 20px");
        return "success";
      }
      console.log("%c failed to refresh the containers", "color: red; font-size: 20px");
      return "fail";
    } catch (e) {
      console.log("%c failed to refresh the containers", "color: red; font-size: 20px");
      return "fail";
    }
  }
}
