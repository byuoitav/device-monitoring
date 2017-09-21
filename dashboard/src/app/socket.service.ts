import { Injectable, EventEmitter } from '@angular/core'
import { Http } from '@angular/http'
import { $WebSocket, WebSocketConfig } from 'angular2-websocket/angular2-websocket'

@Injectable()
export class SocketService {
  private socket: $WebSocket;
  private listener: EventEmitter<any>;
  private http: Http;
  private webSocketConfig: WebSocketConfig = {
 	initialTimeout: 100,
    maxTimeout: 500,
	reconnectIfNotNormalClose: true	
  }

  public screenoff: boolean;

  public constructor() {
	this.socket = new $WebSocket("ws://" + location.hostname +":10000/websocket", null, this.webSocketConfig);
	this.listener = new EventEmitter();
	this.screenoff = false;

	this.socket.onMessage((msg) => {
	  if (msg.data.includes("keepalive")) {
		console.log("keep alive message recieved.");
	  } else if (msg.data.includes("refresh")) {
	 	console.log("refreshing!");
		location.assign("http://" + location.hostname + ":10000/dash");
	  } else if (msg.data.includes("screenoff")) {
		 console.log("adding screenoff element");
		 this.screenoff = true;
	  } else {
	  	this.listener.emit({ "type": MESSAGE, "data": msg });
	  }
	}, {autoApply: false}
	);

	this.socket.onOpen((msg) => {
		console.log("websocket opened", msg);	
		this.listener.emit({"type": OPEN})
	});

	this.socket.onError((msg) => {
		console.log("websocket closed.", msg);	
		this.listener.emit({"type": CLOSE})
	});

	this.socket.onClose((msg) => {
		console.log("trying again", msg);	
	});
  }

  public close() {
    this.socket.close(false);
  }

  public getEventListener() {
    return this.listener;
  }
}

export const OPEN: string = "open";
export const CLOSE: string = "close";
export const MESSAGE: string = "message";
