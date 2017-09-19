import { Injectable } from '@angular/core';
import { Http, Response, Headers, RequestOptions } from '@angular/http';
import { Observable } from 'rxjs/Rx';
import 'rxjs/add/operator/map';

@Injectable()
export class APIService {
	public hostname;
	public ipaddress;
	public network;

	private options: RequestOptions;

	constructor(private http: Http) {
		let headers = new Headers();
		headers.append('content-type', 'application/json');
		this.options = new RequestOptions({ headers: headers });

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

	public put(url: string, data: any): Observable<Object> {
		return this.http.put(url, data, this.options)
						.map(res => res.json());
	}

	public localput(endpoint: string, data: any): Observable<Object> {
		return this.http.put("http://" + location.hostname + endpoint, data, this.options)
						.map(res => res.json());
	}
}
