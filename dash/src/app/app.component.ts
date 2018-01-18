import { Component, OnInit } from '@angular/core';

import { Microservice, Event } from './objects';
import { APIService } from './api.service';
import { SocketService, OPEN, CLOSE, MESSAGE } from './socket.service';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss'],
  providers: [APIService, SocketService],
})
export class AppComponent implements OnInit {
	micros: Microservice[];	

	red: string = "#d9534f";
	green: string = "#8bd22f";
	yellow: string = "#e7ba08";
	blue: string = "#2196F3";

	dhcpstate: string = "";
	gettingDHCP: boolean;

	// testing events
	responses: Event[];

    rebootViaMessages = [];

	constructor(private api: APIService, private socket: SocketService) {
		this.micros = ms;
		this.responses = [];
	}

	ngOnInit() {
		this.getDHCPState();
		this.socketSetup();
	}

	socketSetup() {
		this.socket.getEventListener().subscribe(
			event => {
				if (event.type == MESSAGE) {
					let data = JSON.parse(event.data.data);

					let e = new Event();
					Object.assign(e, data);

					console.log("e", e);
					setTimeout(() => {
						this.responses.push(e);
					}, 500);
				} else if (event.type == CLOSE) {
					console.log("BAD NEWS");
				} else if (event.type == OPEN) {
					console.log("YAY! SOCKET OPENED!");
				}
			} 
		);
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

	testEvents() {
		// clear responses
		this.responses = [];

		console.log("testing events");		
		this.api.localget(":10000/testevents").subscribe();	

		setTimeout(() => {
			console.log("responses:" , this.responses);	
		}, 5000);
	}

    rebootVias() {
        this.rebootViaMessages = [];    

        this.rebootViaMessages.push("getting ip addresses of vias...");

        let split = this.api.hostname.split("-");
        let endpoint = ":8000/buildings/" + split[0] + "/rooms/" + split[1] + "/configuration";

        this.api.localget(endpoint).subscribe(
            data => {
                console.log("data", data);

                for(let device of data["devices"]) {

                    if (device.type == "via") {
                        console.log("found a via", device);

                        this.rebootViaMessages.push("rebooting " + device.name + " @ " + device.address);

                        this.api.localget(":8014/via/" + device.address + "/reboot").subscribe(
                            data => {
                                this.rebootViaMessages.push("successfully rebooted " + device.name);
                            }, err => {
                                this.rebootViaMessages.push("failed rebooting " + device.name);
                            }
                        )
                    }
                }
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
		endpoint: ":7000/mstatus",
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
