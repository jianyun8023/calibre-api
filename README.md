# Calibre-Api

基于meilisearch搭建的calibre书籍搜索、下载、预览
文件存储支持webdav、local、s3

## TODO
- 使用 [bleve](https://github.com/blevesearch/bleve) 替换掉 meilisearch 搜索，降低依赖
- 支持直接读取calibre db数据到bleve
- 支持EPUB虚拟路径

## 接口

```text
GET    /get/cover/:id            --> 获取书籍封面
GET    /get/book/:id             --> 下载书籍文件
GET    /read/:id/toc             --> 获取书籍目录（包含元信息、目录和地址）
GET    /read/:id/file/*path      --> 读取书籍中的文件
GET    /book/:id                 --> 获取书籍信息
GET    /search                   --> 搜索书籍(参数q 搜索词，更多参数参考[meilisearch search](https://docs.meilisearch.com/reference/api/search.html))
POST   /search                   --> 搜索书籍(参数q 搜索词，更多参数参考[meilisearch search](https://docs.meilisearch.com/reference/api/search.html))
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
    "formats",
    "id",
    "last_modified",
    "pubdate",
    "publisher",
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

使用calibredb将数据导出为json，
然后上传到meilisearch中

```shell
maxID=(curl -s "http://localhost:7700/indexes/books/search?q&limit=1&sort=id%3Adesc&attributesToRetrieve=id" | jq '.hits[0].id')
calibredb --with-library=library list -f all  --for-machine --search="id:>$maxID" >> data.json
curl \
  -X PUT 'http://localhost:7700/indexes/books/documents' \
  -H 'Content-Type: application/json' \
  --data-binary @data.json
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
search:
  host: http://127.0.0.1:7700
  apikey:
  index: books
  trimPath: /data/book/calibre/library
storage:
  ## 选择文件存储的位置 webdav\local\minio
  use: webdav
  tmpdir: ".files"
  ## webdav配置
  webdav:
    host: http://xxxxxx:2344
    user: xxxxx
    password: "xxxxxxxxxxxxx"
    path: /book/calibre/library
  ## 本地路径
  local:
    path: /book/calibre/library
  ## minio s3或者s3兼容存储
  minio:
    endpoint: xxx
    accessKeyID: xxx
    secretAccessKey: xxxx
    useSSL: true
    bucketName: bucket
    path: /book/calibre/library




```

### 环境变量

环境变量优先于配置文件，可以使用环境变量覆盖配置文件中的参数

```text
CALIBRE_ADDRESS
CALIBRE_DEBUG

## search
CALIBRE_SEARCH_HOST
CALIBRE_SEARCH_APIKEY
CALIBRE_SEARCH_INDEX
CALIBRE_SEARCH_TRIMPATH

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
