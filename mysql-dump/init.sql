CREATE DATABASE IF NOT EXISTS duplicates;
USE duplicates;

CREATE TABLE IF NOT EXISTS forbidden_image
(
    external_reference VARCHAR(255),
    sift_descriptor    MEDIUMBLOB,
    orb_descriptor     MEDIUMBLOB,
    brisk_descriptor   MEDIUMBLOB,
    p_hash             BIGINT UNSIGNED,
    rotation_phash      BIGINT UNSIGNED,
    PRIMARY KEY (external_reference)
);

CREATE TABLE IF NOT EXISTS search_image
(
    id INT AUTO_INCREMENT,
    external_reference VARCHAR(255),
    original_reference VARCHAR(255),
    scenario ENUM('identical', 'scaled', 'rotated', 'mirrored', 'moved', 'background', 'motive', 'part', 'mixed'),
    notes VARCHAR(255),
    PRIMARY KEY(id)
);