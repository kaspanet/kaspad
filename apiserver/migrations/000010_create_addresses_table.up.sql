CREATE TABLE `addresses`
(
    `id`                BIGINT UNSIGNED NOT NULL,
    `address`    CHAR(50) NOT NULL,
    `index`             INT UNSIGNED    NOT NULL,
    `value`             BIGINT UNSIGNED NOT NULL,
    `pk_script` BLOB            NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_addresses_address` (`address`)
)