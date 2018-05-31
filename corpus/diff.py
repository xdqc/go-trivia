import csv

stdCorpus = {}
myCorpus = {}

with open('ctx.csv', encoding='utf-8', newline='') as csvfile:
    reader = csv.reader(csvfile)
    for row in reader:
        myCorpus[row[0]] = int(row[1])

with open('CorpusWordPOSlist.csv', encoding='utf-8', newline='') as csvfile:
    reader = csv.reader(csvfile)
    for row in reader:
        stdCorpus[row[0]] = int(row[3])

oldWord = stdCorpus.keys() - filter(lambda p:myCorpus[p]>100, myCorpus)
oldList = {}

for word in oldWord:
    oldList[word] = stdCorpus[word]
oldList = sorted(oldList.items(), key=lambda w: w[1], reverse=True)


newWord = filter(lambda p:myCorpus[p]>5000, myCorpus) - stdCorpus.keys()
newList = {}

for word in newWord:
    newList[word] = myCorpus[word]
newList = sorted(newList.items(), key=lambda w: w[1], reverse=True)


print([i[0] for i in newList])