for n in $(seq 1 10); do ./a.out s   1; ./clean.sh; done
for n in $(seq 1 10); do ./a.out s   2; ./clean.sh; done
for n in $(seq 1 10); do ./a.out s   4; ./clean.sh; done
for n in $(seq 1 10); do ./a.out s   8; ./clean.sh; done
for n in $(seq 1 10); do ./a.out s  16; ./clean.sh; done
for n in $(seq 1 10); do ./a.out s  32; ./clean.sh; done
for n in $(seq 1 10); do ./a.out s  64; ./clean.sh; done
for n in $(seq 1 10); do ./a.out s 128; ./clean.sh; done
for n in $(seq 1 10); do ./a.out s 256; ./clean.sh; done
for n in $(seq 1 10); do ./a.out s 512; ./clean.sh; done

for n in $(seq 1 10); do ./a.out c   1; done
for n in $(seq 1 10); do ./a.out c   2; done
for n in $(seq 1 10); do ./a.out c   4; done
for n in $(seq 1 10); do ./a.out c   8; done
for n in $(seq 1 10); do ./a.out c  16; done
for n in $(seq 1 10); do ./a.out c  32; done
for n in $(seq 1 10); do ./a.out c  64; done
for n in $(seq 1 10); do ./a.out c 128; done
for n in $(seq 1 10); do ./a.out c 256; done
for n in $(seq 1 10); do ./a.out c 512; done