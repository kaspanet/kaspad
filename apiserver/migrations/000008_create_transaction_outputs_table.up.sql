CREATE TABLE `transaction_outputs`
(
    `id`             BIGINT UNSIGNED NOT NULL,
    `transaction_id` BIGINT UNSIGNED NOT NULL,
    `index`          INT UNSIGNED    NOT NULL,
    `value`          BIGINT UNSIGNED NOT NULL,
    `pk_script`      BLOB            NOT NULL,
    `address_id`     BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (`id`),
    INDEX `idx_transaction_outputs_transaction_id` (`transaction_id`),
    CONSTRAINT `fk_transaction_outputs_transaction_id`
        FOREIGN KEY (`transaction_id`)
            REFERENCES `transactions` (`id`),
    CONSTRAINT `fk_transaction_outputs_address_id`
        FOREIGN KEY (`address_id`)
            REFERENCES `addresses` (`id`)
);
