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
		endpoint: ":8000/health"
	},	
	{
		name: "event router",
		healthy: false,
		endpoint: ":8000/health"
	},	
	{
		name: "event translator",
		healthy: false,
		endpoint: ":8000/health"
	},	
	{
		name: "ui",
		healthy: false,
		endpoint: ":8000/health"
	},	
	{
		name: "config-db",
		healthy: false,
		endpoint: ":8000/health"
	},	
]
