import matplotlib.pyplot as plt
from wordcloud import WordCloud
from random import shuffle
import math


words = []
corpus = {}
with open('CorpusWordPOSlist.csv','r', encoding='utf-8') as r:
    for line in r.readlines():
        wc = tuple(line.strip().split(','))
        corpus[wc[0]] = wc

# print(corpus)
 
with open('ctx.csv', 'r', encoding='utf-8') as r:
    for line in r.readlines():
        word, count = line.split(',')[0], int(line.split(',')[1])
        if len(word) > 1 and 900000 > count > 100:
            if not word in corpus.keys():
                words.extend([word]*int(math.log(count)))
            elif any(x in corpus[word][1] for x in ['a', 'n']):
                words.extend([word]*int(math.log(count)))

shuffle(words)
words = ' '.join(words)
stopwords = ['百度','百科']
my_wordcloud = WordCloud(relative_scaling=0.6, width=2560, height=1440, max_words=20000, margin=0, background_color='white', stopwords=stopwords).generate(words)

plt.imshow(my_wordcloud)
plt.axis("off")
plt.savefig('wordcloud.png')
