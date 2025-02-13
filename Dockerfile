FROM alpine:3.18

RUN apk add --no-cache tzdata 
RUN mkdir /data
RUN mkdir /app
WORKDIR /app
ENV TRENDING_CMC_DB /data/crypto-analytics.db

ADD  bin/crypto-analytics.arm64  /app/crypto-analytics
CMD ["/app/crypto-analytics"]

