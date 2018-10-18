import { NgModule } from "@angular/core";
import {
  IconChevronsLeft,
  IconWifi,
  IconWifiOff,
  IconAlertCircle,
  IconCheckCircle
} from "angular-feather";

const icons = [
  IconChevronsLeft,
  IconWifi,
  IconWifiOff,
  IconAlertCircle,
  IconCheckCircle
];

@NgModule({
  exports: icons
})
export class IconsModule {}
