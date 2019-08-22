CREATE TABLE `subnetworks`
(
    `id`            BIGINT UNSIGNED NOT NULL,
    `subnetwork_id` VARCHAR(32)     NOT NULL,
    `gas_limit`     BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_subnetworks_subnetwork_id` (`subnetwork_id`)
);
