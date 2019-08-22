CREATE TABLE `utxos`
(
    `transaction_output_id` BIGINT UNSIGNED NOT NULL,
    `accepting_block_id`    INT UNSIGNED    NULL,
    PRIMARY KEY (`transaction_output_id`),
    INDEX `idx_utxos_accepting_block_id` (`accepting_block_id`),
    CONSTRAINT `fk_utxos_transaction_output_id`
        FOREIGN KEY (`transaction_output_id`)
            REFERENCES `transactions_outputs` (`id`)
);
