CREATE TABLE `accepting_blocks`
(
    `block_id`           BIGINT UNSIGNED NOT NULL,
    `accepting_block_id` BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (`block_id`, `accepting_block_id`),
    CONSTRAINT `fk_accepting_blocks_block_id`
        FOREIGN KEY (`block_id`)
            REFERENCES `blocks` (`id`),
    CONSTRAINT `fk_accepting_blocks_accepting_block_id`
        FOREIGN KEY (`accepting_block_id`)
            REFERENCES `blocks` (`id`)
);
