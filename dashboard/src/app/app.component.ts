import { Component } from '@angular/core';

import { Microservice } from './objects';
import { APIService } from './api.service';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss']
})
export class AppComponent {
	micros: Microservice[];	

	red: string = "#d9534f";
	green: string = "#8bd22f";
	yellow: string = "#e7ba08";

	constructor(private api: APIService) {
		this.micros = ms;	
	}

	reboot() {
		this.api.localget(":10000/reboot").subscribe()	
	}

	toUI() {
		window.location.assign("http://" + location.hostname + ":8888/")
	}
}

const ms: Microservice[] = [
	{
		name: "av api",
		endpoint: ":8000/mstatus",
	},	
	{
		name: "event router",
		endpoint: ":7000/health",
	},	
	{
		name: "event translator",
		endpoint: ":7000/health",
	},	
	{
		name: "ui",
		endpoint: ":8888/health",
	},	
	{
		name: "config db",
		endpoint: ":8006/health",
	},	
]
