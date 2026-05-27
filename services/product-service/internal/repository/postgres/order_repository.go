package postgres

import (
	"context"

	"github.com/TheDeutsch13/b2-project/services/product-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type orderRowScanner interface {
	Scan(dest ...any) error
}

func scanOrderRow(scanner orderRowScanner, order *domain.Order) error {
	return scanner.Scan(
		&order.ID,
		&order.UserID,
		&order.Status,
		&order.ContactName,
		&order.ContactPhone,
		&order.ContactEmail,
		&order.DeliveryAddress,
		&order.DeliveryType,
		&order.DeliveryCost,
		&order.DeliveryCity,
		&order.CdekPvzCode,
		&order.DeliveryPayment,
		&order.PaymentMethod,
		&order.Comment,
		&order.TotalAmount,
		&order.CreatedAt,
	)
}

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

func orderReservesStock(status domain.OrderStatus) bool {
	return status != domain.OrderStatusCancelled
}

func decrementProductStock(ctx context.Context, tx pgx.Tx, productID int64, quantity int) error {
	var stock int
	err := tx.QueryRow(
		ctx,
		`UPDATE products SET stock = stock - $2, updated_at = NOW()
		 WHERE id = $1 AND stock >= $2
		 RETURNING stock`,
		productID,
		quantity,
	).Scan(&stock)
	if err == pgx.ErrNoRows {
		return domain.ErrInsufficientStock
	}
	return err
}

func incrementProductStock(ctx context.Context, tx pgx.Tx, productID int64, quantity int) error {
	_, err := tx.Exec(
		ctx,
		`UPDATE products SET stock = stock + $2, updated_at = NOW() WHERE id = $1`,
		productID,
		quantity,
	)
	return err
}

