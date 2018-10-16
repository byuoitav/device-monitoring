import { NgModule } from "@angular/core";
import { BrowserModule } from "@angular/platform-browser";
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { RouterModule, Routes } from "@angular/router";
import { HttpClientModule } from "@angular/common/http";
import {
  MatSidenavModule,
  MatButtonModule,
  MatToolbarModule
} from "@angular/material";
import "hammerjs";

import { AppComponent } from "./components/app/app.component";
import { OverviewComponent } from "./components/overview/overview.component";

import { APIService } from "./services/api.service";

const routes: Routes = [
  {
    path: "overview",
    component: OverviewComponent
  },
  {
    path: "",
    redirectTo: "/overview",
    pathMatch: "full"
  }
];

@NgModule({
  declarations: [AppComponent, OverviewComponent],
  imports: [
    BrowserModule,
    BrowserAnimationsModule,
    HttpClientModule,
    MatSidenavModule,
    MatButtonModule,
    MatToolbarModule,
    RouterModule.forRoot(routes)
  ],
  providers: [APIService],
  bootstrap: [AppComponent]
})
export class AppModule {}
