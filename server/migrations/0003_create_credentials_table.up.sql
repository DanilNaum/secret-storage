
CREATE TABLE credentials (
    record_id VARCHAR(255) UNIQUE,
    login  VARCHAR(255),
    password  VARCHAR(255),
    CONSTRAINT fk_record_id
        FOREIGN KEY (record_id)
        REFERENCES records(id)
        ON DELETE RESTRICT
);

CREATE TABLE text(
    record_id VARCHAR(255) UNIQUE,
    content VARCHAR(255),
    CONSTRAINT fk_record_id
        FOREIGN KEY (record_id)
        REFERENCES records(id)
        ON DELETE RESTRICT
)