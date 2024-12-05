FROM golang:bookworm AS build

RUN apt update && apt install -y cmake make git build-essential gperf libssl-dev zlib1g-dev nodejs npm

RUN mkdir -p /src
WORKDIR /src

RUN git config --global http.version HTTP/1.1
RUN git clone https://github.com/YouROK/TorrSBot.git
RUN git clone https://github.com/YouROK/TorrServer.git
RUN ls

# Build bot
RUN cd /src/TorrSBot && ./build_tgbot.sh
RUN cd /src/TorrSBot && go build -ldflags='-s -w' -o torrsbot ./cmd

# Build TS
RUN npm install --global yarn
RUN go install github.com/swaggo/swag/cmd/swag@latest

RUN cd /src/TorrServer && NODE_OPTIONS=--openssl-legacy-provider go run gen_web.go
RUN cd /src/TorrServer/server && swag init -g web/server.go
RUN cd /src/TorrServer/server && go build -ldflags='-s -w' -tags=nosqlite -trimpath -o torrserver ./cmd

# Build container
FROM debian:12-slim

EXPOSE 8090

RUN mkdir -p /opt/torrserver
RUN mkdir -p /opt/torrsbot

RUN echo '{"BitTorr":{"CacheSize":96468992,"ConnectionsLimit":30,"DisableDHT":false,"DisablePEX":false,"DisableTCP":false,"DisableUPNP":false,"DisableUTP":false,"DisableUpload":false,"DownloadRateLimit":0,"EnableDLNA":false,"EnableDebug":false,"EnableIPv6":false,"EnableRutorSearch":false,"ForceEncrypt":false,"FriendlyName":"","PeersListenPort":0,"PreloadCache":13,"ReaderReadAHead":87,"RemoveCacheOnDrop":false,"RetrackersMode":1,"SslCert":"","SslKey":"","SslPort":0,"TorrentDisconnectTimeout":120,"TorrentsSavePath":"","UploadRateLimit":0,"UseDisk":false}}' > /opt/torrserver/settings.json

COPY --from=build /src/TorrServer/server/torrserver /opt/torrserver/torrserver
COPY --from=build /src/TorrSBot/tgbotapi/build/telegram-bot-api /opt/torrsbot/telegram-bot-api
COPY --from=build /src/TorrSBot/torrsbot /opt/torrsbot/torrsbot

#ENTRYPOINT ./TorrServer-linux-amd64 -r -k

