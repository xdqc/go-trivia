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

  playerName: string;

  data: matchInfo.MatchInfo;
  matches: matchInfo.MatchInfo['matches'];
  userNames: matchInfo.Participant['userName'][];
  letterGrids: matchInfo.Match['letters'][];
  tileGrids: matchInfo.Tile[][];

  selected: string[];

  constructor(private http: Http) {
    console.log("constructor do...");
    this.playerName = "semiconductor"
  }

  ngOnInit() {
    this.http.get('http://localhost:8080/match')
      .map((resp) => resp.text() !== "" ? resp.json() : "")
      .subscribe(
        (data) => {
          this.data = data;
          if (this.data !== undefined) {
            if (data.matches !== undefined) {
              // fetch game list
              this.matches = data.matches;
            } else if (data.match !== undefined) {
              // fetch newly created game
              this.matches = [data.match]
            }
            // sort new turns on top
            this.matches.sort((b, a) => Math.max(a.participants[0].turnDate.valueOf(), a.participants[1].turnDate.valueOf())
              - Math.max(b.participants[0].turnDate.valueOf(), b.participants[1].turnDate.valueOf()))
            console.log(this.matches);

            // only show matches on my turn
            this.matches = this.matches.filter(m => m.participants[m.currentPlayerIndex].userName === this.playerName)

            this.userNames = this.matches.map(m => m.participants[0]['userName']);
            console.log(this.userNames);
            this.letterGrids = this.matches.map(m => m['letters']);
            console.log(this.letterGrids);
            this.tileGrids = this.matches.map(m => m['serverData']['tiles'])
              .map(t => t.slice(20, 25).concat(t.slice(15, 20).concat(t.slice(10, 15).concat(t.slice(5, 10).concat(t.slice(0, 5))))));
            console.log(this.tileGrids);

            // set surronded tiles
            for (let k = 0; k < this.tileGrids.length; k++) {
              const tg = this.tileGrids[k];
              for (let i = 0; i < 25; i++) {
                tg[i].s = [
                  ([4, 9, 14, 19, 24].includes(i)) ? undefined : tg[i + 1],
                  ([0, 5, 10, 15, 20].includes(i)) ? undefined : tg[i - 1],
                  tg[i + 5],
                  tg[i - 5]
                ]
                  .filter(t => t !== undefined)
                  .every(t => t.o == tg[i].o)
              }
            }

            this.selected = Array<string>(this.matches.length);
            for (let i = 0; i < this.matches.length; i++) {
              this.selected[i] = "_".repeat(25);
            }

          }
        })
  }

  fetchGames() {
    this.ngOnInit()
  }

  selectLetter(e: HTMLInputElement, i: number, n: number) {
    if (e.checked) {
      this.selected[i] = this.selected[i].substr(0, n) + e.value + this.selected[i].substr(n + 1);
    } else {
      this.selected[i] = this.selected[i].substr(0, n) + '_' + this.selected[i].substr(n + 1);
    }
  }

}
