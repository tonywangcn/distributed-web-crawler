FROM apify/actor-node-puppeteer-chrome:latest
COPY ./node /home/myuser
WORKDIR /home/myuser
USER root
RUN npm i
USER myuser