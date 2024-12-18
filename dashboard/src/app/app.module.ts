import { NgModule } from "@angular/core";
import { APP_BASE_HREF } from '@angular/common';
import { BrowserModule } from '@angular/platform-browser';
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { RouterModule, Routes } from "@angular/router";
import { HttpClientModule } from "@angular/common/http";

import { MatSidenavModule } from "@angular/material/sidenav";
import { MatButtonModule } from "@angular/material/button";
import { MatToolbarModule } from "@angular/material/toolbar";
import { MatCardModule } from "@angular/material/card";
import { MatDividerModule } from "@angular/material/divider";
import { MatListModule } from "@angular/material/list";
import { MatExpansionModule } from "@angular/material/expansion";
import { MatIconModule } from "@angular/material/icon";
import { MatProgressBarModule } from "@angular/material/progress-bar";
import { MatDialogModule } from "@angular/material/dialog";
import { MatProgressSpinnerModule } from "@angular/material/progress-spinner";

import "hammerjs";

import { AppComponent } from "./components/app/app.component";
import { OverviewComponent } from "./components/overview/overview.component";
import { ProvisioningComponent } from "./components/provisioning/provisioning.component";
import { SoftwareInfoComponent } from "./components/software-info/software-info.component";
import { ReachableDevicesComponent } from "./components/reachable-devices/reachable-devices.component";
import { APIService } from "./services/api.service";
import { RunnerInfoComponent } from "./components/runner-info/runner-info.component";;
import { RebootComponent } from './popups/reboot/reboot.component';
import { DeviceInfoComponent } from './device-info/device-info.component';

import { AlertModule } from "./_alert/alert.module";
import { provideAnimationsAsync } from '@angular/platform-browser/animations/async';

const routes: Routes = [
  {
    path: "",
    redirectTo: "/overview",
    pathMatch: "full"
  },
  {
    path: "overview",
    component: OverviewComponent
  },
  {
    path: "provisioning",
    component: ProvisioningComponent
  },
  {
    path: "software",
    component: SoftwareInfoComponent
  },
  {
    path: "reachable-devices",
    component: ReachableDevicesComponent
  },
  {
    path: "runner-info",
    component: RunnerInfoComponent
  },
  {
    path: "device-info",
    component: DeviceInfoComponent
  }
];

@NgModule({
  declarations: [
    AppComponent,
    OverviewComponent,
    ProvisioningComponent,
    SoftwareInfoComponent,
    ReachableDevicesComponent,
    RunnerInfoComponent,
    RebootComponent,
    DeviceInfoComponent,
  ],
  imports: [
    BrowserModule,
    BrowserAnimationsModule,
    HttpClientModule,
    MatSidenavModule,
    MatButtonModule,
    MatToolbarModule,
    MatCardModule,
    MatDividerModule,
    MatListModule,
    MatExpansionModule,
    MatIconModule,
    MatProgressBarModule,
    MatDialogModule,
    MatProgressSpinnerModule,
    RouterModule.forRoot(routes),
    AlertModule
  ],
  providers: [
    APIService,
    {
      // set the base route for the router
      provide: APP_BASE_HREF,
      useValue: "/dashboard"
    },
    provideAnimationsAsync()
  ],

  bootstrap: [AppComponent]
})
export class AppModule {}
