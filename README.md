# Indexer


[https://docs.indexs.io/](https://docs.indexs.io/)

## Module
- [x] Indexer
- [x] APIServer

## Supported
- [x] ASC-20 on Avalanche
- [x] BSC-20
- [x] PRC-20 
- [x] ERC-20 


## How to Run Indexer

### Prepare database

```
mysql -uroot -p < db/init_mysql.sql
```

### Modify config.json

### Build indexer
```
make dev-indexer-build-darwin-arm64
./bin/indexer-alpha-0.0.1 -config config.json
```


## How to Run Indexer JSONRPC API
### Modify config.json

### Build apiserver
```
make dev-apiserver-build-darwin-arm64:
./bin/apiserver-alpha-0.0.1 -config config.json
```



