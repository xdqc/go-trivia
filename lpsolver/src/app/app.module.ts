import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { HttpModule } from '@angular/http';
import { FormsModule } from "@angular/forms";

import { AppComponent } from './app.component';
import { MatchComponent } from './components/match/match.component';
import { BrainComponent } from './components/brain/brain.component';

@NgModule({
  declarations: [
    AppComponent,
    MatchComponent,
    BrainComponent,
  ],
  imports: [
    BrowserModule,
    HttpModule,
    FormsModule
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }
