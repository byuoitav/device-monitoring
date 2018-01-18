import { Component, Input } from '@angular/core';

@Component({
	selector: 'mybtn',
	template: `
		<div class="container" [style.background-color]="color">
			<i class="material-icons">{{icon}}</i>
			<span>{{name}}</span>
		</div>
	`,
	styles: [`
		.container {
			height: 9vh;
			width: 20vh;
			border-radius: 2vh;
			border: 1px solid transparent;

			display: inline-flex;
			vertical-align: middle;
			flex-direction: row;
			justify-content: space-around;	
			align-items: center;
		}
	`]
})
export class ButtonComponent {
	@Input() name: string;	
	@Input() icon: string;	
	@Input() color: string;
}
