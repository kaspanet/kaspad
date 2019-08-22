CREATE TABLE `transactions_inputs`
(
    `id`                    BIGINT UNSIGNED NOT NULL,
    `transaction_id`        BIGINT UNSIGNED NULL,
    `transaction_output_id` BIGINT UNSIGNED NOT NULL,
    `index`                 INT UNSIGNED    NOT NULL,
    `signature_script`      BLOB            NOT NULL,
    `sequence`              BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (`id`),
    INDEX `transaction_id_IDX` (`transaction_id`),
    INDEX `transaction_output_id_IDX` (`transaction_output_id`),
    CONSTRAINT `fk_transactions_inputs_transaction_id`
        FOREIGN KEY (`transaction_id`)
            REFERENCES `transactions` (`id`),
    CONSTRAINT `fk_transactions_inputs_transaction_output_id`
        FOREIGN KEY (`transaction_output_id`)
            REFERENCES `transactions_outputs` (`id`)
);
