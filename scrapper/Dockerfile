FROM python:3.12-alpine

RUN apk add --no-cache bash ttf-dejavu ttf-liberation firefox busybox-suid && \
    addgroup --system --gid 2001 scrapper && \
    adduser --system --uid 2001 scrapper

USER scrapper:scrapper

WORKDIR /app
COPY --chown=scrapper:scrapper . .

RUN ./install.sh && \
    export CRON_CONTENT="# Execute every 5 minutes.\n*/5 * * * * /app/run.sh\n" && \
    if crontab -l; then crontab -l | sed "\$a$CRON_CONTENT" | crontab -; else printf "$CRON_CONTENT" | crontab -; fi

USER root:root
CMD ["crond", "-f"]
