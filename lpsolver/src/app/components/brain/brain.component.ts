import { Component, OnInit, OnDestroy, Injectable } from '@angular/core';
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
@Injectable()
export class BrainComponent implements OnInit, OnDestroy {

  total: number;
  q: qInfo.QuestionInfo;
  quiz: qInfo.QuestionInfo['data']['quiz'];
  options: qInfo.QuestionInfo['data']['options'];
  qNum: qInfo.QuestionInfo['data']['num'];
  qSchool: qInfo.QuestionInfo['data']['school'];
  qType: qInfo.QuestionInfo['data']['type'];
  qChoice: qInfo.Caldata['Choice'];
  qVoice: qInfo.Caldata['Voice'];
  ans: qInfo.Caldata['Answer'];
  ansPos: qInfo.Caldata['AnswerPos'];
  ansUser: qInfo.Caldata['User'];
  odds: qInfo.Caldata['Odds'];
  highestOdd: number;
  imgTime: qInfo.Caldata['ImageTime'] = 0;

  speakOn: boolean;
  volume: number;
  language: string;
  quotes:Quotes.Quote[];

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
    this.qVoice = 3;
    this.getQuoteData().subscribe(data => {this.quotes=(data as Quotes.Quote[]);console.log(this.quotes)}, error => console.log(error));
  }

  ngOnInit() {
    console.log("brain start...");

    this.fetch = setInterval(() => this.fetchQuestion(), 1000);
    // this.fetch = setInterval(() => this.fetchIdiom(), 3000)

  }



  ngOnDestroy(): void {
    console.log("brain end");
    clearInterval(this.fetch);
    speechSynthesis.cancel();
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
            this.qSchool = this.q.data.school;
            this.qType = this.q.data.type;
            this.qChoice = this.q.caldata.Choice;
            this.ansUser = this.q.caldata.User;
            this.odds = this.q.caldata.Odds;
            if (this.q.caldata.Voice > 0 && this.q.caldata.Voice <= 6){
              this.qVoice = this.q.caldata.Voice - 1;
            }
            if (this.odds != null) {              
              for (let i = 0; i < this.odds.length; i++) {
                let n = parseFloat(this.odds[i])
                this.odds[i] = n >= 999 ? "999" : n >= 666 ? n.toFixed(0) : n > 0.005 ? n.toFixed(2) : "0";
              }
            }
            this.changeQuizAnsBackground(this.q.caldata.ImageTime)
            this.speechText(this)
            this.quiz = this.q.data.quiz;
            this.ansPos = this.q.caldata.AnswerPos;
            this.qNum = this.q.data.num;
          }
        }
      );
  }

  fetchOCR(){
    this.http.post('http://' + env.host + ':' + env.port + '/brain-ocr', null).subscribe();
  }

  voiceOn() {
    if (this.speakOn) {
      speechSynthesis.speak(new SpeechSynthesisUtterance("Voice on"))
    } else (
      speechSynthesis.speak(new SpeechSynthesisUtterance("Voice off"))
    )
  }

  speechText(that) {
    if (that.speakOn && (that.quiz !== that.q.data.quiz || that.quiz+that.ansPos !== that.q.data.quiz+that.q.caldata.AnswerPos)) {
      const en = speechSynthesis.getVoices().filter(v => v.lang.indexOf('en') >= 0);
      const zh = speechSynthesis.getVoices().filter(v => v.lang.indexOf('zh') >= 0);

      // speak out game over
      if (that.q.data.quiz == "game over") {
        let quote = this.quotes[Math.floor(Math.random()*this.quotes.length)];
        let sayGG = new SpeechSynthesisUtterance(quote.text+" "+(quote.author=="Unknown"?"":quote.author));
        sayGG.voice = en[Math.floor(Math.random() * en.length)];
        sayGG.volume = (that.volume || 100) / 80;
        sayGG.rate = 0.9;
        speechSynthesis.speak(sayGG)
        console.log(that.q.data.quiz)
        return
      }

      // speak out quiz school
      // if (that.q.data.school != "") {        
      //   let saySchool = new SpeechSynthesisUtterance(that.q.data.type + "题");
      //   saySchool.voice = zh[0];
      //   speechSynthesis.speak(saySchool);
      // }

      // speak out new question answer
      this.highestOdd = 0;
      that.odds.forEach(n => this.highestOdd = parseFloat(n) > this.highestOdd ? parseFloat(n) : this.highestOdd)
      let utterance = this.highestOdd == 444 ? 'google ' : this.highestOdd == 333 ? 'should be ' : this.highestOdd == 888||that.highestOdd == 666 ? 'choose ' :that.highestOdd > 100? 'absolutely ':that.highestOdd > 10? 'definitely ':that.highestOdd > 3? 'exactly ': that.highestOdd > 1 ? 'probably ': that.highestOdd > 0.5 ? 'possibly ' : 'perhaps ';
      if (that.q.data.school == '理科' && this.highestOdd < 1) {
        utterance = 'Attention, ' + utterance
      }

      let sayNumber = new SpeechSynthesisUtterance(utterance + that.q.caldata.AnswerPos + '. ');
      
      sayNumber.voice = en[Math.floor(Math.random() * en.length)];
      sayNumber.volume = (that.volume || 100) / 80;

      if (this.highestOdd >= 1) {
        let sayChoice = new SpeechSynthesisUtterance(that.q.data.quiz+that.q.caldata.Answer);//
        sayChoice.voice = zh[this.qVoice];
          // /[\u4E00-\u9FA5\uF900-\uFA2D]/.test(that.q.caldata.Answer)
          // ? zh[Math.floor(Math.random() * zh.length)]
          // : sayNumber.voice;
        sayChoice.rate = 1.05;
        sayChoice.pitch = 1;
        sayChoice.volume = (that.volume || 100) / 100;
        that.language = sayNumber.voice.lang
        speechSynthesis.speak(sayChoice)
      } else {
        let sayChoice = new SpeechSynthesisUtterance(that.q.data.quiz);//
        sayChoice.voice = zh[this.qVoice];
          // /[\u4E00-\u9FA5\uF900-\uFA2D]/.test(that.q.caldata.Answer)
          // ? zh[Math.floor(Math.random() * zh.length)]
          // : sayNumber.voice;
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

  getQuoteData(){
    let apiUrl = 'solver/assets/quotes.json';
    return this.http.get(apiUrl)
    .map((res:any) => res.json());
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
