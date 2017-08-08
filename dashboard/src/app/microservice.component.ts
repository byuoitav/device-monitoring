import { Component, Input, ViewChild, Output, EventEmitter } from '@angular/core';
import { Http, Response, Headers, RequestOptions } from '@angular/http';
import { Observable, Subject } from 'rxjs/Rx';
import 'rxjs/add/operator/map';

import { Microservice, Status } from './objects';
import { APIService } from './api.service';

@Component({ selector: 'microservice',
  template: `
 	<div class="container" (click)="modal.show()">
		<div class="icon">
			<i *ngIf="s.statuscode == 0" class="material-icons healthy">check_circle</i>
			<i *ngIf="s.statuscode == 1" class="material-icons sick">warning</i>
			<i *ngIf="s.statuscode == 2" class="material-icons dead">cancel</i>
		</div>
		<div class="info">
			<span class="name">{{m.name}}</span>
			<span class="version">{{s.version}}</span>
		</div>
    </div>	
	<side-modal #modal [vertical]="false" [opposite]="true">
		<span>{{s.statusinfo}}</span>
	</side-modal>
  `,
  styles: [`
	  .container {
		background-color: rgba(255,255,255,0.1);
		height: 25vh;
		width: 50vh;
	 	box-shadow: 1px 1px 4px rgba(0,0,0,0.20); 
		border: 0;
		border-radius: 2vh;

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
		position: absolute;
	 	font-size: 13vh; 
	  }

	  .info {
		width: 60%;
		height: 100%;

		display: flex;
		flex-direction: column;
		justify-content: space-around;
		align-items: center;
	  }

	  .name {
		font-size: 3.5vh;
	  }

	  .version {
	 	font-size: 2.5vh; 
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
	@Output() modalVisible = new EventEmitter<boolean>();

	s: Status; 

	constructor(private api: APIService) {
		this.s = new Status();
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
			this.s.statuscode = 1;
			Object.assign(this.s, data);
			if (this.s.version == null || this.s.statusinfo == null) {
				this.s.statusinfo = "Incorrect response from server: " + JSON.stringify(data);
			}
			console.log("obj", this.s)
		}, err => {
			console.error("error!", err);
			this.s.statuscode = 2;
		})
	}
}
