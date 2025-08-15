-- Create releases table (updated to include project_id and composite unique constraint)
CREATE TABLE IF NOT EXISTS releases (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    tag VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, tag)
);

-- Create release_prs join table
CREATE TABLE IF NOT EXISTS release_prs (
    release_id INTEGER NOT NULL REFERENCES releases(id) ON DELETE CASCADE,
    pull_request_id INTEGER NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
    PRIMARY KEY (release_id, pull_request_id)
);

-- Create index on releases.status for faster lookups
CREATE INDEX IF NOT EXISTS idx_releases_status ON releases(status);
