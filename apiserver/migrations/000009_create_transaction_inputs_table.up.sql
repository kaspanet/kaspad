CREATE TABLE `transaction_inputs`
(
    `id`                             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `transaction_id`                 BIGINT UNSIGNED NULL,
    `previous_transaction_output_id` BIGINT UNSIGNED NOT NULL,
    `index`                          INT UNSIGNED    NOT NULL,
    `signature_script`               BLOB            NOT NULL,
    `sequence`                       BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (`id`),
    INDEX `idx_transaction_inputs_transaction_id` (`transaction_id`),
    INDEX `idx_transaction_inputs_previous_transaction_output_id` (`previous_transaction_output_id`),
    CONSTRAINT `fk_transaction_inputs_transaction_id`
        FOREIGN KEY (`transaction_id`)
            REFERENCES `transactions` (`id`),
    CONSTRAINT `fk_transaction_inputs_previous_transaction_output_id`
        FOREIGN KEY (`previous_transaction_output_id`)
            REFERENCES `transaction_outputs` (`id`)
);
