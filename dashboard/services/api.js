class APIService {
  constructor() {
    this.urlParams = new URLSearchParams(window.location.search);
  }

  api(path) {
    return "http://localhost:8000/" + path;
    if (/^https?:\/\//i.test(path)) return path;
    return path.startsWith("/") ? path : `/${path}`;
  }

  async request(url, options = {}) {
    const res = await fetch(url, {
      credentials: "same-origin",
      ...options,
    });

    if (!res.ok) {
      const text = await res.text();
      const err = new Error(text || res.statusText);
      err.status = res.status;
      err.error = text;
      throw err;
    }

    return res;
  }

  switchToUI() {
    window.location.assign(`${location.protocol}//${location.hostname}:8888/`);
  }

  refresh() {
    window.location.reload();
  }

  async reboot() {
    try {
      await this.request(this.api("device/reboot"), {
        method: "PUT",
      });
      return "success";
    } catch (e) {
      if (e?.status === 200) {
        return "success";
      }
      console.error("error rebooting device:", e);
      return "fail";
    }
  }

  async getDeviceInfo() {
    try {
      const res = await this.request(this.api("device"));
      return await res.json();
    } catch (e) {
      if (e?.error) {
        return JSON.parse(e.error);
      }
      throw new Error("error getting device info: " + e.message);
    }
  }

  async getSoftwareStati() {
    const res = await this.request(this.api("device/status"));
    return res.json();
  }

  async getDeviceID() {
    const res = await this.request(this.api("device/id"));
    return res.text();
  }

  async getRoomPing() {
    const res = await this.request(this.api("room/ping"));
    const data = await res.json();

    const result = new Map();
    for (const [key, val] of Object.entries(data ?? {})) {
      if (key && val) {
        result.set(key, val);
      }
    }
    return result;
  }

  async getRoomHealth() {
    const res = await this.request(this.api("room/health"));
    const data = await res.json();

    const result = new Map();
    for (const [key, val] of Object.entries(data ?? {})) {
      if (key && val) {
        result.set(key, val);
      }
    }
    return result;
  }

  async getRunnerInfo() {
    const res = await this.request(this.api("device/runners"));
    return res.json();
  }

  async getViaInfo() {
    const res = await this.request(this.api("room/viainfo"));
    return res.json();
  }

  async resetVia(address) {
    await this.request(`http://${location.hostname}:8014/via/${address}/reset`);
  }

  async rebootVia(address) {
    await this.request(`http://${location.hostname}:8014/via/${address}/reboot`);
  }

  async getDividerSensorsStatus(address) {
    const res = await this.request(`http://${address}:10000/divider/state`);
    const data = await res.json();

    for (const key of Object.keys(data ?? {})) {
      if (key.includes("disconnected")) return false;
      if (key.includes("connected")) return true;
    }
    return undefined;
  }

  async getHardwareInfo() {
    const res = await this.request(this.api("/device/hardwareinfo"));
    return res.json();
  }

  async flushDNS() {
    try {
      const controller = new AbortController();
      setTimeout(() => controller.abort(), 2500);

      const res = await this.request(this.api("/dns"), {
        signal: controller.signal,
      });
      const text = await res.text();

      if (text.toLowerCase().includes("success")) {
        console.log("successfully flushed the dns cache");
        return "success";
      }
      return "fail";
    } catch (e) {
      console.error("error flushing dns:", e);
      throw e;
    }
  }

  async reSyncDB() {
    try {
      const controller = new AbortController();
      setTimeout(() => controller.abort(), 2500);

      const res = await this.request(this.api("/resyncDB"), {
        signal: controller.signal,
      });
      const text = await res.text();

      return text.toLowerCase().includes("success") ? "success" : "fail";
    } catch (e) {
      console.log("resync likely in progress / service restarting");
      throw e;
    }
  }

  async refreshContainers() {
    try {
      const controller = new AbortController();
      setTimeout(() => controller.abort(), 2500);

      const res = await this.request(this.api("/refreshContainers"), {
        signal: controller.signal,
      });
      const text = await res.text();

      return text.toLowerCase().includes("success") ? "success" : "fail";
    } catch (e) {
      console.log("refresh likely triggered; connection dropped during restart");
      throw e;
    }
  }
}
