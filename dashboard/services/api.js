class APIService {
  constructor() {
    this.urlParams = new URLSearchParams(window.location.search);
  }

  api(path) {
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

      if (text?.toLowerCase().includes("success")) {
        console.log("successfully flushed the dns cache");
      }
      return text || "No response";
    } catch (e) {
      console.error("error flushing dns:", e);
      throw e;
    }
  }

  async reSyncDB() {
    try {
      const controller = new AbortController();
      setTimeout(() => controller.abort(), 4500);

      const res = await this.request(this.api("/resyncDB"), {
        signal: controller.signal,
      });
      const text = await res.text();

      if (text) {
        return text;
      }
      if (res.status === 202) {
        return "Accepted (202)";
      }
      return "No response";
    } catch (e) {
      console.log("resync likely in progress / service restarting");
      if (e?.name === "AbortError") {
        return "Timed out locally; resync may still be in progress";
      }
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

      return text || "No response";
    } catch (e) {
      console.log("refresh likely triggered; connection dropped during restart");
      if (e?.name === "AbortError") {
        return "Timed out locally; refresh may still be in progress";
      }
      throw e;
    }
  }

  async hasDividerSensor() {
    const address = await this.getDividerSensorAddress();
    return !!address;
  }
  
  async getDividerSensorAddress() {
    // get the hostname
    const info = await this.getDeviceInfo();
    const hostname = info?.hostname;
    if (!hostname) {
      return;
    }

    const building = hostname.split("-")[0];
    const number = hostname.split("-")[1];
    if (!building || !number) {
      return;
    }

    const uiConfigUrl = `http://${location.hostname}:8000/buildings/${building}/rooms/${number}/configuration`;
    const res = await this.request(uiConfigUrl);
    const jsonResponse = await res.json();
    console.log("Response", jsonResponse);

    const devices = [...(jsonResponse?.devices || [])];
    for (const device of devices) {
      const roles = [...(device?.roles || [])];
      for (const role of roles) {
        if (role._id === "DividerSensor") {
          return device.address;
        }
      }
    }
  }

  async getDividerSensorPreset(address) {
    const hostname = await this.getDeviceInfo().then(info => info?.hostname);
    const res = await this.request(`http://${address}:10000/divider/preset/` + hostname); 
    const data = await res.text();
    return data;
  }

  async getDividerPin(address) {
    // get the system ID from the address
    const systemID = address.split(".")[0];

    const res = await this.request(this.api(`/divider/pins/${systemID}`));
    const data = await res.json();
    return data[0]?.pin;
  }

  async getDividerSensorInfo() {
    try {
      // check to see if there is a divider sensor
      const address = await this.getDividerSensorAddress();
      if (!address) {
        throw new Error("Divider sensor address not found");
      }
      const status = await this.getDividerSensorsStatus(address);
      const preset = await this.getDividerSensorPreset(address);
      const pin = await this.getDividerPin(address);

      const dividerSensorData = {
        address: address,
        status: status !== undefined ? (status ? "connected" : "disconnected") : "unknown",
        preset: preset || "unknown",
        pin: pin || "unknown",
      };
      console.log("Divider Sensor Info:", dividerSensorData);
      return new dividerSensorInfo(dividerSensorData);
    } catch (e) {
      console.error("Error getting divider sensor info:", e);
      return undefined;
    }
  }
} 
