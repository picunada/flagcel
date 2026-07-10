-- +goose Up
ALTER TABLE flags DROP CONSTRAINT flags_context_id_fkey;
ALTER TABLE flags
    ADD CONSTRAINT flags_context_id_fkey
    FOREIGN KEY (context_id) REFERENCES contexts(id) ON DELETE RESTRICT;

-- +goose Down
ALTER TABLE flags DROP CONSTRAINT flags_context_id_fkey;
ALTER TABLE flags
    ADD CONSTRAINT flags_context_id_fkey
    FOREIGN KEY (context_id) REFERENCES contexts(id) ON DELETE SET NULL;
