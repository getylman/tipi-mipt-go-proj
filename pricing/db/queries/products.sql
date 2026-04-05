-- name: GetProductsByIDs :many
-- Загружает продукты по списку id одним запросом.
-- Возвращает map[id]Product в репозитории.
SELECT id, name, price_per_unit, unit, updated_at
FROM   products
WHERE  id = ANY(@ids::text[]);

-- name: ListProducts :many
-- Полный справочник продуктов — для GET /v1/products.
SELECT id, name, price_per_unit, unit, updated_at
FROM   products
ORDER  BY id;
