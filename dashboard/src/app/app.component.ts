import { Component, OnInit } from '@angular/core';

import { Microservice } from './objects';
import { APIService } from './api.service';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss']
})
export class AppComponent implements OnInit {
	micros: Microservice[];	

	red: string = "#d9534f";
	green: string = "#8bd22f";
	yellow: string = "#e7ba08";

	dhcpstate: string = "";
	gettingDHCP: boolean;

	constructor(private api: APIService) {
		this.micros = ms;
	}

	ngOnInit() {
		this.getDHCPState();
	}

	reboot() {
		this.api.localget(":10000/reboot").subscribe();
	}

	refresh() {
		window.location.assign("http://" + location.hostname + ":10000/dash");
	}

	toUI() {
		window.location.assign("http://" + location.hostname + ":8888/");
	}

	getDHCPState() {
		this.gettingDHCP = true;
		this.api.localget(":7010/dhcp").subscribe(
			data => {
				this.dhcpstate = data.toString();
			},
			err => {
				setTimeout(() => {
					this.getDHCPState();
				}, 5000)
			}, () => {
				this.gettingDHCP = false;
			}
		);
	}

	toggleDHCP() {
		this.gettingDHCP = true;

		this.api.localput(":7010/dhcp", null).subscribe(
			data => {
				this.dhcpstate = data.toString();
			}, err => {
					
			}, () => {
				this.gettingDHCP = false;
			}
		);	
	}
}

const ms: Microservice[] = [
	{
		name: "av api",
		endpoint: ":8000/mstatus",
	},	
	{
		name: "event router",
		endpoint: ":6999/mstatus",
	},	
	{
		name: "event translator",
		endpoint: ":6998/mstatus",
	},	
	{
		name: "ui",
		endpoint: ":8888/mstatus",
	},	
	{
		name: "config db",
		endpoint: ":8006/mstatus",
	},	
]
