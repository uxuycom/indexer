FROM mysql:8.0

MAINTAINER Max N <max@uxuy.com>

VOLUME ["/var/lib/mysql"]

RUN sed -i "s/^user.*/user = root/g" /etc/mysql/my.cnf

RUN chown -R mysql /var/lib/mysql
RUN chgrp -R mysql /var/lib/mysql