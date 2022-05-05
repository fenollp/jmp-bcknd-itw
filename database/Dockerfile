FROM postgres:14.2-alpine

ENV POSTGRES_USER jump
ENV POSTGRES_PASSWORD password

ADD schema.sql /docker-entrypoint-initdb.d/schema.sql

