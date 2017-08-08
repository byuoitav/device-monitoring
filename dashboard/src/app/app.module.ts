import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { HttpModule } from '@angular/http';

import { AppComponent } from './app.component';
import { MicroserviceComponent } from './microservice.component'
import { APIService } from './api.service'

@NgModule({
  declarations: [
    AppComponent,
	MicroserviceComponent
  ],
  imports: [
    BrowserModule,
	HttpModule
  ],
  providers: [APIService],
  bootstrap: [AppComponent]
})
export class AppModule { }
