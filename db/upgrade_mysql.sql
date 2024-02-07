Use
    tap_indexer;

create index idx_tx_hash
    on balance_txn (tx_hash(12));