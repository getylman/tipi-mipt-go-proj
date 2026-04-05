-- name: InsertConsumption :exec
-- Одна строка потребления.
-- Вызывается в цикле внутри транзакции через WithTx.
INSERT INTO consumption
    (user_id, product_id, quantity, unit_price, total_price)
VALUES
    (@user_id::uuid, @product_id, @quantity, @unit_price, @total_price);

-- name: GetConsumptionByUser :many
-- История потребления пользователя, свежие сверху.
SELECT product_id, quantity, unit_price, total_price, calculated_at
FROM   consumption
WHERE  user_id = @user_id::uuid
ORDER  BY calculated_at DESC;

-- name: SumConsumptionByUser :one
-- Накопленная сумма затрат пользователя.
SELECT COALESCE(SUM(total_price), 0)::numeric AS total
FROM   consumption
WHERE  user_id = @user_id::uuid;
