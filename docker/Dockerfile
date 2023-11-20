FROM alpine:3.17.1

Copy root /

RUN set -eux;  \
    chown root.root /opt/reverse -R; \
    chown root.root  /usr/local/bin/docker-entrypoint.sh; \
    chmod a+x /usr/local/bin/docker-entrypoint.sh;

WORKDIR /opt/reverse
ENTRYPOINT ["docker-entrypoint.sh"]
VOLUME /data

EXPOSE 1080
EXPOSE 8118
EXPOSE 80
CMD ["default-command"]