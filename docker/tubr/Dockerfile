FROM alpine:3.20.2 AS build

RUN apk update
RUN apk add go

RUN mkdir /usr/app
WORKDIR /usr/app/
COPY . .

RUN go build -mod vendor -o tubr

FROM alpine:3.20.2 
WORKDIR /usr/app

COPY --from=build /usr/app/tubr ./server
COPY --from=build /usr/app/blacklist.txt .
COPY --from=build /usr/app/game-list.json .

RUN apk update
RUN apk add ffmpeg

CMD ["/usr/app/server"]