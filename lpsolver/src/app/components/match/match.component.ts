import { Component, OnInit } from '@angular/core';
import { Http } from "@angular/http";
import * as matchInfo from 'matchInfo';
import 'rxjs/add/operator/map';

@Component({
  selector: 'app-match',
  templateUrl: './match.component.html',
  styleUrls: ['./match.component.css']
})
export class MatchComponent implements OnInit {
  matchIdNumber:number;
  name:string;
  data:matchInfo.MatchInfo;
  matches:matchInfo.Match[];


  constructor(private http: Http) {
    console.log("constructor do...");
  }

  ngOnInit() {
    console.log("ngOnInit...");

    this.http.get('http://localhost:8080/match')
    .map((resp) => resp.json())
    .subscribe(
      (data) => this.data = data
    )
    this.matches = this.data.matches;
    
  }

}
