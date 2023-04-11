#!/bin/python3
import matplotlib.pyplot as plt
import csv
import numpy as np

TYPE = 'latency'
METHOD = 'C TCP epoll'

y = []
count = 0

with open(TYPE + '.csv', 'r') as csvfile:
    lines = csv.reader(csvfile, delimiter=',', quoting=csv.QUOTE_NONNUMERIC)
    for row in lines:
        y.extend(row)

# plt.plot(list(range(len(y))), y, color='g',
#          linestyle='dashed', marker='.', label=TYPE)
array_y = np.asarray(y)
percentile_y = np.percentile(array_y, 97)
avg = sum(y)/len(y)

plt.plot(list(range(len(y))), y, '.', label=TYPE)
plt.xticks(rotation=25)
plt.xlabel('times')
plt.ylabel('latency(ns)')
plt.ylim(0, 200000)
plt.axhline(avg, color='red',
            linestyle='--', linewidth=3, label='Avg: ' + str(avg))
plt.title(METHOD+' total latency (10 thread)', fontsize=20)
plt.grid()
plt.legend()
plt.show()
plt.savefig(TYPE + '.png')
