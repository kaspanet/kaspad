CREATE TABLE `raw_blocks`
(
    `block_id`   BIGINT UNSIGNED NOT NULL,
    `block_data` BLOB            NOT NULL,
    PRIMARY KEY (`block_id`),
    CONSTRAINT `fk_raw_blocks_block_id`
        FOREIGN KEY (`block_id`)
            REFERENCES `blocks` (`id`)
);
