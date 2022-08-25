# Calibre-Api


## 打包

`docker build -t calibre-api:latest -f ./Dockerfile .`

## 配置

### 配置文件
配置文件会按优先级从下面查找:
- `/etc/calibre-api/config.yaml`
- `$HOME/.calibre-api`
- `./config.yaml`

配置内容
```yaml
address: :8080
debug: false
search:
  host: http://127.0.0.1:7700
  apikey:
  index: books
  trimPath: /data/book/calibre/library
storage:
  use: webdav
  tmpdir: ".files"
  webdav:
    host: http://xxxxxx
    user: xxxxx
    password: "xxxxxxxxxxxxx"
    path: /book/calibre/library

```

### 环境变量
环境变量优先于配置文件，可以使用环境变量覆盖配置文件中的参数

```
CALIBRE_ADDRESS
CALIBRE_DEBUG

## search
CALIBRE_SEARCH_HOST
CALIBRE_SEARCH_APIKEY
CALIBRE_SEARCH_INDEX
CALIBRE_SEARCH_TERMPATH

## storage
CALIBRE_STORAGE_USE
CALIBRE_STORAGE_TMPDIR

## webdav
CALIBRE_STORAGE_WEBDAV_HOST
CALIBRE_STORAGE_WEBDAV_USER
CALIBRE_STORAGE_WEBDAV_PASSWORD
CALIBRE_STORAGE_WEBDAV_PATH
```

## 适配阅读书源
```json
{
    "bookSourceUrl": "http://192.168.2.4:8080",
    "bookSourceType": 0,
    "bookSourceName": "calibre书库",
    "bookSourceGroup": "calibre",
    "bookSourceComment": "",
    "loginUrl": "",
    "loginUi": "",
    "loginCheckJs": "",
    "concurrentRate": "",
    "header": "",
    "bookUrlPattern": "",
    "searchUrl": "search?q={{key}}&sort=id:desc",
    "exploreUrl": "",
    "enabled": true,
    "enabledExplore": false,
    "weight": 0,
    "customOrder": 0,
    "lastUpdateTime": 1661322926750,
    "ruleSearch": {
        "bookList": "$.hits",
        "name": "$.title",
        "author": "$.authors",
        "intro": "$.comments",
        "coverUrl": "/get/cover/{{$.id}}.jpg",
        "bookUrl": "/book/{{$.id}}"
    },
    "ruleExplore": {},
    "ruleBookInfo": {
        "name": "$.title",
        "author": "$.authors",
        "intro": "$.comments",
        "coverUrl": "/get/cover/{{$.id}}.jpg",
        "tocUrl": "/read/{{$.id}}/toc"
    },
    "ruleToc": {
        "chapterList": "$.points",
        "chapterName": "$.text",
        "chapterUrl": "$.content.src"
    },
    "ruleContent": {
        "content": "//body"
    }
}
```