func (r *OrderRepository) Create(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	for _, item := range order.Items {
		if err := decrementProductStock(ctx, tx, item.ProductID, item.Quantity); err != nil {
			return nil, err
		}
	}

	query := `
		INSERT INTO orders (
			user_id, status, contact_name, contact_phone, contact_email,
			delivery_address, delivery_type, delivery_cost, delivery_city, cdek_pvz_code,
			delivery_payment, payment_method, comment, total_amount
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at
	`

	err = tx.QueryRow(
		ctx,
		query,
		order.UserID,
		order.Status,
		order.ContactName,
		order.ContactPhone,
		order.ContactEmail,
		order.DeliveryAddress,
		order.DeliveryType,
		order.DeliveryCost,
		order.DeliveryCity,
		order.CdekPvzCode,
		order.DeliveryPayment,
		order.PaymentMethod,
		order.Comment,
		order.TotalAmount,
	).Scan(&order.ID, &order.CreatedAt)
	if err != nil {
		return nil, err
	}

	itemQuery := `
		INSERT INTO order_items (order_id, product_id, quantity, price)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	for index := range order.Items {
		item := &order.Items[index]
		item.OrderID = order.ID

		err = tx.QueryRow(
			ctx,
			itemQuery,
			item.OrderID,
			item.ProductID,
			item.Quantity,
			item.Price,
		).Scan(&item.ID)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return order, nil
}

const orderSelectColumns = `
		id, user_id, status, contact_name, contact_phone, contact_email,
		delivery_address,
		COALESCE(delivery_type, 'custom'),
		COALESCE(delivery_cost, 0)::float8,
		COALESCE(delivery_city, ''),
		COALESCE(cdek_pvz_code, ''),
		COALESCE(delivery_payment, 'on_receipt'),
		payment_method, COALESCE(comment, ''), total_amount::float8, created_at
`

func (r *OrderRepository) ListByUserID(ctx context.Context, userID int64) ([]domain.Order, error) {
	query := `
		SELECT ` + orderSelectColumns + `
		FROM orders
		WHERE user_id = $1
		ORDER BY id DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		if isUndefinedTable(err) {
			return []domain.Order{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	orders := make([]domain.Order, 0)

	for rows.Next() {
		var order domain.Order

		if err := scanOrderRow(rows, &order); err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return r.attachItems(ctx, orders)
}

func (r *OrderRepository) ListAll(ctx context.Context) ([]domain.Order, error) {
	query := `
		SELECT ` + orderSelectColumns + `
		FROM orders
		ORDER BY id DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		if isUndefinedTable(err) {
			return []domain.Order{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	orders := make([]domain.Order, 0)

	for rows.Next() {
		var order domain.Order

		if err := scanOrderRow(rows, &order); err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return r.attachItems(ctx, orders)
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID int64, status domain.OrderStatus) (*domain.Order, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var oldStatus domain.OrderStatus
	err = tx.QueryRow(
		ctx,
		`SELECT status FROM orders WHERE id = $1 FOR UPDATE`,
		orderID,
	).Scan(&oldStatus)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	itemRows, err := tx.Query(
		ctx,
		`SELECT product_id, quantity FROM order_items WHERE order_id = $1`,
		orderID,
	)
	if err != nil {
		return nil, err
	}

	var items []domain.OrderItem
	for itemRows.Next() {
		var item domain.OrderItem
		if err := itemRows.Scan(&item.ProductID, &item.Quantity); err != nil {
			itemRows.Close()
			return nil, err
		}
		items = append(items, item)
	}
	itemRows.Close()
	if err := itemRows.Err(); err != nil {
		return nil, err
	}

	wasReserved := orderReservesStock(oldStatus)
	willReserve := orderReservesStock(status)

	if wasReserved && !willReserve {
		for _, item := range items {
			if err := incrementProductStock(ctx, tx, item.ProductID, item.Quantity); err != nil {
				return nil, err
			}
		}
	}

	if !wasReserved && willReserve {
		for _, item := range items {
			if err := decrementProductStock(ctx, tx, item.ProductID, item.Quantity); err != nil {
				return nil, err
			}
		}
	}

	query := `
		UPDATE orders
		SET status = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING ` + orderSelectColumns + `
	`

	var order domain.Order
	err = scanOrderRow(tx.QueryRow(ctx, query, status, orderID), &order)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	orders, err := r.attachItems(ctx, []domain.Order{order})
	if err != nil {
		return nil, err
	}

	return &orders[0], nil
}

func (r *OrderRepository) attachItems(ctx context.Context, orders []domain.Order) ([]domain.Order, error) {
	if len(orders) == 0 {
		return orders, nil
	}

	orderIDs := make([]int64, 0, len(orders))
	orderMap := make(map[int64]*domain.Order, len(orders))

	for index := range orders {
		orderIDs = append(orderIDs, orders[index].ID)
		orderMap[orders[index].ID] = &orders[index]
	}

	query := `
		SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price, p.title
		FROM order_items oi
		JOIN products p ON p.id = oi.product_id
		WHERE oi.order_id = ANY($1)
	`

	rows, err := r.db.Query(ctx, query, orderIDs)
	if err != nil {
		if isUndefinedTable(err) {
			return orders, nil
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.OrderItem

		if err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Quantity,
			&item.Price,
			&item.Title,
		); err != nil {
			return nil, err
		}

		order := orderMap[item.OrderID]
		order.Items = append(order.Items, item)
	}

	return orders, rows.Err()
}

func (r *OrderRepository) GetByID(ctx context.Context, orderID int64) (*domain.Order, error) {
	query := `
		SELECT ` + orderSelectColumns + `
		FROM orders
		WHERE id = $1
	`

	var order domain.Order

	err := scanOrderRow(r.db.QueryRow(ctx, query, orderID), &order)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	orders, err := r.attachItems(ctx, []domain.Order{order})
	if err != nil {
		return nil, err
	}

	return &orders[0], nil
}

func (r *OrderRepository) UserHasReceivedProduct(
	ctx context.Context,
	userID int64,
	productID int64,
) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM orders o
			INNER JOIN order_items oi ON oi.order_id = o.id
			WHERE o.user_id = $1
				AND oi.product_id = $2
				AND o.status = $3
		)
	`

	var exists bool

	err := r.db.QueryRow(ctx, query, userID, productID, domain.OrderStatusReceived).Scan(&exists)
	if err != nil {
		if isUndefinedTable(err) {
			return false, nil
		}

		return false, err
	}

	return exists, nil
}
