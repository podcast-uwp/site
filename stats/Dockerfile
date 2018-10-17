FROM umputun/baseimage:micro-latest

RUN apk add --update --no-cache vnstat

COPY exec.sh /exec.sh
RUN chmod +x /exec.sh

CMD ["/exec.sh"]
