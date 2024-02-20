CREATE
    DATABASE `tap_indexer` DEFAULT COLLATE = `utf8mb4_general_ci`;

Use
    tap_indexer;
-- inscription table ---------
CREATE TABLE `inscriptions`
(
    `id`             int unsigned                                                  NOT NULL AUTO_INCREMENT,
    `sid`            int unsigned                                                  NOT NULL, -- sid
    `chain`          varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci  NOT NULL, -- chain code, eth / avax / btc / doge
    `protocol`       varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_bin    NOT NULL, -- protocol code, POLS, ETHS, BRC20
    `tick`           varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_bin    NOT NULL, -- ticker code
    `name`           varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci  NOT NULL, -- ticker name
    `limit_per_mint` DECIMAL(38, 18)                                               NOT NULL, -- mint amount limit by per mint
    `deploy_by`      varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL, -- deployed address
    `total_supply`   DECIMAL(38, 18)                                               NOT NULL, -- total supply
    `decimals`       tinyint(1) unsigned                                           NOT NULL, -- decimals
    `deploy_hash`    varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL, -- deployed tx hash
    `deploy_time`    timestamp                                                     NOT NULL, -- deployed time
    `transfer_type`  tinyint(1)                                                    NOT NULL, -- transfer type
    `created_at`     timestamp                                                     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     timestamp                                                     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uq_chain_protocol_name` (`chain`, `protocol`, `tick`),
    UNIQUE KEY `uq_chain_sid` (`chain`, `sid`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;

-- inscription statics table ---------
CREATE TABLE `inscriptions_stats`
(
    `id`                  int unsigned                                                 NOT NULL AUTO_INCREMENT,
    `sid`                 int unsigned                                                 NOT NULL,             -- sid
    `chain`               varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,             -- chain code
    `protocol`            varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_bin   NOT NULL,             -- protocol code, POLS, ETHS, BRC20
    `tick`                varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_bin   NOT NULL,             -- ticker code
    `minted`              DECIMAL(38, 18) unsigned                                     NOT NULL DEFAULT '0', -- minted amount
    `mint_completed_time` timestamp                                                    NULL,                 -- mint completed time
    `mint_first_block`    bigint unsigned                                              NOT NULL,             -- mint start block
    `mint_last_block`     bigint unsigned                                              NOT NULL,             -- mint completed block
    `last_sn`             int unsigned                                                 NOT NULL,             -- last sn
    `holders`             int unsigned                                                 NOT NULL,             -- total holders
    `tx_cnt`              bigint unsigned                                              NOT NULL,             -- total txs
    `created_at`          timestamp                                                    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`          timestamp                                                    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uq_chain_protocol_name` (`chain`, `protocol`, `tick`),
    UNIQUE KEY `uq_chain_sid` (`chain`, `sid`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;

-- tx raw table ---------
CREATE TABLE `txs`
(
    `id`                bigint unsigned NOT NULL AUTO_INCREMENT,
    `chain`             varchar(32)     NOT NULL COMMENT 'chain name',
    `protocol`          varchar(32)     NOT NULL COMMENT 'protocol name',
    `block_height`      bigint unsigned NOT NULL COMMENT 'block height',
    `position_in_block` bigint unsigned NOT NULL COMMENT 'Position in Block',
    `block_time`        timestamp       NOT NULL COMMENT 'block time',
    `tx_hash`           varchar(128)    NOT NULL COMMENT 'tx hash',
    `from`              varchar(128)    NOT NULL COMMENT 'from address',
    `to`                varchar(128)    NOT NULL COMMENT 'to address',
    `op`                varchar(32)     NOT NULL COMMENT 'op code',
    `tick`              varchar(32)     NOT NULL COMMENT 'inscription code',
    `amt`               DECIMAL(38, 18) NOT NULL COMMENT 'amount',
    `gas`               bigint          NOT NULL COMMENT 'gas, spend fee',
    `gas_price`         bigint          NOT NULL COMMENT 'gas price',
    `status`            tinyint(1)      NOT NULL COMMENT 'tx status',
    `created_at`        timestamp       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`        timestamp       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_tx_hash_chain` (`tx_hash`(12), `chain`(4))
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;

-- address ticks balances ---------
CREATE TABLE `balances`
(
    `id`         int unsigned                            NOT NULL AUTO_INCREMENT,
    `sid`        int unsigned                            NOT NULL COMMENT 'sid',
    `chain`      varchar(32) COLLATE utf8mb4_general_ci  NOT NULL COMMENT 'chain name',
    `protocol`   varchar(32) COLLATE utf8mb4_0900_bin    NOT NULL COMMENT 'protocol name',
    `address`    varchar(128) COLLATE utf8mb4_general_ci NOT NULL COMMENT 'address',
    `tick`       varchar(32) COLLATE utf8mb4_0900_bin    NOT NULL COMMENT 'inscription code',
    `available`  DECIMAL(38, 18)                         NOT NULL COMMENT 'available',
    `balance`    DECIMAL(38, 18)                         NOT NULL COMMENT 'balance',
    `created_at` timestamp                               NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp                               NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `address` (`address`, `chain`, `protocol`, `tick`),
    UNIQUE KEY `uqx_chain_sid` (`chain`, `sid`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;

-- address related txs ---------
CREATE TABLE `address_txs`
(
    `id`         bigint unsigned                                               NOT NULL AUTO_INCREMENT,
    `chain`      varchar(32) COLLATE utf8mb4_general_ci                        NOT NULL COMMENT 'chain name',
    `event`      tinyint(1)                                                    NOT NULL,
    `protocol`   varchar(32) COLLATE utf8mb4_0900_bin                          NOT NULL COMMENT 'protocol name',
    `operate`    varchar(32) COLLATE utf8mb4_0900_bin                          NOT NULL COMMENT 'operate',
    `tx_hash`    varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_bin   NOT NULL COMMENT 'tx hash',
    `address`    varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'from address',
    `amount`     DECIMAL(38, 18)                                               NOT NULL COMMENT 'amount',
    `tick`       varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci  NOT NULL COMMENT 'inscription name',
    `created_at` timestamp                                                     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp                                                     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_tx_hash` (`tx_hash`(12)),
    KEY `idx_address` (`address`(12))
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;

-- address balances change logs ---------
CREATE TABLE `balance_txn`
(
    `id`         bigint unsigned                                               NOT NULL AUTO_INCREMENT,
    `chain`      varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci  NOT NULL,
    `protocol`   varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_bin    NOT NULL,
    `event`      tinyint(1)                                                    NOT NULL,
    `address`    varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
    `tick`       varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_bin    NOT NULL,
    `amount`     DECIMAL(38, 18)                                               NOT NULL,
    `available`  DECIMAL(38, 18)                                               NOT NULL COMMENT 'available',
    `balance`    DECIMAL(38, 18)                                               NOT NULL,
    `tx_hash`    varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
    `created_at` timestamp                                                     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp                                                     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_address` (`address`(12))
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;


-- address utxos ------------------------------
CREATE TABLE `utxos`
(
    `id`         bigint unsigned                                               NOT NULL AUTO_INCREMENT,
    `sn`         varchar(255)                                                  NOT NULL COMMENT 'tx sn',
    `chain`      varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci  NOT NULL,
    `protocol`   varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_bin    NOT NULL,
    `address`    varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
    `tick`       varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_bin    NOT NULL,
    `amount`     DECIMAL(38, 18)                                               NOT NULL,
    `root_hash`  varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
    `tx_hash`    varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
    `status`     tinyint(1)                                                    NOT NULL COMMENT 'tx status',
    `created_at` timestamp                                                     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp                                                     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_address` (`address`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;

CREATE TABLE `block`
(
    `chain`        varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci  NOT NULL,
    `block_hash`   varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
    `block_number` bigint                                                        NOT NULL,
    `block_time`   timestamp                                                     NOT NULL,
    `updated_at`   timestamp                                                     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`chain`) USING BTREE,
    UNIQUE KEY `uqx_chain` (`chain`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;

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
