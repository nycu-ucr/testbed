#include <time.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#define BUFFER_SIZE 2048

void server(int port) {
    int listen_sock, client_sock ;
    int nfds;
    fd_set afds, rfds;
    char buffer[BUFFER_SIZE] = {'\0'};
    struct sockaddr_in server_address; memset(&server_address, 0, sizeof(struct sockaddr_in));
    
    listen_sock = socket(AF_INET, SOCK_STREAM, IPPROTO_TCP);

    server_address.sin_family = AF_INET;
    server_address.sin_port = htons(port);
    server_address.sin_addr.s_addr = INADDR_ANY;

    int reuse = 1;
    setsockopt(listen_sock, SOL_SOCKET, SO_REUSEADDR, (const char*)&reuse, sizeof(reuse));
    bind(listen_sock, (const struct sockaddr *) &server_address, sizeof(struct sockaddr_in));
    listen(listen_sock, 2048);

    nfds = listen_sock;
    FD_SET(listen_sock, &afds);
    
	int max = 0;
    int cn = 0;
    for(;;) {
        memcpy(&rfds, &afds, sizeof(rfds));
        if (select(nfds+1, &rfds, NULL, NULL, NULL) < 0) {
            perror("select error");
            exit(0);
        }
		max = (max > nfds) ? max : nfds;
		// printf("max nfds: %d\r", max);

        for (int fd=0; fd < nfds+1; ++fd) {
            if (FD_ISSET(fd, &rfds)) {
                if (fd == listen_sock) {
                    client_sock = accept(listen_sock, NULL, NULL);
                    FD_SET(client_sock, &afds);

                    nfds = (client_sock > nfds) ? client_sock : nfds;
                    ++cn;
                    // printf("Accept connection: %d\n", cn);
                } else {
                    int bytes_received = read(fd, buffer, BUFFER_SIZE);
                    
                    if (bytes_received > 0) {
                        write(fd, buffer, bytes_received);
                    } else if (bytes_received == 0) {
                        close(fd);
                        FD_CLR(fd, &afds);
                        if (fd == nfds) {
                            nfds -= 1;
                        }
                    } else {
                        perror("Read failed");
                    }
                }
            }
        }
    }

    close(listen_sock);
}

int main(int argc, char *argv[]) {
    int portno;
    if (argc != 2) {
        fprintf(stderr, "usage: %s <port>\n", argv[0]);
        exit(1);
    }
    portno = atoi(argv[1]);

    server(portno);

    return 0;
}