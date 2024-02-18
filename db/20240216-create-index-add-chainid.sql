/*
 * Copyright (C) 2024 Baidu, Inc. All Rights Reserved.
 */
Use
tap_indexer;

create index idx_tx_hash on balance_txn (tx_hash(12));

ALTER TABLE block ADD chain_id BIGINT NOT NULL DEFAULT 0  COMMENT "chain id";