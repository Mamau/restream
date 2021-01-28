# Restream service

Сервси позволяет рестримить потоковые видео такие как `hls`, `mpeg`

Компиляция
```bash
make build
```

Запуск
```bash
docker-compose up
``` 
После `запуска` станет доступен хост http://localhost:89  

Доступные роуты:  
* /api/v1/stream-start (method: POST, params: manifest, stream)
* /api/v1/stream-stop (method: POST, params: manifest, stream)
* /api/v1/streams (method: GET)
* /api/v1/start-channel (method: POST, params: name[tnt|1tv|match])

Пример запуска стрима:
```bash
curl -d '{"name":"tnt"}' -H "Content-Type: application/json" -X POST http://localhost:89/api/v1/stream-start
``` 
При запуске роута start-channel селениум идет на соответствующий сайт, находит файл манифеста и начинает его рестримить.
