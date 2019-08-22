CREATE TABLE `transactions_outputs`
(
    `id`                BIGINT UNSIGNED NOT NULL,
    `transaction_id`    BIGINT UNSIGNED NOT NULL,
    `index`             INT UNSIGNED    NOT NULL,
    `value`             BIGINT UNSIGNED NOT NULL,
    `public_key_script` BLOB            NOT NULL,
    PRIMARY KEY (`id`),
    INDEX `transaction_id_IDX` (`transaction_id`),
    CONSTRAINT `fk_transactions_outputs_transaction_id`
        FOREIGN KEY (`transaction_id`)
            REFERENCES `transactions` (`id`)
);
