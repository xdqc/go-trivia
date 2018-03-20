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
  data:matchInfo.MatchInfo;
  matches:matchInfo.MatchInfo['matches'];
  userNames:matchInfo.Participant['userName'][];
  letterGrids:matchInfo.Match['letters'][];
  tileGrids:matchInfo.Tile[][];

  constructor(private http: Http) {
    console.log("constructor do...");
  
  }

  ngOnInit() {
    this.http.get('http://localhost:8080/match')
    .map((resp) => resp.text() !== "" ? resp.json(): "")
    .subscribe(
      (data) => {this.data = data;
        if (this.data != undefined){
          this.matches = data.matches;
          console.log("this.matches");
          this.userNames = this.matches.map(m => m.participants[0]['userName']);
          console.log(this.userNames);
          this.letterGrids = this.matches.map(m => m['letters']);
          console.log(this.letterGrids);
          this.tileGrids = this.matches.map(m => m['serverData']['tiles']);
          console.log(this.tileGrids);
        }
      })
  }


}
