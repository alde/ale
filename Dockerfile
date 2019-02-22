FROM golang:1.11-alpine3.9

RUN mkdir -p /opt/ale
COPY ale /opt/ale/ale
WORKDIR /opt/ale
CMD [ "/opt/ale/ale" ]
