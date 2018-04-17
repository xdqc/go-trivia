import { Component } from '@angular/core';
import { env } from '../environments/environment';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent {
  title = 'LP solver';
  showBrain:boolean;
  constructor(){
    this.showBrain = false;
  }
}
