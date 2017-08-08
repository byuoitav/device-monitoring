import { NgModule } from '@angular/core';
import { HttpModule } from '@angular/http';
import { BrowserModule } from '@angular/platform-browser';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';

import { AppComponent } from './app.component';
import { MicroserviceComponent } from './microservice.component'
import { ButtonComponent } from './button.component'
import { SideModalComponent } from './side-modal.component'
import { APIService } from './api.service'

@NgModule({
  declarations: [
    AppComponent,
	MicroserviceComponent,
	ButtonComponent,
	SideModalComponent
  ],
  imports: [
    BrowserModule,
	HttpModule,
	BrowserAnimationsModule
  ],
  providers: [APIService],
  bootstrap: [AppComponent]
})
export class AppModule { }
