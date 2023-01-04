#!/bin/python3
# ./calc.py c_tcp_wireshark.txt 
# ./calc.py go_tcp_wireshark.txt 
import sys

if len(sys.argv) == 2:
    fp = open(sys.argv[1], "r")

    print("Wireshark\tOne-Way\t\tRoundTrip")
    try:
        for line in fp:
            line =  line[:-1]
            x, y, z = line.split(",")
            print("{}\t\t{}\t\t{}".format(eval(x), eval(y), z))
    except SyntaxError:
        pass
    
    fp.close()