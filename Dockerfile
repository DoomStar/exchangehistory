FROM busybox

ENTRYPOINT /sre-test-semion

ENV QUOTE_PORT 80

COPY sre-test-semion /
