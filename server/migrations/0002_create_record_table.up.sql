
CREATE TABLE record_types (
    id SERIAL PRIMARY KEY,
    type_name VARCHAR(50) NOT NULL UNIQUE
);


CREATE TABLE records(
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255),
    type_id INTEGER,
    CONSTRAINT fk_record_type
        FOREIGN KEY (type_id) 
        REFERENCES record_types(id)
        ON DELETE RESTRICT
);

CREATE TABLE user_records(
    user_id int,
    record_id VARCHAR(255),
    CONSTRAINT fk_user
        FOREIGN KEY (user_id) 
        REFERENCES users(uuid)
        ON DELETE CASCADE,
    CONSTRAINT fk_record
        FOREIGN KEY (record_id) 
        REFERENCES records(id)
        ON DELETE CASCADE
);


INSERT INTO record_types (type_name) VALUES 
('credentials'),
('text'), 
('file');


CREATE INDEX idx_records_type_id ON records(type_id);