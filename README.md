# Restream service

Сервси позволяет рестримить потоковые видео такие как `hls`, `mpeg`

Компиляция
```bash
go build
```

Запуск
```bash
go run restream
``` 
После `запуска` станет доступен хост http://localhost:89  

Доступные роуты:  
* /api/v1/stream-start (method: POST, params: filename, stream)
* /api/v1/stream-stop (method: POST, params: filename, stream)
* /api/v1/streams (method: GET)

Пример запуска стрима:
```bash
curl -d '{"stream":"myStream", "filename":"https://matchtv.ru/vdl/playlist/133529/adaptive/1603646236/003b4d95f7db681249f9b6252da9ecdc/web.m3u8"}' -H "Content-Type: application/json" -X POST http://localhost:89/api/v1/stream-start
```
#### Обязательно требуется ffmpeg
#### Используется образ alfg/docker-nginx-rtmp 
Для стриминга используется это https://github.com/alfg/docker-nginx-rtmp