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
  MatIconModule
} from "@angular/material";
import "hammerjs";

import { IconsModule } from "./icons.module";
import { AppComponent } from "./components/app/app.component";
import { OverviewComponent } from "./components/overview/overview.component";
import { SoftwareInfoComponent } from "./components/software-info/software-info.component";
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
  }
];

@NgModule({
  declarations: [AppComponent, OverviewComponent, SoftwareInfoComponent],
  imports: [
    BrowserModule,
    BrowserAnimationsModule,
    HttpClientModule,
    IconsModule,
    MatSidenavModule,
    MatButtonModule,
    MatToolbarModule,
    MatCardModule,
    MatDividerModule,
    MatListModule,
    MatExpansionModule,
    MatIconModule,
    RouterModule.forRoot(routes)
  ],
  providers: [APIService],
  bootstrap: [AppComponent]
})
export class AppModule {}
