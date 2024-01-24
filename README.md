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

### Build & Install
```
make build install
```

### Build indexer
```
make build install-indexer
indexer -config config.json
```


## How to Run Indexer JSONRPC API
### Modify config_jsonrpc.json

### Build apiserver
```
make build install-jsonrpc
apiserver -config config_jsonrpc.json
```



