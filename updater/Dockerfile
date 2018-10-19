FROM umputun/baseimage:app-latest

ADD updater.sh /srv/updater.sh
ADD entrypoint.sh /srv/entrypoint.sh

RUN \
    apk add --update openssh-client && \
    mkdir -p /home/app/.ssh && \
    echo "StrictHostKeyChecking=no" > /home/app/.ssh/config && \
    echo "LogLevel=quiet" >> /home/app/.ssh/config && \
    chmod 600 /home/app/.ssh/* && \
    chmod 700 /home/app/.ssh  && \
    chmod +x /srv/updater.sh

CMD ["/srv/entrypoint.sh"]