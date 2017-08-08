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

	constructor(private api: APIService) {
		this.micros = ms;	
	}
}

const ms: Microservice[] = [
	{
		name: "av api",
		healthy: false,
		endpoint: ":8000/health",
		version: "v1.0"
	},	
	{
		name: "event router",
		healthy: false,
		endpoint: ":7000/health",
		version: "v1.0"
	},	
	{
		name: "event translator",
		healthy: false,
		endpoint: ":7000/health",
		version: "v1.0"
	},	
	{
		name: "ui",
		healthy: false,
		endpoint: ":8888/health",
		version: "v1.0"
	},	
	{
		name: "config-db",
		healthy: false,
		endpoint: ":8006/health",
		version: "v1.0"
	},	
]
