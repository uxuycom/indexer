

Use
tap_indexer;

CREATE INDEX idx_chain_protocol_tick ON  address_txs(chain,protocol,operate);
CREATE INDEX idx_chain_protocol_tick ON balance_txn(chain, protocol, tick);
CREATE INDEX idx_chain_protocol_tick ON balances(chain, protocol, tick);
CREATE INDEX idx_chain_protocol_tick ON txs(chain, protocol, tick);
CREATE INDEX idx_chain_block_height ON txs(chain, block_height);



