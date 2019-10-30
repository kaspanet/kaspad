CREATE TABLE `ip_uses`
(
    `ip`       CHAR(15) NOT NULL,
    `last_use` DATETIME NOT NULL,
    PRIMARY KEY (`ip`)
);