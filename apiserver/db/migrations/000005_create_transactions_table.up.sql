CREATE TABLE `transactions`
(
    `id`               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `accepting_block_id`         BIGINT UNSIGNED NULL,
    `transaction_hash` CHAR(64)     NOT NULL,
    `transaction_id`   CHAR(64)     NOT NULL,
    `lock_time`        BIGINT UNSIGNED NOT NULL,
    `subnetwork_id`    BIGINT UNSIGNED NOT NULL,
    `gas`              BIGINT UNSIGNED NOT NULL,
    `payload_hash`     CHAR(64)     NOT NULL,
    `payload`          BLOB            NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_transactions_transaction_hash` (`transaction_hash`),
    INDEX `idx_transactions_transaction_id` (`transaction_id`),
    CONSTRAINT `fk_transactions_accepting_block_id`
        FOREIGN KEY (`accepting_block_id`)
            REFERENCES `blocks` (`id`)
);
