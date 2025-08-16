package store

const (
	// Insert into 'orders' table
	qInsertOrders = `
		INSERT INTO orders (
			order_uid, track_number, entry, locale, internal_signature, customer_id, 
			delivery_service, shard_key, sm_id, date_created, oof_shard, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW()
		) RETURNING id;
	`

	// Insert into 'deliveries' table, here order_id references id, returned by order insertion
	qInsertDeliveries = `
		INSERT INTO deliveries (
			order_id, name, phone, zip, city, address, region, email
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		);
	`
	// Insert into 'payments' table, here order_id references id, returned by order insertion
	qInsertPayments = `
		INSERT INTO payments (
			order_id, transaction, request_id, currency, provider, amount, 
			payment_dt, bank, delivery_cost, goods_total, custom_fee
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		);
	`
	// Same stuff
	qInsertItems = `
		INSERT INTO items (
			order_id, chrt_id, track_number, price, rid, name, 
			sale, size, total_price, nm_id, brand, status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		);
	`

	// Retrieves a JSON object from the database.
	qRetrieveJSON = `
		SELECT
			json_build_object(
				'order_uid', o.order_uid,
				'track_number', o.track_number,
				'entry', o.entry,
				'delivery', json_build_object(
					'name', d.name,
					'phone', d.phone,
					'zip', d.zip,
					'city', d.city,
					'address', d.address,
					'region', d.region,
					'email', d.email
				),
				'payment', json_build_object(
					'transaction', p.transaction,
					'request_id', p.request_id,
					'currency', p.currency,
					'provider', p.provider,
					'amount', p.amount,
					'payment_dt', p.payment_dt,
					'bank', p.bank,
					'delivery_cost', p.delivery_cost,
					'goods_total', p.goods_total,
					'custom_fee', p.custom_fee
				),
				'items', COALESCE(i.items_json, '[]'::json),
				'locale', o.locale,
				'internal_signature', o.internal_signature,
				'customer_id', o.customer_id,
				'delivery_service', o.delivery_service,
				'shardkey', o.shard_key,
				'sm_id', o.sm_id,
				'date_created', o.date_created,
				'oof_shard', o.oof_shard
			)
		FROM
			orders o
		JOIN
			deliveries d ON o.id = d.order_id
		JOIN
			payments p ON o.id = p.order_id
		LEFT JOIN
			(
				SELECT
					order_id,
					json_agg(json_build_object(
						'chrt_id', chrt_id,
						'track_number', track_number,
						'price', price,
						'rid', rid,
						'name', name,
						'sale', sale,
						'size', size,
						'total_price', total_price,
						'nm_id', nm_id,
						'brand', brand,
						'status', status
					) ORDER BY id) AS items_json
				FROM
					items
				GROUP BY
					order_id
			) i ON o.id = i.order_id
		WHERE
			o.order_uid = $1;
	`

	// Retrieves a full order object by joining the necessary tables.
	// This query can return multiple rows for a single order if it has multiple items.
	qRetrieveOrder = `
		SELECT
			-- orders
			o.order_uid, o.track_number AS order_track_number, o.entry, o.locale, o.internal_signature, o.customer_id, 
			o.delivery_service, o.shard_key, o.sm_id, o.date_created, o.oof_shard,
			-- delivery
			d.name AS delivery_name, d.phone, d.zip, d.city, d.address, d.region, d.email,
			-- payment
			p.transaction, p.request_id, p.currency, p.provider, p.amount, 
			p.payment_dt, p.bank, p.delivery_cost, p.goods_total, p.custom_fee,
			-- item
			i.chrt_id, i.track_number AS item_track_number, i.price, i.rid, i.name AS item_name, 
			i.sale, i.size, i.total_price, i.nm_id, i.brand, i.status
		FROM
			orders o
		JOIN
			deliveries d ON o.id = d.order_id
		JOIN
			payments p ON o.id = p.order_id
		LEFT JOIN
			items i ON o.id = i.order_id
		WHERE
			o.order_uid = $1;
	`

	// Retrieves the latest N orders as JSONs based on update timestamp.
	qGetLatestOrdersAsJSON = `
		SELECT
			json_build_object(
				'order_uid', o.order_uid,
				'track_number', o.track_number,
				'entry', o.entry,
				'delivery', json_build_object(
					'name', d.name,
					'phone', d.phone,
					'zip', d.zip,
					'city', d.city,
					'address', d.address,
					'region', d.region,
					'email', d.email
				),
				'payment', json_build_object(
					'transaction', p.transaction,
					'request_id', p.request_id,
					'currency', p.currency,
					'provider', p.provider,
					'amount', p.amount,
					'payment_dt', p.payment_dt,
					'bank', p.bank,
					'delivery_cost', p.delivery_cost,
					'goods_total', p.goods_total,
					'custom_fee', p.custom_fee
				),
				'items', COALESCE(i.items_json, '[]'::json),
				'locale', o.locale,
				'internal_signature', o.internal_signature,
				'customer_id', o.customer_id,
				'delivery_service', o.delivery_service,
				'shardkey', o.shard_key,
				'sm_id', o.sm_id,
				'date_created', o.date_created,
				'oof_shard', o.oof_shard
			)
		FROM
			orders o
		JOIN
			deliveries d ON o.id = d.order_id
		JOIN
			payments p ON o.id = p.order_id
		LEFT JOIN
			(
				SELECT
					order_id,
					json_agg(json_build_object(
						'chrt_id', chrt_id,
						'track_number', track_number,
						'price', price,
						'rid', rid,
						'name', name,
						'sale', sale,
						'size', size,
						'total_price', total_price,
						'nm_id', nm_id,
						'brand', brand,
						'status', status
					) ORDER BY id) AS items_json
				FROM
					items
				GROUP BY
					order_id
			) i ON o.id = i.order_id
		ORDER BY
			o.updated_at DESC
		LIMIT $1;
	`
)
