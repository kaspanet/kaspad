CREATE TABLE `blocks`
(
    `id`                      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `block_hash`              VARCHAR(32)     NOT NULL,
    `version`                 INT             NOT NULL,
    `hash_merkle_root`        VARCHAR(32)     NOT NULL,
    `accepted_id_merkle_root` VARCHAR(32)     NOT NULL,
    `utxo_commitment`         VARCHAR(32)     NOT NULL,
    `timestamp`               DATETIME        NOT NULL,
    `bits`                    INT UNSIGNED    NOT NULL,
    `nonce`                   BIGINT UNSIGNED NOT NULL,
    `blue_score`              BIGINT UNSIGNED NOT NULL,
    `is_chain_block`          TINYINT         NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `block_hash_UNIQUE` (`block_hash`),
    INDEX `timestamp_IDX` (`timestamp`),
    INDEX `is_chain_block_IDX` (`is_chain_block`)
);
