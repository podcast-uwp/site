FROM alpine:3.8

RUN apk add --update --no-cache vnstat

COPY exec.sh /exec.sh
RUN chmod +x /exec.sh
ADD index.html /index.html

CMD ["/exec.sh"]
