import { Component, OnInit, OnDestroy } from '@angular/core';
import { Http, Headers } from '@angular/http';
import { environment } from '../../../environments/environment';
import 'rxjs/add/operator/map';
import { Observable } from 'rxjs/Rx';
import * as qInfo from 'questionInfo';
import * as idiomInfo from 'IdiomInfo';


@Component({
  selector: 'app-brain',
  templateUrl: './brain.component.html',
  styleUrls: ['./brain.component.scss']
})
export class BrainComponent implements OnInit, OnDestroy {

  total: number;
  q: qInfo.QuestionInfo;
  quiz: qInfo.QuestionInfo['data']['quiz'];
  options: qInfo.QuestionInfo['data']['options'];
  qNum: qInfo.QuestionInfo['data']['num'];
  ans: qInfo.Caldata['Answer']
  ansPos: qInfo.Caldata['AnswerPos']

  speakOn: boolean;

  fetch

  //Idioms
  rawIdioms: string;
  idiom: idiomInfo.Idiom;
  idioms: idiomInfo.Idiom[];

  constructor(private http: Http) {
    this.total = 5;
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
    this.http.get('http://localhost:8080/answer')
      .map((resp) => resp.text() !== '' ? resp.json() : '')
      .subscribe(
        data => {
          if (data) {
            this.q = data;
            if (this.speakOn && this.quiz !== this.q.data.quiz) {
              // speak out new question answer
              let msg = new SpeechSynthesisUtterance('选' + this.q.caldata.AnswerPos + '。 ');//+ this.q.data.quiz + this.q.caldata.Answer
              msg.voice = speechSynthesis.getVoices().filter(v => v.lang === 'zh-CN')[0]
              msg.rate = 1.2
              msg.pitch = 0.96
              msg.volume = 0.50
              // console.log(msg);
              speechSynthesis.speak(msg)
            }

            this.quiz = this.q.data.quiz;
            this.qNum = this.q.data.num;
            this.options = this.q.data.options;
            this.ans = this.q.caldata.Answer;
            this.ansPos = this.q.caldata.AnswerPos;
          }
        }
      );
  }

  //process idioms json
  fetchIdiom() {
    this.http.get("http://localhost:8080/idiom")
      .map((resp) => resp.text() !== '' ? resp.json() : '')
      .subscribe(data => {
        console.log('ddd' + data['data']);
        if (data['data'] != 1) {
          this.idiom = data['data'];
        }
      });
  }

  showIdioms(){
    let data = JSON.parse(this.rawIdioms);
    this.idioms = data;
    this.idioms.forEach(i => i['words'] = i['words'].split(''))
    console.log(data);
  }

}
