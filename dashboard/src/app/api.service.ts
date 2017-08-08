import { Injectable } from '@angular/core';
import { Http, Response, Headers, RequestOptions } from '@angular/http';
import { Observable, Subject } from 'rxjs/Rx';
import 'rxjs/add/operator/map';

@Injectable()
export class APIService {
	public hostname;

	constructor(private http: Http) {
		this.get("http://" + location.hostname + ":8888/hostname").subscribe(data => {
			this.hostname = String(data);		
		})
	}
	
	public get(url): Observable<Object> {
		return this.http.get(url).map(res => res.json());	
	}
}
