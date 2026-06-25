ALTER TABLE users
    ADD COLUMN IF NOT EXISTS encrypted_dek BYTEA;

CREATE TABLE IF NOT EXISTS vms (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    datanode_id            UUID NOT NULL REFERENCES datanodes(id) ON DELETE RESTRICT,
    name                   TEXT NOT NULL,
    host                   TEXT NOT NULL,
    port                   INT NOT NULL DEFAULT 22,
    encrypted_credentials  BYTEA NOT NULL,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, name)
);

CREATE INDEX IF NOT EXISTS idx_vms_user_id ON vms(user_id);
CREATE INDEX IF NOT EXISTS idx_vms_datanode_id ON vms(datanode_id);
