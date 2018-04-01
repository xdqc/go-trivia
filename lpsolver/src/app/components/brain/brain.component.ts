import { Component, OnInit, OnDestroy } from '@angular/core';
import { Http, Headers } from '@angular/http';
import { environment } from '../../../environments/environment';
import * as qInfo from 'questionInfo';
import 'rxjs/add/operator/map';
import { Observable } from 'rxjs/Rx';


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

  speakOn:boolean;

  fetch

  constructor(private http: Http) {
    this.total = 5;
  }

  ngOnInit() {
    console.log("brain start...");
    this.fetch = setInterval(()=>this.fetchQuestion(), 1000)
  }


  ngOnDestroy(): void {
    console.log("brain end");
    clearInterval(this.fetch)
  }

  fetchQuestion() {
    this.http.get('http://localhost:8080/answer')
      .map((resp) => resp.text() !== '' ? resp.json() : '')
      .subscribe(
        data => {
          if (data) {
            this.q = data;
            if (this.speakOn && this.qNum !== this.q.data.num) {
              // speak out new question answer
              let msg = new SpeechSynthesisUtterance('选' + this.q.caldata.AnswerPos + '。 ' );//+ this.q.data.quiz + this.q.caldata.Answer
              msg.voice = speechSynthesis.getVoices().filter(v => v.lang === 'zh-CN')[0]
              msg.rate = 1.2
              msg.pitch = 0.96
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
      )
  }

}
