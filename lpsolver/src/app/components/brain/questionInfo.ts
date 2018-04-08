declare module QuestionInfo {

    export interface Data {
        quiz: string;
        options: string[];
        num: number;
        school: string;
        type: string;
        contributor: string;
        endTime: number;
        curTime: number;
    }

    export interface Caldata {
        RoomID: string;
        Answer: string;
        AnswerPos: number;
        TrueAnswer: string;
        Odds: string[];
    }

    export interface QuestionInfo {
        data: Data
        caldata: Caldata;
        errcode: number;
    }
}

declare module 'questionInfo' {
    import Q = QuestionInfo;
    export = Q;
}