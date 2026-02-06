-- name: CreateModelPricing :exec
INSERT INTO model_pricing (id, display_name, input_per_million, output_per_million, cache_read_per_million, cache_write_per_million, long_context_input_per_million, long_context_output_per_million, long_context_threshold, is_default, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetModelPricingByID :one
SELECT * FROM model_pricing WHERE id = ?;

-- name: GetDefaultModelPricing :one
SELECT * FROM model_pricing WHERE is_default = 1 LIMIT 1;

-- name: ListModelPricing :many
SELECT * FROM model_pricing ORDER BY display_name ASC;

-- name: UpdateModelPricing :exec
UPDATE model_pricing
SET display_name = ?, input_per_million = ?, output_per_million = ?, cache_read_per_million = ?, cache_write_per_million = ?, long_context_input_per_million = ?, long_context_output_per_million = ?, long_context_threshold = ?, is_default = ?
WHERE id = ?;

-- name: SetDefaultModelPricing :exec
UPDATE model_pricing SET is_default = CASE WHEN id = ? THEN 1 ELSE 0 END;

-- name: DeleteModelPricing :exec
DELETE FROM model_pricing WHERE id = ?;
