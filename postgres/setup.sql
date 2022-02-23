drop table blocks;
drop table transactions;
drop table address_transaction;
drop table identities;
CREATE TABLE blocks (
    id SERIAL PRIMARY KEY,
    height NUMERIC(30, 0) NOT NULL,
    tx_count NUMERIC(30, 0) NOT NULL,
    value_total NUMERIC(30, 0) NOT NULL,
    mined_coins NUMERIC(30, 0) NOT NULL,
    block_fee NUMERIC(30, 0) NOT NULL,
    block_hash VARCHAR(255) NOT NULL
);

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    tx_hash VARCHAR(255), 
    num_vin NUMERIC(30, 0) NOT NULL,
    num_vout NUMERIC(30, 0) NOT NULL,
    tx_fee NUMERIC(30, 0) NOT NULL,
    block_hash VARCHAR(255) NOT NULL,
    height NUMERIC(30, 0) NOT NULL,
    value NUMERIC(30, 0) NOT NULL
);


-- Rethink this mapping for the future, maybe instead of id use hash or the other way around
CREATE TABLE address_transaction (
    id SERIAL PRIMARY KEY,
    address VARCHAR(255) NOT NULL,
    transaction_hash VARCHAR(255) NOT NULL,
    vin BOOLEAN NOT NULL,
    vout BOOLEAN NOT NULL,
    block_hash VARCHAR(255) NOT NULL,
    height NUMERIC(30, 0) NOT NULL
);

CREATE TABLE identities(
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address VARCHAR(255) NOT NULL,
    height NUMERIC(30, 0) NOT NULL,
    block_hash VARCHAR(255) NOT NULL
);
