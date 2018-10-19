# Сайт подкаста UWP 

[Еженедельный подкаст от Umputun](https://podcast.umputun.com)

* Построение контейнера с hugo: `docker-compose -f docker-compose-publisher.yml build`. Это надо сделать один раз, чтоб построить image который будет использоваться для построения сайта. При обновлении версии hugo процедуру надо будет повторить.
* Генерация сайта: `git pull && docker-compose -f docker-compose-publisher.yml run --rm hugo`
* Автоматическое обновление (fetch каждые 10 секунд): `nohup ./updater.sh > updater.log 2>&1 &`

### скрипты публикации подкаста

- `publisher/make_new_episode.sh` - создает шаблон нового выпуска
- `publisher/upload_mp3.sh` – загружает подкаст во все места, предварительно добавляет mp3 теги и картинку
- `publisher/deploy.sh` – добавляет в гит