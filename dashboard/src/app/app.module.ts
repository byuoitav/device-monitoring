import { NgModule } from "@angular/core";
import { BrowserModule } from "@angular/platform-browser";
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { RouterModule, Routes } from "@angular/router";
import { HttpClientModule } from "@angular/common/http";
import {
  MatSidenavModule,
  MatButtonModule,
  MatToolbarModule,
  MatCardModule,
  MatDividerModule,
  MatListModule,
  MatExpansionModule,
  MatIconModule,
  MatProgressBarModule
} from "@angular/material";
import "hammerjs";

import { AppComponent } from "./components/app/app.component";
import { OverviewComponent } from "./components/overview/overview.component";
import { SoftwareInfoComponent } from "./components/software-info/software-info.component";
import { ReachableDevicesComponent } from "./components/reachable-devices/reachable-devices.component";
import { APIService } from "./services/api.service";

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
    path: "software",
    component: SoftwareInfoComponent
  },
  {
    path: "reachable-devices",
    component: ReachableDevicesComponent
  }
];

@NgModule({
  declarations: [
    AppComponent,
    OverviewComponent,
    SoftwareInfoComponent,
    ReachableDevicesComponent
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
    RouterModule.forRoot(routes)
  ],
  providers: [APIService],
  bootstrap: [AppComponent]
})
export class AppModule {}
