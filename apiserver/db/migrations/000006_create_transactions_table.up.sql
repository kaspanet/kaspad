CREATE TABLE `transactions`
(
    `id`               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `block_id`         BIGINT UNSIGNED NOT NULL,
    `transaction_hash` VARCHAR(32)     NOT NULL,
    `transaction_id`   VARCHAR(32)     NOT NULL,
    `lock_time`        BIGINT UNSIGNED NOT NULL,
    `subnetwork_id`    BIGINT UNSIGNED NOT NULL,
    `gas`              BIGINT UNSIGNED NOT NULL,
    `payload_hash`     VARCHAR(32)     NOT NULL,
    `payload`          BLOB            NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `transaction_hash_UNIQUE` (`transaction_hash`),
    INDEX `transaction_id_IDX` (`transaction_id`),
    CONSTRAINT `fk_transactions_block_id`
        FOREIGN KEY (`block_id`)
            REFERENCES `blocks` (`id`)
);
