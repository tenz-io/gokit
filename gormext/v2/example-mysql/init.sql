CREATE DATABASE IF NOT EXISTS gormext_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE gormext_db;

CREATE TABLE IF NOT EXISTS `user_tab`
(
    `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `username`   VARCHAR(100)         DEFAULT NULL comment 'username',
    `password`   VARCHAR(100)         DEFAULT NULL comment 'password',
    `created_at` TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP       NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uniq_username` (`username`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci
    COMMENT 'user table';

INSERT INTO `user_tab` (`username`, `password`)
VALUES ('admin', 'admin');
INSERT INTO `user_tab` (`username`, `password`)
VALUES ('alice', 'alice');
INSERT INTO `user_tab` (`username`, `password`)
VALUES ('bob', 'bob');