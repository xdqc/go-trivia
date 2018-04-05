declare module 'IdiomInfo' {

    export interface AllWaitSelect {
        hasChoose: boolean;
        word: string;
    }

    export interface Idiom {
        id: string;
        uniacid?: any;
        subjectno: string;
        subjectresult: string;
        pictureurl: string;
        subjectdesc: string;
        jamwords: string;
        userId: string;
        currentNo: number;
        scores: string;
        allWaitSelect: AllWaitSelect[];
        soldmoney: string;
        soldscores: string;
        scorespershare: string;
        scorespersubject: string;
        scorespertip: string;
        wxappname: string;
        scoresforaskforhelp: string;
    }

    export interface IdiomInfo {
        errno: number;
        message: string;
        data: Idiom;
    }
}