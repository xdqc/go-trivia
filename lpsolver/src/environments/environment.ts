// The file contents for the current environment will overwrite these during build.
// The build system defaults to the dev environment which uses `environment.ts`, but if you do
// `ng build --env=prod` then `environment.prod.ts` will be used instead.
// The list of which env maps to which file can be found in `.angular-cli.json`.

export const environment = {
  production: false,
  // 2018/03/22 18:12:37 GET https://solebonapi.com:443/api/1.0/lplist_matches.json?appkey=VI2ET0P4&offset=0&count=50 HTTP/1.1
  player: [
    {
      name: "Sammuel",
      Host: 'https://solebonapi.com:443',
      header: {
        'accept': '*/*',
        'accept-language': 'en-us',
        'accept-encoding': 'gzip, deflate',
        'connection': 'keep-alive',
        'authorization': 'Basic NFdVRDRYODUyMkhTUk02WDphTDY3UGJJRVJD',
        'x-solebonapi-token': 'ef5869e1954712dde93ed7bf8e865970e7b2b03374ed8c27365b1e3ff05513b9',
        'user-agent': 'Letterpress/248 CFNetwork/811.5.4 Darwin/16.7.0'
      },
      GET:{
        matchlist: '/api/1.0/lplist_matches.json?appkey=VI2ET0P4&offset=0&count=50',
        checkword: '/api/1.0/lp_check_word.json?appkey=VI2ET0P4&word=',
      }
    },
    {
      name: "semiconductor",
      Host: 'https://solebonapi.com:443',
      header: {
        'accept': '*/*',
        'accept-language': 'en-us',
        'accept-encoding': 'gzip, deflate',
        'connection': 'keep-alive',
        'authorization': 'Basic TU1TSlVLSDlTWUxRVlpEWTpFSTBDOUtiSkQw',
        'x-solebonapi-token': '35096253fa358ae5ac6d377c6680f1d974a9533c090580b9e648e6445f624c87',
        'user-agent': 'Letterpress/248 CFNetwork/811.5.4 Darwin/16.7.0'
      },
      GET:{
        matchlist: '/api/1.0/lplist_matches.json?appkey=VI2ET0P4&offset=0&count=50',
        checkword: '/api/1.0/lp_check_word.json?appkey=VI2ET0P4&word=',
      }
    },
  ]
}
