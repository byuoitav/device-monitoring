import { Component, Input } from '@angular/core';
import { Http, Response, Headers, RequestOptions } from '@angular/http';
import { Observable, Subject } from 'rxjs/Rx';
import 'rxjs/add/operator/map';

import { Microservice } from './objects';
import { APIService } from './api.service';

@Component({
  selector: 'microservice',
  template: `
 	<div class="container">
		<div class="icon">
			<i *ngIf="m.healthy" class="material-icons healthy">check_circle</i>
			<i *ngIf="!m.healthy" class="material-icons dead">cancel</i>
			<i *ngIf="false" class="material-icons sick">warning</i>
		</div>
		<div class="name">
			<span>{{m.name}}</span>
		</div>
    </div>	
  `,
  styles: [`
	  .container {
		background-color: rgba(255,255,255,0.1);
		height: 25vh;
		width: 50vh;
	 	box-shadow: 1px 1px 4px rgba(0,0,0,0.20); 
		border: 0;
		border-radius: 6%;

		display: flex;
		flex-direction: row;
	  }

	  .icon {
		width: 40%;
		height: 100%;

		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: center;
	  }

	  .icon i {
	 	font-size: 500%; 
	  }

	  .name {
		width: 60%;
		height: 100%;

		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: center;
	  }

	  .healthy {
	 	color: #8bd22f !important; 
	  }

	  .sick {
	 	color: #f0ad4e !important; 
	  }

	  .dead {
	 	color: #d9534f !important; 
	  }
  `]
})
export class MicroserviceComponent {
	@Input('microservice') m: Microservice;

	constructor(private api: APIService) {
		setTimeout(() => {
			this.checkHealth();
		}, 0);

		setInterval(() => {
			this.checkHealth();	
		}, 10000);
	}

	checkHealth() {
		this.api.get("http://" + location.hostname + this.m.endpoint)
		.subscribe(data => {
			console.log("data", data);
			// if response is good
			if (data != null) {
				this.m.healthy = true;	
			}	
		}, err => {
			console.log("error!", err);	
			this.m.healthy = false;
		})
	}
}
