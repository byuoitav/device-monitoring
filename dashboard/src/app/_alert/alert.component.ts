import { Component, OnInit, OnDestroy, Input } from '@angular/core';
import { Router, NavigationStart } from '@angular/router';
import { Subscription } from 'rxjs';
import { trigger, style, transition, animate } from '@angular/animations';

import { Alert, AlertType } from './alert.model';
import { AlertService } from './alert.service';

@Component({
  selector: 'alert',
  templateUrl: 'alert.component.html',
  standalone: false,
  animations: [
    trigger('slideInOut', [
      transition(':enter', [
        style({ transform: 'translateY(8px)', opacity: 0 }),
        animate('160ms ease-out', style({ transform: 'translateY(0)', opacity: 1 }))
      ]),
      transition(':leave', [
        animate('180ms ease-in', style({ transform: 'translateY(8px)', opacity: 0 }))
      ])
    ])
  ]
})
export class AlertComponent implements OnInit, OnDestroy {
  @Input() id = 'default-alert';
  @Input() fade = true;
  /** how long autoClose lasts (ms). keep in sync with setTimeout */
  @Input() autoCloseMs = 5000;

  alerts: Alert[] = [];
  alertSubscription!: Subscription;
  routeSubscription!: Subscription;

  AlertType = AlertType; // expose enum for template if needed

  constructor(private router: Router, private alertService: AlertService) {}

  ngOnInit() {
    this.alertSubscription = this.alertService.onAlert(this.id).subscribe(alert => {
      if (!alert.message) {
        this.alerts = this.alerts.filter(x => x.keepAfterRouteChange);
        this.alerts.forEach(x => delete x.keepAfterRouteChange);
        return;
      }

      this.alerts.push(alert);

      if (alert.autoClose) {
        setTimeout(() => this.removeAlert(alert), this.autoCloseMs);
      }
    });

    this.routeSubscription = this.router.events.subscribe(event => {
      if (event instanceof NavigationStart) {
        this.alertService.clear(this.id);
      }
    });
  }

  ngOnDestroy() {
    this.alertSubscription?.unsubscribe();
    this.routeSubscription?.unsubscribe();
  }

  removeAlert(alert: Alert) {
    if (!this.alerts.includes(alert)) return;
    const timeout = this.fade ? 200 : 0;
    alert.fade = this.fade;
    setTimeout(() => {
      this.alerts = this.alerts.filter(x => x !== alert);
    }, timeout);
  }

  // ----- Presentation helpers -----

  typeClass(alert: Alert) {
    return {
      'is-success': alert.type === AlertType.Success,
      'is-error': alert.type === AlertType.Error,
      'is-info': alert.type === AlertType.Info,
      'is-warning': alert.type === AlertType.Warning
    };
  }

  iconFor(alert: Alert): string {
    // using inline SVG for crisp icons without extra deps
    const base = 'width="18" height="18" viewBox="0 0 24 24" fill="currentColor"';
    switch (alert.type) {
      case AlertType.Success:
        return `<svg ${base} aria-hidden="true"><path d="M12 2a10 10 0 1 0 .001 20.001A10 10 0 0 0 12 2zm-1 14l-4-4 1.414-1.414L11 12.172l5.586-5.586L18 8l-7 8z"/></svg>`;
      case AlertType.Error:
        return `<svg ${base} aria-hidden="true"><path d="M11 7h2v6h-2V7zm0 8h2v2h-2v-2z"/><path d="M12 2C6.477 2 2 6.477 2 12s4.477 10 10 10
        10-4.477 10-10S17.523 2 12 2z"/></svg>`;
      case AlertType.Warning:
        return `<svg ${base} aria-hidden="true"><path d="M1 21h22L12 2 1 21zm12-3h-2v-2h2v2zm0-4h-2v-4h2v4z"/></svg>`;
      default:
        return `<svg ${base} aria-hidden="true"><path d="M11 17h2v2h-2v-2zm0-10h2v8h-2V7z"/><path d="M12 2C6.48 2 2 6.48
        2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2z"/></svg>`;
    }
  }

  alertRole(alert: Alert) {
    return alert.type === AlertType.Error || alert.type === AlertType.Warning ? 'alert' : 'status';
  }

  ariaLive(alert: Alert) {
    return alert.type === AlertType.Error || alert.type === AlertType.Warning ? 'assertive' : 'polite';
  }
}
