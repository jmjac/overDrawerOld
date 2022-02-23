drop table blocks;
drop table transactions;
drop table address_transaction;
CREATE TABLE blocks (
    id SERIAL PRIMARY KEY,
    height NUMERIC NOT NULL,
    tx_count NUMERIC NOT NULL,
    value_total NUMERIC NOT NULL,
    mined_coins NUMERIC NOT NULL,
    block_fee NUMERIC NOT NULL,
    block_hash VARCHAR(255) NOT NULL
);

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    tx_hash VARCHAR(255), 
    num_vin NUMERIC NOT NULL,
    num_vout NUMERIC NOT NULL,
    tx_fee NUMERIC NOT NULL,
    block_hash VARCHAR(255) NOT NULL,
    height NUMERIC NOT NULL,
    value NUMERIC NOT NULL
);


-- Rethink this mapping for the future, maybe instead of id use hash or the other way around
CREATE TABLE address_transaction (
    id SERIAL PRIMARY KEY,
    address VARCHAR(255) NOT NULL,
    transaction_hash VARCHAR(255) NOT NULL,
    vin BOOLEAN NOT NULL,
    vout BOOLEAN NOT NULL,
    block_hash VARCHAR(255) NOT NULL,
    height NUMERIC NOT NULL
);
