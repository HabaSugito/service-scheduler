CREATE TABLE managers (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name       VARCHAR(255)    NOT NULL,
    email      VARCHAR(255)    NOT NULL UNIQUE,
    created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
);

CREATE TABLE technicians (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name       VARCHAR(255)    NOT NULL,
    email      VARCHAR(255)    NOT NULL UNIQUE,
    created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
);

CREATE TABLE quotes (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    title       VARCHAR(255)    NOT NULL,
    description TEXT,
    status      ENUM('unscheduled','scheduled') NOT NULL DEFAULT 'unscheduled',
    created_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    INDEX idx_status (status)
);

CREATE TABLE jobs (
    id             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    quote_id       BIGINT UNSIGNED NOT NULL,
    technician_id  BIGINT UNSIGNED NOT NULL,
    manager_id     BIGINT UNSIGNED NOT NULL,
    start_at       DATETIME        NOT NULL,
    end_at         DATETIME        NOT NULL,
    status         ENUM('scheduled','completed') NOT NULL DEFAULT 'scheduled',
    created_at     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    CONSTRAINT fk_jobs_quote       FOREIGN KEY (quote_id)      REFERENCES quotes(id),
    CONSTRAINT fk_jobs_technician  FOREIGN KEY (technician_id) REFERENCES technicians(id),
    CONSTRAINT fk_jobs_manager     FOREIGN KEY (manager_id)    REFERENCES managers(id),
    INDEX idx_technician_time (technician_id, start_at, end_at)
);

CREATE TABLE notifications (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_type  ENUM('manager','technician') NOT NULL,
    user_id    BIGINT UNSIGNED NOT NULL,
    message    TEXT            NOT NULL,
    job_id     BIGINT UNSIGNED NOT NULL,
    read_at    DATETIME        NULL,
    created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    INDEX idx_user (user_type, user_id)
);
