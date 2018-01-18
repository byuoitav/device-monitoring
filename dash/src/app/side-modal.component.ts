import { Component, Input, OnInit, Output, EventEmitter } from '@angular/core';
import { trigger, state, style, animate, transition } from '@angular/animations';

@Component({
  selector: 'side-modal',
  template: `
  	<div class="modal-container" [class.hidden]="!visible" (click)="onContainerClicked($event)">
		<div class="sidebar" [@slideInOut]="animate" [class.vertical]="vertical" [class.right]="vertical && opposite" [class.horizontal]="!vertical" [class.bottom]="!vertical && opposite">
			<ng-content></ng-content>
		</div>
	</div>
  `,
  styles: [`
	  .modal-container {
		position: fixed;
		top: 0;
		left: 0;
	 	width: 100%; 
		height: 100%;

		z-index: 1500;
		background-color: rgba(0,0,0,.3);
	  }

	  .hidden {
	 	display: none; 
	  }

	  .sidebar {
	 	position: absolute; 
		background-color: white;

		display: flex;
	  }

	  .sidebar.vertical {
	 	min-height: 100%;
   		top: 0;	
		min-width: 30%;
		border-radius: 0vh .5vh .5vh 0vh;
	  }
	  .sidebar.vertical.right {
	 	right: 0; 
		border-radius: .5vh 0vh 0vh .5vh;
	  }

	  .sidebar.horizontal {
		min-width: 100%;
	 	left: 0; 
		min-height: 30%;
		border-radius: 0vh 0vh .5vh .5vh;
	  }
	  .sidebar.horizontal.bottom {
	 	bottom: 0; 
		border-radius: .5vh .5vh 0vh 0vh;
	  }
  `], 
  animations: [
	  trigger('slideInOut', [
		state('v-out', style({minWidth: '0'})),
		state('v-in', style({minWidth: '30%'})),
		state('h-out', style({minHeight: '0'})),
		state('h-in', style({minHeight: '30%'})),
	 	transition('* => *', [
			style({
			}),
			animate('300ms ease-out')
		]),
	  ]),
  ]
})
export class SideModalComponent {
	// inputs 
	@Input() vertical:boolean;
	@Input() opposite:boolean;

	// outputs
	@Output() visibleStatus = new EventEmitter<boolean>()

  	public visible: boolean;
	public animate: string;

	ngOnInit() {
		this.setVisible(false);
		if(this.vertical) {
			this.animate = "v-out";	
		} else {
			this.animate = "h-out";	
		}
	}
	
  	public show(): void {
		this.setVisible(true);

		if(this.vertical) {
			this.animate = "v-in";
		} else {
			this.animate = "h-in";	
		}
  	}

  	public hide(): void {
		if(this.vertical) {
			this.animate = "v-out";	
		} else {
			this.animate = "h-out";	
		}

		this.visibleStatus.emit(false);
		setTimeout(() => this.visible = false, 300)
  	}

  	public onContainerClicked(event: MouseEvent): void {
    	if ((<HTMLElement>event.target).classList.contains('modal-container')) {
      	this.hide();
    	}
  	}

	private setVisible(b: boolean) {
		this.visible = b;
		this.visibleStatus.emit(this.visible);
	}
}
