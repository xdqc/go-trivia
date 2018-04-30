import { Component, OnInit, OnDestroy } from '@angular/core';
import { Http, Headers } from '@angular/http';
import { env } from '../../../environments/environment';
import 'rxjs/add/operator/map';
import * as qInfo from 'questionInfo';
import * as idiomInfo from 'IdiomInfo';


@Component({
  selector: 'app-brain',
  templateUrl: './brain.component.html',
  styleUrls: ['./brain.component.scss'],
})

export class BrainComponent implements OnInit, OnDestroy {

  total: number;
  q: qInfo.QuestionInfo;
  quiz: qInfo.QuestionInfo['data']['quiz'];
  options: qInfo.QuestionInfo['data']['options'];
  qNum: qInfo.QuestionInfo['data']['num'];
  ans: qInfo.Caldata['Answer'];
  ansPos: qInfo.Caldata['AnswerPos'];
  odds: qInfo.Caldata['Odds'];
  imgTime: qInfo.Caldata['ImageTime'] = 0;

  speakOn: boolean;
  volume: number;
  language: string;

  showImage: boolean = true;
  imgPath: string = 'solver/assets/quiz.jpg?';

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
    this.fetch = setInterval(() => this.fetchQuestion(), 500)
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
            this.changeQuizAnsBackground(this.q.caldata.ImageTime)
            this.speechText(this)
            this.quiz = this.q.data.quiz;
            this.qNum = this.q.data.num;
          }
        }
      );
  }

  fetchOCR() {
    this.http.put('http://' + env.host + ':' + env.port + '/brain-ocr', null).subscribe();
  }

  voiceOn() {
    if (this.speakOn) {
      speechSynthesis.speak(new SpeechSynthesisUtterance("Voice on"))
    } else (
      speechSynthesis.speak(new SpeechSynthesisUtterance("Voice off"))
    )
  }

  speechText(that) {
    if (that.speakOn && that.quiz !== that.q.data.quiz) {
      // speak out new question answer
      let higestOdd = 0
      that.odds.forEach(n => higestOdd = parseFloat(n) > higestOdd ? parseFloat(n) : higestOdd)
      let utterance = higestOdd == 444 ? 'google ' : higestOdd == 333 ? 'should be ' : higestOdd == 888 ? 'choose ' :higestOdd > 20? 'absolutely ':higestOdd > 7? 'definitely ':higestOdd > 3? 'exactly ': higestOdd > 1 ? 'probably ': higestOdd > 0.5 ? 'possibly ' : 'perhaps ';
      if (that.q.data.school == '理科' && higestOdd < 1) {
        utterance = 'Attention, ' + utterance
      }

      let sayNumber = new SpeechSynthesisUtterance(utterance + that.q.caldata.AnswerPos + '. ');
      const en = speechSynthesis.getVoices().filter(v => v.lang.indexOf('en') >= 0);
      
      sayNumber.voice = en[Math.floor(Math.random() * en.length)];
      sayNumber.volume = (that.volume || 100) / 80;
      speechSynthesis.speak(sayNumber)

      if (higestOdd >= 1) {
        const zh = speechSynthesis.getVoices().filter(v => v.lang.indexOf('zh') >= 0);
        let sayChoice = new SpeechSynthesisUtterance(that.q.caldata.Answer);//+ that.q.data.quiz 
        sayChoice.voice = /[\u4E00-\u9FA5\uF900-\uFA2D]/.test(that.q.caldata.Answer)
          ? zh[Math.floor(Math.random() * zh.length)]
          : sayNumber.voice;
        sayChoice.rate = 1.05;
        sayChoice.pitch = 1;
        sayChoice.volume = (that.volume || 100) / 100;
        that.language = sayNumber.voice.lang
        speechSynthesis.speak(sayChoice)
      }
    }
  }

  changeQuizAnsBackground(newImgTime: number) {
    if (newImgTime > this.imgTime && this.showImage) {
      this.imgTime = newImgTime;
      let sheet = document.styleSheets[document.styleSheets.length - 1] as CSSStyleSheet
      sheet.addRule('.bg-img[_ngcontent-c1]::before', 'background-image: url("' + this.imgPath + this.imgTime + '")', 0);
      sheet.deleteRule(1)
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
