sslcheck
========

sslcheck allows you to check ssl keys with or without DNS resolution. The company I work for requires that all servers
have different ssl keys.  So if you've got two, four or more servers that sit behind a load balancer, it makes it hard
to inspect the keys on each server individually as the load balancer is where the DNS name for that SSL key resides.

Sslcheck takes two parameters -i <ip address> -p <port> -d <dns name>.  The -i is optional, and if it's passed sslcheck will not
perform any dns lookups, while the -d is manditory. The -p is for servers that might be running on a port other than 443.

Sslcheck is written in Go and should compile without the need of additional packages.