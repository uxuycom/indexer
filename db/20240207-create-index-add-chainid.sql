Use
    tap_indexer;

create index idx_tx_hash on balance_txn (tx_hash(12));

ALTER TABLE block ADD chain_id BIGINT NOT NULL DEFAULT 0  COMMENT "chain id";

ALTER TABLE txs ADD  chain_id BIGINT NOT NULL DEFAULT 0  COMMENT "chain id";
