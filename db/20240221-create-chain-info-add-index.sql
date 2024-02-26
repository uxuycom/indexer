Use
tap_indexer;
-- chain statics by hour table ---------
CREATE TABLE `chain_stats_hour`
(
    `id`                 int unsigned                           NOT NULL AUTO_INCREMENT,
    `chain`              varchar(32) COLLATE utf8mb4_general_ci NOT NULL COMMENT 'chain name',
    `date_hour`          int unsigned                           NOT NULL COMMENT 'date_hour',
    `address_count`      int unsigned                           NOT NULL COMMENT 'address_count',
    `address_last_id`    bigint unsigned                        NOT NULL COMMENT 'address_last_id',
    `inscriptions_count` int unsigned                           NOT NULL COMMENT 'inscriptions_count',
    `balance_sum`        DECIMAL(38, 18)                        NOT NULL COMMENT 'balance_sum',
    `balance_last_id`    bigint unsigned                        NOT NULL COMMENT 'balance_last_id',
    `created_at`         timestamp                              NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`         timestamp                              NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uqx_chain_date_hour` (`chain`, `date_hour`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;

-- chain info ---------
CREATE TABLE `chain_info`
(
    `id`          int unsigned                             NOT NULL AUTO_INCREMENT,
    `chain_id`    int unsigned                             NOT NULL COMMENT 'chain id',
    `chain`       varchar(32) COLLATE utf8mb4_general_ci   NOT NULL COMMENT 'inner chain name',
    `outer_chain` varchar(32) COLLATE utf8mb4_general_ci   NOT NULL COMMENT 'outer chain name',
    `name`        varchar(32) COLLATE utf8mb4_general_ci   NOT NULL COMMENT 'name',
    `logo`        varchar(1024) COLLATE utf8mb4_general_ci NOT NULL COMMENT 'logo url',
    `network_id`  int unsigned                             NOT NULL COMMENT 'network id',
    `ext`         varchar(4098)                            NOT NUll COMMENT 'ext',
    `created_at`  timestamp                                NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  timestamp                                NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uqx_chain_id_chain_name` (`chain_id`, `chain`, `name`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;

INSERT INTO chain_info (chain_id, chain, outer_chain, name, logo, network_id,ext)VALUES (0, 'btc', 'btc', 'BTC', '', 0, '');
INSERT INTO chain_info (chain_id, chain, outer_chain, name, logo, network_id,ext)VALUES (1, 'eth', 'eth', 'Ethereum', '', 1, '');
INSERT INTO chain_info (chain_id, chain, outer_chain, name, logo, network_id,ext)VALUES (43114, 'avalanche', 'avax', 'Avalanche', '', 43114, '');
INSERT INTO chain_info (chain_id, chain, outer_chain, name, logo, network_id,ext)VALUES (42161, 'arbitrum', 'ETH', 'Arbitrum One', '', 42161, '');
INSERT INTO chain_info (chain_id, chain, outer_chain, name, logo, network_id,ext)VALUES (56, 'bsc', 'BSC', 'BNB Smart Chain Mainnet', '', 56, '');
INSERT INTO chain_info (chain_id, chain, outer_chain, name, logo, network_id,ext)VALUES (250, 'fantom', 'FTM', 'Fantom Opera', '', 250, '');
INSERT INTO chain_info (chain_id, chain, outer_chain, name, logo, network_id,ext)VALUES (137, 'polygon', 'Polygon', 'Polygon Mainnet', '', 137, '');