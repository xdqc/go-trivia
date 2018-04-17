import { Component, OnInit, OnDestroy, state, style, trigger, transition, animate } from '@angular/core';
import { Http, Headers } from '@angular/http';
import { env } from '../../../environments/environment';
import 'rxjs/add/operator/map';
import { Observable } from 'rxjs/Rx';
import * as qInfo from 'questionInfo';
import * as idiomInfo from 'IdiomInfo';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { timeout } from 'rxjs/operator/timeout';



@Component({
  selector: 'app-brain',
  templateUrl: './brain.component.html',
  styleUrls: ['./brain.component.scss'],
  animations:[
    trigger('langAnimation', [
        state('hide', style({
          opacity: 0,
        })),
        state('emerge', style({
          opacity: 1,
        })),
        transition('hide <=> emerge', animate('3000ms ease-in-out')),
    ]),
  ]
  
})

export class BrainComponent implements OnInit, OnDestroy {

  total: number;
  q: qInfo.QuestionInfo;
  quiz: qInfo.QuestionInfo['data']['quiz'];
  options: qInfo.QuestionInfo['data']['options'];
  qNum: qInfo.QuestionInfo['data']['num'];
  ans: qInfo.Caldata['Answer']
  ansPos: qInfo.Caldata['AnswerPos']
  odds: qInfo.Caldata['Odds']

  speakOn: boolean;
  volume: number;
  language: string;
  state: string = 'hide';

  fetch

  //Idioms
  rawIdioms: string;
  idiom: idiomInfo.Idiom;
  idioms: idiomInfo.Idiom[];

  constructor(private http: Http) {
    this.total = 5;
    this.speakOn = false;
  }

  ngOnInit() {
    console.log("brain start...");
    this.fetch = setInterval(() => this.fetchQuestion(), 1000)
    // this.fetch = setInterval(() => this.fetchIdiom(), 3000)
  }



  ngOnDestroy(): void {
    console.log("brain end");
    clearInterval(this.fetch)
  }

  //periodically fetch new question data
  fetchQuestion() {
    this.http.get('http://' + env.host + ':' + env.port + '/answer')
      .map((resp) => resp.text() !== '' ? resp.json() : '')
      .subscribe(
        data => {
          if (data) {
            this.q = data;
            this.options = this.q.data.options;
            this.ans = this.q.caldata.Answer;
            this.ansPos = this.q.caldata.AnswerPos;
            this.odds = this.q.caldata.Odds;
            for (let i = 0; i < this.odds.length; i++) {
              let n = parseFloat(this.odds[i])
              this.odds[i] = n >= 999 ? "999" : n >= 888 ? n.toFixed(0) : n > 0.005 ? n.toFixed(2) : "0";
            }
            this.speech_text(this)
            this.quiz = this.q.data.quiz;
            this.qNum = this.q.data.num;

          }
        }
      );
  }

  fetchOCR() {
    this.http.put('http://' + env.host + ':' + env.port + '/brain-ocr', null).subscribe();
  }

  voiceOn(){
    if (this.speakOn){
      speechSynthesis.speak(new SpeechSynthesisUtterance("Voice on"))
    } else (
      speechSynthesis.speak(new SpeechSynthesisUtterance("Voice off"))
    )
  }

  speech_text(that) {
    if (that.speakOn && that.quiz !== that.q.data.quiz) {
      // speak out new question answer
      let higestOdd = 0
      that.odds.forEach(n => higestOdd = parseFloat(n) > higestOdd ? parseFloat(n) : higestOdd)
      let utterance = higestOdd == 444 ? 'google ' : higestOdd == 333 ? 'record ' : higestOdd > 5 ? 'choose ' : 'could be ';
      if (that.q.data.school == '理科' && higestOdd < 5) {
        utterance = 'Attention, ' + utterance
      }
      let sayNumber = new SpeechSynthesisUtterance(utterance + that.q.caldata.AnswerPos + '. ')
      let sayChoice = new SpeechSynthesisUtterance(that.q.caldata.Answer + '。'+ that.q.data.quiz);//+ that.q.data.quiz 
      let chinesesNum = speechSynthesis.getVoices().filter(v => v.lang.indexOf('zh')>=0).length
      sayChoice.voice = speechSynthesis.getVoices().filter(v => v.lang.indexOf('zh')>=0)[Math.floor(Math.random()*chinesesNum)]
      sayChoice.rate = 1.05
      sayChoice.pitch = 1
      sayChoice.volume = (that.volume || 100) / 100
      // console.log(msg);
      sayNumber.voice = speechSynthesis.getVoices()[Math.floor(Math.random()*speechSynthesis.getVoices().length)]
      that.language = sayNumber.voice.lang +' '+ sayNumber.voice.name
      that.state = 'emerge'
      setTimeout(function() {
        that.state = 'hide'
      },5000)
      speechSynthesis.speak(sayNumber)
      speechSynthesis.speak(sayChoice)
    }
  }


  //process idioms json
  fetchIdiom() {
    this.http.get('http://' + env.host + ':' + env.port + '/idiom')
      .map((resp) => resp.text() !== '' ? resp.json() : '')
      .subscribe(data => {
        console.log('ddd' + data['data']);
        if (data['data'] != 1) {
          this.idiom = data['data'];
        }
      });
  }
  showIdioms() {
    let data = JSON.parse(this.rawIdioms);
    this.idioms = data;
    this.idioms.forEach(i => i['words'] = i['words'].split(''))
    console.log(data);
  }
}
