import { Component, OnInit } from "@angular/core";
import {
  JsonConvert,
  OperationMode,
  ValueCheckingMode,
  JsonObject,
  JsonProperty,
  Any,
  JsonCustomConvert,
  JsonConverter
} from "json2typescript";

@Component({
  selector: "provisioning",
  templateUrl: "./provisioning.component.html",
  styleUrls: ["./provisioning.component.scss"]
})
export class ProvisioningComponent implements OnInit {
  public socket: WebSocket;
  private jsonConvert: JsonConvert;

  constructor() {
    this.jsonConvert = new JsonConvert();
    this.jsonConvert.ignorePrimitiveChecks = false;
  }

  ngOnInit() {
    this.openSocket();
  }

  private openSocket() {
    this.socket = new WebSocket(
      "ws://" + window.location.host + "/provisioning/ws"
    );

    this.socket.onopen = () => {
      console.log("socket connection successfully opened.");
    };

    this.socket.onmessage = message => {
      const data = JSON.parse(message.data);
      const event = this.jsonConvert.deserializeObject(data, Event);

      this.onEvent(event);
    };

    // try to reconnect when socket is closed
    this.socket.onclose = () => {
      console.log("socket connection closed. retrying in 5 seconds...");
      setTimeout(() => {
        this.openSocket();
      }, 5 * 1000);
    };
  }

  private onEvent(event: Event) {
    console.log("received event:", event);
  }
}

@JsonObject("BasicRoomInfo")
export class BasicRoomInfo {
  @JsonProperty("buildingID", String, true)
  BuildingID = "";

  @JsonProperty("roomID", String, true)
  RoomID = "";

  constructor(roomID: string) {
    if (roomID == null || roomID === undefined) {
      return;
    }

    const split = roomID.split("-");

    if (split.length === 2) {
      this.BuildingID = split[0];
      this.RoomID = split[0] + "-" + split[1];
    }
  }
}

@JsonObject("BasicDeviceInfo")
export class BasicDeviceInfo {
  @JsonProperty("buildingID", String, true)
  BuildingID = "";

  @JsonProperty("roomID", String, true)
  RoomID = "";

  @JsonProperty("deviceID", String, true)
  DeviceID = "";

  constructor(deviceID: string) {
    if (deviceID == null || deviceID === undefined) {
      return;
    }

    const split = deviceID.split("-");

    if (split.length === 3) {
      this.BuildingID = split[0];
      this.RoomID = split[0] + "-" + split[1];
      this.DeviceID = split[0] + "-" + split[1] + "-" + split[2];
    }
  }
}

@JsonConverter
class DateConverter implements JsonCustomConvert<Date> {
  serialize(date: Date): any {
    function pad(n) {
      return n < 10 ? "0" + n : n;
    }

    return (
      date.getUTCFullYear() +
      "-" +
      pad(date.getUTCMonth() + 1) +
      "-" +
      pad(date.getUTCDate()) +
      "T" +
      pad(date.getUTCHours()) +
      ":" +
      pad(date.getUTCMinutes()) +
      ":" +
      pad(date.getUTCSeconds()) +
      "Z"
    );
  }

  deserialize(date: any): Date {
    return new Date(date);
  }
}

@JsonObject("Event")
export class Event {
  @JsonProperty("generating-system", String, true)
  GeneratingSystem = "";

  @JsonProperty("timestamp", DateConverter, true)
  Timestamp: Date = undefined;

  @JsonProperty("event-tags", [String], true)
  EventTags: string[] = [];

  @JsonProperty("target-device", BasicDeviceInfo, true)
  TargetDevice = new BasicDeviceInfo(undefined);

  @JsonProperty("affected-room", BasicRoomInfo, true)
  AffectedRoom = new BasicRoomInfo(undefined);

  @JsonProperty("key", String, true)
  Key = "";

  @JsonProperty("value", String, true)
  Value = "";

  @JsonProperty("user", String, true)
  User = "";

  @JsonProperty("data", Any, true)
  Data: any = undefined;

  public hasTag(tag: string): boolean {
    return this.EventTags.includes(tag);
  }
}
