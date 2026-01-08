class device {
  constructor(data = {}) {
    this.hostname = data.hostname ?? "";
    this.id = data.id ?? "";
    this.ip = data.ip ?? "";
    this["internet-connectivity"] = data["internet-connectivity"] ?? false;
    this.dhcp = {
      enabled: data.dhcp?.enabled ?? false,
      toggleable: data.dhcp?.toggleable ?? false,
    };
  }
}

