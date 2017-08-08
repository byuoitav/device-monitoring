import { Injectable } from '@angular/core';
import { Http, Response, Headers, RequestOptions } from '@angular/http';
import { Observable, Subject } from 'rxjs/Rx';
import 'rxjs/add/operator/map';

@Injectable()
export class APIService {
	public hostname;
	public ipaddress;
	public network;

	constructor(private http: Http) {
		// hostname
		this.get("http://" + location.hostname + ":10000/hostname").subscribe(data => {
			this.hostname = String(data);		
		})

		// ipaddr
		this.get("http://" + location.hostname + ":10000/ip").subscribe(data => {
			this.ipaddress = String(data);		
		})

		// network
		this.get("http://" + location.hostname + ":10000/network").subscribe(data => {
			this.network = String(data);
		})
	}
	
	public get(url: string): Observable<Object> {
		return this.http.get(url).map(res => res.json());	
	}

	public localget(endpoint: string): Observable<Object> {
		return this.http.get("http://" + location.hostname + endpoint).map(res => res.json());	
	}
}
