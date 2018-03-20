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
  playerName = "Samuell"

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
        if (this.data !== undefined){
          if (data.matches !== undefined){
            // fetch game list
            this.matches = data.matches;
          } else if (data.match !== undefined) {
            // fetch newly created game
            this.matches = [data.match]
          }
          // sort new turns on top
          this.matches.sort((b,a) => Math.max(a.participants[0].turnDate.valueOf(), a.participants[1].turnDate.valueOf())
            - Math.max(b.participants[0].turnDate.valueOf(), b.participants[1].turnDate.valueOf()))
          console.log(this.matches);

          // only show matches on my turn
          this.matches = this.matches.filter(m=>m.participants[m.currentPlayerIndex].userName === this.playerName)

          this.userNames = this.matches.map(m => m.participants[0]['userName']);
          console.log(this.userNames);
          this.letterGrids = this.matches.map(m => m['letters']);
          console.log(this.letterGrids);
          this.tileGrids = this.matches.map(m => m['serverData']['tiles']);
          console.log(this.tileGrids);
        }
      })
  }

  fetchGames(){
    this.ngOnInit()
  }


}
