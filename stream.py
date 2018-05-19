#!/usr/bin/python

import requests
import time
import csv
import operator
from collections import Counter


def getCtxStream(freq, noCxtCounter):
    url = 'http://localhost:8080/quizContextStream'

    try:
        resp = requests.get(url=url)
        data = resp.json()
        words = data['words']
    except requests.ConnectionError as err:
        words = None
        print(err)


    if words is not None:
        noCxtCounter = 0
        counts = Counter(words)
        for key in counts.keys():
            if key in freq.keys():
                freq[key] = int(freq[key]) + int(counts[key])
            else:
                freq[key] = int(counts[key])

        freq = sorted(freq.items(), key=operator.itemgetter(1), reverse=True)
    else:
        noCxtCounter += 1
        print('no more ctx found', noCxtCounter)
    return noCxtCounter


def main():
    reader = csv.reader(open('ctx.csv', encoding='utf-8'))
    freq = {}
    for row in reader:
        if len(row) > 0:
            key = row[0]
            freq[key] = int(row[1])

    noCxtCounter = 0
    while noCxtCounter < 17:
        noCxtCounter = getCtxStream(freq, noCxtCounter)
        time.sleep(10)
        if noCxtCounter % 10 == 8:
            freqList = sorted(freq.items(), key=operator.itemgetter(1), reverse=True)
            with open('ctx.csv', 'w', encoding='utf-8') as w:
                for key, value in freqList:
                    w.write(key+','+str(value)+'\n')
                print("ctx.csv written")


main()