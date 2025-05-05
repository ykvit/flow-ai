FROM node:22.14.0-alpine

WORKDIR /usr/src/app

ENV TZ=Europe/Kyiv
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

COPY package.json ./
COPY package-lock.json ./

RUN npm install --omit=dev --no-optional

COPY . .

RUN mkdir -p data && chown node:node data

USER node

EXPOSE 3001

CMD [ "node", "server.js" ]