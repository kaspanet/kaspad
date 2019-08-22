CREATE TABLE `transactions_to_blocks`
(
    `transaction_id`    BIGINT UNSIGNED NOT NULL,
    `block_id`          BIGINT UNSIGNED NOT NULL,
    `location_in_block` INT UNSIGNED    NOT NULL,
    PRIMARY KEY (`transaction_id`, `block_id`),
    INDEX `idx_transactions_to_blocks_location_in_block` (`location_in_block`),
    CONSTRAINT `fk_transactions_to_blocks_block_id`
        FOREIGN KEY (`block_id`)
            REFERENCES `blocks` (`id`),
    CONSTRAINT `fk_transactions_to_blocks_transaction_id`
        FOREIGN KEY (`transaction_id`)
            REFERENCES `transactions` (`id`)
);
