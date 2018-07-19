import jieba
from collections import Counter

def get_single_char(file,noCut=False):
    words = []
    setences = set()
    with open(file,'r',encoding='utf8') as f:
        for line in f.readlines():
            setences.add(line[:300])
        for line in setences:
            if noCut:
                words.append(line)
                continue
            line = jieba.cut(line)
            for w in line:
                words.append(w)
    return Counter(words)

passed = get_single_char('passed.txt')
passed = set(passed)
failed = get_single_char('failed.txt')
failed = set(filter(lambda t: failed[t] > 0, failed))
sen = set(get_single_char('sen.txt', noCut=True))

sensored = failed.difference(passed).difference(sen)

print(len(sensored))
sensored = sorted(sensored)
with open('sensor.txt', 'w', encoding='utf8') as f:
    f.writelines([w+'\n' for w in sensored])