import { Component, OnInit } from '@angular/core';
import { APIService } from '../services/api.service';
import { DeviceInfo } from '../objects';

@Component({
  selector: 'device-info',
  templateUrl: './device-info.component.html',
  styleUrls: ['./device-info.component.scss']
})
export class DeviceInfoComponent implements OnInit {

  public hardwareInfo
  public hostList = {};
  public memoryList = {};
  public CPUList = {};
  public diskList = {};
  public dockerList = {};
  public tempList = {};
  public highestTemp;

  constructor(private api : APIService) { }

  async ngOnInit() {
    this.hardwareInfo = await this.api.getHardwareInfo();
    console.log("hardware info", this.hardwareInfo);
    this.getHostList();
    console.log("host list", this.hostList);
    this.getMemoryList();
    console.log("memory list", this.memoryList);
    this.getCPUList();
    console.log("cpu list", this.CPUList);
    this.getDiskList();
    console.log("disk list", this.diskList);
    this.getDockerList();
    console.log("docker list", this.dockerList);
    this.getTempList();
    console.log("temp list", this.tempList);
    this.getHighestTemp();
    console.log("highest temp", this.highestTemp);
  }

  getHostList() {
    this.hostList["Hostname"] = this.hardwareInfo.host.os.hostname
    let uptime = this.hardwareInfo.host.os.uptime/3600;
    this.hostList["Uptime"] = uptime.toFixed(2).toString() + " hours"
    this.hostList["OS"] = this.hardwareInfo.host.os.os
    this.hostList["Platform"] = this.hardwareInfo.host.os.platform
    this.hostList["Platform Version"] = this.hardwareInfo.host.os.platformVersion
    this.hostList["Kernel Version"] = this.hardwareInfo.host.os.kernelVersion
    this.hostList["Host ID"] = this.hardwareInfo.host.os.hostid
  }

  getMemoryList() {
    this.memoryList["Swap Total"] = this.formatBytes(this.hardwareInfo.memory.swap.total)
    this.memoryList["Swap Free"] = this.formatBytes(this.hardwareInfo.memory.swap.free)
    this.memoryList["Swap Used"] = this.formatBytes(this.hardwareInfo.memory.swap.used)
    let swapPecent = this.hardwareInfo.memory.swap.usedPercent;
    this.memoryList["Swap Used Percent"] = swapPecent.toFixed(2).toString()+"%"
    this.memoryList["Virtual Total"] = this.formatBytes(this.hardwareInfo.memory.virtual.total)
    this.memoryList["Virtual Free"] = this.formatBytes(this.hardwareInfo.memory.virtual.free)
    this.memoryList["Virtual Used"] = this.formatBytes(this.hardwareInfo.memory.virtual.used)
    let virtualPercent = this.hardwareInfo.memory.virtual.usedPercent;
    this.memoryList["Virtual Used Percent"] = virtualPercent.toFixed(2).toString()+"%"
  }

  getCPUList() {    
    let thing = this.hardwareInfo.cpu.usage;
    for (let [key] of Object.entries(thing)) {
      this.CPUList[key.toUpperCase()] = thing[key]
    }
  }

  getDiskList() {
    let thing = this.hardwareInfo.disk["io-counters"];
    for (let [key] of Object.entries(thing)) {
      this.diskList["Write Count"] = thing[key].writeCount
    }
    this.diskList["Total"] = this.formatBytes(this.hardwareInfo.disk.usage.total)
    this.diskList["Free"] = this.formatBytes(this.hardwareInfo.disk.usage.free)
    this.diskList["Used"] = this.formatBytes(this.hardwareInfo.disk.usage.used)
    let usedPercent = this.hardwareInfo.disk.usage.usedPercent;
    this.diskList["Used Percent"] = usedPercent.toFixed(2).toString()+"%"
  }

  getDockerList() {
    for (let docker of this.hardwareInfo.docker.stats) {
      if (docker.running) {
        this.dockerList[docker.name] = "running "
      } else {
        this.dockerList[docker.name] = "stopped "
      }
      this.dockerList[docker.name] += docker.status
    }
  }

  getTempList() {
    let thing = this.hardwareInfo.host.temperature;
    for (let [key] of Object.entries(thing)) {
      this.tempList[key] = thing[key]
    }
  }

  getHighestTemp() {
    this.highestTemp = 0
    for (let [key] of Object.entries(this.tempList)) {
      if (this.tempList[key] > this.highestTemp) {
        this.highestTemp = this.tempList[key];
      }
    }
  }

  formatBytes(bytes, decimals = 2) {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
  }
}
