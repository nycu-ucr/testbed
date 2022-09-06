"2022-08-31T03:58:49Z [DEBU][ONVM][client] Decode time: 0.015058"

import sys

fname = sys.argv[1]

fp = open(fname, "r")
res = []
for line in fp:
    line = line.replace("\n", "")
    t = float(line.split()[-1])
    res.append(t)

print(min(res), max(res), sum(res)/len(res))
