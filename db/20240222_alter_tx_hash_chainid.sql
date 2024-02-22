
Use
tap_indexer;

CREATE INDEX idx_tx_hash ON balance_txn (tx_hash(12));
ALTER TABLE txs MODIFY COLUMN tx_hash VARBINARY(128);
ALTER TABLE address_txs MODIFY COLUMN tx_hash VARBINARY(128);
ALTER TABLE balance_txn MODIFY COLUMN tx_hash VARBINARY(128);
ALTER TABLE address_txs ADD related_address varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL
    COMMENT 'related address';
ALTER TABLE block ADD chain_id BIGINT NOT NULL DEFAULT 0  COMMENT "chain id";
