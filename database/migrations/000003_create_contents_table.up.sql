CREATE TABLE IF NOT EXISTS contents (
    id SERIAL PRIMARY KEY,
    category_id INT NOT NULL REFERENCES categories(id),
    created_by_id INT NOT NULL REFERENCES users(id),
    title VARCHAR(200) NOT NULL,
    excerpt VARCHAR(250) NOT NULL,
    description TEXT NOT NULL,
    image TEXT NULL,
    status VARCHAR(20) DEFAULT 'PUBLISH',
    tags TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_contents_category_id ON contents(category_id);
CREATE INDEX IF NOT EXISTS idx_contents_created_by_id ON contents(created_by_id);
CREATE INDEX IF NOT EXISTS idx_contents_status ON contents(status);
CREATE INDEX IF NOT EXISTS idx_contents_title ON contents(title);