declare namespace MatchInfo {

    export interface Tile {
        t: string;
        o: number;
        s: boolean;
    }

    export interface ServerData {
        language: number;
        usedTiles: number[];
        tiles: Tile[];
        usedWords: string[];
        minVersion: number;
    }

    export interface Participant {
        userId: string;
        userName: string;
        playerIndex: number;
        playerStatus: string;
        lastTurnStatus: string;
        matchOutcome: string;
        turnDate: Date;
        timeoutDate?: any;
        avatarURL: string;
        isFavorite: boolean;
        useBadWords: boolean;
        blockChat: boolean;
        deletedFromPlayerList: boolean;
        online: boolean;
        chatsUnread: number;
        muteChat: boolean;
        abandonedMatch: boolean;
        isBot: boolean;
        bannedChat: boolean;
    }

    export interface Match {
        matchId: string;
        matchIdNumber: number;
        matchURL: string;
        createDate: Date;
        updateDate: Date;
        matchStatus: number;
        currentPlayerIndex: number;
        letters: string;
        rowCount: number;
        columnCount: number;
        turnCount: number;
        matchData: string;
        serverData: ServerData;
        participants: Participant[];
    }

    export interface MatchInfo {
        success: boolean;
        matches: Match[];
        match: Match;
    }
}

declare module 'matchInfo' {
    import M = MatchInfo;
    export = M;
}
