import { Component, OnInit } from '@angular/core';
import { Http, Headers } from '@angular/http';
import { environment } from '../../../environments/environment';
import * as matchInfo from 'matchInfo';
import 'rxjs/add/operator/map';

@Component({
  selector: 'app-match',
  templateUrl: './match.component.html',
  styleUrls: ['./match.component.css']
})
export class MatchComponent implements OnInit {

  playersId: string[];

  matches: matchInfo.MatchInfo['matches'];
  opponentNames: matchInfo.Participant['userName'][];
  letterGrids: matchInfo.Match['letters'][];
  tileGrids: matchInfo.Tile[][];

  selectedTile: boolean[][];
  foundWords: string[][];
  choosingWord: string[];

  constructor(private http: Http) {
    console.log('constructor do...');
    this.playersId = environment.player.map(p => p.id);
  }

  ngOnInit() {
    this.fetchGames();
  }

  fetchGames() {
    this.http.get('http://localhost:8080/match')
      .map((resp) => resp.text() !== '' ? resp.json() : '')
      .subscribe(
        (data) => {
          if (data) {
            this.processGameData(data);
          }
        });
  }

  processGameData(data) {
    if (data.matches) {
      // fetch exsisting game list
      this.matches = data.matches;
    } else if (data.match) {
      // fetch newly created game
      this.matches = [data.match];
    }

    // only show matches on my turn
    this.matches = this.matches.filter(m => this.playersId.includes(m.participants[m.currentPlayerIndex].userId));

    // sort new turns on top
    this.matches.sort((b, a) => Math.max(a.participants[0].turnDate.valueOf(), a.participants[1].turnDate.valueOf())
      - Math.max(b.participants[0].turnDate.valueOf(), b.participants[1].turnDate.valueOf()));
    console.log(this.matches);

    this.opponentNames = this.matches.map(m => this.playersId.includes(m.participants[0].userId) ? m.participants[1].userName : m.participants[0].userName);
    console.log(this.opponentNames);
    this.letterGrids = this.matches.map(m => m['letters']);
    console.log(this.letterGrids);
    this.tileGrids = this.matches.map(m => m['serverData']['tiles'])
      .map(t => t.slice(20, 25).concat(t.slice(15, 20).concat(t.slice(10, 15).concat(t.slice(5, 10).concat(t.slice(0, 5))))));


    for (let i = 0; i < this.tileGrids.length; i++) {
      const tg = this.tileGrids[i];
      // reverse owner, if player moves first
      if (this.playersId.includes(this.matches[i].participants[0].userId)) {
        tg.forEach(t => t.o == 1 ? t.o = 0 : t.o == 0 ? t.o = 1 : t.o)
      }

      // set surronded tiles
      for (let k = 0; k < 25; k++) {
        tg[k].s = [
          ([4, 9, 14, 19, 24].includes(k)) ? undefined : tg[k + 1],
          ([0, 5, 10, 15, 20].includes(k)) ? undefined : tg[k - 1],
          tg[k + 5],
          tg[k - 5]
        ]
          .filter(t => t)
          .every(t => t.o === tg[k].o);
      }
    }

    this.selectedTile = Array<boolean[]>(this.matches.length);

    // initialze seleted tiles
    for (let i = 0; i < this.matches.length; i++) {
      this.selectedTile[i] = Array<boolean>(25);
      const tg = this.tileGrids[i];
      for (let k = 0; k < 25; k++) {
        // auto select unsurrounded opponent's tiles
        this.selectedTile[i][k] = (tg[k].o == 0 && !tg[k].s);
      }
    }


    this.foundWords = Array<string[]>(this.matches.length);
    this.choosingWord = Array<string>(this.matches.length);
  }


  findWords(i: number) {
    const letters = this.tileGrids[i].map(t => t.t).join('').toUpperCase();
    let selected = [];
    for (let k = 0; k < 25; k++) {
      if (this.selectedTile[i][k]) {
        selected.push(letters[k])
      }
    }
    console.log(letters);
    console.log(selected.join(''));
    this.http.get('http://localhost:8080/words?selected=' + selected.join('') + '&letters=' + letters)
      .map(resp => resp.json())
      .subscribe(data => {
        this.foundWords[i] = data;
        const usedWords = this.matches[i].serverData.usedWords
        // filter out usedWords
        this.foundWords[i] = this.foundWords[i].filter(w => !usedWords.some(uw => uw.indexOf(w) === 0));
        this.choosingWord[i] = this.foundWords[i][0];
      });
  }

  clearSelected(i: number) {
    this.selectedTile[i].fill(false)
  }

  deleteWord(i: number) {
    this.http.delete('http://localhost:8080/word?delete=' + this.choosingWord[i])
    .subscribe()
    console.log(this.choosingWord, 'deleted');
  }

}
