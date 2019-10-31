CREATE TABLE `ip_uses`
(
    `ip`       VARCHAR(39) NOT NULL,
    `last_use` DATETIME NOT NULL,
    PRIMARY KEY (`ip`)
);