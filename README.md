# Calibre-Api

基于 MeiliSearch 搭建的 calibre 书籍搜索、下载、预览
数据来源 Calibre Content Server

## 特性
- 使用 Calibre Content Server 作为数据来源
- MeiliSearch 增强查询响应速度
- 支持书籍元数据读取
- 支持下载、编辑元数据、删除书籍
- 支持在线搜索元数据


## 接口

```text
GET    /api/get/cover/:id            --> 获取书籍封面
GET    /api/get/book/:id             --> 下载书籍文件
GET    /api/read/:id/toc             --> 获取书籍目录（包含元信息、目录和地址）
GET    /api/read/:id/file/*path      --> 读取书籍中的文件
GET    /api/book/:id                 --> 获取书籍信息
GET    /api/search                   --> 搜索书籍(参数q 搜索词，更多参数参考[meilisearch search](https://docs.meilisearch.com/reference/api/search.html))
POST   /api/search                   --> 搜索书籍(参数q 搜索词，更多参数参考[meilisearch search](https://docs.meilisearch.com/reference/api/search.html))
POST   /api/index/update             --> 更新索引
```

## 数据导入

创建索引，更新索引设置，该命令仅第一次使用需要执行。
```shell
## Create index
curl -X "POST" "http://localhost:7700/indexes" \
     -H 'Content-Type: application/json' \
     -d $'{
  "uid": "books"
}'
## Update settings
curl -X "PATCH" "http://localhost:7700/indexes/books/settings" \
     -H 'Content-Type: application/json' \
     -d $'{
  "displayedAttributes": [
    "*"
  ],
  "filterableAttributes": [
    "authors",
    "file_path",
    "id",
    "last_modified",
    "pubdate",
    "publisher",
    "isbn",
    "tags"
  ],
  "searchableAttributes": [
    "title",
    "authors"
  ],
  "sortableAttributes": [
    "authors_sort",
    "id",
    "last_modified",
    "pubdate",
    "publisher"
  ]
}'
```

使用下面命令更新索引
```shell
curl -X "POST" "http://localhost:8080/index/update" -H 'Content-Type: application/json' 
```

## 接口

## 打包
- 直接打包 `make build`
- 使用Docker `docker build -t calibre-api:latest -f ./Dockerfile .`

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
staticDir: "/app/static"
tmpDir: ".files"
content:
  server: https://lib.pve.icu
search:
  host: http://127.0.0.1:7700
  apikey:
  index: books
```

### 环境变量

环境变量优先于配置文件，可以使用环境变量覆盖配置文件中的参数

```text
CALIBRE_ADDRESS
CALIBRE_DEBUG
CALIBRE_STATICDIR
CALIBRE_TMP_DIR

## search
CALIBRE_SEARCH_HOST
CALIBRE_SEARCH_APIKEY
CALIBRE_SEARCH_INDEX
```

## 适配阅读书源

支持添加书源的APP可以使用下面配置，将本服务引入

```json
{
  "bookSourceUrl": "http://localhost:8080",
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