"2022-08-31T04:23:20Z [INFO][client][App] [Delay] Write=0.003643, Read=2.337646"

import sys
ws, rs = [], []
avgs = ""
with open(sys.argv[1], "r") as fp:
    for line in fp:
        if "Average" in line:
            avgs = " ".join(line.split()[-3:])
        else: 
            line = line.replace("\n", "")
            t = line.split()
            ws.append(float(t[3][6:].replace(",", "")))
            rs.append(float(t[4][5:]))

print("Write:")
print("\tmin = {}\n\tmax = {}\n\tavg = {}".format(min(ws), max(ws), sum(ws)/len(ws)))
print("Read:")
print("\tmin = {}\n\tmax = {}\n\tavg = {}".format(min(rs), max(rs), sum(rs)/len(rs)))
print(avgs)
