import { Injectable } from "@angular/core";
import { HttpClient } from "@angular/common/http";
import { JsonConvert } from "json2typescript";

import { DeviceInfo } from "../objects";

@Injectable({
  providedIn: "root"
})
export class APIService {
  //  private options: RequestOptions;
  private jsonConvert: JsonConvert = new JsonConvert();

  constructor(private http: HttpClient) {
    this.jsonConvert = new JsonConvert();
    this.jsonConvert.ignorePrimitiveChecks = false;

    //  const headers = new Headers();
    //  headers.append("content-type", "application/json");
    //  this.options = new RequestOptions({ headers: headers });
  }

  public getDeviceInfo() {
    return this.http.get("device").map(resp => resp.json());
  }
}
